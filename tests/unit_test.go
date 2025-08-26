// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

//go:build unit

package e2e

import (
	"testing"

	redir "github.com/Bl4cky99/caddy-redirector"
	"github.com/caddyserver/caddy/v2"
)

func Test_Exact_ToHost_JSON(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_exact.json")

	resp := s.RunOnce(r, &RequestSpec{
		Host: "exact.example",
		Path: "/old",
	}, nil)

	AssertRedirect(t, resp, 301, "https://success.example/new")
}

func Test_Exact_NoToHost_Inline(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorInline(308, []redir.HostBlock{
		{
			Pattern: "samehost.example",
			Exact:   map[string]string{"/from": "/to"},
		},
	})

	resp := s.RunOnce(r, &RequestSpec{
		Host: "samehost.example",
		Path: "/from",
	}, nil)

	AssertRedirect(t, resp, 308, "/to")
}

func Test_Exact_AbsoluteTarget_IgnoresToHost(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/absolute_exact.yaml")
	resp := s.RunOnce(r, &RequestSpec{Host: "abs.example", Path: "/foo"}, nil)
	AssertRedirect(t, resp, 308, "https://absolute.example/foo")
}

func Test_Prefix_Longest_YAML(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_prefix.yaml")

	resp := s.RunOnce(r, &RequestSpec{
		Host: "prefix.example",
		Path: "/a/long/thing",
	}, nil)

	AssertRedirect(t, resp, 308, "https://success.example/b/long/thing")
}

func Test_Prefix_Miss_PassesNext(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_prefix.yaml")

	resp := s.RunOnce(r, &RequestSpec{
		Host: "prefix.example",
		Path: "/nope",
	}, NextOK{})

	AssertPassedThrough(t, resp, 204)
}

func Test_Prefix_ToWithoutTrailingSlash_AddsSlashWhenNeeded(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorInline(308, []redir.HostBlock{
		{Pattern: "slash.example", ToHost: "success.example", Prefix: []redir.PrefixRule{{From: "/p/", To: "/to"}}},
	})
	resp := s.RunOnce(r, &RequestSpec{Host: "slash.example", Path: "/p/x"}, nil)
	AssertRedirect(t, resp, 308, "https://success.example/to/x")
}

func Test_Prefix_EqualLength_StableOrder(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorInline(308, []redir.HostBlock{
		{
			Pattern: "eq.example",
			ToHost:  "success.example",
			Prefix: []redir.PrefixRule{
				{From: "/aa/", To: "/X/"},
				{From: "/ab/", To: "/Y/"},
			},
		},
	})
	resp := s.RunOnce(r, &RequestSpec{Host: "eq.example", Path: "/ab/thing"}, nil)
	AssertRedirect(t, resp, 308, "https://success.example/Y/thing")
}

func Test_Regex_Capture_TOML(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_regex.toml")

	resp := s.RunOnce(r, &RequestSpec{
		Host: "regex.example",
		Path: "/u/12345",
	}, nil)

	AssertRedirect(t, resp, 308, "https://success.example/users/12345")
}

func Test_Regex_Absolute_IgnoresToHost(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_regex.toml")

	resp := s.RunOnce(r, &RequestSpec{
		Host: "regex.example",
		Path: "/abs/any/path",
	}, nil)

	AssertRedirect(t, resp, 308, "https://absolute.example/any/path")
}

func Test_Regex_Miss_PassesNext(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_regex.toml")

	resp := s.RunOnce(r, &RequestSpec{
		Host: "regex.example",
		Path: "/u/abc",
	}, NextOK{})

	AssertPassedThrough(t, resp, 204)
}

func Test_Regex_TargetMissingLeadingSlash_IsNormalized(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/regex_no_slash.json")

	resp := s.RunOnce(r, &RequestSpec{Host: "noslash.example", Path: "/u/42"}, nil)
	AssertRedirect(t, resp, 308, "https://success.example/users/42")
}

func Test_HostPrecedence_ExactOverWildcardOverCatchAll(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_wildcards.yaml")

	resp1 := s.RunOnce(r, &RequestSpec{
		Host: "special.prec.example",
		Path: "/e",
	}, nil)
	AssertRedirect(t, resp1, 308, "https://success.example/x")

	resp2 := s.RunOnce(r, &RequestSpec{
		Host: "foo.prec.example",
		Path: "/e",
	}, nil)
	AssertRedirect(t, resp2, 308, "https://wildcard.example/w")

	resp3 := s.RunOnce(r, &RequestSpec{
		Host: "other.com",
		Path: "/ping",
	}, nil)
	AssertRedirect(t, resp3, 308, "/health")
}

func Test_HostCaseInsensitivity(t *testing.T) {
	r := &redir.Redirector{
		DefaultCode: 308,
		RulesFiles:  []redir.RulesFile{{Path: ConfigPath(t, "configs/case.json")}},
	}
	if err := r.Provision(caddy.Context{}); err != nil {
		t.Fatalf("provision failed: %v", err)
	}
	resp := NewSuite(t).RunOnce(r, &RequestSpec{Host: "TeSt.ExAmPlE", Path: "/old"}, nil)
	AssertRedirect(t, resp, 308, "https://success.example/new")
}

func Test_Merge_Order_Files_LastWinsOnConflicts(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/merge_a.json", "configs/merge_b.json")

	resp := s.RunOnce(r, &RequestSpec{
		Host: "merge.example",
		Path: "/x",
	}, nil)
	AssertRedirect(t, resp, 301, "https://a.example/z")

	resp2 := s.RunOnce(r, &RequestSpec{
		Host: "merge.example",
		Path: "/a/123",
	}, nil)
	AssertRedirect(t, resp2, 301, "https://a.example/b/123")
}

func Test_StatusCodeResolution_HostOverridesGlobal(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorInline(308, []redir.HostBlock{
		{
			Pattern: "status.example",
			Status:  301,
			Exact:   map[string]string{"/old": "/new"},
		},
	})

	resp := s.RunOnce(r, &RequestSpec{
		Host: "status.example",
		Path: "/old",
	}, nil)
	AssertRedirect(t, resp, 301, "/new")
}

func Test_Scheme_Inference_XForwardedProto(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/scheme.yaml")

	resp1 := s.RunOnce(r, &RequestSpec{
		Host: "scheme.example",
		Path: "/h",
	}, nil)
	AssertRedirect(t, resp1, 308, "https://success.example/h")

	resp2 := s.RunOnce(r, &RequestSpec{
		Host:         "scheme.example",
		Path:         "/h",
		ForwardProto: "http",
	}, nil)
	AssertRedirect(t, resp2, 308, "http://success.example/h")
}

func Test_Scheme_TlsBeats_XForwardedProto(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorInline(308, []redir.HostBlock{
		{Pattern: "tls.example", ToHost: "success.example", Exact: map[string]string{"/h": "/h"}},
	})

	resp := s.RunOnce(r, &RequestSpec{Host: "tls.example", Path: "/h", ForwardProto: "http", UseTLS: true}, nil)
	AssertRedirect(t, resp, 308, "https://success.example/h")
}

func Test_Precedence_ExactOverPrefixOverRegex(t *testing.T) {
	s := NewSuite(t)
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
	AssertRedirect(t, resp, 308, "https://success.example/E")
}

func Test_Wildcard_DoesNotMatchApex(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorInline(308, []redir.HostBlock{
		{Pattern: "*.wild.example", ToHost: "wild.example", Exact: map[string]string{"/p": "/q"}},
		{Pattern: "wild.example", ToHost: "apex.example", Exact: map[string]string{"/p": "/q"}},
	})

	resp := s.RunOnce(r, &RequestSpec{Host: "wild.example", Path: "/p"}, nil)
	AssertRedirect(t, resp, 308, "https://apex.example/q")
}

func Test_Error_BadRegex_ProvisionFails(t *testing.T) {
	r := &redir.Redirector{
		DefaultCode: 308,
		RulesFiles:  []redir.RulesFile{{Path: ConfigPath(t, "configs/bad_regex.yaml")}},
	}
	if err := r.Provision(caddy.Context{}); err == nil {
		t.Fatalf("expected provision error for bad regex, got nil")
	}
}

func Test_Error_UnknownFormat_NoExt(t *testing.T) {
	r := &redir.Redirector{
		DefaultCode: 308,
		RulesFiles:  []redir.RulesFile{{Path: ConfigPath(t, "configs/unknown.data")}},
	}
	if err := r.Provision(caddy.Context{}); err == nil {
		t.Fatalf("expected provision error for unknown format")
	}
}

func Test_Error_MissingFile(t *testing.T) {
	r := &redir.Redirector{
		DefaultCode: 308,
		RulesFiles:  []redir.RulesFile{{Path: ConfigPath(t, "configs/does_not_exist.json")}},
	}
	if err := r.Provision(caddy.Context{}); err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func Test_UnmatchedHost_PassesToNext(t *testing.T) {
	s := NewSuite(t)
	r := s.BuildRedirectorFromFiles(308, "configs/rules_exact.json")
	resp := s.RunOnce(r, &RequestSpec{Host: "no-match.example", Path: "/anything"}, NextOK{})
	AssertPassedThrough(t, resp, 204)
}

func Test_ExplicitFormat_WithNoExt_Succeeds(t *testing.T) {
	r := &redir.Redirector{
		DefaultCode: 308,
		RulesFiles: []redir.RulesFile{
			{Path: ConfigPath(t, "configs/noext"), Format: "json"},
		},
	}
	if err := r.Provision(caddy.Context{}); err != nil {
		t.Fatalf("provision failed: %v", err)
	}
	resp := NewSuite(t).RunOnce(r, &RequestSpec{Host: "noext.example", Path: "/old"}, nil)
	AssertRedirect(t, resp, 308, "https://success.example/new")
}
