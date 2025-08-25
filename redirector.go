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

	compiled = make([]compiledHostBlock, 0, len(r.Hosts))
	for _, hb := range r.Hosts {
		ch := compiledHostBlock{
			toHost:      hb.ToHost,
			exactPaths:  hb.Exact,
			status:      hb.Status,
			prefixRules: hb.Prefix,
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

		if len(ch.prefixRules) > 1 {
			sort.SliceStable(ch.prefixRules, func(i, j int) bool {
				return len(ch.prefixRules[i].From) > len(ch.prefixRules[j].To)
			})
		}

		compiled = append(compiled, ch)
	}

	return nil
}

func (r *Redirector) Validate() error { return nil }

func (r *Redirector) ServeHTTP(w http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	host := strings.ToLower(req.Host)
	path := req.URL.Path

	var block *compiledHostBlock
	for i := range compiled {
		ch := &compiled[i]
		switch {
		case ch.exactHost != "" && host == ch.exactHost:
			block = ch
			goto FOUND
		case ch.suffix != "" && strings.HasSuffix(host, ch.suffix):
			block = ch
			goto FOUND
		case ch.matchAll:
			if block == nil {
				block = ch
			}
		}
	}

FOUND:
	if block != nil {
		if to, ok := block.exactPaths[path]; ok {
			target := to
			if block.toHost != "" {
				scheme := "https"
				if req.TLS == nil && req.Header.Get("X-Forwarded-Proto") == "http" {
					scheme = "http"
				}

				if !strings.HasPrefix(to, "http://") && !strings.HasPrefix(to, "https://") {
					if !strings.HasPrefix(to, "/") {
						to = "/" + to
					}
					target = scheme + "://" + block.toHost + to
				} else {
					target = to
				}
			}
			http.Redirect(w, req, target, block.status)
			return nil
		}

		if len(block.prefixRules) > 0 {
			for _, pr := range block.prefixRules {
				if strings.HasPrefix(path, pr.From) {
					rest := strings.TrimPrefix(path, pr.From)
					to := pr.To
					if !strings.HasSuffix(to, "/") && rest != "" && !strings.HasPrefix(rest, "/") {
						to += "/"
					}
					newPath := to + rest

					target := newPath
					if block.toHost != "" && !strings.HasPrefix(newPath, "http://") && !strings.HasPrefix(newPath, "https://") {
						scheme := "https"
						if req.TLS == nil && req.Header.Get("X-Forwarded-Proto") == "http" {
							scheme = "http"
						}
						target = scheme + "://" + block.toHost + newPath
					}

					http.Redirect(w, req, target, block.status)
					return nil
				}
			}
		}

		if len(block.regexRules) > 0 {
			for _, rr := range block.regexRules {
				if rr.re.MatchString(path) {
					out := rr.re.ReplaceAllString(path, rr.to)
					target := out

					if !strings.HasPrefix(out, "http://") && !strings.HasPrefix(out, "https://") {
						if block.toHost != "" {
							scheme := "https"
							if req.TLS == nil && req.Header.Get("X-Forwarded-Proto") == "http" {
								scheme = "http"
							}

							if !strings.HasPrefix(out, "/") {
								out = "/" + out
							}

							target = scheme + "://" + block.toHost + out
						} else {
							target = out
						}
					}

					http.Redirect(w, req, target, block.status)
					return nil
				}
			}
		}
	}

	return next.ServeHTTP(w, req)
}
