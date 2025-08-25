// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package redirector

import "regexp"

type Redirector struct {
	Hosts       []HostBlock
	DefaultCode int
}

type HostBlock struct {
	Pattern string
	ToHost  string
	Status  int
	Exact   map[string]string
	Prefix  []PrefixRule
	Regex   []RegexRule
}

type PrefixRule struct {
	From string
	To   string
}

type RegexRule struct {
	Pattern string
	To      string
}

type compiledHostBlock struct {
	matchAll    bool
	suffix      string
	exactHost   string
	toHost      string
	status      int
	exactPaths  map[string]string
	prefixRules []PrefixRule
	regexRules  []compiledRegexRule
}

type compiledRegexRule struct {
	re *regexp.Regexp
	to string
}
