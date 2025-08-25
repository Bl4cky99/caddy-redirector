// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package redirector

import "regexp"

type Redirector struct {
	Hosts       []HostBlock
	DefaultCode int
	RulesFiles  []RulesFile

	baseDir string `json:"-"`
}

type RulesFile struct {
	Path   string `json:"path"   yaml:"path"   toml:"path"`
	Format string `json:"format,omitempty" yaml:"format,omitempty" toml:"format,omitempty"`
}

type ExternalRules struct {
	Hosts []HostBlock `json:"hosts" yaml:"hosts" toml:"hosts"`
}

type HostBlock struct {
	Pattern string            `json:"pattern" yaml:"pattern" toml:"pattern"`
	ToHost  string            `json:"to_host" yaml:"to_host" toml:"to_host"`
	Status  int               `json:"status" yaml:"status" toml:"status"`
	Exact   map[string]string `json:"exact" yaml:"exact" toml:"exact"`
	Prefix  []PrefixRule      `json:"prefix" yaml:"prefix" toml:"prefix"`
	Regex   []RegexRule       `json:"regex" yaml:"regex" toml:"regex"`
}

type PrefixRule struct {
	From string `json:"from" yaml:"from" toml:"from"`
	To   string `json:"to"      yaml:"to"      toml:"to"`
}

type RegexRule struct {
	Pattern string
	To      string
}

type compiledHostBlock struct {
	matchAll      bool
	suffix        string
	exactHost     string
	toHost        string
	status        int
	exactPaths    map[string]string
	prefixBuckets map[string][]PrefixRule
	regexRules    []compiledRegexRule
}

type compiledRegexRule struct {
	re *regexp.Regexp
	to string
}
