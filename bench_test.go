package redirector

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2"
)

type nullRW struct {
	h http.Header
}

func (w *nullRW) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}
func (w *nullRW) Write(p []byte) (int, error) { return len(p), nil }
func (w *nullRW) WriteHeader(statusCode int)  {}

type nopNext struct{}

func (nopNext) ServeHTTP(http.ResponseWriter, *http.Request) error { return nil }

func buildExactHost(n int, withToHost bool, code int) *Redirector {
	hb := HostBlock{
		Pattern: "bench.example",
		Exact:   make(map[string]string, n),
		Status:  code,
	}
	if withToHost {
		hb.ToHost = "target.example"
	}
	for i := 0; i < n; i++ {
		from := "/e" + strconv.Itoa(i)
		to := "/t" + strconv.Itoa(i)
		hb.Exact[from] = to
	}
	r := &Redirector{
		DefaultCode: code,
		Hosts:       []HostBlock{hb},
	}
	_ = r.Provision(caddy.Context{})
	return r
}

func buildPrefixHost(n int, withToHost bool, code int) *Redirector {
	hb := HostBlock{
		Pattern: "bench.example",
		Status:  code,
	}
	if withToHost {
		hb.ToHost = "target.example"
	}

	for i := 0; i < n; i++ {
		hb.Prefix = append(hb.Prefix, PrefixRule{
			From: "/p/" + strings.Repeat("x", i) + "/",
			To:   "/q/" + strings.Repeat("y", i) + "/",
		})
	}
	r := &Redirector{
		DefaultCode: code,
		Hosts:       []HostBlock{hb},
	}
	_ = r.Provision(caddy.Context{})
	return r
}

func buildRegexHost(n int, withToHost bool, code int) *Redirector {
	hb := HostBlock{
		Pattern: "bench.example",
		Status:  code,
	}
	if withToHost {
		hb.ToHost = "target.example"
	}

	for i := 0; i < n; i++ {
		hb.Regex = append(hb.Regex, RegexRule{
			Pattern: "^/u/([0-9]{" + strconv.Itoa(2+i%3) + "})/post$",
			To:      "/users/$1/article",
		})
	}
	r := &Redirector{
		DefaultCode: code,
		Hosts:       []HostBlock{hb},
	}
	_ = r.Provision(caddy.Context{})
	return r
}

func BenchmarkExact_Hit_1e3(b *testing.B) {
	r := buildExactHost(1000, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/e500", nil)
	w := &nullRW{}
	next := nopNext{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(w, req, next)
	}
}

func BenchmarkExact_Miss_1e3(b *testing.B) {
	r := buildExactHost(1000, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/notfound", nil)
	w := &nullRW{}
	next := nopNext{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(w, req, next)
	}
}

func BenchmarkPrefix_Longest_Hit_1e3(b *testing.B) {
	r := buildPrefixHost(1000, true, 308)

	req := httptest.NewRequest("GET", "http://bench.example/p/"+strings.Repeat("x", 999)+"/rest", nil)
	w := &nullRW{}
	next := nopNext{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(w, req, next)
	}
}

func BenchmarkPrefix_Miss_1e3(b *testing.B) {
	r := buildPrefixHost(1000, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/none/at/all", nil)
	w := &nullRW{}
	next := nopNext{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(w, req, next)
	}
}

func BenchmarkRegex_Hit_1e2(b *testing.B) {
	r := buildRegexHost(100, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/u/123/post", nil)
	w := &nullRW{}
	next := nopNext{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(w, req, next)
	}
}

func BenchmarkRegex_Miss_1e2(b *testing.B) {
	r := buildRegexHost(100, true, 308)
	req := httptest.NewRequest("GET", "http://bench.example/u/abc/post", nil)
	w := &nullRW{}
	next := nopNext{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ServeHTTP(w, req, next)
	}
}
