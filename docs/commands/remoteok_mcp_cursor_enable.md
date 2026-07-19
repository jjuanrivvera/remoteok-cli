## remoteok mcp cursor enable

Add server to Cursor config

### Synopsis

Add this application as an MCP server in Cursor

```
remoteok mcp cursor enable [flags]
```

### Options

```
      --config-path string   Path to Cursor config file
  -e, --env stringToString   Environment variables (e.g., --env KEY1=value1 --env KEY2=value2) (default [])
  -h, --help                 help for enable
      --log-level string     Log level (debug, info, warn, error)
      --server-name string   Name for the MCP server (default: derived from executable name)
      --workspace            Add to workspace settings (.cursor/mcp.json) instead of user settings
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

* [remoteok mcp cursor](remoteok_mcp_cursor.md)	 - Manage Cursor MCP servers

