package manifest

type Manifest struct {
	Models   map[string]Model   `json:"models"`
	Runtimes map[string]Runtime `json:"runtimes"`
}

type Model struct {
	Type    string      `json:"type"`
	Runtime string      `json:"runtime,omitempty"`
	Files   []ModelFile `json:"files"`
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
			"seaco-paraformer-trilingual": {
				Type:    "asr",
				Runtime: "sherpa-onnx",
				Files: []ModelFile{{
					Path: "sherpa-onnx-paraformer-trilingual-zh-cantonese-en.tar.bz2",
					URL:  "https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-paraformer-trilingual-zh-cantonese-en.tar.bz2",
				}},
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
					Path: "silero_vad.onnx",
					URL:  "",
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
						Binaries:   []string{"sherpa-onnx-offline", "sherpa-onnx-offline-punctuation"},
					},
					"linux-amd64": {
						ArchiveURL: "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-linux-x64-shared-no-tts.tar.bz2",
						URL:        "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-linux-x64-shared-no-tts.tar.bz2",
						SHA256:     "79f792f0a3ac0521b451a6ec3da6f446798c49924603320aaa69e6a7b281ec24",
						Binaries:   []string{"sherpa-onnx-offline", "sherpa-onnx-offline-punctuation"},
					},
					"windows-amd64": {
						ArchiveURL: "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-win-x64-shared-MD-Release-no-tts.tar.bz2",
						URL:        "https://github.com/k2-fsa/sherpa-onnx/releases/download/v1.13.0/sherpa-onnx-v1.13.0-win-x64-shared-MD-Release-no-tts.tar.bz2",
						SHA256:     "21e86a35fd2cd7b61b0f0d7f2dcb6dfac3e4e97910690a0be213ff9f02b86f19",
						Binaries:   []string{"sherpa-onnx-offline.exe", "sherpa-onnx-offline-punctuation.exe"},
					},
				},
			},
		},
	}
}
