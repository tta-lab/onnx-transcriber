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
