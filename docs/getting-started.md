# Getting started

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/remoteok-cli/main/install.sh | sh
```

Or `brew install jjuanrivvera/remoteok-cli/remoteok-cli`, or
`go install github.com/jjuanrivvera/remoteok-cli/cmd/remoteok@latest`.

## First run

```sh
remoteok init                 # writes a config file and smoke-tests the feed
remoteok jobs list --tag golang --limit 20
remoteok jobs list --search kubernetes -o json
remoteok jobs get 1135010 -o json
```

## Filtering

All filters are client-side over the live feed and compose freely: `--tag`/`--tags` (AND),
`--search`, `--company`, `--min-salary`, `--limit`. Combine with `-o json|yaml|csv|id`,
`--columns`, and `--jq` for scripting.

## Configuration

No secrets are stored. Optional overrides:

```sh
remoteok config set user_agent "Mozilla/5.0 …"   # Remote OK 403s a default/bot UA
remoteok config view
```
