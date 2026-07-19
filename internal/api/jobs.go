package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Job is one Remote OK listing. Field names mirror the API's JSON keys exactly (see
// DECISIONS.md for the confirmed live schema). Values are kept verbatim from the API — HTML
// entities in text fields (e.g. "&amp;") are preserved so json/yaml/csv stay byte-faithful.
type Job struct {
	ID          ID       `json:"id,omitempty"`
	Slug        string   `json:"slug,omitempty"`
	Position    string   `json:"position,omitempty"`
	Company     string   `json:"company,omitempty"`
	CompanyLogo string   `json:"company_logo,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	Location    string   `json:"location,omitempty"`
	SalaryMin   Int      `json:"salary_min,omitempty"`
	SalaryMax   Int      `json:"salary_max,omitempty"`
	Date        string   `json:"date,omitempty"`
	Epoch       Int      `json:"epoch,omitempty"`
	URL         string   `json:"url,omitempty"`
	ApplyURL    string   `json:"apply_url,omitempty"`
	Logo        string   `json:"logo,omitempty"`
}

// Legal is the FIRST element of the Remote OK feed: an attribution/terms notice, not a job.
// Remote OK's Terms of Service require a follow backlink to remoteok.com when their data is
// displayed, so the CLI surfaces this (see the `jobs` commands and README).
type Legal struct {
	Legal       string `json:"legal,omitempty"`
	LastUpdated Int    `json:"last_updated,omitempty"`
}

// isLegal reports whether a raw feed element is the legal/attribution notice rather than a
// job. Detection is by the presence of a "legal" key AND the absence of a job id, so it is
// robust whether or not the notice stays the literal first element.
func isLegal(raw json.RawMessage) bool {
	var probe struct {
		Legal *string          `json:"legal"`
		ID    *json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return false
	}
	return probe.Legal != nil && probe.ID == nil
}

// JobListOptions filter and shape a `jobs list` request. All filtering is client-side over
// the full feed (see DECISIONS.md): Remote OK's server-side ?tags= filter returns loosely
// relevant results, so the CLI treats client-side filtering as authoritative and only uses
// ?tags= as a payload optimization when exactly one tag is requested.
type JobListOptions struct {
	Tags      []string  // ALL must be present on a job (AND), case-insensitive
	Search    string    // substring over position/company/description/tags, case-insensitive
	Company   string    // substring over company, case-insensitive
	MinSalary int64     // keep jobs whose salary_max ≥ this (0 = no filter)
	Since     time.Time // keep jobs posted on/after this instant (zero = no filter)
	Limit     int       // cap the number of returned jobs (0 = no cap)
}

// fetchFeed GETs /api and splits the response into the attribution notice and the jobs,
// sorted newest-first by epoch. serverTag, when non-empty, is sent as ?tags= to reduce the
// payload; results are still filtered client-side by the caller.
func (c *Client) fetchFeed(ctx context.Context, serverTag string) ([]Job, *Legal, error) {
	q := url.Values{}
	if serverTag != "" {
		q.Set("tags", serverTag)
	}
	var raw []json.RawMessage
	if err := c.GetJSON(ctx, apiPath, q, &raw); err != nil {
		return nil, nil, err
	}

	var jobs []Job
	var legal *Legal
	for _, el := range raw {
		if isLegal(el) {
			var l Legal
			if json.Unmarshal(el, &l) == nil {
				legal = &l
			}
			continue
		}
		var j Job
		if err := json.Unmarshal(el, &j); err != nil {
			// A single malformed element must not sink the whole feed.
			continue
		}
		jobs = append(jobs, j)
	}

	// Newest-first by epoch is deterministic; a stable sort keeps equal-epoch feed order.
	sort.SliceStable(jobs, func(i, j int) bool { return jobs[i].Epoch.Int64() > jobs[j].Epoch.Int64() })
	return jobs, legal, nil
}

// Jobs fetches recent listings and applies the client-side filters in opts. It also returns
// the Remote OK attribution notice so the caller can honor the backlink requirement, and the
// count of listings that --min-salary dropped only because they publish no salary (see
// countNoSalaryExcluded) so the CLI can explain a surprisingly small result set.
func (c *Client) Jobs(ctx context.Context, opts JobListOptions) ([]Job, *Legal, int, error) {
	serverTag := ""
	if len(opts.Tags) == 1 {
		serverTag = opts.Tags[0]
	}
	jobs, legal, err := c.fetchFeed(ctx, serverTag)
	if err != nil {
		return nil, legal, 0, err
	}

	// Count before filtering: the loop below reuses the jobs backing array (jobs[:0]).
	noSalaryExcluded := countNoSalaryExcluded(jobs, opts)

	filtered := jobs[:0]
	for _, j := range jobs {
		if !matches(j, opts) {
			continue
		}
		filtered = append(filtered, j)
	}
	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}
	return filtered, legal, noSalaryExcluded, nil
}

// countNoSalaryExcluded reports how many listings the --min-salary filter drops SOLELY
// because they publish no salary, as opposed to publishing one that fell below the bar.
// Remote OK rarely publishes salary (usually a handful of listings per feed), so a
// `--min-salary 80000` can silently discard almost the whole feed and look like "no matches".
// The CLI surfaces this count as an on-stderr hint. Only listings that would otherwise satisfy
// every OTHER filter are counted, and the result is 0 when --min-salary is unset — a job whose
// exclusion is a genuine below-bar drop (it published a salary_max under the bar) is NOT
// counted, keeping the hint honest.
func countNoSalaryExcluded(jobs []Job, opts JobListOptions) int {
	if opts.MinSalary <= 0 {
		return 0
	}
	rest := opts
	rest.MinSalary = 0 // isolate the "would pass if salary were ignored" set
	n := 0
	for _, j := range jobs {
		if j.SalaryMax.Int64() > 0 {
			continue // published a salary — any drop here is a real below-bar exclusion
		}
		if matches(j, rest) {
			n++
		}
	}
	return n
}

// ErrJobNotFound is returned by GetJob when no listing in the feed has the given id.
var ErrJobNotFound = errors.New("job not found in the current Remote OK feed")

// GetJob returns the single listing with the given id from the current feed. Remote OK has
// no per-id endpoint, so this fetches the feed and selects locally.
func (c *Client) GetJob(ctx context.Context, id string) (*Job, *Legal, error) {
	jobs, legal, err := c.fetchFeed(ctx, "")
	if err != nil {
		return nil, legal, err
	}
	// In dry-run the feed was never fetched (the curl was printed instead), so there is
	// nothing to select — return cleanly rather than a spurious "not found".
	if c.DryRun {
		return nil, legal, nil
	}
	for i := range jobs {
		if jobs[i].ID.String() == id {
			return &jobs[i], legal, nil
		}
	}
	return nil, legal, ErrJobNotFound
}

// matches reports whether a job satisfies every filter in opts.
func matches(j Job, opts JobListOptions) bool {
	for _, want := range opts.Tags {
		if !hasTag(j.Tags, want) {
			return false
		}
	}
	if opts.Company != "" && !strings.Contains(strings.ToLower(j.Company), strings.ToLower(opts.Company)) {
		return false
	}
	if opts.MinSalary > 0 && j.SalaryMax.Int64() < opts.MinSalary {
		return false
	}
	if opts.Search != "" && !matchesSearch(j, opts.Search) {
		return false
	}
	if !opts.Since.IsZero() {
		// A job with no determinable posting time cannot be proven recent, so a date
		// filter drops it rather than leaking a possibly-stale listing through.
		posted, ok := jobPostedAt(j)
		if !ok || posted.Before(opts.Since) {
			return false
		}
	}
	return true
}

// dateLayouts are the accepted forms of a listing's `date` field. Remote OK's live feed
// uses RFC3339 (DECISIONS.md §7), but a bare YYYY-MM-DD is tolerated so a date-only value
// still keys the `--since` filter instead of silently dropping to the epoch fallback.
var dateLayouts = []string{time.RFC3339, "2006-01-02"}

// jobPostedAt resolves a listing's posting instant. It prefers the `date` string (the field
// the `--since` filter is documented against) and falls back to the numeric `epoch` so a
// listing with a missing/garbled `date` is still datable. The bool is false when neither
// field yields a usable time.
func jobPostedAt(j Job) (time.Time, bool) {
	if j.Date != "" {
		for _, layout := range dateLayouts {
			if t, err := time.Parse(layout, j.Date); err == nil {
				return t, true
			}
		}
	}
	if e := j.Epoch.Int64(); e > 0 {
		return time.Unix(e, 0).UTC(), true
	}
	return time.Time{}, false
}

func hasTag(tags []string, want string) bool {
	want = strings.ToLower(strings.TrimSpace(want))
	for _, t := range tags {
		if strings.ToLower(t) == want {
			return true
		}
	}
	return false
}

// matchesSearch does a case-insensitive substring match of term over the searchable fields.
func matchesSearch(j Job, term string) bool {
	term = strings.ToLower(term)
	if strings.Contains(strings.ToLower(j.Position), term) ||
		strings.Contains(strings.ToLower(j.Company), term) ||
		strings.Contains(strings.ToLower(j.Description), term) {
		return true
	}
	for _, t := range j.Tags {
		if strings.Contains(strings.ToLower(t), term) {
			return true
		}
	}
	return false
}
