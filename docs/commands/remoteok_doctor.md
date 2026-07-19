## remoteok doctor

Diagnose configuration and Remote OK connectivity

### Synopsis

Run local and remote health checks: config file, resolved base URL and User-Agent,
and a live fetch of the Remote OK feed. Exits non-zero when any check fails, so it is
scriptable.

```
remoteok doctor [flags]
```

### Examples

```
  remoteok doctor
  remoteok doctor --json
```

### Options

```
  -h, --help   help for doctor
      --json   output as JSON
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

