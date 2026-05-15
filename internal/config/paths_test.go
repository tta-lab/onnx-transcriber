package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestHomeDirPrefersEnvironmentOverride(t *testing.T) {
	t.Setenv("ONNX_TRANSCRIBER_HOME", "/tmp/custom-onnx-home")

	got, err := HomeDir()
	if err != nil {
		t.Fatal(err)
	}

	if got != "/tmp/custom-onnx-home" {
		t.Fatalf("HomeDir() = %q, want override", got)
	}
}

func TestPlatformKeyMapsGoRuntimeNames(t *testing.T) {
	got := PlatformKey("darwin", "arm64")
	if got != "darwin-arm64" {
		t.Fatalf("PlatformKey(darwin, arm64) = %q", got)
	}

	got = PlatformKey("windows", "amd64")
	if got != "windows-amd64" {
		t.Fatalf("PlatformKey(windows, amd64) = %q", got)
	}
}

func TestHomeDirUsesOSDefault(t *testing.T) {
	t.Setenv("ONNX_TRANSCRIBER_HOME", "")

	got, err := HomeDir()
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(got) != "onnx-transcriber" {
		t.Fatalf("HomeDir() = %q, want path ending in onnx-transcriber", got)
	}

	if runtime.GOOS == "darwin" && filepath.Base(filepath.Dir(got)) != "Application Support" {
		t.Fatalf("HomeDir() = %q, want macOS Application Support path", got)
	}
}
