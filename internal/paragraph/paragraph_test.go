package paragraph

import "testing"

func TestFormatSplitsChineseSentencesIntoReadableParagraphs(t *testing.T) {
	input := "第一句。第二句！第三句？第四句。第五句。"

	got := Format(input, Options{SentencesPerParagraph: 2, MaxChars: 100})
	want := "第一句。第二句！\n\n第三句？第四句。\n\n第五句。"

	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatRespectsMaxChars(t *testing.T) {
	input := "这是第一段很长的内容没有必要和后面放在一起。这里是第二句。"

	got := Format(input, Options{SentencesPerParagraph: 10, MaxChars: 18})
	want := "这是第一段很长的内容没有必要和后面放在一起。\n\n这里是第二句。"

	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}
