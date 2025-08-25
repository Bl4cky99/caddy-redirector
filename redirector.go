// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package redirector

import (
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

var compiled []compiledHostBlock

func init() {
	caddy.RegisterModule(Redirector{})
	httpcaddyfile.RegisterHandlerDirective("redirector", parseCaddyFile)
}

func (Redirector) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.redirector",
		New: func() caddy.Module { return new(Redirector) },
	}
}

func (r *Redirector) Provision(caddy.Context) error {
	if r.DefaultCode == 0 {
		r.DefaultCode = http.StatusPermanentRedirect
	}

	if len(r.RulesFiles) > 0 {
		if err := r.loadExternalRules(); err != nil {
			return err
		}
	}

	compiled = make([]compiledHostBlock, 0, len(r.Hosts))
	for _, hb := range r.Hosts {
		ch := compiledHostBlock{
			toHost:     hb.ToHost,
			exactPaths: hb.Exact,
			status:     hb.Status,
		}

		p := strings.ToLower(strings.TrimSpace(hb.Pattern))
		switch {
		case p == "*":
			ch.matchAll = true
		case strings.HasPrefix(p, "*.") && len(p) > 2:
			ch.suffix = p[1:]
		default:
			ch.exactHost = p
		}

		if ch.status == 0 {
			ch.status = r.DefaultCode
		}

		for _, rr := range hb.Regex {
			re, err := regexp.Compile(rr.Pattern)
			if err != nil {
				return err
			}

			ch.regexRules = append(ch.regexRules, compiledRegexRule{re: re, to: rr.To})
		}

		if len(hb.Prefix) > 0 {
			ch.prefixBuckets = make(map[string][]PrefixRule, len(hb.Prefix))
			for _, pr := range hb.Prefix {
				k := bucketKey(pr.From)
				ch.prefixBuckets[k] = append(ch.prefixBuckets[k], pr)
			}

			for k := range ch.prefixBuckets {
				sort.SliceStable(ch.prefixBuckets[k], func(i, j int) bool {
					return len(ch.prefixBuckets[k][i].From) > len(ch.prefixBuckets[k][j].From)
				})
			}
		}

		compiled = append(compiled, ch)
	}

	return nil
}

func (r *Redirector) Validate() error { return nil }

func (r *Redirector) ServeHTTP(w http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	host := strings.ToLower(req.Host)
	path := req.URL.Path

	block := findHostBlock(host)
	if block != nil {
		if to, ok := block.exactPaths[path]; ok {
			return doRedirect(w, req, buildTarget(block.toHost, to, req), block.status)
		}

		if target, ok := matchPrefix(block, path, req); ok {
			return doRedirect(w, req, target, block.status)
		}

		if target, ok := matchRegex(block, path, req); ok {
			return doRedirect(w, req, target, block.status)
		}
	}

	return next.ServeHTTP(w, req)
}

func findHostBlock(host string) *compiledHostBlock {
	var fallback *compiledHostBlock
	for i := range compiled {
		ch := &compiled[i]
		switch {
		case ch.exactHost != "" && host == ch.exactHost:
			return ch
		case ch.suffix != "" && strings.HasSuffix(host, ch.suffix):
			return ch
		case ch.matchAll:
			if fallback == nil {
				fallback = ch
			}
		}
	}
	return fallback
}

func matchPrefix(block *compiledHostBlock, path string, req *http.Request) (string, bool) {
	if block.prefixBuckets == nil {
		return "", false
	}
	lst := block.prefixBuckets[bucketKey(path)]
	if len(lst) == 0 {
		return "", false
	}
	for _, pr := range lst {
		if strings.HasPrefix(path, pr.From) {
			rest := path[len(pr.From):]
			to := pr.To
			if !strings.HasSuffix(to, "/") && rest != "" && !strings.HasPrefix(rest, "/") {
				to += "/"
			}
			newPath := to + rest
			return buildTarget(block.toHost, newPath, req), true
		}
	}
	return "", false
}

func matchRegex(block *compiledHostBlock, path string, req *http.Request) (string, bool) {
	if len(block.regexRules) == 0 {
		return "", false
	}
	for _, rr := range block.regexRules {
		if rr.re.MatchString(path) {
			out := rr.re.ReplaceAllString(path, rr.to)
			return buildTarget(block.toHost, out, req), true
		}
	}
	return "", false
}

func doRedirect(w http.ResponseWriter, req *http.Request, target string, status int) error {
	http.Redirect(w, req, target, status)
	return nil
}

func buildTarget(toHost string, candidate string, req *http.Request) string {
	if isAbsoluteURL(candidate) {
		return candidate
	}

	if toHost == "" {
		return candidate
	}
	p := candidate
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return schemeFromRequest(req) + "://" + toHost + p
}

func isAbsoluteURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func schemeFromRequest(req *http.Request) string {
	if req.TLS == nil && strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "http") {
		return "http"
	}
	return "https"
}

func bucketKey(s string) string {
	if len(s) < 2 || s[0] != '/' {
		return ""
	}
	s = s[1:]
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 8 {
		s = s[:8]
	}
	return s
}
