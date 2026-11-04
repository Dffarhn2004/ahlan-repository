package crawler

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// DiffLines counts the lines present in after but not before (added) and the reverse (removed). It compares lines as
// a multiset, ignoring order, which is what matters when telling a pesantren "what changed on the site".
func DiffLines(before, after string) (added, removed int) {
	counts := map[string]int{}
	for _, line := range meaningfulLines(before) {
		counts[line]++
	}
	for _, line := range meaningfulLines(after) {
		if counts[line] > 0 {
			counts[line]--
		} else {
			added++
		}
	}
	for _, n := range counts {
		removed += n
	}
	return added, removed
}

func meaningfulLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// RenderDocument builds the stored file for a text page: a header the assistant can cite, then the body. editedAt marks a
// body the pesantren corrected by hand.
func RenderDocument(pageURL, title, body string, fetchedAt time.Time, editedAt *time.Time) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "Sumber: %s\n", pageURL)
	if title != "" {
		fmt.Fprintf(&b, "Judul: %s\n", title)
	}
	fmt.Fprintf(&b, "Diambil: %s\n", fetchedAt.In(wib).Format("2 Jan 2006 15:04 WIB"))
	if editedAt != nil {
		fmt.Fprintf(&b, "Dikoreksi oleh pengelola pesantren: %s\n", editedAt.In(wib).Format("2 Jan 2006 15:04 WIB"))
	}
	b.WriteString("\n---\n\n")
	b.WriteString(body)
	b.WriteByte('\n')
	return b.Bytes()
}
