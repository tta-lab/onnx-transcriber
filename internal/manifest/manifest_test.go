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

	asr, ok := m.Models["seaco-paraformer-trilingual"]
	if !ok {
		t.Fatal("missing seaco-paraformer-trilingual model")
	}
	if len(asr.Files) == 0 || asr.Files[0].URL == "" {
		t.Fatalf("asr model is not installable: %#v", asr)
	}

	punct, ok := m.Models["ct-transformer-zh-en"]
	if !ok {
		t.Fatal("missing ct-transformer-zh-en model")
	}
	if len(punct.Files) == 0 || punct.Files[0].SHA256 == "" {
		t.Fatalf("punctuation model file must include sha256: %#v", punct)
	}
}
