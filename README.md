<div align="center">

# remoteok

[![CI](https://github.com/jjuanrivvera/remoteok-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/jjuanrivvera/remoteok-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/jjuanrivvera/remoteok-cli)](https://github.com/jjuanrivvera/remoteok-cli/releases/latest)
[![Coverage](https://img.shields.io/badge/coverage-%E2%89%A580%25-brightgreen)](https://github.com/jjuanrivvera/remoteok-cli/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jjuanrivvera/remoteok-cli.svg)](https://pkg.go.dev/github.com/jjuanrivvera/remoteok-cli)
[![Go version](https://img.shields.io/github/go-mod/go-version/jjuanrivvera/remoteok-cli)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/jjuanrivvera/remoteok-cli)
[![Built with cliwright](https://img.shields.io/badge/built_with-cliwright-1f6feb)](https://cliwright.jjuanrivvera.com)

**Remote OK jobs from your terminal — search and filter the public jobs feed, agent-friendly output (JSON/YAML/CSV/MCP).**

[Documentation](https://jjuanrivvera.github.io/remoteok-cli/) · [Command reference](https://jjuanrivvera.github.io/remoteok-cli/commands/remoteok/)

</div>

A fast, scriptable, **read-only** command-line client for the [Remote OK](https://remoteok.com)
jobs API. Built for machines: JSON/YAML/CSV output, an `-o id` mode for piping, a `--jq`
filter, and an MCP server — so an AI assistant or a shell pipeline can discover remote work
programmatically.

Remote OK is a free public API and needs **no account or token**.

> ### Attribution requirement — please read
> Remote OK's [API Terms of Service](https://remoteok.com/api) require a **follow backlink**
> to <https://remoteok.com> (with follow, *without* `nofollow`) and a mention of *Remote OK*
> as the source whenever you display their data — *"If you do not we'll have to suspend API
> access."* When you surface these jobs in a UI, site, or bot, include that backlink. The CLI
> prints a `Source: Remote OK — https://remoteok.com` line on **stderr** as a reminder
> (suppress with `--quiet`); in machine formats the attribution is your responsibility.

## Install

**Install script (macOS/Linux)** — downloads the release archive, verifies its SHA-256, and
installs the binary:

```sh
curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/remoteok-cli/main/install.sh | sh
```

**Homebrew:**

```sh
brew install jjuanrivvera/remoteok-cli/remoteok-cli
```

**Scoop (Windows):**

```powershell
scoop bucket add remoteok https://github.com/jjuanrivvera/scoop-remoteok-cli
scoop install remoteok
```

**Go:**

```sh
go install github.com/jjuanrivvera/remoteok-cli/cmd/remoteok@latest
```

## Quickstart

```sh
# Recent Go jobs, newest first
remoteok jobs list --tag golang --limit 20

# Multiple tags are AND-ed; add a salary floor
remoteok jobs list --tags golang,remote --min-salary 120000

# Keyword search across position/company/description/tags
remoteok jobs list --search kubernetes -o json

# Filter by company, export CSV
remoteok jobs list --company stripe -o csv

# Only postings from the last week (relative), or since an absolute date
remoteok jobs list --since 7d --tag golang
remoteok jobs list --since 2026-07-12 -o json

# One job by id (from the current feed)
remoteok jobs get 1135010 -o json

# Pipe ids into other tools
remoteok jobs list --tag golang -o id | head

# See exactly what would be requested (no network)
remoteok jobs list --tag golang --dry-run
```

## Output & filtering

- `-o table` (default, colored on a TTY), `-o json`, `-o yaml`, `-o csv`, `-o id`.
- `--columns id,position,company` selects table/CSV columns; `--jq '.[].company'` runs a
  gojq expression over the response.
- Notes/attribution go to **stderr** so stdout stays pipe-clean.
- `--quiet` suppresses chatter (including the attribution line).

## Filters (`jobs list`)

| Flag | Meaning |
|---|---|
| `--tag` / `--tags` | Require every listed tag (AND), case-insensitive |
| `--search` | Keyword over position, company, description, and tags |
| `--company` | Substring over the company name |
| `--min-salary` | Keep jobs whose advertised `salary_max` clears the amount |
| `--since` / `--posted-after` | Keep jobs posted on/after a date (`YYYY-MM-DD`) or window (`Nd`/`Nw`, e.g. `7d`, `2w`) |
| `--limit` | Cap the number of results |

All filters are applied client-side over the live feed, so they compose freely.

**`--tag` matches Remote OK's fixed tag vocabulary** (e.g. `golang`, `react`, `devops`,
`remote`). A term that is *not* a real tag — an industry like `fintech`, a role, or a free
keyword — matches nothing via `--tag`; use `--search` for those instead.

**Remote OK rarely publishes salary** (only a few listings per feed carry one), so
`--min-salary` drops every listing with no published minimum — i.e. most of them. Use it to
narrow a broad query, not as a primary filter. When it excludes listings for lack of a
published salary, the CLI notes how many on stderr (suppress with `--quiet`).

## Configuration

There are no secrets. Optional overrides live in `~/.remoteok-cli/config.yaml`
(or `$XDG_CONFIG_HOME/remoteok/config.yaml`):

```sh
remoteok config path
remoteok config view
remoteok config set user_agent "Mozilla/5.0 (X11; Linux x86_64) …"
remoteok config set base_url https://remoteok.com
```

Precedence is flag > env (`REMOTEOK_BASE_URL`, `REMOTEOK_USER_AGENT`) > config file > default.
A real browser **User-Agent** is required — Remote OK returns 403 for a default/bot UA — and
the CLI ships a working default.

## For AI agents

`remoteok` is agent-ready:

```sh
remoteok mcp start                       # run as an MCP server (tools: remoteok_jobs_list, remoteok_jobs_get)
remoteok agent guard --host claude-code  # emit safety rails for an agent driving the CLI
```

Every command carries MCP read/write annotations. Because `remoteok` is read-only, all job
tools are reads; the guard still blocks the raw `api` escape hatch on write methods and
`alias set`.

## Meta commands

`jobs`, `config`, `init`, `doctor`, `completion`, `alias`, `api`, `version`, `update`, `mcp`,
`agent`. Run `remoteok <command> --help` for details, or see the
[command reference](docs/commands/remoteok.md).

## Contributing & build

See [AGENTS.md](AGENTS.md). The gate is `make verify` (fmt, vet, lint, tests, spec-check,
coverage ≥80%, DoD). `make build` produces `bin/remoteok`.

## License

MIT — see [LICENSE](LICENSE). Job data is © Remote OK; see the attribution requirement above.
