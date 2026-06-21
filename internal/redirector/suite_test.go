// SPDX-License-Identifier: MIT
// Copyright (c) 2025-2026 Jason Giese (Bl4cky99)

package redirector_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRedirector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redirector Suite")
}
