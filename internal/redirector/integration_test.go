// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 Jason Giese (Bl4cky99)

//go:build integration

package redirector_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func findCaddy(root string) (string, error) {
	if p := os.Getenv("CADDY_BIN"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	candidates := []string{
		filepath.Join(root, "caddy"),
		filepath.Join(root, "bin", "caddy"),
		filepath.Join(root, "dist", "caddy"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}

	if p, err := exec.LookPath("caddy"); err == nil {
		return p, nil
	}

	return "", errors.New("caddy binary not found (set CADDY_BIN or place ./caddy at repo root)")
}

var _ = Describe("Caddyfile", Label("integration"), func() {
	It("serves requests end-to-end", func() {
		root := ProjectRoot()
		cfg := filepath.Join(root, "tests", "configs", "Caddyfile")

		_, err := os.Stat(cfg)
		Expect(err).NotTo(HaveOccurred(), "Caddyfile not found at %s", cfg)

		caddyBin, err := findCaddy(root)
		if err != nil {
			Skip(err.Error())
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, caddyBin, "run", "--config", cfg, "--adapter", "caddyfile")
		cmd.Dir = filepath.Dir(cfg)

		stdout, err := cmd.StdoutPipe()
		Expect(err).NotTo(HaveOccurred())
		stderr, err := cmd.StderrPipe()
		Expect(err).NotTo(HaveOccurred())

		var outBuf, errBuf bytes.Buffer
		go func() { _, _ = io.Copy(&outBuf, stdout) }()
		go func() { _, _ = io.Copy(&errBuf, stderr) }()

		Expect(cmd.Start()).To(Succeed())
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		healthURL := "http://localhost:8080/health"
		deadline := time.Now().Add(10 * time.Second)
		var lastErr error
		for time.Now().Before(deadline) {
			resp, err := http.Get(healthURL) //nolint:noctx
			if err == nil {
				_ = resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return
				}
			} else {
				lastErr = err
			}
			time.Sleep(150 * time.Millisecond)
		}

		Fail(fmt.Sprintf("caddy did not become ready; lastErr=%v\n--- STDOUT ---\n%s\n--- STDERR ---\n%s",
			lastErr, outBuf.String(), errBuf.String()))
	})
})
