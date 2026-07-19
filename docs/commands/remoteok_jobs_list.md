## remoteok jobs list

List recent remote jobs (newest first)

### Synopsis

List recent remote jobs from Remote OK, newest first. All filters are applied
client-side over the live feed, so they compose freely:

  --tag/--tags   keep only jobs carrying EVERY requested tag (AND), case-insensitive
  --search       case-insensitive keyword over position, company, description, and tags
  --company      case-insensitive substring over the company name
  --min-salary   keep jobs whose advertised salary_max is at least this amount
  --since        keep jobs posted on/after a date (YYYY-MM-DD) or window (Nd/Nw)
  --limit        cap the number of results

Remote OK's Terms require a follow backlink to https://remoteok.com when you display
their data; a Source attribution is printed on stderr (suppress with --quiet).

```
remoteok jobs list [flags]
```

### Examples

```
  remoteok jobs list --tag golang --limit 20
  remoteok jobs list --tags golang,remote --min-salary 100000
  remoteok jobs list --search kubernetes -o json
  remoteok jobs list --company stripe -o csv
  remoteok jobs list --since 7d --tag golang
  remoteok jobs list --since 2026-07-12 -o json
  remoteok jobs list -o id | head
```

### Options

```
      --company string        filter by company name (substring)
  -h, --help                  help for list
      --min-salary int        keep jobs with salary_max ≥ this amount
      --posted-after string   alias for --since
      --search string         keyword over position/company/description/tags
      --since string          keep jobs posted on/after a date (YYYY-MM-DD) or window (Nd/Nw, e.g. 7d, 2w)
      --tag strings           require this tag (repeatable); alias --tags
      --tags strings          comma-separated tags to require (AND)
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

* [remoteok jobs](remoteok_jobs.md)	 - Browse remote job listings

