## remoteok jobs get

Show one job by id

### Synopsis

Fetch a single listing by its Remote OK id. Remote OK exposes no per-id endpoint,
so the CLI fetches the current feed and selects the matching job locally — an id
that has aged out of the feed will not be found.

```
remoteok jobs get <id> [flags]
```

### Examples

```
  remoteok jobs get 1135010
  remoteok jobs list -o id | head -1 | xargs remoteok jobs get -o json
```

### Options

```
  -h, --help   help for get
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

* [remoteok jobs](remoteok_jobs.md)	 - Browse remote job listings

