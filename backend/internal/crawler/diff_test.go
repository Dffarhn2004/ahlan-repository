package crawler

import (
	"strings"
	"testing"
	"time"
)

func TestDiffLines(t *testing.T) {
	cases := []struct {
		before, after   string
		added, removed int
	}{
		{"", "a\nb", 2, 0},
		{"a\nb", "", 0, 2},
		{"a\nb\nc", "a\nb\nc", 0, 0},
		{"a\nb\nc", "a\nx\nc", 1, 1},
		{"  a  \n\n\nb", "a\nb", 0, 0},     // whitespace-only differences are not changes
		{"a\na", "a", 0, 1},                // duplicate lines are counted
		{"biaya 500", "biaya 750\nbaru", 2, 1},
	}
	for _, c := range cases {
		added, removed := DiffLines(c.before, c.after)
		if added != c.added || removed != c.removed {
			t.Errorf("DiffLines(%q,%q) = %d,%d want %d,%d", c.before, c.after, added, removed, c.added, c.removed)
		}
	}
}

func TestRenderDocument(t *testing.T) {
	at := time.Date(2026, 10, 31, 3, 0, 0, 0, time.UTC)
	plain := string(RenderDocument("https://x.example/a", "Judul", "isi", at, nil))
	if !strings.Contains(plain, "Sumber: https://x.example/a\nJudul: Judul\nDiambil: 31 Oct 2026 10:00 WIB\n") || strings.Contains(plain, "Dikoreksi") || !strings.HasSuffix(plain, "---\n\nisi\n") {
		t.Errorf("unexpected document:\n%s", plain)
	}
	edited := time.Date(2026, 11, 1, 5, 30, 0, 0, time.UTC)
	withEdit := string(RenderDocument("https://x.example/a", "", "isi", at, &edited))
	if strings.Contains(withEdit, "Judul:") || !strings.Contains(withEdit, "Dikoreksi oleh pengelola pesantren: 1 Nov 2026 12:30 WIB") {
		t.Errorf("edited document wrong:\n%s", withEdit)
	}
}
