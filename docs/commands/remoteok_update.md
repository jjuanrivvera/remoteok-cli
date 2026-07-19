## remoteok update

Update remoteok to the latest GitHub release

### Synopsis

Download the latest remoteok release, verify it against checksums.txt, and replace
the running binary in place. Use 'remoteok update check' to see what's available
without installing.

```
remoteok update [flags]
```

### Examples

```
  remoteok update
  remoteok update check
```

### Options

```
  -h, --help   help for update
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
* [remoteok update check](remoteok_update_check.md)	 - Check for a newer release without installing it

