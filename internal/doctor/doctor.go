package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/guion-opensource/onnx-transcriber/internal/config"
	"github.com/guion-opensource/onnx-transcriber/internal/manifest"
)

type Result struct {
	OK       bool
	Messages []string
}

type Options struct {
	ModelName string
	UseVAD    bool
}

func Run(home string, opts ...Options) Result {
	options := Options{ModelName: "sensevoice-small", UseVAD: true}
	if len(opts) > 0 {
		options = opts[0]
	}
	if options.ModelName == "" {
		options.ModelName = "sensevoice-small"
	}

	result := Result{OK: true}
	check := func(ok bool, msg string) {
		if !ok {
			result.OK = false
		}
		result.Messages = append(result.Messages, msg)
	}

	if ffmpeg, err := exec.LookPath("ffmpeg"); err == nil {
		check(true, "ok: ffmpeg found at "+ffmpeg)
	} else {
		check(false, "missing: ffmpeg not found in PATH")
	}

	binDir := config.RuntimeBinDir(home, config.CurrentPlatformKey())
	foundBins := map[string]string{}
	for _, name := range runtimeBinaryNames() {
		path, err := findExecutable(binDir, name)
		check(err == nil, status(err == nil, pathOrName(path, name)))
		if err == nil {
			foundBins[trimExe(name)] = path
		}
	}

	m := manifest.Default()
	model, ok := m.Models[options.ModelName]
	check(ok, status(ok, "ASR model manifest "+options.ModelName))
	if ok {
		modelDir := config.ModelDir(home, options.ModelName)
		asrOK := pathExists(modelDir)
		check(asrOK, status(asrOK, "ASR model directory "+options.ModelName))
		for _, file := range model.RequiredFiles {
			path := filepath.Join(modelDir, file)
			if !pathExists(path) {
				path = ""
				_ = filepath.WalkDir(modelDir, func(candidate string, d os.DirEntry, err error) error {
					if err != nil || d.IsDir() {
						return err
					}
					if strings.HasSuffix(candidate, file) {
						path = candidate
						return filepath.SkipAll
					}
					return nil
				})
			}
			check(path != "", status(path != "", "ASR model file "+file))
		}
		checkCapabilities(check, foundBins, model.Backend, options.UseVAD)
	}
	punctOK := pathExists(config.ModelDir(home, "ct-transformer-zh-en"))
	check(punctOK, status(punctOK, "punctuation model directory"))
	if options.UseVAD {
		vadOK := pathExists(filepath.Join(config.ModelDir(home, "silero-vad"), "silero_vad.onnx"))
		check(vadOK, status(vadOK, "VAD model"))
	}

	return result
}

func checkCapabilities(check func(bool, string), foundBins map[string]string, backend string, useVAD bool) {
	required := capabilityFlags(backend, useVAD)
	if len(required) == 0 {
		return
	}

	for bin, flags := range required {
		path := foundBins[bin]
		if path == "" {
			continue
		}
		help, err := exec.Command(path, "--help").CombinedOutput()
		if err != nil {
			check(false, fmt.Sprintf("missing: %s --help failed: %v", bin, err))
			continue
		}
		text := string(help)
		for _, flag := range flags {
			check(strings.Contains(text, flag), status(strings.Contains(text, flag), bin+" capability "+flag))
		}
	}
}

func capabilityFlags(backend string, useVAD bool) map[string][]string {
	asrFlags := map[string][]string{
		"sensevoice": {"sense-voice-model"},
		"nano":       {"funasr-nano-encoder-adaptor", "funasr-nano-hotwords"},
	}
	flags, ok := asrFlags[backend]
	if !ok {
		return nil
	}
	result := map[string][]string{"sherpa-onnx-offline": flags}
	if useVAD {
		vadFlags := append([]string{"silero-vad-model"}, flags...)
		result["sherpa-onnx-vad-with-offline-asr"] = vadFlags
	}
	return result
}

func Write(w io.Writer, r Result) {
	for _, msg := range r.Messages {
		fmt.Fprintln(w, msg)
	}
	if r.OK {
		fmt.Fprintln(w, "doctor: ok")
		return
	}
	fmt.Fprintln(w, "doctor: incomplete setup")
}

func runtimeBinaryNames() []string {
	if filepath.Separator == '\\' {
		return []string{"sherpa-onnx-offline.exe", "sherpa-onnx-offline-punctuation.exe", "sherpa-onnx-vad-with-offline-asr.exe"}
	}
	return []string{"sherpa-onnx-offline", "sherpa-onnx-offline-punctuation", "sherpa-onnx-vad-with-offline-asr"}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findExecutable(root, name string) (string, error) {
	var fallback string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Base(path) != name {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) == "bin" {
			fallback = path
			return filepath.SkipAll
		}
		if fallback == "" {
			fallback = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if fallback == "" {
		return "", os.ErrNotExist
	}
	return fallback, nil
}

func status(ok bool, name string) string {
	if ok {
		return "ok: " + name
	}
	return "missing: " + name
}

func pathOrName(path, name string) string {
	if path != "" {
		return path
	}
	return name
}

func trimExe(name string) string {
	return strings.TrimSuffix(name, ".exe")
}
