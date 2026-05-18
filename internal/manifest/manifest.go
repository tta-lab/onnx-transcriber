package manifest

type Manifest struct {
	Models   map[string]Model   `json:"models"`
	Runtimes map[string]Runtime `json:"runtimes"`
}

type Model struct {
	Type          string      `json:"type"`
	Runtime       string      `json:"runtime,omitempty"`
	Backend       string      `json:"backend,omitempty"`
	Files         []ModelFile `json:"files"`
	RequiredFiles []string    `json:"requiredFiles,omitempty"`
}

type ModelFile struct {
	Path   string `json:"path"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256,omitempty"`
}

type Runtime struct {
	Platforms map[string]RuntimeAsset `json:"platforms"`
}

type RuntimeAsset struct {
	ArchiveURL string   `json:"archiveUrl"`
	URL        string   `json:"url"`
	SHA256     string   `json:"sha256"`
	Binaries   []string `json:"binaries"`
}

func Default() Manifest {
	return Manifest{
		Models: map[string]Model{
			"funasr-nano-int8": {
				Type:    "asr",
				Runtime: "sherpa-onnx",
				Backend: "nano",
				Files: []ModelFile{{
					Path: "sherpa-onnx-funasr-nano-int8-2025-12-30.tar.bz2",
					URL:  "https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-funasr-nano-int8-2025-12-30.tar.bz2",
				}},
				RequiredFiles: []string{
					"encoder_adaptor.int8.onnx",
					"embedding.int8.onnx",
					"llm.int8.onnx",
					"Qwen3-0.6B/tokenizer.json",
					"Qwen3-0.6B/vocab.json",
					"Qwen3-0.6B/merges.txt",
				},
			},
			"ct-transformer-zh-en": {
				Type:    "punctuation",
				Runtime: "sherpa-onnx",
				Files: []ModelFile{{
					Path:   "sherpa-onnx-punct-ct-transformer-zh-en-vocab272727-2024-04-12-int8.tar.bz2",
					URL:    "https://github.com/k2-fsa/sherpa-onnx/releases/download/punctuation-models/sherpa-onnx-punct-ct-transformer-zh-en-vocab272727-2024-04-12-int8.tar.bz2",
					SHA256: "c0d5aa5f8eeb686032345e180bedf39319dc2e0556781c6264bcadba8328a6e1",
				}},
			},
			"silero-vad": {
				Type:    "vad",
				Runtime: "sherpa-onnx",
				Files: []ModelFile{{
					Path:   "silero_vad.onnx",
					URL:    "https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/silero_vad.onnx",
					SHA256: "9e2449e1087496d8d4caba907f23e0bd3f78d91fa552479bb9c23ac09cbb1fd6",
				}},
			},
		},
		Runtimes: map[string]Runtime{
			"sherpa-onnx": {
				Platforms: map[string]RuntimeAsset{
					"darwin-arm64": {
						ArchiveURL: "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-osx-arm64-shared-no-tts.tar.bz2",
						URL:        "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-osx-arm64-shared-no-tts.tar.bz2",
						SHA256:     "5eb3ff7ed1e17fd54b0aba5d9b7b1ca7911715582aa811d825c94e84e216abe1",
						Binaries:   []string{"sherpa-onnx-offline", "sherpa-onnx-offline-punctuation", "sherpa-onnx-vad-with-offline-asr"},
					},
					"linux-amd64": {
						ArchiveURL: "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-linux-x64-shared-no-tts.tar.bz2",
						URL:        "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-linux-x64-shared-no-tts.tar.bz2",
						SHA256:     "79f792f0a3ac0521b451a6ec3da6f446798c49924603320aaa69e6a7b281ec24",
						Binaries:   []string{"sherpa-onnx-offline", "sherpa-onnx-offline-punctuation", "sherpa-onnx-vad-with-offline-asr"},
					},
					"windows-amd64": {
						ArchiveURL: "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-win-x64-shared-MD-Release-no-tts.tar.bz2",
						URL:        "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-win-x64-shared-MD-Release-no-tts.tar.bz2",
						SHA256:     "21e86a35fd2cd7b61b0f0d7f2dcb6dfac3e4e97910690a0be213ff9f02b86f19",
						Binaries:   []string{"sherpa-onnx-offline.exe", "sherpa-onnx-offline-punctuation.exe", "sherpa-onnx-vad-with-offline-asr.exe"},
					},
				},
			},
		},
	}
}
