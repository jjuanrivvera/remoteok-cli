## remoteok mcp

MCP server management

### Synopsis

Manage MCP servers for AI assistants and code editors

### Options

```
  -h, --help   help for mcp
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
* [remoteok mcp claude](remoteok_mcp_claude.md)	 - Manage Claude Desktop MCP servers
* [remoteok mcp cursor](remoteok_mcp_cursor.md)	 - Manage Cursor MCP servers
* [remoteok mcp start](remoteok_mcp_start.md)	 - Start the MCP server
* [remoteok mcp stream](remoteok_mcp_stream.md)	 - Stream the MCP server over HTTP
* [remoteok mcp tools](remoteok_mcp_tools.md)	 - Export tools as JSON
* [remoteok mcp vscode](remoteok_mcp_vscode.md)	 - Manage VSCode MCP servers

