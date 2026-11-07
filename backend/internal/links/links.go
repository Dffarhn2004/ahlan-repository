// Package links holds the rules for a pesantren's official links: which kinds exist, which domain each kind may point at,
// which links the crawler can learn from, and how the shareable ones are described to the assistant.
package links

import (
	"errors"
	"net/netip"
	"net/url"
	"strings"
	"unicode"

	"ponpes-saas/internal/crawler"
)

var (
	ErrInvalidKind  = errors.New("LINK_KIND_INVALID")
	ErrInvalidURL   = errors.New("LINK_URL_INVALID")
	ErrWrongDomain  = errors.New("LINK_DOMAIN_MISMATCH")
	ErrCannotLearn  = errors.New("LINK_CANNOT_LEARN")
	MaxLinksPerSite = 20
)

type kindRule struct {
	label string
	hosts []string // empty: any public https domain
	learn bool     // the crawler can read this kind of page
}

var kinds = map[string]kindRule{
	"PENDAFTARAN": {label: "Pendaftaran online", learn: true},
	"BROSUR":      {label: "Brosur", learn: true},
	"JADWAL":      {label: "Jadwal", learn: true},
	"BIAYA":       {label: "Biaya", learn: true},
	"WEBSITE":     {label: "Website", learn: true},
	"LAINNYA":     {label: "Tautan lain", learn: true},
	"WHATSAPP":    {label: "WhatsApp", hosts: []string{"wa.me", "api.whatsapp.com", "chat.whatsapp.com", "wa.link"}},
	"INSTAGRAM":   {label: "Instagram", hosts: []string{"instagram.com", "instagr.am"}},
	"YOUTUBE":     {label: "YouTube", hosts: []string{"youtube.com", "youtu.be"}},
	"FACEBOOK":    {label: "Facebook", hosts: []string{"facebook.com", "fb.com", "fb.me"}},
	"TIKTOK":      {label: "TikTok", hosts: []string{"tiktok.com"}},
}

// Hosts that need a login or render everything in script, so reading them as text would only yield an empty shell.
var unlearnable = []string{"forms.gle", "wa.me", "instagram.com", "facebook.com", "fb.com", "fb.me", "tiktok.com", "youtube.com", "youtu.be", "linktr.ee"}

func hostMatches(host, domain string) bool {
	return host == domain || strings.HasSuffix(host, "."+domain)
}

// Label is the Indonesian name of a kind, shown to people and to the assistant.
func Label(kind string) string { return kinds[kind].label }

// KindExists reports whether kind is one of the supported kinds.
func KindExists(kind string) bool { _, ok := kinds[kind]; return ok }

// Validate checks a pasted link against its kind and returns it normalised.
func Validate(kind, raw string) (string, error) {
	rule, ok := kinds[kind]
	if !ok {
		return "", ErrInvalidKind
	}
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return "", ErrInvalidURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Hostname() == "" || (u.Port() != "" && u.Port() != "443") {
		return "", ErrInvalidURL
	}
	host := strings.ToLower(u.Hostname())
	if _, err := netip.ParseAddr(host); err == nil || host == "localhost" || !strings.Contains(host, ".") {
		return "", ErrInvalidURL
	}
	if len(rule.hosts) > 0 {
		matched := false
		for _, domain := range rule.hosts {
			if hostMatches(host, domain) {
				matched = true
				break
			}
		}
		if !matched {
			return "", ErrWrongDomain
		}
	}
	return u.String(), nil
}

// CanLearn reports whether a link of this kind and address can also be registered as a crawl source.
func CanLearn(kind, link string) bool {
	if !kinds[kind].learn {
		return false
	}
	target, err := crawler.ParseSource(link)
	if err != nil {
		return false
	}
	host := strings.ToLower(target.Start.Hostname())
	for _, domain := range unlearnable {
		if hostMatches(host, domain) {
			return false
		}
	}
	return true
}

// Clean makes admin-written text safe to place inside the assistant's instructions: one line, no control characters,
// and a hard length limit, so a label can never smuggle in extra instructions on new lines.
func Clean(s string, max int) string {
	var b strings.Builder
	space := false
	for _, r := range s {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			space = b.Len() > 0
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(r)
	}
	out := b.String()
	if runes := []rune(out); len(runes) > max {
		out = strings.TrimSpace(string(runes[:max]))
	}
	return out
}

// Shared is a link the assistant may hand to a prospective student.
type Shared struct{ Kind, Label, URL, Hint string }

// PersonaSection describes the shareable links to the assistant. It tells the assistant to use only these links.
func PersonaSection(items []Shared) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nTautan resmi pesantren. Bagikan hanya tautan dari daftar ini, apa adanya. Jangan membuat, menebak, atau mengubah tautan; jika tidak ada yang cocok, katakan bahwa Anda tidak memiliki tautannya dan arahkan ke kontak pesantren.\n")
	for _, item := range items {
		b.WriteString("- " + Label(item.Kind) + " — " + Clean(item.Label, 80) + ": " + item.URL + "\n")
		if hint := Clean(item.Hint, 160); hint != "" {
			b.WriteString("  Bagikan saat: " + hint + "\n")
		}
	}
	return b.String()
}
