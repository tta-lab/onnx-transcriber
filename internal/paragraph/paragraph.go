package paragraph

import (
	"strings"
	"unicode/utf8"
)

type Options struct {
	SentencesPerParagraph int
	MaxChars              int
}

func Format(input string, opts Options) string {
	text := strings.TrimSpace(input)
	if text == "" {
		return ""
	}
	if opts.SentencesPerParagraph <= 0 {
		opts.SentencesPerParagraph = 3
	}
	if opts.MaxChars <= 0 {
		opts.MaxChars = 500
	}

	sentences := splitSentences(text)
	var paragraphs []string
	var current strings.Builder
	count := 0

	flush := func() {
		if current.Len() == 0 {
			return
		}
		paragraphs = append(paragraphs, strings.TrimSpace(current.String()))
		current.Reset()
		count = 0
	}

	for _, sentence := range sentences {
		if sentence == "" {
			continue
		}
		nextLen := utf8.RuneCountInString(current.String()) + utf8.RuneCountInString(sentence)
		if current.Len() > 0 && (count >= opts.SentencesPerParagraph || nextLen > opts.MaxChars) {
			flush()
		}
		current.WriteString(sentence)
		count++
	}
	flush()

	return strings.Join(paragraphs, "\n\n")
}

func splitSentences(text string) []string {
	var out []string
	var current strings.Builder
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' {
			r = ' '
		}
		current.WriteRune(r)
		if isStrongPunctuation(r) {
			out = append(out, strings.TrimSpace(current.String()))
			current.Reset()
		}
	}
	if tail := strings.TrimSpace(current.String()); tail != "" {
		out = append(out, tail)
	}
	return out
}

func isStrongPunctuation(r rune) bool {
	switch r {
	case '。', '！', '？', '.', '!', '?':
		return true
	default:
		return false
	}
}
