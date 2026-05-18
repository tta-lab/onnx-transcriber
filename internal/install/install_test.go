package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHumanBytes(t *testing.T) {
	tests := map[int64]string{
		512:        "512 B",
		1024:       "1.0 KiB",
		1048576:    "1.0 MiB",
		1073741824: "1.0 GiB",
	}

	for input, want := range tests {
		if got := humanBytes(input); got != want {
			t.Fatalf("humanBytes(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestPromoteBinariesCopiesAdjacentWindowsDLLs(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "sherpa-onnx", "bin")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "sherpa-onnx-offline.exe"), []byte("exe"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "onnxruntime.dll"), []byte("dll"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PromoteBinaries(root, []string{"sherpa-onnx-offline.exe"}); err != nil {
		t.Fatalf("PromoteBinaries returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "sherpa-onnx-offline.exe")); err != nil {
		t.Fatalf("promoted exe missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "onnxruntime.dll")); err != nil {
		t.Fatalf("adjacent dll missing: %v", err)
	}
}

func TestPromoteBinariesRepairsMissingDLLWhenBinaryAlreadyPromoted(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "sherpa-onnx", "bin")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sherpa-onnx-offline.exe"), []byte("old exe"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "sherpa-onnx-offline.exe"), []byte("exe"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "onnxruntime.dll"), []byte("dll"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := PromoteBinaries(root, []string{"sherpa-onnx-offline.exe"}); err != nil {
		t.Fatalf("PromoteBinaries returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "onnxruntime.dll")); err != nil {
		t.Fatalf("adjacent dll missing after repair: %v", err)
	}
}
