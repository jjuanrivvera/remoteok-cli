## remoteok mcp cursor

Manage Cursor MCP servers

### Synopsis

Manage MCP server configuration for Cursor

### Options

```
  -h, --help   help for cursor
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

* [remoteok mcp](remoteok_mcp.md)	 - MCP server management
* [remoteok mcp cursor disable](remoteok_mcp_cursor_disable.md)	 - Remove server from Cursor config
* [remoteok mcp cursor enable](remoteok_mcp_cursor_enable.md)	 - Add server to Cursor config
* [remoteok mcp cursor list](remoteok_mcp_cursor_list.md)	 - Show Cursor MCP servers

