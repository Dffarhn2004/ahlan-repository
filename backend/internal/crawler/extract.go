package crawler

import (
	"bytes"
	"net/url"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

const maxTextBytes = 200 << 10

// Extracted is the readable content of one HTML page.
type Extracted struct {
	Title   string
	Text    string
	Links   []string
	NoIndex bool
}

var skipTags = map[string]bool{"script": true, "style": true, "noscript": true, "template": true, "svg": true, "iframe": true, "head": true, "nav": true, "canvas": true, "object": true, "embed": true}

var blockTags = map[string]bool{"p": true, "div": true, "br": true, "li": true, "ul": true, "ol": true, "tr": true, "table": true, "section": true, "article": true, "header": true, "footer": true, "main": true, "aside": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true, "blockquote": true, "pre": true, "form": true, "dd": true, "dt": true}

// ExtractHTML pulls the visible text, title and links out of an HTML document. Scripts, styles, frames, comments and
// navigation menus are dropped, so hidden instructions in markup never reach the assistant as text. Relative links
// are resolved against base.
func ExtractHTML(body []byte, base *url.URL) (Extracted, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return Extracted{}, err
	}
	var out Extracted
	var text strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if out.Title == "" {
					out.Title = collapse(nodeText(n))
				}
				return
			case "meta":
				if strings.EqualFold(attr(n, "name"), "robots") && strings.Contains(strings.ToLower(attr(n, "content")), "noindex") {
					out.NoIndex = true
				}
			case "a":
				if link := resolve(base, attr(n, "href")); link != "" {
					out.Links = append(out.Links, link)
				}
			}
			if skipTags[n.Data] {
				// <head> holds the title and robots meta, which were handled above through a separate pass.
				if n.Data == "head" {
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						walk(c)
					}
				}
				// Menus are noise as text but are the best way to discover the rest of the site.
				if n.Data == "nav" {
					collectLinks(n, base, &out.Links)
				}
				return
			}
			if blockTags[n.Data] {
				text.WriteByte('\n')
			}
		}
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
			text.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
		if n.Type == html.ElementNode && blockTags[n.Data] {
			text.WriteByte('\n')
		}
	}
	walk(doc)
	out.Text = tidy(text.String())
	return out, nil
}

func collectLinks(n *html.Node, base *url.URL, links *[]string) {
	if n.Type == html.ElementNode && n.Data == "a" {
		if link := resolve(base, attr(n, "href")); link != "" {
			*links = append(*links, link)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectLinks(c, base, links)
	}
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func nodeText(n *html.Node) string {
	var b strings.Builder
	var rec func(*html.Node)
	rec = func(x *html.Node) {
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			rec(c)
		}
	}
	rec(n)
	return b.String()
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

// tidy collapses runs of spaces within lines and blank lines between them, and caps the size on a rune boundary.
func tidy(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = collapse(line)
		if line == "" {
			if !blank && len(out) > 0 {
				out = append(out, "")
			}
			blank = true
			continue
		}
		blank = false
		out = append(out, line)
	}
	text := strings.TrimSpace(strings.Join(out, "\n"))
	if len(text) > maxTextBytes {
		text = text[:maxTextBytes]
		for !utf8.ValidString(text) {
			text = text[:len(text)-1]
		}
	}
	return text
}

func resolve(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	abs.Fragment = ""
	return abs.String()
}
