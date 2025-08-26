// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package e2e

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func ProjectRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determin caller path")
	}

	testDir := filepath.Dir(file)
	return filepath.Clean(filepath.Join(testDir, ".."))
}

func ConfigPath(t *testing.T, rel string) string {
	t.Helper()
	return filepath.Join(ProjectRoot(t), "tests", rel)
}

type RequestSpec struct {
	Method       string
	Host         string
	Path         string
	ForwardProto string
	UseTLS       bool
	Header       http.Header
}

func (r *RequestSpec) Build() *http.Request {
	if r.Method == "" {
		r.Method = http.MethodGet
	}
	url := "http://" + r.Host + r.Path
	req := httptest.NewRequest(r.Method, url, nil)
	req.Host = r.Host
	if r.Header != nil {
		for k, vv := range r.Header {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}

	if r.ForwardProto != "" {
		req.Header.Set("X-Forwarded-Proto", r.ForwardProto)
	}

	if r.UseTLS {
		req.TLS = &tls.ConnectionState{}
	}

	if strings.EqualFold(r.ForwardProto, "http") {
		req.Header.Set("X-Forwarded-Proto", "http")
	}
	return req
}

type recorder struct {
	h      http.Header
	code   int
	wroteH bool
	body   []byte
}

func newRecorder() *recorder { return &recorder{h: make(http.Header)} }

func (r *recorder) Header() http.Header { return r.h }
func (r *recorder) Write(p []byte) (int, error) {
	if !r.wroteH {
		r.WriteHeader(http.StatusOK)
	}
	r.body = append(r.body, p...)
	return len(p), nil
}
func (r *recorder) WriteHeader(statusCode int) { r.wroteH, r.code = true, statusCode }

type Response struct{ Recorder *recorder }

func (r *Response) Status() int              { return r.Recorder.code }
func (r *Response) Header(key string) string { return r.Recorder.h.Get(key) }
func (r *Response) Location() string         { return r.Header("Location") }
func (r *Response) Body() string             { return string(r.Recorder.body) }

type NextOK struct{}

func (NextOK) ServeHTTP(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("X-Next", "hit")
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type NextCapture struct{}

func (NextCapture) ServeHTTP(w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("X-Next", "hit")
	_, _ = w.Write([]byte("next"))
	return nil
}

func AssertRedirect(t *testing.T, resp *Response, wantStatus int, wantLoc string) {
	t.Helper()
	if resp.Status() != wantStatus {
		t.Fatalf("status mismatch: got=%d want=%d", resp.Status(), wantStatus)
	}
	if got := resp.Location(); got != wantLoc {
		t.Fatalf("location mismatch: got=%q want=%q", got, wantLoc)
	}
}

func AssertPassedThrough(t *testing.T, resp *Response, wantStatus int) {
	t.Helper()
	if resp.Header("X-Next") != "hit" {
		t.Fatalf("expected request to pass to next handler")
	}
	if resp.Status() != wantStatus {
		t.Fatalf("unexpected next status: got=%d want=%d", resp.Status(), wantStatus)
	}
}
