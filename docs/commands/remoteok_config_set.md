## remoteok config set

Set a non-secret option (base_url, user_agent)

### Synopsis

Set a non-secret option in the config file.
Keys: base_url (API endpoint override), user_agent (browser UA sent on every request —
Remote OK 403s a default/bot UA, so override only with a real browser string).

```
remoteok config set <key> <value> [flags]
```

### Examples

```
  remoteok config set user_agent "Mozilla/5.0 (X11; Linux x86_64) …"
  remoteok config set base_url https://remoteok.com
```

### Options

```
  -h, --help   help for set
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

* [remoteok config](remoteok_config.md)	 - Inspect and edit remoteok configuration

