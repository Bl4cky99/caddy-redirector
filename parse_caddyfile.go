// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Jason Giese (Bl4cky99)

package redirector

import (
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func parseCaddyFile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var r Redirector
	if err := r.UnmarshalCaddyfile(h.Dispenser); err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *Redirector) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "host":
				if err := parseHost(d, r); err != nil {
					return err
				}
			case "status":
				if err := parseStatus(d, r); err != nil {
					return err
				}
			default:
				return d.Errf("unknown directive %q in redirector", d.Val())
			}
		}
	}
	return nil
}

func parseStatus(d *caddyfile.Dispenser, r *Redirector) error {
	var code string
	if !d.Args(&code) {
		return d.ArgErr()
	}

	r.DefaultCode = parseRedirectCode(code)
	if r.DefaultCode == 0 {
		return d.Errf("status must be 301, 307 or 308, %s given", d.Val())
	}
	return nil
}

func parseHost(d *caddyfile.Dispenser, r *Redirector) error {
	var pat string
	if !d.Args(&pat) {
		return d.ArgErr()
	}
	hb := HostBlock{Pattern: pat, Exact: make(map[string]string)}

	for d.NextBlock(1) {
		switch d.Val() {
		case "status":
			if err := parseHostStatus(d, &hb); err != nil {
				return err
			}
		case "to_host":
			if err := parseHostToHost(d, &hb); err != nil {
				return err
			}
		case "exact":
			if err := parseHostExact(d, &hb); err != nil {
				return err
			}
		case "prefix":
			if err := parseHostPrefix(d, &hb); err != nil {
				return err
			}
		case "regex":
			if err := parseHostRegex(d, &hb); err != nil {
				return err
			}
		default:
			return d.Errf("unknown subdirective %q in host block", d.Val())
		}
	}
	r.Hosts = append(r.Hosts, hb)
	return nil
}

func parseHostToHost(d *caddyfile.Dispenser, hb *HostBlock) error {
	var th string
	if !d.Args(&th) {
		return d.ArgErr()
	}
	hb.ToHost = th
	return nil
}

func parseHostExact(d *caddyfile.Dispenser, hb *HostBlock) error {
	var from, to string
	if !d.Args(&from, &to) {
		return d.ArgErr()
	}
	hb.Exact[from] = to
	return nil
}

func parseHostStatus(d *caddyfile.Dispenser, hb *HostBlock) error {
	var code string
	if !d.Args(&code) {
		return d.ArgErr()
	}

	hb.Status = parseRedirectCode(code)
	if hb.Status == 0 {
		return d.Errf("status must be 301, 307 or 308, %s given", d.Val())
	}
	return nil
}

func parseHostPrefix(d *caddyfile.Dispenser, hb *HostBlock) error {
	var from, to string
	if !d.Args(&from, &to) {
		return d.ArgErr()
	}

	hb.Prefix = append(hb.Prefix, PrefixRule{From: from, To: to})
	return nil
}

func parseHostRegex(d *caddyfile.Dispenser, hb *HostBlock) error {
	var pat, to string
	if !d.Args(&pat, &to) {
		return d.ArgErr()
	}

	hb.Regex = append(hb.Regex, RegexRule{Pattern: pat, To: to})
	return nil
}

func parseRedirectCode(code string) int {
	switch code {
	case "301":
		return 301
	case "307":
		return 307
	case "308":
		return 308
	default:
		return 0
	}
}
