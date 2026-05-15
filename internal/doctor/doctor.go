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
		path := filepath.Join(binDir, name)
		info, err := os.Stat(path)
		ok := err == nil && !info.IsDir()
		check(ok, status(ok, path))
	}

	asrOK := pathExists(config.ModelDir(home, "seaco-paraformer-trilingual"))
	check(asrOK, status(asrOK, "ASR model directory"))
	punctOK := pathExists(config.ModelDir(home, "ct-transformer-zh-en"))
	check(punctOK, status(punctOK, "punctuation model directory"))

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
		return []string{"sherpa-onnx-offline.exe", "sherpa-onnx-offline-punctuation.exe"}
	}
	return []string{"sherpa-onnx-offline", "sherpa-onnx-offline-punctuation"}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func status(ok bool, name string) string {
	if ok {
		return "ok: " + name
	}
	return "missing: " + name
}
