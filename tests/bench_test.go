// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

//go:build bench

package e2e

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	redir "github.com/Bl4cky99/caddy-redirector"
	"github.com/caddyserver/caddy/v2"
)

type benchNext struct{}

func (benchNext) ServeHTTP(http.ResponseWriter, *http.Request) error { return nil }

func buildExactHost(n int, withToHost bool, code int) *redir.Redirector {
	hb := redir.HostBlock{
		Pattern: "bench.example",
		Exact:   make(map[string]string, n),
		Status:  code,
	}
	if withToHost {
		hb.ToHost = "success.example"
	}
	for i := 0; i < n; i++ {
		hb.Exact["/e"+strconv.Itoa(i)] = "/t" + strconv.Itoa(i)
	}
	r := &redir.Redirector{DefaultCode: code, Hosts: []redir.HostBlock{hb}}
	_ = r.Provision(caddy.Context{})
	return r
}

func buildPrefixHost(n int, withToHost bool, code int) *redir.Redirector {
	hb := redir.HostBlock{
		Pattern: "bench.example",
		Status:  code,
	}
	if withToHost {
		hb.ToHost = "success.example"
	}

	for i := 0; i < n; i++ {
		hb.Prefix = append(hb.Prefix, redir.PrefixRule{
			From: "/p/" + strings.Repeat("x", i) + "/",
			To:   "/q/" + strings.Repeat("y", i) + "/",
		})
	}
	r := &redir.Redirector{DefaultCode: code, Hosts: []redir.HostBlock{hb}}
	_ = r.Provision(caddy.Context{})
	return r
}

func buildRegexHost(n int, withToHost bool, code int) *redir.Redirector {
	hb := redir.HostBlock{
		Pattern: "bench.example",
		Status:  code,
	}
	if withToHost {
		hb.ToHost = "success.example"
	}
	for i := 0; i < n; i++ {
		hb.Regex = append(hb.Regex, redir.RegexRule{
			Pattern: "^/u/([0-9]{" + strconv.Itoa(2+(i%3)) + "})/post$",
			To:      "/users/$1/article",
		})
	}
	r := &redir.Redirector{DefaultCode: code, Hosts: []redir.HostBlock{hb}}
	_ = r.Provision(caddy.Context{})
	return r
}

func BenchmarkExact_Hit_1e3(b *testing.B) {
	r := buildExactHost(1000, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/e500", nil)
	req.Host = "bench.example"
	next := benchNext{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(httptest.NewRecorder(), req, next)
	}
}

func BenchmarkExact_Miss_1e3(b *testing.B) {
	r := buildExactHost(1000, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/notfound", nil)
	req.Host = "bench.example"
	next := benchNext{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(httptest.NewRecorder(), req, next)
	}
}

func BenchmarkPrefix_Longest_Hit_1e3(b *testing.B) {
	r := buildPrefixHost(1000, true, 308)
	path := "/p/" + strings.Repeat("x", 999) + "/rest"
	req := httptest.NewRequest("GET", "http://bench.example"+path, nil)
	req.Host = "bench.example"
	next := benchNext{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(httptest.NewRecorder(), req, next)
	}
}

func BenchmarkPrefix_Miss_1e3(b *testing.B) {
	r := buildPrefixHost(1000, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/none/at/all", nil)
	req.Host = "bench.example"
	next := benchNext{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(httptest.NewRecorder(), req, next)
	}
}

func BenchmarkRegex_Hit_1e2(b *testing.B) {
	r := buildRegexHost(100, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/u/123/post", nil)
	req.Host = "bench.example"
	next := benchNext{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(httptest.NewRecorder(), req, next)
	}
}

func BenchmarkRegex_Miss_1e2(b *testing.B) {
	r := buildRegexHost(100, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/u/abc/post", nil)
	req.Host = "bench.example"
	next := benchNext{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(httptest.NewRecorder(), req, next)
	}
}
