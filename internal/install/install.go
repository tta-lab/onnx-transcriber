package install

import (
	"archive/tar"
	"compress/bzip2"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/guion-opensource/onnx-transcriber/internal/config"
	"github.com/guion-opensource/onnx-transcriber/internal/manifest"
)

type Installer struct {
	Home     string
	Manifest manifest.Manifest
	Stdout   io.Writer
}

func (i Installer) EnsureDirs() error {
	for _, dir := range []string{
		filepath.Join(i.Home, "cache"),
		filepath.Join(i.Home, "logs"),
		filepath.Join(i.Home, "models"),
		config.RuntimeBinDir(i.Home, config.CurrentPlatformKey()),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (i Installer) InstallRuntime(name, fromDir string) error {
	rt, ok := i.Manifest.Runtimes[name]
	if !ok {
		return fmt.Errorf("unknown runtime %q", name)
	}
	asset, ok := rt.Platforms[config.CurrentPlatformKey()]
	if !ok {
		return fmt.Errorf("runtime %q does not support %s", name, config.CurrentPlatformKey())
	}

	archive, err := i.fetch(asset.URL, asset.SHA256, fromDir)
	if err != nil {
		return err
	}
	dest := config.RuntimeBinDir(i.Home, config.CurrentPlatformKey())
	if err := ExtractTarBz2(archive, dest); err != nil {
		return err
	}
	return PromoteBinaries(dest, asset.Binaries)
}

func (i Installer) InstallModel(name, fromDir string) error {
	model, ok := i.Manifest.Models[name]
	if !ok {
		return fmt.Errorf("unknown model %q", name)
	}
	if len(model.Files) == 0 {
		return fmt.Errorf("model %q has no files", name)
	}
	outDir := config.ModelDir(i.Home, name)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, file := range model.Files {
		if file.URL == "" && fromDir == "" {
			return fmt.Errorf("model %q file %q has no verified download URL yet", name, file.Path)
		}
		path, err := i.fetch(file.URL, file.SHA256, fromDir)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".tar.bz2") {
			if err := ExtractTarBz2(path, outDir); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(path, filepath.Join(outDir, file.Path)); err != nil {
			return err
		}
	}
	return nil
}

func (i Installer) fetch(url, wantSHA, fromDir string) (string, error) {
	cache := filepath.Join(i.Home, "cache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", err
	}

	name := filepath.Base(url)
	if fromDir != "" {
		if name == "." || name == "/" || name == "" {
			return "", errors.New("cannot resolve local asset name without a URL")
		}
		local := filepath.Join(fromDir, name)
		if err := verifySHA(local, wantSHA); err != nil {
			return "", err
		}
		return local, nil
	}

	out := filepath.Join(cache, name)
	if _, err := os.Stat(out); err == nil {
		if err := verifySHA(out, wantSHA); err != nil {
			return "", err
		}
		return out, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("download %s: %s", url, resp.Status)
	}

	tmp := out + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	if err = verifySHA(tmp, wantSHA); err != nil {
		return "", err
	}
	return out, os.Rename(tmp, out)
}

func ExtractTarBz2(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	tr := tar.NewReader(bzip2.NewReader(f))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := safeJoin(dest, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := hdr.FileInfo().Mode()
			if mode&0o111 != 0 {
				mode |= 0o755
			} else {
				mode = 0o644
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				_ = out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		}
	}
}

func safeJoin(root, name string) (string, error) {
	clean := filepath.Clean(name)
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) || clean == ".." {
		return "", fmt.Errorf("unsafe archive path %q", name)
	}
	target := filepath.Join(root, clean)
	if !strings.HasPrefix(target, filepath.Clean(root)+string(os.PathSeparator)) && target != filepath.Clean(root) {
		return "", fmt.Errorf("unsafe archive path %q", name)
	}
	return target, nil
}

func verifySHA(path, want string) error {
	if want == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if got != want {
		return fmt.Errorf("sha256 mismatch for %s: got %s want %s", path, got, want)
	}
	return nil
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func PromoteBinaries(root string, names []string) error {
	for _, name := range names {
		target := filepath.Join(root, name)
		if _, err := os.Stat(target); err == nil {
			continue
		}
		found, err := findUnder(root, name)
		if err != nil {
			return err
		}
		if err := copyFile(found, target); err != nil {
			return err
		}
		if err := os.Chmod(target, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func findUnder(root, name string) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Base(path) == name {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("binary %q not found in archive", name)
	}
	return found, nil
}
