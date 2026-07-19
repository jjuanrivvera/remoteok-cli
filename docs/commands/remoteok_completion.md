## remoteok completion

Generate shell completion scripts

```
remoteok completion [bash|zsh|fish|powershell]
```

### Examples

```
  remoteok completion zsh > "${fpath[1]}/_remoteok"
  remoteok completion bash > /etc/bash_completion.d/remoteok
  remoteok completion fish > ~/.config/fish/completions/remoteok.fish
```

### Options

```
  -h, --help   help for completion
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

