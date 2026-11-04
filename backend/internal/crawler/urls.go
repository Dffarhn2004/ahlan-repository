// Package crawler fetches the public web pages and Google documents a pesantren registers as knowledge sources.
// It is deliberately self-contained: it knows nothing about tenants, files or Hermes, so its safety rules
// (SSRF guard, robots.txt, size limits) can be tested in isolation. See PRD 25_crawl_sumber_web.md.
package crawler

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// Kind says how a source is fetched. The server always derives it from the URL; clients cannot choose it.
type Kind string

const (
	KindSite        Kind = "SITE"
	KindSinglePage  Kind = "SINGLE_PAGE"
	KindGoogleDoc   Kind = "GOOGLE_DOC"
	KindGoogleSheet Kind = "GOOGLE_SHEET"
	KindGoogleSlide Kind = "GOOGLE_SLIDES"
	KindDriveFile   Kind = "GOOGLE_DRIVE_FILE"
)

var (
	// ErrForbiddenURL means the URL, a redirect or the address it resolves to is not allowed.
	ErrForbiddenURL = errors.New("CRAWL_URL_FORBIDDEN")
	// ErrNeedsPublicSharing means a Google document exists but is not shared with "anyone with the link".
	ErrNeedsPublicSharing = errors.New("NEEDS_PUBLIC_SHARING")
	// ErrUnreachable covers network failures, timeouts and unexpected HTTP statuses.
	ErrUnreachable = errors.New("CRAWL_UNREACHABLE")
	// ErrTooLarge means the response exceeded the byte limit.
	ErrTooLarge = errors.New("CRAWL_TOO_LARGE")
)

var (
	googleID = regexp.MustCompile(`^[A-Za-z0-9_-]{20,}$`)
	sheetGID = regexp.MustCompile(`^[0-9]{1,12}$`)
)

// Target is a validated source. For Google kinds FetchURL is built by the server from the document ID and never
// taken from the pasted text.
type Target struct {
	Kind     Kind
	Start    *url.URL
	FetchURL string
	DocID    string
}

func (t Target) IsGoogle() bool {
	switch t.Kind {
	case KindGoogleDoc, KindGoogleSheet, KindGoogleSlide, KindDriveFile:
		return true
	}
	return false
}

// ParseSource validates a pasted URL and decides how it will be fetched.
func ParseSource(raw string) (Target, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return Target{}, ErrForbiddenURL
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return Target{}, ErrForbiddenURL
	}
	if p := u.Port(); p != "" && p != "80" && p != "443" {
		return Target{}, ErrForbiddenURL
	}
	host := strings.ToLower(u.Hostname())
	switch host {
	case "docs.google.com":
		return parseDocs(u)
	case "drive.google.com":
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) >= 3 && parts[0] == "file" && parts[1] == "d" && googleID.MatchString(parts[2]) {
			return Target{Kind: KindDriveFile, Start: u, DocID: parts[2], FetchURL: "https://drive.google.com/uc?export=download&id=" + parts[2]}, nil
		}
		return Target{}, ErrForbiddenURL // folders and share pages are not supported
	}
	if isGoogleMaps(u) {
		return Target{}, ErrForbiddenURL // maps are stored on the profile, never fetched
	}
	if host == "localhost" || !strings.Contains(host, ".") {
		return Target{}, ErrForbiddenURL
	}
	u.Fragment = ""
	return Target{Kind: KindSite, Start: u, FetchURL: u.String()}, nil
}

func parseDocs(u *url.URL) (Target, error) {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 3 || parts[1] != "d" || !googleID.MatchString(parts[2]) {
		return Target{}, ErrForbiddenURL
	}
	id := parts[2]
	base := "https://docs.google.com/" + parts[0] + "/d/" + id + "/export?format="
	switch parts[0] {
	case "document":
		return Target{Kind: KindGoogleDoc, Start: u, DocID: id, FetchURL: base + "txt"}, nil
	case "presentation":
		return Target{Kind: KindGoogleSlide, Start: u, DocID: id, FetchURL: base + "txt"}, nil
	case "spreadsheets":
		fetch := base + "csv"
		gid := u.Query().Get("gid")
		if gid == "" {
			if frag, err := url.ParseQuery(u.Fragment); err == nil {
				gid = frag.Get("gid")
			}
		}
		if sheetGID.MatchString(gid) {
			fetch += "&gid=" + gid
		}
		return Target{Kind: KindGoogleSheet, Start: u, DocID: id, FetchURL: fetch}, nil
	}
	return Target{}, ErrForbiddenURL
}

func isGoogleMaps(u *url.URL) bool {
	switch strings.ToLower(u.Hostname()) {
	case "maps.google.com", "maps.app.goo.gl":
		return true
	case "www.google.com", "google.com", "goo.gl":
		return strings.HasPrefix(u.Path, "/maps")
	}
	return false
}

// ErrInvalidMapsURL is returned by ValidateMapsURL.
var ErrInvalidMapsURL = errors.New("MAPS_URL_INVALID")

// ValidateMapsURL accepts an empty string or an https link that points at Google Maps. The URL is only stored
// and shown; the server never opens it.
func ValidateMapsURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	if len(raw) > 2048 {
		return ErrInvalidMapsURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Hostname() == "" || (u.Port() != "" && u.Port() != "443") {
		return ErrInvalidMapsURL
	}
	if !isGoogleMaps(u) {
		return ErrInvalidMapsURL
	}
	return nil
}
