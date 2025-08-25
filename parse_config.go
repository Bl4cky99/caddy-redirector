// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package redirector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	yaml "gopkg.in/yaml.v3"
)

func (r *Redirector) loadExternalRules() error {
	for _, rf := range r.RulesFiles {
		abs := resolvePath(r.baseDir, rf.Path)

		data, err := os.ReadFile(abs)
		if err != nil {
			return err
		}

		format := pickFormat(rf.Format, abs)

		var er ExternalRules
		if err := unmarshalByFormat(format, data, &er, rf.Path); err != nil {
			return err
		}

		r.Hosts = mergeHosts(r.Hosts, er.Hosts)
	}
	return nil
}

func mergeHosts(dst, src []HostBlock) []HostBlock {
	index := make(map[string]int, len(dst))
	for i, hb := range dst {
		index[strings.ToLower(hb.Pattern)] = i
	}

	for _, s := range src {
		key := strings.ToLower(s.Pattern)
		if i, ok := index[key]; ok {
			mergeHostBlock(&dst[i], s)
			continue
		}
		dst = append(dst, s)
		index[key] = len(dst) - 1
	}
	return dst
}

func resolvePath(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	if base == "" {
		base = "."
	}
	return filepath.Join(base, p)
}

func pickFormat(explicit, path string) string {
	f := strings.ToLower(strings.TrimSpace(explicit))
	if f != "" {
		return f
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yml", ".yaml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".json":
		return "json"
	default:
		return "json"
	}
}

func unmarshalByFormat(format string, data []byte, out *ExternalRules, pathForErr string) error {
	switch format {
	case "json":
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("rule_file %q: %w", pathForErr, err)
		}
	case "yaml":
		if err := yaml.Unmarshal(data, out); err != nil {
			return fmt.Errorf("rule_file %q: %w", pathForErr, err)
		}
	case "toml":
		if err := toml.Unmarshal(data, out); err != nil {
			return fmt.Errorf("rule_file %q: %w", pathForErr, err)
		}
	default:
		return fmt.Errorf("rule_file %q: unsupported format %q", pathForErr, format)
	}
	return nil
}

func mergeHostBlock(dst *HostBlock, s HostBlock) {
	if s.Status != 0 {
		dst.Status = s.Status
	}
	if s.ToHost != "" {
		dst.ToHost = s.ToHost
	}
	if len(s.Exact) != 0 {
		if dst.Exact == nil {
			dst.Exact = make(map[string]string, len(s.Exact))
		}
		for k, v := range s.Exact {
			dst.Exact[k] = v
		}
	}
	if len(s.Prefix) != 0 {
		dst.Prefix = append(dst.Prefix, s.Prefix...)
	}
	if len(s.Regex) != 0 {
		dst.Regex = append(dst.Regex, s.Regex...)
	}
}
