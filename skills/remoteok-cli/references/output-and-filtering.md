# Output & filtering

## Formats (`-o`)
- `table` (default; colored on a TTY, honors `NO_COLOR`/`--no-color`)
- `json` — full records, pipe-clean on stdout
- `yaml`
- `csv` — spreadsheet-injection-safe (leading `= + @ -` neutralized)
- `id` — one id per line, ideal for `xargs`/`head`

## Column control
- `--columns id,position,company` selects table/CSV columns (deterministic order otherwise:
  id, position, company, tags, location, salary_min, salary_max, date, url).
- Wide cells truncate in the table with `…`; use `-o json` for full values.

## `--jq`
A gojq expression runs over the response before rendering:
```sh
remoteok jobs list --tag golang --jq '.[] | {position, company, salary_max}'
remoteok jobs list -o json --jq '[.[].company] | unique'
```

## Filters (all client-side, AND-ed)
| Flag | Field |
|---|---|
| `--tag`/`--tags` | `tags[]` — every listed tag must be present (case-insensitive) |
| `--search` | position, company, description, tags (substring) |
| `--company` | company (substring) |
| `--min-salary` | `salary_max` ≥ amount |
| `--limit` | result cap |

A single `--tag` is also sent to the server as `?tags=` to reduce payload; correctness never
depends on the server filter.
