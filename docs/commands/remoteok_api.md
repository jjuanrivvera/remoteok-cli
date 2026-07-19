## remoteok api

Send a raw request to the Remote OK API (escape hatch)

### Synopsis

Call the Remote OK site directly. The path is relative to the configured base URL
(default https://remoteok.com), e.g. api.

This is the documented escape hatch for anything remoteok does not wrap as a
first-class command. It honors --dry-run, -o/--output, and --jq like every other
command. The public API is read-only; in practice only GET /api serves data.

```
remoteok api <METHOD> <PATH> [-d body] [-q key=value ...] [flags]
```

### Examples

```
  remoteok api GET api
  remoteok api GET api -q tags=golang
  remoteok api GET api --dry-run
```

### Options

```
  -d, --data string         request body: inline, @file, or - for stdin
  -h, --help                help for api
  -q, --query stringArray   query parameter key=value (repeatable)
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

