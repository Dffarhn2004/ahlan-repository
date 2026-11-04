package httpapi

import (
	"fmt"
	"strings"
	"testing"

	"ponpes-saas/internal/crawler"
)

func TestCrawlFilenameIsStableUniqueAndSafe(t *testing.T) {
	a := crawlFilename(crawler.Page{URL: "https://pesantren.example.id/profil", Title: "Profil Pesantren: Sejarah & Visi!", Ext: ".txt"})
	b := crawlFilename(crawler.Page{URL: "https://pesantren.example.id/profil", Title: "Judul sudah diganti", Ext: ".txt"})
	c := crawlFilename(crawler.Page{URL: "https://pesantren.example.id/lain", Title: "Profil Pesantren: Sejarah & Visi!", Ext: ".txt"})
	if !strings.HasPrefix(a, "web-profil-pesantren-sejarah-visi-") || !strings.HasSuffix(a, ".txt") {
		t.Errorf("unexpected name %q", a)
	}
	if a == c {
		t.Error("same title on different URLs must not collide")
	}
	if strings.Split(a, "-")[len(strings.Split(a, "-"))-1] != strings.Split(b, "-")[len(strings.Split(b, "-"))-1] {
		t.Error("the URL hash part must not depend on the title")
	}
	for _, name := range []string{a, crawlFilename(crawler.Page{URL: "https://x.example/../../etc/passwd", Title: "../../SOUL.md", Ext: ".txt"}), crawlFilename(crawler.Page{URL: "https://x.example/", Ext: ".pdf"})} {
		if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") || len(name) > 80 {
			t.Errorf("unsafe filename %q", name)
		}
	}
}

func TestCrawlFailureMessages(t *testing.T) {
	cases := map[error]string{
		crawler.ErrNeedsPublicSharing: "NEEDS_PUBLIC_SHARING",
		crawler.ErrForbiddenURL:       "FAILED",
		crawler.ErrTooLarge:           "FAILED",
		crawler.ErrUnsupportedType:    "FAILED",
		fmt.Errorf("dial tcp: boom"):  "FAILED",
	}
	for err, want := range cases {
		status, code, message := crawlFailure(err)
		if status != want || code == "" || message == "" {
			t.Errorf("crawlFailure(%v) = %s %s %q", err, status, code, message)
		}
		if strings.Contains(message, "dial tcp") || strings.Contains(message, "boom") {
			t.Errorf("internal error text leaked into the public message: %q", message)
		}
	}
}
