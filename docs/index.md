# remoteok

A fast, scriptable, read-only client for the [Remote OK](https://remoteok.com) jobs API.
Built for machines — JSON/YAML/CSV output, `-o id` for piping, `--jq`, and an MCP server.

Remote OK is a free public API and needs no account or token.

!!! warning "Attribution requirement"
    Remote OK's [Terms of Service](https://remoteok.com/api) require a **follow backlink** to
    <https://remoteok.com> and a mention of *Remote OK* as the source whenever you display
    their data. The CLI prints a `Source: Remote OK` line on stderr as a reminder; in machine
    formats the backlink is your responsibility.

See [Getting started](getting-started.md) or the
[command reference](commands/remoteok.md).
