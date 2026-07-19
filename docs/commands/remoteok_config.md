## remoteok config

Inspect and edit remoteok configuration

### Synopsis

The config file holds only non-secret settings (base URL, User-Agent, aliases).
Remote OK is an unauthenticated public API, so nothing here is ever a secret.

### Options

```
  -h, --help   help for config
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
* [remoteok config path](remoteok_config_path.md)	 - Print the config file path
* [remoteok config set](remoteok_config_set.md)	 - Set a non-secret option (base_url, user_agent)
* [remoteok config view](remoteok_config_view.md)	 - Show the resolved configuration

