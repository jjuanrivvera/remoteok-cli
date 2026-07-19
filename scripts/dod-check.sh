#!/usr/bin/env bash
# dod-check.sh — deterministic Definition-of-Done checks (cliwright GOAL.md §9/§12).
# One concrete check per atomic criterion. Tailored for remoteok: Remote OK is an
# UNAUTHENTICATED, read-only public API, so there is no auth/keyring/secret surface to check
# (see DECISIONS.md). Usage: ./scripts/dod-check.sh <binary-name>
set -uo pipefail
BIN="${1:-remoteok}"
fail=0

ok()   { printf "  ✓ %s\n" "$1"; }
bad()  { printf "  ✗ %s\n" "$1"; fail=1; }
# have <description> <test-command...>
have() { if eval "${@:2}" >/dev/null 2>&1; then ok "$1"; else bad "$1"; fi; }

echo "Definition-of-Done checks for '$BIN':"

# Agent surface
have "mcp server command present"        "rg -lq 'ophis|mcp' commands/mcp.go"
have "agent guard command present"       "test -f commands/agent.go"

# Output formats (atomic — one per format)
for f in json yaml csv table; do
  have "output format: $f"               "rg -liq '\"$f\"|format$f|$f *format' internal/output"
done

# Resilience & safety
have "--dry-run prints equivalent curl"  "rg -lq 'dry-run' . && rg -lq 'curl' internal/api"
have "Ctrl-C: signal.NotifyContext"      "rg -lq 'signal.NotifyContext' cmd"
have "no stray context.Background()"     "! rg -lq 'context.Background()' commands internal/api"
have "browser User-Agent pinned"         "rg -lq 'DefaultUserAgent' internal/api"
have "attribution/backlink surfaced"     "rg -liq 'attribution|link back|remoteok.com' commands internal/api"
have "no fmt.Scan reads"                 "! rg -q 'fmt\\.Scan(ln|f)?\\(' commands internal"
have "idempotent-only retry"             "rg -lq 'idempotent|MethodGet|MethodPut|MethodDelete' internal/api"
have "flexible JSON types + fuzz"        "rg -lq 'func Fuzz' internal/api"

# Meta commands (atomic — one per command). Remote OK is auth-free, so no auth command.
for c in config init doctor completion alias api version update; do
  have "meta command: $c"                "test -f commands/$c.go || rg -lq '\"$c\"' commands"
done

# Distribution & CI
have "GoReleaser config present"         "test -f .goreleaser.yaml || test -f .goreleaser.yml"
have "goreleaser check clean"            "! command -v goreleaser >/dev/null || goreleaser check"
have "install script present"            "test -f install.sh"
have "CI workflow present"               "test -f .github/workflows/ci.yml"
have "release workflow present"          "test -f .github/workflows/release.yml"

# Hygiene
for doc in README.md LICENSE CHANGELOG.md SECURITY.md AGENTS.md DECISIONS.md; do
  have "doc: $doc"                       "test -f $doc"
done
have "attribution requirement in README" "rg -liq 'remote ?ok|backlink|attribution' README.md"

if [[ $fail -ne 0 ]]; then
  echo "✗ Definition-of-Done incomplete"; exit 1
fi
echo "✓ Definition-of-Done satisfied"
