# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.2] - 2026-07-19

### Added
- `jobs list` now prints a one-line hint on stderr when `--min-salary` excluded one or more
  listings that publish no salary (same channel and `--quiet` suppression as the attribution
  line): `Note: --min-salary excluded N listing(s) with no published salary (Remote OK rarely
  publishes salary).` — so a near-empty `--min-salary` result is not mistaken for "no matches".

### Documentation
- Documented that Remote OK rarely publishes salary, so `--min-salary` drops every listing
  without a published minimum (most of them) — pair it with a broad query, not as a primary
  filter (help text, README, SKILL cheatsheet, DECISIONS).
- Documented that `--tag` matches Remote OK's fixed tag vocabulary; for non-tag terms
  (industries like `fintech`, roles, keywords) use `--search` instead.

## [0.1.1] - 2026-07-19

### Added
- `jobs list --since` (alias `--posted-after`) — a posting-date filter. Accepts an absolute
  date `YYYY-MM-DD` or a relative window `Nd`/`Nw` (days/weeks, e.g. `7d`, `2w`) meaning
  "posted within the last N days/weeks". Keeps jobs whose `date` is on/after the threshold and
  composes (AND) with the existing client-side filters.

## [0.1.0] - 2026-07-18

### Added
- Initial release of `remoteok`, a read-only CLI for the Remote OK jobs API.
- `jobs list` — recent remote jobs (newest first) with client-side `--tag`/`--tags` (AND),
  `--search`, `--company`, `--min-salary`, and `--limit` filters; a single `--tag` is also
  sent server-side as `?tags=` for payload size.
- `jobs get <id>` — one job selected from the current feed (Remote OK has no per-id endpoint).
- Output formats: table, json, yaml, csv, and `-o id`; `--jq`, `--columns`, `--no-color`,
  `--quiet`, `--dry-run`.
- A pinned browser User-Agent (Remote OK 403s a default/bot UA) with `--user-agent` override.
- Remote OK attribution/backlink surfaced on stderr per their Terms of Service.
- Meta commands: `config`, `init`, `doctor`, `completion`, `alias`, `api`, `version`, `update`.
- MCP server (`mcp`) exposing the read-only job tools, and `agent guard` safety config.
