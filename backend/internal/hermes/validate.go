package hermes

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ErrInvalidInput is returned for any parameter that fails validation. The message never contains the offending value.
var ErrInvalidInput = errors.New("invalid Hermes parameter")

// Profile keys are created by the provisioning job as 'ponpes_' + tenant slug. Nothing else may be addressed.
var profileKeyPattern = regexp.MustCompile(`^ponpes_[a-z0-9][a-z0-9_-]{0,56}$`)

// ValidateProfileKey accepts only provisioning-generated keys, so a profile key can never carry a path or option.
func ValidateProfileKey(key string) error {
	if !profileKeyPattern.MatchString(key) {
		return fmt.Errorf("%w: profile key", ErrInvalidInput)
	}
	return nil
}

// Knowledge names are "file.ext" or "folder_<uuid>/file.ext" (optionally + TranscriptSuffix). Each segment must be a
// plain name: no traversal, no empty segment, no control characters or backslash.
const maxKnowledgeNameBytes = 255

// ValidateKnowledgeName rejects absolute paths, '..' segments, option-like names and unsafe characters.
func ValidateKnowledgeName(name string) error {
	if name == "" || len(name) > maxKnowledgeNameBytes || !utf8.ValidString(name) || strings.HasPrefix(name, "/") {
		return fmt.Errorf("%w: file name", ErrInvalidInput)
	}
	segments := strings.Split(name, "/")
	if len(segments) > 2 {
		return fmt.Errorf("%w: file name", ErrInvalidInput)
	}
	for _, segment := range segments {
		// Names always live under an absolute profile directory, so a leading dash or dot is harmless; traversal is not.
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("%w: file name", ErrInvalidInput)
		}
		for _, r := range segment {
			// Shell metacharacters are fine: every name is single-quoted or passed as a plain argument.
			if r < 0x20 || r == 0x7f || r == '\\' {
				return fmt.Errorf("%w: file name", ErrInvalidInput)
			}
		}
	}
	return nil
}

var envPrefixes = map[string]bool{"TELEGRAM_": true, "WHATSAPP_": true}

// ValidateChannelPrefix accepts only the two channel prefixes the dashboard manages.
func ValidateChannelPrefix(prefix string) error {
	if !envPrefixes[prefix] {
		return fmt.Errorf("%w: channel", ErrInvalidInput)
	}
	return nil
}

// noControl rejects NUL, newlines and other control characters, which could inject extra .env lines.
func noControl(value string) bool {
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

var telegramTokenPattern = regexp.MustCompile(`^[0-9]{3,20}:[A-Za-z0-9_-]{16,128}$`)

// ValidateTelegramToken checks the BotFather token shape, which also rules out any env-line injection.
func ValidateTelegramToken(token string) error {
	if !telegramTokenPattern.MatchString(strings.TrimSpace(token)) {
		return fmt.Errorf("%w: Telegram token", ErrInvalidInput)
	}
	return nil
}

var phonePattern = regexp.MustCompile(`^\+?[0-9]{5,20}$`)

// ValidateWhatsAppNumber accepts digits with an optional leading plus.
func ValidateWhatsAppNumber(number string) error {
	if !phonePattern.MatchString(strings.TrimSpace(number)) {
		return fmt.Errorf("%w: WhatsApp number", ErrInvalidInput)
	}
	return nil
}

var eventKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._~+/=-]{8,256}$`)

// ValidateEventKey checks the shared secret that the log hook sends back to the API.
func ValidateEventKey(key string) error {
	if !eventKeyPattern.MatchString(key) {
		return fmt.Errorf("%w: event key", ErrInvalidInput)
	}
	return nil
}

// ValidateEventURL accepts a plain http(s) URL without credentials, query or whitespace.
func ValidateEventURL(raw string) error {
	if len(raw) > 512 || !noControl(raw) || strings.ContainsAny(raw, " \"'`\\") {
		return fmt.Errorf("%w: event URL", ErrInvalidInput)
	}
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%w: event URL", ErrInvalidInput)
	}
	return nil
}

var containerPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,127}$`)

// ValidateContainerName checks the configured container name before it is handed to the Docker CLI.
func ValidateContainerName(name string) error {
	if !containerPattern.MatchString(name) {
		return fmt.Errorf("%w: container name", ErrInvalidInput)
	}
	return nil
}

// Size limits for content written into a profile.
const (
	MaxSoulBytes        = 400 * 1024
	MaxDescriptionBytes = 8 * 1024
	MaxKnowledgeBytes   = 64 << 20
)
