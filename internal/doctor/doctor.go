package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/guion-opensource/onnx-transcriber/internal/config"
)

type Result struct {
	OK       bool
	Messages []string
}

func Run(home string) Result {
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
	for _, name := range runtimeBinaryNames() {
		path, err := findExecutable(binDir, name)
		check(err == nil, status(err == nil, pathOrName(path, name)))
	}

	asrOK := pathExists(config.ModelDir(home, "seaco-paraformer-trilingual"))
	check(asrOK, status(asrOK, "ASR model directory"))
	punctOK := pathExists(config.ModelDir(home, "ct-transformer-zh-en"))
	check(punctOK, status(punctOK, "punctuation model directory"))
	vadOK := pathExists(filepath.Join(config.ModelDir(home, "silero-vad"), "silero_vad.onnx"))
	check(vadOK, status(vadOK, "VAD model"))

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
