# DECISIONS.md — remoteok

Pinned, ambiguous, or non-obvious decisions for the Remote OK CLI. Each is a
question → decision → why, read back on every iteration so the build stays
deterministic (cliwright GOAL.md §11). Never silently re-decide.

## API shape & quirks (verified live 2026-07-18)

1. **Base URL / endpoint.** The Remote OK public API is a single endpoint:
   `GET https://remoteok.com/api`. It returns a JSON array. There is no OpenAPI /
   Postman / llms.txt. Base URL is the site root (`https://remoteok.com`); the JSON
   lives at `/api`. `api_method_total: 1` in the manifest reflects this — it is the
   complete documented surface.

2. **The first array element is a legal/attribution notice, not a job.** Confirmed
   live: element `[0]` is `{"last_updated":<epoch>,"legal":"API Terms of Service:
   Please link back (with follow, and without nofollow!) to … Remote OK …"}`. The
   parser skips any element that has a `legal` key AND no job `id` (`isLegal`), so
   detection is robust even if the notice ever stops being strictly first. The notice
   is surfaced to the caller (`Legal`) and the CLI prints a Source attribution line.

3. **Attribution / backlink is REQUIRED by Remote OK's Terms.** Their notice requires a
   *follow* backlink to remoteok.com when you display their data ("If you do not we'll
   have to suspend API access"). Decision: `jobs list`/`jobs get` print
   `Source: Remote OK — https://remoteok.com …` to **stderr** (keeps stdout pipe-clean),
   suppressible with `--quiet`. Documented prominently in the README. When emitting
   machine formats the consumer is responsible for the backlink in their own UI.

4. **User-Agent is REQUIRED (403 otherwise).** Remote OK (behind Cloudflare) is
   documented to reject a missing/default/bot-like User-Agent with 403. The client
   pins a real browser UA (`api.DefaultUserAgent`, a Chrome string) on every request;
   overridable via `--user-agent`, `REMOTEOK_USER_AGENT`, or `config set user_agent`.
   *Note:* on the build network the default Go UA did not 403 (Cloudflare behavior is
   geo/edge-variable), but the browser UA is pinned defensively per the documented
   requirement — cheap and harmless, and the 403 hint tells the user how to fix a UA
   block if their network hits one.

5. **Tag filtering mechanism.** Server-side `?tags=<tag>` works (returns 200 with jobs
   carrying that tag) but its relevance is loose. Decision: **client-side filtering is
   authoritative** — `--tag/--tags` (AND across all requested tags, case-insensitive),
   `--search`, `--company`, `--min-salary` are all applied over the full feed so
   results are deterministic and testable. A single `--tag` is *also* sent as `?tags=`
   purely as a payload optimization; correctness never depends on the server filter.

6. **`jobs get <id>` selects over the feed.** Remote OK exposes no per-id endpoint, so
   `get` fetches the feed and matches locally. An id that has aged out of the current
   feed returns `ErrJobNotFound`. Documented in the command help.

7. **Field schema (confirmed live).** `id, slug, position, company, company_logo,
   tags[], description, location, salary_min, salary_max, date, epoch, url, apply_url,
   logo`. `id` arrives as string OR number and `salary_*` as number OR quoted string,
   so `ID`/`Int` flexible types (fuzz-tested) absorb the variance. Text fields keep
   HTML entities (`&amp;`) verbatim so json/yaml/csv stay byte-faithful; the human
   table view passes through the terminal-escape sanitizer.

8. **Sort order.** Newest-first by `epoch` (stable sort preserves feed order on ties).

## Architecture decisions (determinism, §11)

9. **Resource pattern — neither pure A nor B.** Remote OK is a *single read-only,
   non-CRUD endpoint*, so the generic-core CRUD builder (`Resource[T]` with
   list/get/create/update/delete against REST ids) does not apply, and Pattern B's
   per-resource service objects would be one trivial service. Decision: a thin bespoke
   `jobs` command over a shared `Client` + the shared `output` renderer. This keeps the
   §1 standard (one renderer, flexible types, retry, dry-run, MCP annotations) without
   forcing a CRUD abstraction onto an endpoint that has no writes and no per-id GET.

10. **No auth / no keyring / no profiles.** Remote OK is unauthenticated and a single
    fixed public instance. Per GOAL §2 ("single fixed instance → no profiles";
    "auth only when the API supports it") the CLI ships **no** `auth` command, no
    keyring, and no profile selector. `config` holds only non-secret overrides
    (`base_url`, `user_agent`) and user aliases. `dod-check.sh` is tailored accordingly
    (no auth/keyring/secret checks). This is the honest scale-down of the standard, not
    a gap.

11. **Meta command set.** Shipped: `jobs`, `config` (path/view/set), `init`, `doctor`,
    `completion`, `alias`, `api`, `version`, `update`, `mcp`, `agent`. Omitted with
    cause: `auth` (no credentials, #10); `config use`/`list-profiles` (no profiles).

## Conditional-patterns catalog (§3d) — decisions

- **Event-store / offline-cache:** N/A. Remote OK is re-fetchable and not an event
  stream or time-series; a `--limit` over the live feed is sufficient.
- **Spec-contract test / spec-sync / smoke.yml:** N/A. No machine spec exists and there
  are no durable credentials (public API); a scheduled live smoke would add little over
  `doctor`.
- **Universal write flags / multi-group creds / adopt-a-library:** N/A (read-only,
  single unauth endpoint, no mature typed Go client for Remote OK).
- **Terminal-escape sanitization:** YES — job text is free-form API data; the shared
  renderer strips ANSI/OSC in the human table + error path (json/yaml/csv stay faithful).
- **Binary self-update:** YES — distributed by direct download; `update`/`update check`
  self-replace from GitHub releases (checksum-verified).

## Completeness

- `spec-completeness`: the manifest wraps the single enumerated endpoint via the `jobs`
  resource (list + get). Coverage is 100% of the enumerated public surface — **no
  coverage-waiver needed**.
