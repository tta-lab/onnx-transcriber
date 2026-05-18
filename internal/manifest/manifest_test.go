package manifest

import "testing"

func TestDefaultManifestContainsVerifiedMacRuntimeAndCoreModels(t *testing.T) {
	m := Default()

	runtimeAsset, ok := m.Runtimes["sherpa-onnx"].Platforms["darwin-arm64"]
	if !ok {
		t.Fatal("missing darwin-arm64 sherpa runtime")
	}
	if runtimeAsset.SHA256 == "" || runtimeAsset.URL == "" {
		t.Fatalf("runtime asset is not installable: %#v", runtimeAsset)
	}

}

func TestDefaultManifestContainsNanoASRModel(t *testing.T) {
	m := Default()
	nano, ok := m.Models["funasr-nano-int8"]
	if !ok {
		t.Fatal("missing funasr-nano-int8 model")
	}
	if nano.Backend != "nano" {
		t.Fatalf("funasr-nano-int8 backend = %q, want nano", nano.Backend)
	}
	want := map[string]bool{
		"encoder_adaptor.int8.onnx": true,
		"embedding.int8.onnx":       true,
		"llm.int8.onnx":             true,
		"Qwen3-0.6B/tokenizer.json": true,
	}
	for _, file := range nano.RequiredFiles {
		delete(want, file)
	}
	if len(want) != 0 {
		t.Fatalf("funasr-nano-int8 missing required files: %#v", want)
	}
}
