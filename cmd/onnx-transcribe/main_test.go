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
