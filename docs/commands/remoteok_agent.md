## remoteok agent

AI-agent integration helpers

### Synopsis

Generate safety configuration for AI agents that drive remoteok.

### Options

```
  -h, --help   help for agent
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

* [remoteok](remoteok.md)	 - Discover remote jobs from the Remote OK public API
* [remoteok agent guard](remoteok_agent_guard.md)	 - Generate agent-safety config for AI agents that drive remoteok

