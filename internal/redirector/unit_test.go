// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package redirector_test

import (
	redir "github.com/Bl4cky99/caddy-redirector"
	"github.com/caddyserver/caddy/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Redirector", func() {
	var s *Suite

	BeforeEach(func() {
		s = NewSuite()
	})

	Describe("Exact rules", func() {
		It("redirects with JSON config and ToHost", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_exact.json")
			resp := s.RunOnce(r, &RequestSpec{Host: "exact.example", Path: "/old"}, nil)
			AssertRedirect(resp, 301, "https://success.example/new")
		})

		It("redirects without ToHost using same host", func() {
			r := s.BuildRedirectorInline(308, []redir.HostBlock{
				{Pattern: "samehost.example", Exact: map[string]string{"/from": "/to"}},
			})
			resp := s.RunOnce(r, &RequestSpec{Host: "samehost.example", Path: "/from"}, nil)
			AssertRedirect(resp, 308, "/to")
		})

		It("ignores ToHost when target is an absolute URL", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/absolute_exact.yaml")
			resp := s.RunOnce(r, &RequestSpec{Host: "abs.example", Path: "/foo"}, nil)
			AssertRedirect(resp, 308, "https://absolute.example/foo")
		})
	})

	Describe("Prefix rules", func() {
		It("matches the longest prefix from YAML config", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_prefix.yaml")
			resp := s.RunOnce(r, &RequestSpec{Host: "prefix.example", Path: "/a/long/thing"}, nil)
			AssertRedirect(resp, 308, "https://success.example/b/long/thing")
		})

		It("passes to next handler on miss", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_prefix.yaml")
			resp := s.RunOnce(r, &RequestSpec{Host: "prefix.example", Path: "/nope"}, NextOK{})
			AssertPassedThrough(resp, 204)
		})

		It("adds slash between To and remaining path when To has no trailing slash", func() {
			r := s.BuildRedirectorInline(308, []redir.HostBlock{
				{Pattern: "slash.example", ToHost: "success.example", Prefix: []redir.PrefixRule{{From: "/p/", To: "/to"}}},
			})
			resp := s.RunOnce(r, &RequestSpec{Host: "slash.example", Path: "/p/x"}, nil)
			AssertRedirect(resp, 308, "https://success.example/to/x")
		})

		It("uses stable order for equal-length prefixes", func() {
			r := s.BuildRedirectorInline(308, []redir.HostBlock{
				{
					Pattern: "eq.example",
					ToHost:  "success.example",
					Prefix:  []redir.PrefixRule{{From: "/aa/", To: "/X/"}, {From: "/ab/", To: "/Y/"}},
				},
			})
			resp := s.RunOnce(r, &RequestSpec{Host: "eq.example", Path: "/ab/thing"}, nil)
			AssertRedirect(resp, 308, "https://success.example/Y/thing")
		})
	})

	Describe("Regex rules", func() {
		It("captures groups from TOML config", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_regex.toml")
			resp := s.RunOnce(r, &RequestSpec{Host: "regex.example", Path: "/u/12345"}, nil)
			AssertRedirect(resp, 308, "https://success.example/users/12345")
		})

		It("ignores ToHost when target is an absolute URL", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_regex.toml")
			resp := s.RunOnce(r, &RequestSpec{Host: "regex.example", Path: "/abs/any/path"}, nil)
			AssertRedirect(resp, 308, "https://absolute.example/any/path")
		})

		It("passes to next handler on miss", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_regex.toml")
			resp := s.RunOnce(r, &RequestSpec{Host: "regex.example", Path: "/u/abc"}, NextOK{})
			AssertPassedThrough(resp, 204)
		})

		It("normalizes target with missing leading slash", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/regex_no_slash.json")
			resp := s.RunOnce(r, &RequestSpec{Host: "noslash.example", Path: "/u/42"}, nil)
			AssertRedirect(resp, 308, "https://success.example/users/42")
		})
	})

	Describe("Host matching", func() {
		It("prefers exact host over wildcard over catch-all", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_wildcards.yaml")

			resp1 := s.RunOnce(r, &RequestSpec{Host: "special.prec.example", Path: "/e"}, nil)
			AssertRedirect(resp1, 308, "https://success.example/x")

			resp2 := s.RunOnce(r, &RequestSpec{Host: "foo.prec.example", Path: "/e"}, nil)
			AssertRedirect(resp2, 308, "https://wildcard.example/w")

			resp3 := s.RunOnce(r, &RequestSpec{Host: "other.com", Path: "/ping"}, nil)
			AssertRedirect(resp3, 308, "/health")
		})

		It("matches host case-insensitively", func() {
			r := &redir.Redirector{
				DefaultCode: 308,
				RulesFiles:  []redir.RulesFile{{Path: ConfigPath("configs/case.json")}},
			}
			Expect(r.Provision(caddy.Context{})).To(Succeed())
			resp := NewSuite().RunOnce(r, &RequestSpec{Host: "TeSt.ExAmPlE", Path: "/old"}, nil)
			AssertRedirect(resp, 308, "https://success.example/new")
		})

		It("does not match wildcard against apex domain", func() {
			r := s.BuildRedirectorInline(308, []redir.HostBlock{
				{Pattern: "*.wild.example", ToHost: "wild.example", Exact: map[string]string{"/p": "/q"}},
				{Pattern: "wild.example", ToHost: "apex.example", Exact: map[string]string{"/p": "/q"}},
			})
			resp := s.RunOnce(r, &RequestSpec{Host: "wild.example", Path: "/p"}, nil)
			AssertRedirect(resp, 308, "https://apex.example/q")
		})

		It("passes to next handler when no host matches", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/rules_exact.json")
			resp := s.RunOnce(r, &RequestSpec{Host: "no-match.example", Path: "/anything"}, NextOK{})
			AssertPassedThrough(resp, 204)
		})
	})

	Describe("Rule file merging", func() {
		It("uses last-wins strategy on conflicting keys", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/merge_a.json", "configs/merge_b.json")

			resp := s.RunOnce(r, &RequestSpec{Host: "merge.example", Path: "/x"}, nil)
			AssertRedirect(resp, 301, "https://a.example/z")

			resp2 := s.RunOnce(r, &RequestSpec{Host: "merge.example", Path: "/a/123"}, nil)
			AssertRedirect(resp2, 301, "https://a.example/b/123")
		})
	})

	Describe("Status code resolution", func() {
		It("host-level status overrides the global default", func() {
			r := s.BuildRedirectorInline(308, []redir.HostBlock{
				{Pattern: "status.example", Status: 301, Exact: map[string]string{"/old": "/new"}},
			})
			resp := s.RunOnce(r, &RequestSpec{Host: "status.example", Path: "/old"}, nil)
			AssertRedirect(resp, 301, "/new")
		})
	})

	Describe("Scheme inference", func() {
		It("defaults to https and respects X-Forwarded-Proto header", func() {
			r := s.BuildRedirectorFromFiles(308, "configs/scheme.yaml")

			resp1 := s.RunOnce(r, &RequestSpec{Host: "scheme.example", Path: "/h"}, nil)
			AssertRedirect(resp1, 308, "https://success.example/h")

			resp2 := s.RunOnce(r, &RequestSpec{Host: "scheme.example", Path: "/h", ForwardProto: "http"}, nil)
			AssertRedirect(resp2, 308, "http://success.example/h")
		})

		It("prefers TLS connection state over X-Forwarded-Proto", func() {
			r := s.BuildRedirectorInline(308, []redir.HostBlock{
				{Pattern: "tls.example", ToHost: "success.example", Exact: map[string]string{"/h": "/h"}},
			})
			resp := s.RunOnce(r, &RequestSpec{Host: "tls.example", Path: "/h", ForwardProto: "http", UseTLS: true}, nil)
			AssertRedirect(resp, 308, "https://success.example/h")
		})
	})

	Describe("Rule type precedence", func() {
		It("exact takes priority over prefix, which takes priority over regex", func() {
			r := s.BuildRedirectorInline(308, []redir.HostBlock{
				{
					Pattern: "prec.example",
					ToHost:  "success.example",
					Exact:   map[string]string{"/a": "/E"},
					Prefix:  []redir.PrefixRule{{From: "/a", To: "/P"}},
					Regex:   []redir.RegexRule{{Pattern: "^/a$", To: "/R"}},
				},
			})
			resp := s.RunOnce(r, &RequestSpec{Host: "prec.example", Path: "/a"}, nil)
			AssertRedirect(resp, 308, "https://success.example/E")
		})
	})

	Describe("Error handling", func() {
		It("fails provision on bad regex", func() {
			r := &redir.Redirector{
				DefaultCode: 308,
				RulesFiles:  []redir.RulesFile{{Path: ConfigPath("configs/bad_regex.yaml")}},
			}
			Expect(r.Provision(caddy.Context{})).To(HaveOccurred())
		})

		It("fails provision on unknown file format", func() {
			r := &redir.Redirector{
				DefaultCode: 308,
				RulesFiles:  []redir.RulesFile{{Path: ConfigPath("configs/unknown.data")}},
			}
			Expect(r.Provision(caddy.Context{})).To(HaveOccurred())
		})

		It("fails provision on missing file", func() {
			r := &redir.Redirector{
				DefaultCode: 308,
				RulesFiles:  []redir.RulesFile{{Path: ConfigPath("configs/does_not_exist.json")}},
			}
			Expect(r.Provision(caddy.Context{})).To(HaveOccurred())
		})
	})

	Describe("Explicit format override", func() {
		It("succeeds for extension-less files with an explicit format", func() {
			r := &redir.Redirector{
				DefaultCode: 308,
				RulesFiles:  []redir.RulesFile{{Path: ConfigPath("configs/noext"), Format: "json"}},
			}
			Expect(r.Provision(caddy.Context{})).To(Succeed())
			resp := NewSuite().RunOnce(r, &RequestSpec{Host: "noext.example", Path: "/old"}, nil)
			AssertRedirect(resp, 308, "https://success.example/new")
		})
	})
})
