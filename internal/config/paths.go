package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const EnvHome = "ONNX_TRANSCRIBER_HOME"

func HomeDir() (string, error) {
	if override := os.Getenv(EnvHome); override != "" {
		return override, nil
	}

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "onnx-transcriber"), nil
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			return "", errors.New("LOCALAPPDATA is not set")
		}
		return filepath.Join(local, "onnx-transcriber"), nil
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "onnx-transcriber"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share", "onnx-transcriber"), nil
	}
}

func PlatformKey(goos, goarch string) string {
	return goos + "-" + goarch
}

func CurrentPlatformKey() string {
	return PlatformKey(runtime.GOOS, runtime.GOARCH)
}

func RuntimeBinDir(home, platform string) string {
	return filepath.Join(home, "bin", platform)
}

func ModelDir(home, name string) string {
	return filepath.Join(home, "models", name)
}
