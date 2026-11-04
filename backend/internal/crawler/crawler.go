package crawler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	maxPageBytes  = 2 << 20
	maxDriveBytes = 20 << 20
	maxRobotsSize = 256 << 10

	DefaultMaxDepth = 2
	DefaultMaxPages = 50
	hardMaxDepth    = 3
	hardMaxPages    = 200
)

// ErrUnsupportedType is returned for Drive files that are neither text, PDF nor an image.
var ErrUnsupportedType = errors.New("CRAWL_UNSUPPORTED_TYPE")

var wib = time.FixedZone("WIB", 7*3600)

// Page is one collected document. Hash covers only the content, so a changed "fetched at" date never counts as an edit.
type Page struct {
	URL    string
	Title  string
	Body   []byte
	Ext    string // file extension to store under, e.g. ".txt"
	Binary bool   // true for PDFs and images, which are stored as received
	Hash   string
}

// Document is the file content to store: text pages get a header so the assistant can cite where and when the
// information was read.
func (p Page) Document(at time.Time) []byte {
	if p.Binary {
		return p.Body
	}
	return RenderDocument(p.URL, p.Title, string(p.Body), at, nil)
}

// Options bounds a run. Zero values mean the defaults.
type Options struct {
	MaxDepth int
	MaxPages int
	Delay    time.Duration // pause between requests to the same site
}

// Result is what a run collected. Truncated means the page limit stopped the crawl, so pages not seen must not be
// treated as removed.
type Result struct {
	Pages     []Page
	Failed    int
	Blocked   int // skipped because robots.txt forbids them
	Truncated bool
}

type Crawler struct{ fetcher *Fetcher }

func New(f *Fetcher) *Crawler { return &Crawler{fetcher: f} }

// Run collects the pages for a validated target.
func (c *Crawler) Run(ctx context.Context, t Target, opt Options) (Result, error) {
	if t.IsGoogle() {
		return c.runGoogle(ctx, t)
	}
	return c.runSite(ctx, t, opt)
}

func hashOf(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (c *Crawler) runGoogle(ctx context.Context, t Target) (Result, error) {
	limit := int64(maxPageBytes)
	if t.Kind == KindDriveFile {
		limit = maxDriveBytes
	}
	resp, err := c.fetcher.Get(ctx, t.FetchURL, limit)
	if err != nil {
		return Result{}, err
	}
	// Private documents bounce to the Google sign-in page or answer 401/403.
	if strings.EqualFold(resp.FinalURL.Hostname(), "accounts.google.com") || resp.Status == 401 || resp.Status == 403 {
		return Result{}, ErrNeedsPublicSharing
	}
	if resp.Status != 200 {
		return Result{}, fmt.Errorf("%w: status %d", ErrUnreachable, resp.Status)
	}
	title := dispositionTitle(resp.Disposition)
	page := Page{URL: t.Start.String(), Title: title}
	switch t.Kind {
	case KindGoogleDoc, KindGoogleSlide, KindGoogleSheet:
		// Text and CSV exports are served as such; an HTML body is the sign-in or error page.
		if strings.Contains(resp.ContentType, "html") {
			return Result{}, ErrNeedsPublicSharing
		}
		page.Body = bytes.TrimPrefix(bytes.ToValidUTF8(resp.Body, nil), []byte("\xef\xbb\xbf"))
		page.Ext = ".txt"
		if t.Kind == KindGoogleSheet {
			page.Ext = ".csv"
		}
		if len(bytes.TrimSpace(page.Body)) == 0 {
			return Result{}, fmt.Errorf("%w: empty document", ErrUnreachable)
		}
	case KindDriveFile:
		switch {
		case strings.HasPrefix(resp.ContentType, "application/pdf"):
			page.Ext, page.Binary = ".pdf", true
		case strings.HasPrefix(resp.ContentType, "image/png"):
			page.Ext, page.Binary = ".png", true
		case strings.HasPrefix(resp.ContentType, "image/jpeg"):
			page.Ext, page.Binary = ".jpg", true
		case strings.HasPrefix(resp.ContentType, "text/plain"):
			page.Ext = ".txt"
		default:
			return Result{}, ErrUnsupportedType
		}
		page.Body = resp.Body
		if !page.Binary {
			page.Body = bytes.ToValidUTF8(page.Body, nil)
		}
	}
	page.Hash = hashOf(page.Body)
	return Result{Pages: []Page{page}}, nil
}

func dispositionTitle(header string) string {
	if header == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}
	name := params["filename"]
	return strings.TrimSuffix(name, path.Ext(name))
}

type queued struct {
	url   string
	depth int
}

func (c *Crawler) runSite(ctx context.Context, t Target, opt Options) (Result, error) {
	if opt.MaxDepth <= 0 && t.Kind != KindSinglePage {
		opt.MaxDepth = DefaultMaxDepth
	}
	if t.Kind == KindSinglePage {
		opt.MaxDepth = 0
	}
	if opt.MaxPages <= 0 {
		opt.MaxPages = DefaultMaxPages
	}
	opt.MaxDepth, opt.MaxPages = min(opt.MaxDepth, hardMaxDepth), min(opt.MaxPages, hardMaxPages)

	site := siteKey(t.Start)
	rules := c.loadRobots(ctx, t.Start)
	queue := []queued{{normalize(t.Start), 0}}
	seen := map[string]bool{queue[0].url: true}
	var res Result
	first := true
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if len(res.Pages) >= opt.MaxPages {
			res.Truncated = true
			break
		}
		item := queue[0]
		queue = queue[1:]
		u, _ := url.Parse(item.url)
		if !rules.allowed(u.RequestURI()) {
			res.Blocked++
			if first {
				return res, fmt.Errorf("%w: blocked by robots.txt", ErrForbiddenURL)
			}
			continue
		}
		if !first && opt.Delay > 0 {
			select {
			case <-time.After(opt.Delay):
			case <-ctx.Done():
				return res, ctx.Err()
			}
		}
		pageCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		resp, err := c.fetcher.Get(pageCtx, item.url, maxPageBytes)
		cancel()
		if err == nil && resp.Status >= 400 {
			err = fmt.Errorf("%w: status %d", ErrUnreachable, resp.Status)
		}
		if err != nil {
			if first {
				return res, err
			}
			res.Failed++
			continue
		}
		wasFirst := first
		first = false
		if wasFirst && siteKey(resp.FinalURL) != site {
			site = siteKey(resp.FinalURL) // the start URL moved to its canonical host
		}
		if siteKey(resp.FinalURL) != site {
			res.Failed++
			continue
		}
		if !strings.Contains(resp.ContentType, "html") && !strings.HasPrefix(resp.ContentType, "text/plain") {
			continue // images, archives and the like are not pages
		}
		page, links, ok := c.toPage(resp)
		if !ok {
			continue
		}
		res.Pages = append(res.Pages, page)
		if item.depth >= opt.MaxDepth {
			continue
		}
		for _, link := range links {
			lu, err := url.Parse(link)
			if err != nil || siteKey(lu) != site || isAsset(lu.Path) {
				continue
			}
			norm := normalize(lu)
			if !seen[norm] {
				seen[norm] = true
				queue = append(queue, queued{norm, item.depth + 1})
			}
		}
	}
	if len(res.Pages) == 0 {
		return res, fmt.Errorf("%w: no readable pages", ErrUnreachable)
	}
	return res, nil
}

func (c *Crawler) toPage(resp *Response) (Page, []string, bool) {
	if strings.HasPrefix(resp.ContentType, "text/plain") {
		text := tidy(string(bytes.ToValidUTF8(resp.Body, nil)))
		if text == "" {
			return Page{}, nil, false
		}
		return Page{URL: normalize(resp.FinalURL), Body: []byte(text), Ext: ".txt", Hash: hashOf([]byte(text))}, nil, true
	}
	ex, err := ExtractHTML(resp.Body, resp.FinalURL)
	if err != nil || ex.NoIndex || ex.Text == "" {
		return Page{}, nil, false
	}
	return Page{URL: normalize(resp.FinalURL), Title: ex.Title, Body: []byte(ex.Text), Ext: ".txt", Hash: hashOf([]byte(ex.Text))}, ex.Links, true
}

func (c *Crawler) loadRobots(ctx context.Context, start *url.URL) *robots {
	robotsURL := (&url.URL{Scheme: start.Scheme, Host: start.Host, Path: "/robots.txt"}).String()
	resp, err := c.fetcher.Get(ctx, robotsURL, maxRobotsSize)
	if err != nil || resp.Status >= 400 && resp.Status < 500 {
		return nil // no usable robots.txt: everything is allowed
	}
	if resp.Status >= 500 {
		return &robots{rules: []robotsRule{{allow: false, re: mustAll(), size: 1}}} // unreachable policy: stay away
	}
	return parseRobots(string(resp.Body), c.fetcher.userAgent)
}

func siteKey(u *url.URL) string { return strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.") }

// normalize drops fragments and a trailing slash so one page is not fetched twice under two spellings.
func normalize(u *url.URL) string {
	c := *u
	c.Fragment = ""
	c.Host = strings.ToLower(c.Host)
	if len(c.Path) > 1 {
		c.Path = strings.TrimSuffix(c.Path, "/")
	}
	return c.String()
}

var assetExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true, ".ico": true, ".pdf": true, ".zip": true, ".rar": true, ".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".css": true, ".js": true, ".xml": true}

func isAsset(p string) bool { return assetExt[strings.ToLower(path.Ext(p))] }
