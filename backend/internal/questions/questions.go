// Package questions collects what the assistant could not answer and turns admin-written answers into persona text.
package questions

import (
	"strings"
	"unicode"

	"ponpes-saas/internal/links"
)

const (
	StatusOpen     = "OPEN"
	StatusAnswered = "ANSWERED"
	StatusResolved = "RESOLVED"
	StatusIgnored  = "IGNORED"

	MaxOpen     = 200 // open questions kept per pesantren; the oldest are dropped first
	MaxAnswered = 100 // answers handed to the assistant per pesantren, so the persona stays small
)

// Marker is the fixed wording the assistant is told to use when it has no answer. Spotting it in a reply is how an
// unanswered question is detected, so the persona rules and this constant must change together.
const Marker = "belum tersedia di data kami"

// Sentence is the full sentence given to the assistant in its rules.
const Sentence = "Mohon maaf, informasi tersebut " + Marker + "."

// IsUnanswered reports whether the assistant's reply says it has no answer.
func IsUnanswered(reply string) bool {
	return strings.Contains(strings.ToLower(reply), Marker)
}

// Key normalises a question so the same question asked in slightly different ways is counted once: lowercase,
// punctuation dropped, spaces collapsed, at most 200 runes.
func Key(question string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(question) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(r)
			continue
		}
		space = true
	}
	out := b.String()
	if runes := []rune(out); len(runes) > 200 {
		out = strings.TrimSpace(string(runes[:200]))
	}
	return out
}

// Answered is an admin-written question and answer the assistant may use.
type Answered struct{ Question, Answer string }

// PersonaSection gives the answers to the assistant. Both fields pass through links.Clean (one line, no control
// characters, length-limited), so admin text can never add instruction lines to the persona.
func PersonaSection(items []Answered) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nTanya jawab resmi pesantren. Jika pertanyaan penanya sama atau mirip dengan salah satu di bawah, jawab berdasarkan jawabannya (boleh dirangkum dengan bahasa Anda sendiri, tanpa mengubah isi). Jangan menambahkan fakta di luar jawaban tersebut.\n")
	for _, item := range items {
		question, answer := links.Clean(item.Question, 200), links.Clean(item.Answer, 1500)
		if question == "" || answer == "" {
			continue
		}
		b.WriteString("- Tanya: " + question + "\n  Jawab: " + answer + "\n")
	}
	return b.String()
}
