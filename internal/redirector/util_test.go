// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 Jason Giese (Bl4cky99)

package redirector_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func ProjectRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		Fail("cannot determine caller path")
	}
	// file is .../internal/redirector/util_test.go — go up two levels to reach module root
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
}

func ConfigPath(rel string) string {
	return filepath.Join(ProjectRoot(), "tests", rel)
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

func AssertRedirect(resp *Response, wantStatus int, wantLoc string) {
	GinkgoHelper()
	Expect(resp.Status()).To(Equal(wantStatus), "status mismatch")
	Expect(resp.Location()).To(Equal(wantLoc), "location mismatch")
}

func AssertPassedThrough(resp *Response, wantStatus int) {
	GinkgoHelper()
	Expect(resp.Header("X-Next")).To(Equal("hit"), "expected request to pass to next handler")
	Expect(resp.Status()).To(Equal(wantStatus), "unexpected next status")
}
