package main

import (
	"reflect"
	"testing"
)

func TestNormalizeTranscribeArgsAllowsInputBeforeFlags(t *testing.T) {
	got := normalizeTranscribeArgs([]string{"input.mp4", "--hotwords", "hotwords.txt", "--out", "out.md"})
	want := []string{"--hotwords", "hotwords.txt", "--out", "out.md", "input.mp4"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeTranscribeArgs() = %#v, want %#v", got, want)
	}
}

func TestExtractASRTextFromJSON(t *testing.T) {
	got := extractASRText(`{"text":"第一段"}`)
	want := "第一段"

	if got != want {
		t.Fatalf("extractASRText() = %q, want %q", got, want)
	}
}

func TestExtractVADText(t *testing.T) {
	input := `3.046 -- 5.196: 虽然
11.526 -- 31.580: 每个人来到地上
noise`

	got := extractVADText(input)
	want := "虽然 每个人来到地上"

	if got != want {
		t.Fatalf("extractVADText() = %q, want %q", got, want)
	}
}

func TestDefaultModelForBackend(t *testing.T) {
	tests := map[string]string{
		"sensevoice": "sensevoice-small",
		"nano":       "funasr-nano-int8",
	}

	for backend, want := range tests {
		got, err := defaultModelForBackend(backend)
		if err != nil {
			t.Fatalf("defaultModelForBackend(%q) returned error: %v", backend, err)
		}
		if got != want {
			t.Fatalf("defaultModelForBackend(%q) = %q, want %q", backend, got, want)
		}
	}
}

func TestBuildASRArgsForSenseVoice(t *testing.T) {
	cfg := asrConfig{
		Backend:       backendSenseVoice,
		Model:         "/models/sense/model.int8.onnx",
		Tokens:        "/models/sense/tokens.txt",
		VADModel:      "/models/vad/silero_vad.onnx",
		Threads:       8,
		UseVAD:        true,
		SenseLanguage: "zh",
	}

	got := buildASRArgs(cfg, "/tmp/audio.wav")
	want := []string{
		"--silero-vad-model=/models/vad/silero_vad.onnx",
		"--tokens=/models/sense/tokens.txt",
		"--sense-voice-model=/models/sense/model.int8.onnx",
		"--sense-voice-language=zh",
		"--sense-voice-use-itn=true",
		"--print-args=false",
		"--num-threads=8",
		"/tmp/audio.wav",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildASRArgs() = %#v, want %#v", got, want)
	}
}

func TestBuildASRArgsForNanoPassesHotwords(t *testing.T) {
	cfg := asrConfig{
		Backend:     backendNano,
		Encoder:     "/models/nano/encoder_adaptor.int8.onnx",
		Embedding:   "/models/nano/embedding.int8.onnx",
		LLM:         "/models/nano/llm.int8.onnx",
		Tokenizer:   "/models/nano/Qwen3-0.6B",
		VADModel:    "/models/vad/silero_vad.onnx",
		Hotwords:    []string{"term1", "term2"},
		HotwordFile: "/tmp/hotwords.txt",
		Threads:     4,
		UseVAD:      true,
	}

	got := buildASRArgs(cfg, "/tmp/audio.wav")
	want := []string{
		"--silero-vad-model=/models/vad/silero_vad.onnx",
		"--funasr-nano-encoder-adaptor=/models/nano/encoder_adaptor.int8.onnx",
		"--funasr-nano-embedding=/models/nano/embedding.int8.onnx",
		"--funasr-nano-llm=/models/nano/llm.int8.onnx",
		"--funasr-nano-tokenizer=/models/nano/Qwen3-0.6B",
		"--funasr-nano-language=中文",
		"--funasr-nano-itn=true",
		"--funasr-nano-hotwords=term1,term2",
		"--print-args=false",
		"--num-threads=4",
		"/tmp/audio.wav",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildASRArgs() = %#v, want %#v", got, want)
	}
}
