## remoteok init

First-run setup: write a config file and smoke-test the feed

### Synopsis

Remote OK is a free public API — there is no account or token to set up. This
wizard writes a config file (creating the config directory) and runs a live
connectivity check so you know the feed is reachable from your network.

Pass --user-agent to pin a custom browser User-Agent in the config (Remote OK 403s
a default/bot UA; the CLI already ships a working default).

```
remoteok init [flags]
```

### Examples

```
  remoteok init
  remoteok init --user-agent "Mozilla/5.0 (X11; Linux x86_64) …"
```

### Options

```
  -h, --help   help for init
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

