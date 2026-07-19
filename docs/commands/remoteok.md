## remoteok

Discover remote jobs from the Remote OK public API

### Synopsis

remoteok is a fast, scriptable, read-only client for the Remote OK jobs API
(https://remoteok.com/api). It is built for machine consumption — JSON/YAML/CSV
output, an -o id mode for piping, a --jq filter, and an MCP server — so an AI
assistant or a shell pipeline can discover remote work programmatically.

Remote OK is a free public API and needs no account or token. Its Terms of Service
require a follow backlink to https://remoteok.com when you display their data — the
CLI prints a Source attribution on stderr for that reason.

Examples:
  remoteok jobs list --tag golang --limit 20
  remoteok jobs list --search kubernetes -o json
  remoteok jobs list --tag golang --tag remote --min-salary 100000
  remoteok jobs list --company stripe -o csv
  remoteok jobs get 1135010
  remoteok jobs list -o id | head

### Options

```
      --base-url string     Remote OK base URL override (default https://remoteok.com)
      --columns strings     comma-separated columns to show
      --dry-run             print the equivalent curl and make no request
  -h, --help                help for remoteok
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
* [remoteok alias](remoteok_alias.md)	 - Manage user-defined command aliases
* [remoteok api](remoteok_api.md)	 - Send a raw request to the Remote OK API (escape hatch)
* [remoteok completion](remoteok_completion.md)	 - Generate shell completion scripts
* [remoteok config](remoteok_config.md)	 - Inspect and edit remoteok configuration
* [remoteok doctor](remoteok_doctor.md)	 - Diagnose configuration and Remote OK connectivity
* [remoteok init](remoteok_init.md)	 - First-run setup: write a config file and smoke-test the feed
* [remoteok jobs](remoteok_jobs.md)	 - Browse remote job listings
* [remoteok mcp](remoteok_mcp.md)	 - MCP server management
* [remoteok update](remoteok_update.md)	 - Update remoteok to the latest GitHub release
* [remoteok version](remoteok_version.md)	 - Print version, commit, and build date

