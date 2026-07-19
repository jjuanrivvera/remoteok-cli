# Auth & configuration

## No auth
Remote OK is an unauthenticated public API. `remoteok` stores and sends **no credentials** —
there is no `auth` command, no keyring, and no secret in the config file.

## User-Agent (required)
Remote OK (behind Cloudflare) returns **403** for a missing/default/bot User-Agent. The CLI
ships a working browser UA by default. Override if your network needs a specific one:
```sh
remoteok config set user_agent "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 …"
# or per-invocation:
remoteok jobs list --user-agent "Mozilla/5.0 …"
# or via env:
REMOTEOK_USER_AGENT="Mozilla/5.0 …" remoteok jobs list
```

## Config file
`~/.remoteok-cli/config.yaml` (or `$XDG_CONFIG_HOME/remoteok/config.yaml`). Non-secret only:
`base_url`, `user_agent`, and user aliases.
```sh
remoteok config path
remoteok config view
remoteok config set base_url https://remoteok.com
remoteok alias set go "jobs list --tag golang"
```

Precedence: flag > env (`REMOTEOK_BASE_URL`, `REMOTEOK_USER_AGENT`) > config file > default.
