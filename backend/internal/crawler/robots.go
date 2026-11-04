package crawler

import (
	"regexp"
	"strings"
)

type robotsRule struct {
	allow bool
	re    *regexp.Regexp
	size  int
}

// robots is the subset of robots.txt the crawler honours: Allow/Disallow with * and $ for our user agent,
// falling back to the "*" group. The longest matching rule wins and Allow wins a tie.
type robots struct{ rules []robotsRule }

type robotsGroup struct {
	agents []string
	rules  []robotsRule
}

func parseRobots(body, agent string) *robots {
	agent = strings.ToLower(agent)
	var groups []*robotsGroup
	var current *robotsGroup
	collecting := false
	for _, line := range strings.Split(body, "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key, value = strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value)
		switch key {
		case "user-agent":
			if !collecting || current == nil {
				current = &robotsGroup{}
				groups = append(groups, current)
			}
			current.agents = append(current.agents, strings.ToLower(value))
			collecting = true
		case "allow", "disallow":
			collecting = false
			if current == nil || value == "" {
				continue // an empty Disallow means "allow everything"
			}
			pattern := "^" + strings.ReplaceAll(strings.ReplaceAll(regexp.QuoteMeta(value), `\*`, ".*"), `\$`, "$")
			if re, err := regexp.Compile(pattern); err == nil {
				current.rules = append(current.rules, robotsRule{allow: key == "allow", re: re, size: len(value)})
			}
		}
	}
	pick := func(match func(string) bool) []robotsRule {
		var out []robotsRule
		for _, g := range groups {
			for _, a := range g.agents {
				if match(a) {
					out = append(out, g.rules...)
					break
				}
			}
		}
		return out
	}
	rules := pick(func(a string) bool { return a != "*" && a != "" && strings.Contains(agent, a) })
	if len(rules) == 0 {
		rules = pick(func(a string) bool { return a == "*" })
	}
	return &robots{rules: rules}
}

func mustAll() *regexp.Regexp { return regexp.MustCompile(`^/`) }

func (r *robots) allowed(path string) bool {
	if r == nil {
		return true
	}
	if path == "" {
		path = "/"
	}
	best, allow := -1, true
	for _, rule := range r.rules {
		if rule.re.MatchString(path) && (rule.size > best || (rule.size == best && rule.allow)) {
			best, allow = rule.size, rule.allow
		}
	}
	return allow
}
