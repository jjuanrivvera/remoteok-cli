## remoteok agent guard

Generate agent-safety config for AI agents that drive remoteok

### Synopsis

Classify every API command (read / write / irreversible) from the live command tree
and emit host safety config. remoteok is a read-only client, so every job command is
a read and runs freely — the rails still hard-block the two ways an agent could reach
past the read surface: the raw "remoteok api" escape hatch with a write METHOD, and
"remoteok alias set" (which could mint a shorthand for a future blocked command). The
guard derives from the LIVE tree, so if a write/destructive command is ever added it is
covered automatically, including its cobra alias paths.

For claude-code the output also includes a PreToolUse hook script
(.claude/hooks/remoteok-guard.sh): it strips quote/backslash obfuscation, matches blocked
subcommand paths at the command position even for path-invoked binaries (./bin/remoteok,
/usr/local/bin/remoteok), and gates the raw "remoteok api <METHOD> <PATH>" escape hatch at
the METHOD position — only GET/HEAD/OPTIONS pass; POST/PUT/PATCH/DELETE are denied
case-insensitively, while a GET whose path merely contains "delete" stays allowed.
"remoteok alias set" is denied so an agent cannot mint a new shorthand for a blocked
command.

MCP-only operation is the hard guarantee; the Bash rails are best-effort — the hook
defeats quoting tricks and path prefixes, but not variable indirection
(a=DELETE; remoteok api $a x) or shell aliases. Conservative false positives are
accepted: a line that merely QUOTES a blocked command is denied.

```
remoteok agent guard --host <claude-code|codex|opencode> [flags]
```

### Examples

```
  remoteok agent guard --host claude-code
  remoteok agent guard --host claude-code --write          # write the files into .claude/
  remoteok agent guard --host codex --out ~/.codex/config.toml
  remoteok agent guard --host opencode --all-writes
```

### Options

```
      --all-writes    also hard-block ordinary writes, not just irreversible ops
  -h, --help          help for guard
      --host string   target agent host: claude-code|codex|opencode (required)
      --out string    write to this file instead of stdout
      --write         claude-code only: write hook + settings fragment under .claude/ (never overwrites)
```

### Options inherited from parent commands

```
      --base-url string     Remote OK base URL override (default https://remoteok.com)
      --columns strings     comma-separated columns to show
      --dry-run             print the equivalent curl and make no request
      --jq string           gojq expression applied to the response before rendering
      --limit int           max jobs to return (list commands)
      --no-color            disable colored output
  -o, --output string       output format: table|json|yaml|csv|id
      --quiet               suppress non-essential chatter (incl. the source attribution)
      --user-agent string   override the browser User-Agent (Remote OK 403s a default/bot UA)
  -v, --verbose             verbose request logging (stderr)
```

### SEE ALSO

* [remoteok agent](remoteok_agent.md)	 - AI-agent integration helpers

