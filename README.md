<a id="readme-top"></a>

<br />
<div align="center">
    <a href="https://github.com/Bl4cky99/caddy-redirector">
        <img src="README_ASSETS/logo.png" width="600">
    </a>
    <h3 align="center">caddy-redirector</h3>
    <p align="center">
        A fast, configurable Caddy <b>HTTP middleware</b> for granular host and path migrations.   
        <br/>
        Designed for large redirect sets after domain or app restructures, with a clean Caddyfile syntax and production-friendly behavior.
        <br/>
        Redirect rules can be configured flexibly, either through the Caddyfile or via external config files.
        <br/><br/>
        <a href="https://github.com/Bl4cky99/caddy-redirector/issues/new?template=bug_report.yml">Report Bug</a>
        &middot;
        <a href="https://github.com/Bl4cky99/caddy-redirector/issues/new?template=feature_request.yml">Request Feature</a>
        <br/><br/>
    </p>
</div>



<details>
<summary>Table of Contents</summary>
<ol>
  <li><a href="#features">Features</a></li>
  <li><a href="#installation">Installation</a>
    <ul>
      <li><a href="#install-build">Build a Caddy binary with this module (recommended)</a></li>
      <li><a href="#install-binary-prebuild">Caddy binary (prebuild)</a></li>
      <li><a href="#install-docker">Docker (build yourself)</a></li>
      <li><a href="#install-docker-prebuild">Docker (prebuild image)</a></li>
    </ul>
  </li>
  <li><a href="#configuration">Configuration</a>
    <ul>
      <li><a href="#host-patterns">Host patterns</a></li>
      <li><a href="#rule-types">Rule types</a></li>
      <li><a href="#targets">Targets</a></li>
      <li><a href="#status-codes">Status code resolution</a></li>
      <li><a href="#precedence">Precedence</a></li>
      <li><a href="#configuration-formats">Alternative formats (YAML / JSON / TOML)</a></li>
    </ul>
  </li>
  <li><a href="#examples">Examples</a></li>
  <li><a href="#quick-testing">Quick testing</a></li>
  <li><a href="#troubleshooting">Troubleshooting</a></li>
  <li><a href="#compatibility">Compatibility</a></li>
  <li><a href="#security-notes">Security notes</a></li>
  <li><a href="#roadmap">Roadmap</a></li>
  <li><a href="#faq">FAQ</a></li>
  <li><a href="#developer-documentation">Developer Documentation</a>
    <ul>
      <li><a href="#architecture">Architecture</a></li>
      <li><a href="#tests">Tests</a></li>
      <li><a href="#benchmarks">Benchmarks</a></li>
      <li><a href="#extending">Extending</a></li>
      <li><a href="#minimal-dev-loop">Minimal dev loop</a></li>
    </ul>
  </li>
  <li><a href="#license">License</a></li>
</ol>
</details>

---

## <span id="features">Features</span>

- **Host-aware rules**: define redirects grouped by source host (including wildcards like `*.example.com` or catch-all `*`).
- **Three match modes** per host:
  - `exact` (path = path)
  - `prefix` (longest prefix wins)
  - `regex` (Go RE2; `$1`, `$2`, … captures)
- **Multiple config formats**: rules can be defined not only in the Caddyfile, but also in external YAML, JSON, or TOML files.

- **Configurable status code**: global default and host-level override (301, 307 or 308).
- **Absolute vs relative targets**:
  - Absolute targets (`https://…`) are used verbatim.
  - Relative targets (`/new/path`) are attached to the configured `to_host` or stay on the same host if none is set.
- **Performance-minded**:
  - Exact lookups via map.
  - Prefix rules pre-sorted by length to pick the most specific match quickly.
  - Regexes compiled once at provision time.
- **Clear precedence**: `exact` > `prefix` > `regex`.

> Current default: query strings are **not** preserved automatically (a per-rule option can be added later).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="installation">Installation</span>

### <span id="install-build">Build a Caddy binary with this module (recommended)</span>

```bash
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

# From your (empty) workspace or any folder:
xcaddy build \
  --with github.com/Bl4cky99/caddy-redirector@latest
```

This produces a `caddy` binary that includes the `redirector` module.

> If you are developing locally inside this repository, you can build with:
> `xcaddy build --with github.com/Bl4cky99/caddy-redirector=.`

### <span id="install-docker">Docker (optional)</span>

The Caddy team recommends building your own Caddy binary with your chosen modules.  
If you still want an image, you can use a multi-stage build to bake this module:

```dockerfile
# Example only. Prefer building your own Caddy binary with xcaddy.
FROM caddy:builder AS builder
RUN xcaddy build --with github.com/Bl4cky99/caddy-redirector@latest

FROM caddy:latest
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
# Provide your Caddyfile via bind mount or COPY
```


### <span id="install-binary-prebuild">Caddy binary (prebuild)</span>

Prebuilt, self-contained Caddy binaries that already include this module are attached to each GitHub Release.

**Download & verify (Linux, amd64/arm64):**

```bash
# Pick a released version tag
VERSION=v1.0.0
OS=linux
ARCH=amd64   # or arm64

# Download the tarball and checksum
curl -sSL -o caddy-${OS}-${ARCH}.tar.gz \
  https://github.com/Bl4cky99/caddy-redirector/releases/download/${VERSION}/caddy-${OS}-${ARCH}.tar.gz
curl -sSL -o caddy-${OS}-${ARCH}.tar.gz.sha256 \
  https://github.com/Bl4cky99/caddy-redirector/releases/download/${VERSION}/caddy-${OS}-${ARCH}.tar.gz.sha256

# Verify checksum (must print: OK)
sha256sum -c caddy-${OS}-${ARCH}.tar.gz.sha256

# Install
tar -xzf caddy-${OS}-${ARCH}.tar.gz
sudo install -m 0755 caddy-${OS}-${ARCH} /usr/local/bin/caddy
caddy version
```

> Notes
> - Binaries are built via `xcaddy` and embed this module; you **do not** need to rebuild Caddy to use it.
> - Ensure you comply with the licenses of Caddy and all included modules in downstream distributions.

---

### <span id="install-docker-prebuild">Docker (prebuild)</span>

Official-style image published to GHCR, containing a Caddy binary prebuilt with this module.

**Pull & run:**

```yaml
IMAGE=ghcr.io/bl4cky99/caddy-redirector
TAG=v1.0.0   # or 'latest'

docker pull ${IMAGE}:${TAG}

# Run with your Caddyfile from the current directory
docker run --rm -it \
  -p 8080:8080 \
  -v "$PWD/Caddyfile:/etc/caddy/Caddyfile:ro" \
  ${IMAGE}:${TAG} run --config /etc/caddy/Caddyfile --adapter caddyfile
```

**docker-compose example:**

```yaml
services:
  caddy:
    image: ghcr.io/bl4cky99/caddy-redirector:v1.0.0
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
```

> Notes
> - Image tags follow release tags: `vX.Y.Z`, plus floating tags `X`, `X.Y`, and `latest`.
> - Mount your own `Caddyfile`; the image only provides the prebuilt Caddy binary with this module.
> - For production, persist `/data` and `/config` volumes as shown above.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="configuration">Configuration (Caddyfile)</span>

Top-level directive:

```caddy
redirector {
  # Optional: global default status code (301 or 308). Default: 308
  status 308

  # One or more host blocks:
  host <pattern> {
    # Optional per-host override:
    status 301

    # Optional target host. If set, relative targets become absolute URLs on this host.
    to_host new.example

    # Rules (any order):
    exact  /old        /new
    prefix /blog/      /news/
    regex  ^/u/([0-9]+)$  /users/$1
  }
}
```

### <span id="host-patterns">Host patterns</span>

- `host example.com` – exact host match.
- `host *.example.com` – matches any **subdomain** of `example.com` (does **not** match the apex).
- `host *` – catch-all (least specific; used only if more specific hosts didn’t match).

### <span id="rule-types">Rule types</span>

- **exact `<from> <to>`**  
  When the request path equals `<from>`, redirect to `<to>`.

- **prefix `<from> <to>`**  
  When the request path starts with `<from>`, redirect to `<to>` plus the remaining suffix.  
  The **longest** matching prefix wins.

- **regex `<pattern> <to>`**  
  When the regex matches the path, produce the target by `regexp.ReplaceAllString(path, <to>)`.  
  Use `$1`, `$2`, … for capture groups. The **first** matching regex (in declaration order) wins.

### <span id="targets">Targets</span>

- **Absolute target (`http://…` or `https://…`)**  
  Used verbatim. `to_host` is ignored for that rule.
- **Relative target (`/something`) with `to_host`**  
  Redirect to `{scheme}://{to_host}{target}`.  
  Scheme is inferred: `https` by default, or `http` if `X-Forwarded-Proto: http` and no TLS.
- **Relative target without `to_host`**  
  Redirect to the same host with the new path.

### <span id="status-codes">Status code resolution</span>

1. Use **host-level** `status` if set.
2. Else use **global** `status` (from the `redirector` block).
3. Else default to **308** (Permanent Redirect).

### <span id="precedence">Precedence</span>

1. `exact` rules  
2. `prefix` rules (longest `from` wins)  
3. `regex` rules (first match wins)  
4. No match → pass to the next handler


### <span id="configuration-formats">Alternative formats (YAML / JSON / TOML)</span>

In addition to the [Caddyfile configuration](#configuration), rules can also be defined in structured formats:

- **YAML** ([example/rules.yaml](example/rules.yaml))  
- **JSON** ([example/rules.json](example/rules.json))  
- **TOML** ([example/rules.toml](example/rules.toml))  

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="examples">Examples</span>

- **Exact migration** 

```caddy
:8080

route {
  redirector {
    status 308

    host old.example {
      to_host new.example
      exact  /start   /getting-started
      exact  /about   /company
      exact  /docs    /documentation
    }
  }
}
```

`curl -i -H 'Host: old.example' http://localhost:8080/docs`  
→ `308 Location: https://new.example/documentation`

---

- **Prefix with longest-prefix**  

```caddy
:8080

route {
  redirector {
    host app.example {
      to_host portal.example
      prefix /blog/        /news/
      prefix /blog/2024/   /archive/2024/
    }
  }
}
```

`/blog/2024/…` goes to `/archive/2024/…` because the longer prefix wins.

---

- **Regex with captures**  

```caddy
:8080

route {
  redirector {
    host accounts.example {
      # Send to a completely different site (absolute URL)
      regex ^/u/([0-9]+)$  https://profiles.example/users/$1
    }
  }
}
```

---

- **Wildcard host and catch-all**

```caddy
:8080

route {
  redirector {
    # Subdomains only (not apex)
    host *.old.example {
      to_host new.example
      prefix /a/ /alpha/
    }

    # Fallback for any other host
    host * {
      exact /ping /health
    }
  }
}
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="quick-testing">Quick testing</span>

Start Caddy:

```bash
./caddy run --config Caddyfile --adapter caddyfile
```

Hit endpoints:

```bash
# Exact
curl -i -H 'Host: old.example' http://localhost:8080/docs

# Prefix
curl -i -H 'Host: app.example' http://localhost:8080/blog/2024/post

# Regex
curl -i -H 'Host: accounts.example' http://localhost:8080/u/42
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="troubleshooting">Troubleshooting</span>

- **“unrecognized directive: redirector”**  
  Ensure you built Caddy with this module:
  `xcaddy build --with github.com/Bl4cky99/caddy-redirector@latest`

- **Config adapts but redirects don’t happen**  
  Confirm your `host` block actually matches the request `Host` header (including ports in dev).  
  For wildcard `*.example.com`, remember it does **not** match `example.com` itself.

- **Scheme is wrong (http vs https)**  
  The module infers the scheme as `https` by default, or `http` if the request is not TLS-terminated and the proxy sets `X-Forwarded-Proto: http`.  
  Make sure your proxy sends `X-Forwarded-Proto` exactly.

- **Regex rule not firing**  
  Regexes use Go’s RE2 syntax. Start with `^` and end with `$` when matching the entire path.  
  The first matching rule wins; check your rule order.

- **Caddyfile parse errors**  
  Unknown subdirectives inside a `host` block will be rejected explicitly. Verify spelling and arguments.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="compatibility">Compatibility</span>

- Go: 1.25+
- Caddy: v2 (build with `xcaddy`)
- Regex engine: Go RE2 (no backtracking)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="security-notes">Security notes</span>

- Be careful with wide regexes that can redirect a large portion of your site; keep exact/prefix rules for common paths.
- Avoid user-controlled rule inputs; store redirects in your config or vetted data files.
- Absolute targets (`http://…`) will downgrade scheme on purpose—use only if you intend that.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---


## <span id="roadmap">Roadmap</span>

- Rule-level status and query preservation flags.
- Optional query passthrough and path templating beyond `$1`.
- Benchmark suite and micro-optimizations for very large redirect tables.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="faq">FAQ</span>

**Q: Does `*.example.com` match `example.com`?**  
A: No. Add a separate `host example.com` block for the apex.

**Q: Do query strings get forwarded?**  
A: Not by default in the current version. This can be added per rule/host later.

**Q: How do I choose 301, 307 and 308?**  
A: Use `status 301`, `status 307` or `status 308` at the global level or inside a host block.  
`308` keeps the HTTP method and is often the safer default for permanent moves.

**Q: Can I use absolute URLs in rules?**  
A: Yes. If the target starts with `http://` or `https://`, it is used verbatim and `to_host` is ignored for that rule.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="developer-documentation">Developer Documentation</span>

### <span id="architecture">Architecture (developer view)</span>

Project layout:

```
.
├─ model.go              # data types: HostBlock, PrefixRule, RegexRule, compiledHostBlock, etc.
├─ parse_caddyfile.go    # Caddyfile parsing (UnmarshalCaddyfile), directive registration
├─ parse_config.go       # Config parsing for external redirect rule files (json, yaml, toml)
├─ redirector.go         # module wiring, Provision/Validate/ServeHTTP, core logic
├─ bench_test.go         # simple benchmark test for different redirect methods
└─ go.mod
```

Lifecycle:

1. **Caddyfile → structs**  
   `UnmarshalCaddyfile` parses `redirector { host … }` blocks into `Redirector.Hosts`.
2. **Provision**  
   - Compute compiled form per host: normalize patterns, **compile regex** once, **sort** prefix rules by descending `From` length, resolve per-host `status` (fallback to global).
3. **ServeHTTP** (hot path)  
   - Pick host block: exact host > wildcard suffix (`*.example.com`) > `*`.  
   - Try `exact`, then `prefix` (longest wins), then `regex` (first wins).  
   - Build the target (absolute vs relative + optional `to_host`).  
   - `http.Redirect(w, req, code)`.

Performance & complexity:

- Exact: O(1) map lookup per host block.
- Prefix: O(P) with small constant factor after sorting. For very large P, consider a radix tree.
- Regex: O(R) scan with precompiled regex; keep R modest, put fast rules earlier (exact/prefix cover most cases).

Reload behavior:

- Caddy will call `Provision` on each reload; regexes are recompiled, prefix lists resorted, old state is discarded.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

### <span id="tests">Tests</span>

This repository ships a self-contained test suite under `tests/` (package `e2e`).  
It exercises the handler directly (no external Caddy process) and uses dedicated rule files under `tests/configs/`.

**Layout**
- `tests/unit_test.go` – core unit tests (exact/prefix/regex, host precedence, merging, scheme inference).
- `tests/integration_test.go` - integration test using build caddy binary (must be present in repo)
- `tests/configs/` – rule files (JSON/YAML/TOML) used by the suite.

> **Note:** The module currently keeps a package-global compiled state. Do **not** use `t.Parallel()`; tests run serially by design.

```bash
# Unit tests
go test -v -race -cover -tags=unit ./tests

# Integration test
go test -v -tags=integration ./tests
```

**Test assets**
All rule files consumed by the suite live in `tests/configs/`:
- `rules_exact.json`, `rules_prefix.yaml`, `rules_regex.toml`, `rules_wildcards.yaml`
- `merge_a.json`, `merge_b.json` (merge order/last-wins)
- `scheme.yaml` (scheme inference)
- `bad_regex.yaml`, `unknown.data`, `noext`, `regex_no_slash.json`, `absolute_exact.yaml`, `case.json` (edge/error cases)


<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

### <span id="benchmarks">Benchmarks</span>

This section documents the micro-benchmarks used to characterize the matching cost of the module’s three rule types (**exact**, **prefix**, **regex**) and to validate the expected time complexity.

**What is measured?**

- The benchmarks run entirely **in-process** (no network) using `ServeHTTP` against synthetic requests.
- Rule sets are generated in-memory:
  - `Exact_Hit/Miss`: 1,000 exact rules
  - `Prefix_Longest_Hit/Miss`: 1,000 prefix rules (pre-sorted longest-first)
  - `Regex_Hit/Miss`: 100 compiled regex rules
- Each test uses a minimal `http.ResponseWriter` and a no-op `next` handler to reduce measurement noise.

---

**How to run locally**

> File: `tests/bench_test.go`

```
#### Run all benchmarks with memory stats
go test -run '^$' -bench . -benchmem ./tests

#### Get more stable numbers (5 repetitions)
go test -run '^$' -bench . -benchmem -count=5 ./tests > bench.txt

#### (Optional) Pin to a single OS thread for reproducibility
GOMAXPROCS=1 go test -run '^$' -bench . -benchmem ./tests
```

---

**Interpreting the output**

Each line shows:
- `ns/op`: average nanoseconds per request processed
- `B/op`: bytes allocated per operation
- `allocs/op`: number of allocations per operation

Benchmark Results (Plattform: Linux, amd64, i7-9700K):
```plain
goos: linux
goarch: amd64
pkg: github.com/Bl4cky99/caddy-redirector
cpu: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz

BenchmarkExact_Hit_1e3-8 ~1500 ns/op 1292 B/op 14 allocs/op
BenchmarkExact_Miss_1e3-8 ~145 ns/op 160 B/op 3 allocs/op

BenchmarkPrefix_Longest_Hit_1e3-8 ~8,000 ns/op 5638 B/op 15 allocs/op
BenchmarkPrefix_Miss_1e3-8 ~145 ns/op 160 B/op 3 allocs/op

BenchmarkRegex_Hit_1e2-8 ~2,400 ns/op 1474 B/op 20 allocs/op
BenchmarkRegex_Miss_1e2-8 ~3,700 ns/op 162 B/op 3 allocs/op
```

---

**What these numbers mean:**

- **Exact (map lookup)**
  - *Hit*: ~0.48 µs with 3 small allocations (headers/redirect path composition). Time is effectively **O(1)** with respect to the number of rules.
  - *Miss*: ~30 ns, zero allocations — very fast early exit, also **O(1)**.

- **Prefix (linear scan, longest-first)**
  - *Longest hit*: ~25 µs, a few allocations for response header/Location. This scales roughly **O(P)** with the number of prefix rules because we scan until we find the first match (you pre-sort for best specificity, not for speed).
  - *Miss*: ~0.92 µs, zero allocations — still **O(P)**, but exits after checking all rules without building a redirect.

- **Regex (linear scan over compiled RE2)**
  - *Hit*: ~1.14 µs, ~336 B / 9 allocs — includes `ReplaceAllString` with captures and building the Location header. Complexity **O(R)** for the number of regex rules.
  - *Miss*: ~3.1 µs, zero allocations — time reflects checking each compiled regex and failing to match (**O(R)**).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

### <span id="extending">Extending the module (ideas)</span>

- **Query handling**: `preserve_query on|off` at rule or host level; or `append_query key=value`.
- **Metrics/logging**: counters per rule, structured logs with rule IDs.
- **Radix tree** for prefix rules when you have very large sets.
- **Dual-stack host matching**: treat `Host: example.com:PORT` gracefully in dev (normalize port).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="minimal-dev-loop">Minimal dev loop</span>

```
# 1) Build a local Caddy with your working copy
xcaddy build --with github.com/Bl4cky99/caddy-redirector=.

# 2) Run with your Caddyfile
./caddy run --config Caddyfile --adapter caddyfile

# 3) Test
curl -i -H 'Host: old.example' http://localhost:8080/old
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## <span id="license">License</span>

This project is licensed under the **MIT License**.

- Copyright © 2025 [Jason Giese (Bl4cky99)](https://github.com/Bl4cky99)
- See the full text in [LICENSE](./LICENSE).

**Notes for users and integrators**
- Commercial use, modification, distribution, and private forks are permitted.
- Keep the copyright and permission notice from the MIT license in all copies/substantial portions.
- (Optional) Add an SPDX header to source files for tooling: `// SPDX-License-Identifier: MIT`.

**Third-party software**
This repository contains only the module’s source. When building Caddy with this module, you must also comply with the licenses of Caddy and any other included modules/dependencies in your final binary or container image.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

**Happy redirecting!**
