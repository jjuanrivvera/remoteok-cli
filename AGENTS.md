# AGENTS.md — working in the remoteok-cli repo

`remoteok` is a read-only command-line client for the **Remote OK jobs API**
(`https://remoteok.com/api`), built to the cliwright standard (Go + Cobra + GoReleaser).
It exists so a life-assistant / AI agent (or a shell pipeline) can discover remote jobs
programmatically — machine output (JSON/YAML/CSV, `-o id`, `--jq`, an MCP server) is the
priority over pretty tables. This file orients an AI agent (or human) contributing.

## The one rule that matters
**`make verify` is the gate.** A change is done only when `make verify` exits `0`. It runs
`make check` (fmt, vet, golangci-lint, tests) + `spec-check` (built surface ⊆
`api-manifest.json`) + `spec-completeness` (manifest wraps the single enumerated endpoint)
+ `cover-check` (≥80% coverage) + `dod-check.sh`. Run the full `make verify` for any change
that touches the command surface or a documented behavior — not just `make check`.

## Architecture (where things live)
- `internal/api/` — the Remote OK client core: a **browser-User-Agent** HTTP client
  (`DefaultUserAgent`; Remote OK 403s a default/bot UA), idempotent-only retry honoring
  `Retry-After` with full-jitter backoff, dry-run curl, `APIError` with actionable hints,
  flexible JSON types (`ID`/`Int`, fuzz-tested), and the `jobs` resource (`Jobs`,
  `GetJob`). The feed's leading legal/attribution element is split out (`isLegal`) and
  surfaced as `Legal`.
- `commands/` — the cobra tree. `init()` appends builders to `registrars`/`metaRegistrars`;
  `NewRootCmd()` drains the queue onto a fresh root (no mutable global root). MCP
  annotations are stamped via `annotate(cmd, kind)` as commands are built.
- `internal/{config,output,version,update}` — non-secret config + manual precedence (no
  Viper), the table/json/yaml/csv/id renderer (CSV formula-injection guard, terminal-escape
  sanitizer, NO_COLOR), build metadata, the checksum-verified self-updater.
- `cmd/remoteok/main.go` — `signal.NotifyContext` (Ctrl-C cancels in-flight work) + alias
  expansion before cobra parses; `run()` is split out so it is testable.

## Remote OK specifics you must not re-derive (see DECISIONS.md)
- **No auth.** Public API — no token, no keyring, no profiles. Do not add an `auth`
  command or secret storage.
- **First feed element is a legal/attribution notice**, skipped in parsing. Remote OK's
  Terms require a **follow backlink** to remoteok.com when displaying their data — the CLI
  prints a Source line on **stderr** (suppress with `--quiet`). Keep it.
- **Browser User-Agent is required** — never send Go's default UA. Pinned in
  `api.DefaultUserAgent`, overridable via `--user-agent` / `REMOTEOK_USER_AGENT` /
  `config set user_agent`.
- **Filtering is client-side and authoritative** (`--tag/--tags` AND, `--search`,
  `--company`, `--min-salary`); a single `--tag` is also sent as `?tags=` for payload size.
- **`jobs get <id>`** selects over the feed (no per-id endpoint); an aged-out id is not found.
- Text fields keep HTML entities verbatim so json/yaml/csv stay byte-faithful.

## House rules
- Comments explain **WHY**, not WHAT.
- Thread `cmd.Context()` everywhere; never `context.Background()` (it breaks Ctrl-C).
- Pin every ambiguous API assumption in `DECISIONS.md`; read it back, never re-decide.
- Surface changes require updating `api-manifest.json` AND regenerating docs
  (`make docs-gen`) in the same commit.
- MCP exclusions are by EXACT path (`commands/mcp.go`), never substring.
- New commands ship with tests in the same commit — coverage is a ratchet (≥80%).
