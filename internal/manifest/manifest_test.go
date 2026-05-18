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

	punct, ok := m.Models["ct-transformer-zh-en"]
	if !ok {
		t.Fatal("missing ct-transformer-zh-en model")
	}
	if len(punct.Files) == 0 || punct.Files[0].SHA256 == "" {
		t.Fatalf("punctuation model file must include sha256: %#v", punct)
	}
}

func TestDefaultManifestContainsSenseVoiceAndNanoModels(t *testing.T) {
	m := Default()
	if _, ok := m.Models["seaco-paraformer-trilingual"]; ok {
		t.Fatal("seaco-paraformer-trilingual should not be in default manifest")
	}

	sense, ok := m.Models["sensevoice-small"]
	if !ok {
		t.Fatal("missing sensevoice-small model")
	}
	if sense.Backend != "sensevoice" {
		t.Fatalf("sensevoice-small backend = %q, want sensevoice", sense.Backend)
	}
	if len(sense.RequiredFiles) == 0 {
		t.Fatal("sensevoice-small must declare required files")
	}

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
