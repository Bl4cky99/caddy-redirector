// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package e2e

import (
	"path/filepath"
	"testing"

	redir "github.com/Bl4cky99/caddy-redirector"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type Suite struct {
	T       *testing.T
	BaseDir string
}

func NewSuite(t *testing.T) *Suite {
	t.Helper()

	base := ConfigPath(t, "")

	return &Suite{T: t, BaseDir: base}
}

func (s *Suite) BuildRedirectorFromFiles(defaultCode int, paths ...string) *redir.Redirector {
	s.T.Helper()

	r := &redir.Redirector{
		DefaultCode: defaultCode,
	}

	r.RulesFiles = make([]redir.RulesFile, 0, len(paths))
	for _, rel := range paths {
		abs := ConfigPath(s.T, rel)
		r.RulesFiles = append(r.RulesFiles, redir.RulesFile{Path: abs})
	}

	if err := r.Provision(caddy.Context{}); err != nil {
		s.T.Fatalf("provision failed: %v", err)
	}
	return r
}

func (s *Suite) BuildRedirectorInline(defaultCode int, hosts []redir.HostBlock) *redir.Redirector {
	s.T.Helper()

	r := &redir.Redirector{
		DefaultCode: defaultCode,
		Hosts:       hosts,
	}
	if err := r.Provision(caddy.Context{}); err != nil {
		s.T.Fatalf("provision failed: %v", err)
	}
	return r
}

func (s *Suite) RunOnce(r *redir.Redirector, req *RequestSpec, next caddyhttp.Handler) *Response {
	s.T.Helper()
	if next == nil {
		next = NextOK{}
	}
	rr := newRecorder()
	if err := r.ServeHTTP(rr, req.Build(), next); err != nil {
		s.T.Fatalf("ServeHTTP error: %v", err)
	}
	return &Response{Recorder: rr}
}

func (s *Suite) Abs(rel string) string {
	return filepath.Join(ProjectRoot(s.T), "tests", rel)
}
