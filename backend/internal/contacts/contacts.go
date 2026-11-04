// Package contacts holds the rules for a pesantren's contact persons: how phone numbers and emails are normalised,
// and how the contacts are described to the assistant, which hands them to people who ask for a contact.
package contacts

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"

	"ponpes-saas/internal/links"
)

var (
	ErrInvalidPhone = errors.New("CONTACT_PHONE_INVALID")
	ErrInvalidEmail = errors.New("CONTACT_EMAIL_INVALID")
	MaxPerSite      = 20
)

// NormalizePhone turns what an admin typed ("0812-3456-7890", "+62 812 3456 7890", "62812…") into +<digits>.
// A leading 0 is read as Indonesia (+62). Anything with fewer than 8 or more than 15 digits is rejected.
func NormalizePhone(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	plus := strings.HasPrefix(raw, "+")
	var digits strings.Builder
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			digits.WriteRune(r)
		case r == '+' || r == ' ' || r == '-' || r == '.' || r == '(' || r == ')':
		default:
			return "", ErrInvalidPhone
		}
	}
	number := digits.String()
	switch {
	case plus:
	case strings.HasPrefix(number, "00"):
		number = strings.TrimLeft(number, "0")
	case strings.HasPrefix(number, "0"):
		number = "62" + strings.TrimPrefix(number, "0")
	case strings.HasPrefix(number, "62"):
	default:
		return "", ErrInvalidPhone
	}
	if len(number) < 8 || len(number) > 15 || strings.HasPrefix(number, "0") {
		return "", ErrInvalidPhone
	}
	return "+" + number, nil
}

// NormalizeEmail accepts an empty address (the field is optional) or a plain address like name@domain.id.
func NormalizeEmail(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	addr, err := mail.ParseAddress(raw)
	if err != nil || addr.Address != raw || len(raw) > 120 || !strings.Contains(raw[strings.LastIndex(raw, "@"):], ".") {
		return "", ErrInvalidEmail
	}
	for _, r := range raw {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return "", ErrInvalidEmail
		}
	}
	return strings.ToLower(raw), nil
}

// Shared is a contact the assistant may hand to a person who asks for one.
type Shared struct{ Name, Role, Phone, Email, Hours, Hint string }

// PersonaSection describes the contacts to the assistant. Every field passes through links.Clean, so admin-written
// text can never add instruction lines to the persona.
func PersonaSection(items []Shared) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nKontak person pesantren. Boleh diberikan kepada penanya yang meminta kontak, nomor, atau ingin berbicara langsung dengan pihak pesantren. Berikan hanya kontak dari daftar ini, apa adanya; pilih yang paling sesuai dengan keperluan penanya. Jangan membuat atau menebak nomor, dan jangan memberikan kontak jika penanya tidak memintanya.\n")
	for _, item := range items {
		line := "- " + links.Clean(item.Name, 80)
		if role := links.Clean(item.Role, 80); role != "" {
			line += " (" + role + ")"
		}
		line += ": " + item.Phone
		if email := links.Clean(item.Email, 120); email != "" {
			line += ", email " + email
		}
		b.WriteString(line + "\n")
		if hours := links.Clean(item.Hours, 80); hours != "" {
			b.WriteString("  Jam layanan: " + hours + "\n")
		}
		if hint := links.Clean(item.Hint, 160); hint != "" {
			b.WriteString("  Hubungi untuk: " + hint + "\n")
		}
	}
	return b.String()
}
