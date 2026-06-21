// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package redirector_test

import (
	"path/filepath"

	redir "github.com/Bl4cky99/caddy-redirector"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Suite struct {
	BaseDir string
}

func NewSuite() *Suite {
	return &Suite{BaseDir: ConfigPath("")}
}

func (s *Suite) BuildRedirectorFromFiles(defaultCode int, paths ...string) *redir.Redirector {
	GinkgoHelper()

	r := &redir.Redirector{DefaultCode: defaultCode}
	r.RulesFiles = make([]redir.RulesFile, 0, len(paths))
	for _, rel := range paths {
		r.RulesFiles = append(r.RulesFiles, redir.RulesFile{Path: ConfigPath(rel)})
	}

	Expect(r.Provision(caddy.Context{})).To(Succeed())
	return r
}

func (s *Suite) BuildRedirectorInline(defaultCode int, hosts []redir.HostBlock) *redir.Redirector {
	GinkgoHelper()

	r := &redir.Redirector{
		DefaultCode: defaultCode,
		Hosts:       hosts,
	}
	Expect(r.Provision(caddy.Context{})).To(Succeed())
	return r
}

func (s *Suite) RunOnce(r *redir.Redirector, req *RequestSpec, next caddyhttp.Handler) *Response {
	GinkgoHelper()

	if next == nil {
		next = NextOK{}
	}
	rr := newRecorder()
	Expect(r.ServeHTTP(rr, req.Build(), next)).To(Succeed())
	return &Response{Recorder: rr}
}

func (s *Suite) Abs(rel string) string {
	return filepath.Join(ProjectRoot(), "tests", rel)
}
