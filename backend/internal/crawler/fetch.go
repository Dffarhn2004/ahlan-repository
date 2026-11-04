package crawler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxRedirects   = 5
	requestTimeout = 20 * time.Second
)

// Response is one fetched resource.
type Response struct {
	Status      int
	FinalURL    *url.URL
	ContentType string
	Disposition string
	Body        []byte
}

// Fetcher performs HTTP GETs with the crawler's network restrictions.
type Fetcher struct {
	client    *http.Client
	userAgent string
}

// FetcherConfig tunes a Fetcher. allowInternal exists only so tests can reach httptest servers on 127.0.0.1.
type FetcherConfig struct {
	UserAgent     string
	allowInternal bool
}

func NewFetcher(cfg FetcherConfig) *Fetcher {
	if cfg.UserAgent == "" {
		cfg.UserAgent = "PonpesAICrawler/1.0"
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	if !cfg.allowInternal {
		dialer.Control = dialControl
	}
	transport := &http.Transport{
		Proxy:                 nil, // a proxy would bypass the dial-time address check
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		MaxIdleConns:          10,
	}
	f := &Fetcher{userAgent: cfg.UserAgent}
	f.client = &http.Client{Transport: transport, Timeout: requestTimeout, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("%w: too many redirects", ErrForbiddenURL)
		}
		if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
			return ErrForbiddenURL
		}
		// A Google document may only redirect inside Google; anything else is a hijack attempt.
		if isGoogleHost(via[0].URL.Hostname()) && !isGoogleHost(req.URL.Hostname()) {
			return fmt.Errorf("%w: redirect leaves Google", ErrForbiddenURL)
		}
		return nil
	}}
	return f
}

func isGoogleHost(host string) bool {
	host = strings.ToLower(host)
	for _, suffix := range []string{"google.com", "googleusercontent.com"} {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}

// Get fetches rawURL and reads at most maxBytes of the body (ErrTooLarge beyond that).
func (f *Fetcher) Get(ctx context.Context, rawURL string, maxBytes int64) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, ErrForbiddenURL
	}
	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept", "text/html,text/plain,text/csv,application/pdf;q=0.9,*/*;q=0.1")
	resp, err := f.client.Do(req)
	if err != nil {
		if errors.Is(err, ErrForbiddenURL) {
			return nil, ErrForbiddenURL
		}
		return nil, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	if int64(len(body)) > maxBytes {
		return nil, ErrTooLarge
	}
	return &Response{Status: resp.StatusCode, FinalURL: resp.Request.URL, ContentType: strings.ToLower(resp.Header.Get("Content-Type")), Disposition: resp.Header.Get("Content-Disposition"), Body: body}, nil
}
