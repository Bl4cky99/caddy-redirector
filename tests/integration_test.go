// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

//go:build integration

package e2e

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func findCaddy(t *testing.T, root string) (string, error) {
	t.Helper()

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

func Test_Caddyfile_EndToEnd(t *testing.T) {
	root := ProjectRoot(t)
	cfg := filepath.Join(root, "tests", "configs", "Caddyfile")

	if _, err := os.Stat(cfg); err != nil {
		t.Fatalf("Caddyfile not found at %s", cfg)
	}

	caddyBin, err := findCaddy(t, root)
	if err != nil {
		t.Skipf("%v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, caddyBin, "run", "--config", cfg, "--adapter", "caddyfile")
	cmd.Dir = filepath.Dir(cfg)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	var outBuf, errBuf bytes.Buffer
	go func() { _, _ = io.Copy(&outBuf, stdout) }()
	go func() { _, _ = io.Copy(&errBuf, stderr) }()

	if err := cmd.Start(); err != nil {
		t.Fatalf("start caddy: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	healthURL := "http://localhost:8080/health"
	deadline := time.Now().Add(10 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := http.Get(healthURL)
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

	t.Fatalf("caddy did not become ready; lastErr=%v\n--- STDOUT ---\n%s\n--- STDERR ---\n%s",
		lastErr, outBuf.String(), errBuf.String())
}
