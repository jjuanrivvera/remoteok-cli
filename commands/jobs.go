package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/remoteok-cli/internal/api"
)

// jobListColumns are the default table columns for job listings — a scannable summary; use
// -o json for the full record (description, urls, logo, epoch).
var jobListColumns = []string{"id", "position", "company", "location", "tags", "date"}

func init() {
	registrars = append(registrars, func(d *deps) *cobra.Command {
		jobsCmd := &cobra.Command{
			Use:     "jobs",
			Aliases: []string{"job"},
			Short:   "Browse remote job listings",
			Long:    "List and inspect remote jobs from the Remote OK public feed. Read-only — no account needed.",
		}
		jobsCmd.AddCommand(newJobsListCmd(d), newJobsGetCmd(d))
		return jobsCmd
	})
}

func newJobsListCmd(d *deps) *cobra.Command {
	var opts api.JobListOptions
	var sinceRaw string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent remote jobs (newest first)",
		Long: `List recent remote jobs from Remote OK, newest first. All filters are applied
client-side over the live feed, so they compose freely:

  --tag/--tags   keep only jobs carrying EVERY requested tag (AND), case-insensitive
  --search       case-insensitive keyword over position, company, description, and tags
  --company      case-insensitive substring over the company name
  --min-salary   keep jobs whose advertised salary_max is at least this amount
  --since        keep jobs posted on/after a date (YYYY-MM-DD) or window (Nd/Nw)
  --limit        cap the number of results

--tag matches Remote OK's fixed tag vocabulary (e.g. golang, react, devops, remote). A term
that is NOT a real tag — an industry like fintech, a role, or a free keyword — matches nothing
via --tag; use --search for those instead.

Remote OK rarely publishes salary (only a few listings per feed carry one), so --min-salary
drops every listing without a published minimum — i.e. most of them. Use it to narrow a broad
query, not as a primary filter; when it excludes listings for lack of a published salary the
CLI notes how many on stderr (suppress with --quiet).

Remote OK's Terms require a follow backlink to https://remoteok.com when you display
their data; a Source attribution is printed on stderr (suppress with --quiet).`,
		Example: `  remoteok jobs list --tag golang --limit 20
  remoteok jobs list --tags golang,remote --min-salary 100000
  remoteok jobs list --search kubernetes -o json
  remoteok jobs list --company stripe -o csv
  remoteok jobs list --since 7d --tag golang
  remoteok jobs list --since 2026-07-12 -o json
  remoteok jobs list -o id | head`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, _, err := d.getAPIClient()
			if err != nil {
				return err
			}
			if opts.Since, err = parseSince(sinceRaw); err != nil {
				return err
			}
			opts.Limit = d.gf.limit
			jobs, legal, noSalaryExcluded, err := c.Jobs(cmd.Context(), opts)
			if err != nil {
				return err
			}
			if d.gf.dryRun {
				return nil
			}
			if err := d.render(cmd, rawJobs(jobs), jobListColumns); err != nil {
				return err
			}
			printAttribution(cmd, d, legal)
			printMinSalaryHint(cmd, d, noSalaryExcluded)
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&opts.Tags, "tag", nil, "require this Remote OK tag (repeatable; alias --tags); use --search for non-tag terms")
	cmd.Flags().StringSliceVar(&opts.Tags, "tags", nil, "comma-separated tags to require (AND)")
	cmd.Flags().StringVar(&opts.Search, "search", "", "keyword over position/company/description/tags")
	cmd.Flags().StringVar(&opts.Company, "company", "", "filter by company name (substring)")
	cmd.Flags().Int64Var(&opts.MinSalary, "min-salary", 0, "keep jobs with salary_max ≥ this; Remote OK rarely publishes salary, so this drops most listings — pair it with a broad query")
	cmd.Flags().StringVar(&sinceRaw, "since", "", "keep jobs posted on/after a date (YYYY-MM-DD) or window (Nd/Nw, e.g. 7d, 2w)")
	cmd.Flags().StringVar(&sinceRaw, "posted-after", "", "alias for --since")
	return annotate(cmd, kindRead)
}

// relSinceRe matches the relative --since shorthand: N days (Nd) or N weeks (Nw).
var relSinceRe = regexp.MustCompile(`^([0-9]+)([dw])$`)

// parseSince turns a --since value into an absolute lower-bound instant, relative to now.
// It accepts an absolute date (YYYY-MM-DD, interpreted as that day's 00:00 UTC) or a
// relative window Nd/Nw (posted within the last N days/weeks). An empty value means no
// filter. Errors are actionable so a mistyped value tells the user the exact grammar.
func parseSince(v string) (time.Time, error) { return parseSinceAt(v, time.Now()) }

func parseSinceAt(v string, now time.Time) (time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}, nil
	}
	if m := relSinceRe.FindStringSubmatch(v); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil { // unreachable: the regex guarantees digits, but never trust silently
			return time.Time{}, fmt.Errorf("invalid --since %q: %w", v, err)
		}
		days := n
		if m[2] == "w" {
			days = n * 7
		}
		return now.AddDate(0, 0, -days), nil
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf(
		"invalid --since %q: want an absolute date YYYY-MM-DD (e.g. 2026-07-12) or a relative window Nd/Nw (e.g. 7d, 2w)", v)
}

func newJobsGetCmd(d *deps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Show one job by id",
		Long: `Fetch a single listing by its Remote OK id. Remote OK exposes no per-id endpoint,
so the CLI fetches the current feed and selects the matching job locally — an id
that has aged out of the feed will not be found.`,
		Example: `  remoteok jobs get 1135010
  remoteok jobs list -o id | head -1 | xargs remoteok jobs get -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := d.getAPIClient()
			if err != nil {
				return err
			}
			job, legal, err := c.GetJob(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if d.gf.dryRun {
				return nil
			}
			if err := d.render(cmd, job, jobListColumns); err != nil {
				return err
			}
			printAttribution(cmd, d, legal)
			return nil
		},
	}
	return annotate(cmd, kindRead)
}

// printAttribution honors Remote OK's Terms of Service: a follow backlink to remoteok.com
// must accompany displayed data. It goes to stderr so stdout stays pipe-clean, and is
// suppressed by --quiet (and in machine formats the user is expected to attribute in their
// own UI). Nothing is printed when the API returned no notice.
func printAttribution(cmd *cobra.Command, d *deps, legal *api.Legal) {
	if d.gf.quiet || legal == nil {
		return
	}
	fmt.Fprintln(cmd.ErrOrStderr(), "Source: Remote OK — https://remoteok.com (please keep a follow backlink when displaying these jobs)")
}

// printMinSalaryHint explains a surprisingly small --min-salary result: Remote OK rarely
// publishes salary, so the filter silently discards every listing with no published minimum.
// It fires only when --min-salary actually excluded one or more no-salary listings
// (noSalaryExcluded is already 0 when --min-salary is unset — see countNoSalaryExcluded), so
// it never appears for an ordinary query. Same stderr channel and --quiet suppression as the
// attribution line, keeping stdout pipe-clean.
func printMinSalaryHint(cmd *cobra.Command, d *deps, noSalaryExcluded int) {
	if d.gf.quiet || noSalaryExcluded == 0 {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(),
		"Note: --min-salary excluded %d listing(s) with no published salary (Remote OK rarely publishes salary).\n",
		noSalaryExcluded)
}

// rawJobs normalizes a job slice into one JSON array for the renderer. A nil slice renders
// as an empty array rather than null so `-o json` output is always a list.
func rawJobs(jobs []api.Job) json.RawMessage {
	if jobs == nil {
		return json.RawMessage("[]")
	}
	b, err := json.Marshal(jobs)
	if err != nil {
		return json.RawMessage("[]")
	}
	return b
}
