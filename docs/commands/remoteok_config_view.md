## remoteok config view

Show the resolved configuration

### Synopsis

Print the config as YAML. No secrets are stored in the config, so nothing needs redacting.

```
remoteok config view [flags]
```

### Options

```
  -h, --help   help for view
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

