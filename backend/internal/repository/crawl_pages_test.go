package repository

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCrawledPageUpdateAvailable(t *testing.T) {
	cases := []struct {
		name string
		page CrawledPage
		want bool
	}{
		{"untouched page never shows a notice", CrawledPage{Hash: "a", SourceHash: "b"}, false},
		{"edited and site unchanged", CrawledPage{UserEdited: true, Hash: "a", SourceHash: "a"}, false},
		{"edited and site changed", CrawledPage{UserEdited: true, Hash: "a", SourceHash: "b"}, true},
		{"edited, nothing fetched yet", CrawledPage{UserEdited: true, Hash: "a"}, false},
	}
	for _, c := range cases {
		if got := c.page.UpdateAvailable(); got != c.want {
			t.Errorf("%s: UpdateAvailable = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestCrawledPageEditable(t *testing.T) {
	id := uuid.New()
	ok := CrawledPage{Status: "ACTIVE", FileID: &id, Ext: ".txt", Body: "isi"}
	if !ok.Editable() {
		t.Error("an active text page with a file must be editable")
	}
	cases := map[string]CrawledPage{
		"pdf":       {Status: "ACTIVE", FileID: &id, Ext: ".pdf"},
		"no file":   {Status: "ACTIVE", Ext: ".txt"},
		"excluded":  {Status: "EXCLUDED", FileID: &id, Ext: ".txt"},
		"gone":      {Status: "GONE", FileID: &id, Ext: ".txt"},
		"oversized": {Status: "ACTIVE", FileID: &id, Ext: ".txt", Body: strings.Repeat("x", MaxEditableBytes+1)},
	}
	for name, page := range cases {
		if page.Editable() {
			t.Errorf("%s must not be editable", name)
		}
	}
	if csv := (CrawledPage{Status: "ACTIVE", FileID: &id, Ext: ".csv"}); !csv.Editable() {
		t.Error("sheets exported as csv are editable")
	}
}
