package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// serveFeed returns a handler that answers with the recorded feed fixture and records the
// query the CLI sent (so tests can assert the server-side ?tags= optimization).
func serveFeed(t *testing.T, gotQuery *string) http.HandlerFunc {
	feed := feedFixture(t)
	return func(w http.ResponseWriter, r *http.Request) {
		if gotQuery != nil {
			*gotQuery = r.URL.RawQuery
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(feed)
	}
}

func TestJobs_SkipsLegalElementAndSortsByEpoch(t *testing.T) {
	c := newTestClient(t, serveFeed(t, nil))
	jobs, legal, _, err := c.Jobs(t.Context(), JobListOptions{})
	require.NoError(t, err)

	// The leading legal/attribution element must NOT be parsed as a job.
	require.Len(t, jobs, 3, "3 jobs, the legal notice skipped")
	require.NotNil(t, legal, "the attribution notice is surfaced")
	assert.Contains(t, legal.Legal, "link back")

	// Newest first by epoch.
	assert.Equal(t, ID("1001"), jobs[0].ID)
	assert.Equal(t, ID("1002"), jobs[1].ID)
	assert.Equal(t, ID("1003"), jobs[2].ID)

	// A numeric id (1002) and quoted salary strings decode via the flexible types.
	assert.Equal(t, ID("1002"), jobs[1].ID)
	assert.Equal(t, int64(95000), jobs[1].SalaryMax.Int64())
}

func TestJobs_TagFilterANDServerSideOptimization(t *testing.T) {
	var q string
	c := newTestClient(t, serveFeed(t, &q))

	// Single tag → sent server-side as ?tags= AND filtered client-side.
	jobs, _, _, err := c.Jobs(t.Context(), JobListOptions{Tags: []string{"golang"}})
	require.NoError(t, err)
	assert.Equal(t, "tags=golang", q)
	require.Len(t, jobs, 2)
	for _, j := range jobs {
		assert.True(t, hasTag(j.Tags, "golang"))
	}

	// Two tags → AND semantics, no single server tag.
	q = ""
	jobs, _, _, err = c.Jobs(t.Context(), JobListOptions{Tags: []string{"golang", "aws"}})
	require.NoError(t, err)
	assert.Empty(t, q, "multi-tag requests are not narrowed server-side")
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1003"), jobs[0].ID)
}

func TestJobs_SearchCompanySalaryLimit(t *testing.T) {
	c := newTestClient(t, serveFeed(t, nil))

	// Search matches description/position/tags.
	jobs, _, _, err := c.Jobs(t.Context(), JobListOptions{Search: "kubernetes"})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1001"), jobs[0].ID)

	// Company substring (case-insensitive).
	jobs, _, _, err = c.Jobs(t.Context(), JobListOptions{Company: "initech"})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1003"), jobs[0].ID)

	// Min salary keeps only jobs whose salary_max clears the bar.
	jobs, _, _, err = c.Jobs(t.Context(), JobListOptions{MinSalary: 190000})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1003"), jobs[0].ID)

	// Limit caps the result set.
	jobs, _, _, err = c.Jobs(t.Context(), JobListOptions{Limit: 2})
	require.NoError(t, err)
	require.Len(t, jobs, 2)
}

func TestJobs_SinceFilter(t *testing.T) {
	// Fixture postings: 1001=2026-07-18, 1002=2026-07-17, 1003=2026-07-16 (all UTC).
	c := newTestClient(t, serveFeed(t, nil))

	// A threshold between the newest and the rest keeps only the recent listing and
	// drops the older ones.
	since := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	jobs, _, _, err := c.Jobs(t.Context(), JobListOptions{Since: since})
	require.NoError(t, err)
	require.Len(t, jobs, 1, "only the 2026-07-18 listing clears the threshold")
	assert.Equal(t, ID("1001"), jobs[0].ID)

	// A boundary exactly on a posting instant is inclusive (>=).
	since = time.Date(2026, 7, 17, 2, 30, 32, 0, time.UTC)
	jobs, _, _, err = c.Jobs(t.Context(), JobListOptions{Since: since})
	require.NoError(t, err)
	require.Len(t, jobs, 2)
	assert.Equal(t, ID("1001"), jobs[0].ID)
	assert.Equal(t, ID("1002"), jobs[1].ID)

	// A zero Since is a no-op: every listing survives (composes with other filters).
	jobs, _, _, err = c.Jobs(t.Context(), JobListOptions{})
	require.NoError(t, err)
	require.Len(t, jobs, 3)

	// Since AND tag compose: golang jobs (1001,1003) intersected with posted >= 07-17.
	jobs, _, _, err = c.Jobs(t.Context(), JobListOptions{
		Tags:  []string{"golang"},
		Since: time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1001"), jobs[0].ID)
}

func TestCountNoSalaryExcluded(t *testing.T) {
	// A small hand-built feed: two listings publish a salary, two do not. Only the no-salary
	// listings that would otherwise pass every filter must be counted.
	golangPaid := Job{ID: "p1", Tags: []string{"golang"}, Company: "Acme", SalaryMax: Int(120000)}
	golangUnpaid := Job{ID: "u1", Tags: []string{"golang"}, Company: "Acme"}
	reactUnpaid := Job{ID: "u2", Tags: []string{"react"}, Company: "Globex"}
	belowBarPaid := Job{ID: "p2", Tags: []string{"golang"}, Company: "Initech", SalaryMax: Int(50000)}
	jobs := []Job{golangPaid, golangUnpaid, reactUnpaid, belowBarPaid}

	cases := []struct {
		name string
		opts JobListOptions
		want int
	}{
		{"min-salary unset counts nothing", JobListOptions{}, 0},
		{
			// u1 and u2 have no salary and pass (no other filter); the two paid ones are
			// excluded for having a salary (one clears the bar, one below it) — not counted.
			"counts every no-salary listing that passes other filters",
			JobListOptions{MinSalary: 80000},
			2,
		},
		{
			// Only golang listings are eligible; of those only u1 lacks a salary.
			"respects other filters — tag narrows the no-salary set",
			JobListOptions{MinSalary: 80000, Tags: []string{"golang"}},
			1,
		},
		{
			// A tag that no listing carries: nothing would pass, so nothing is a no-salary drop.
			"no other-filter matches → zero",
			JobListOptions{MinSalary: 80000, Tags: []string{"rust"}},
			0,
		},
		{
			// belowBarPaid publishes a salary under the bar: a genuine below-bar exclusion,
			// which the hint must NOT attribute to missing salary.
			"below-bar published salary is not a no-salary exclusion",
			JobListOptions{MinSalary: 80000, Company: "Initech"},
			0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, countNoSalaryExcluded(jobs, tc.opts))
		})
	}
}

func TestJobs_ReturnsNoSalaryExcludedCount(t *testing.T) {
	// The fixture's three listings all publish a salary, so a bar under all of them excludes
	// nobody for missing salary even though it is set.
	c := newTestClient(t, serveFeed(t, nil))
	_, _, excluded, err := c.Jobs(t.Context(), JobListOptions{MinSalary: 1})
	require.NoError(t, err)
	assert.Equal(t, 0, excluded, "every fixture listing publishes a salary")

	// A bar above every published salary_max drops all three — all as real below-bar drops,
	// none for missing salary.
	_, _, excluded, err = c.Jobs(t.Context(), JobListOptions{MinSalary: 10_000_000})
	require.NoError(t, err)
	assert.Equal(t, 0, excluded)
}

func TestJobPostedAt(t *testing.T) {
	// RFC3339 `date` is the primary source.
	got, ok := jobPostedAt(Job{Date: "2026-07-18T02:30:32+00:00"})
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 7, 18, 2, 30, 32, 0, time.UTC), got.UTC())

	// A garbled `date` falls back to epoch.
	got, ok = jobPostedAt(Job{Date: "not-a-date", Epoch: Int(1784341832)})
	require.True(t, ok)
	assert.Equal(t, int64(1784341832), got.Unix())

	// Neither field usable → not datable.
	_, ok = jobPostedAt(Job{})
	assert.False(t, ok)
}

func TestGetJob(t *testing.T) {
	c := newTestClient(t, serveFeed(t, nil))
	job, legal, err := c.GetJob(t.Context(), "1002")
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, "Frontend Developer", job.Position)
	assert.NotNil(t, legal)

	_, _, err = c.GetJob(t.Context(), "does-not-exist")
	assert.ErrorIs(t, err, ErrJobNotFound)
}

func TestJobs_MalformedElementSkipped(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		// A legal notice, a good job, and a junk element that must be skipped.
		_, _ = w.Write([]byte(`[{"legal":"x"},{"id":"9","position":"OK"},{"id":{"bad":1}}]`))
	})
	jobs, _, _, err := c.Jobs(t.Context(), JobListOptions{})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("9"), jobs[0].ID)
}

func TestIsLegal(t *testing.T) {
	assert.True(t, isLegal([]byte(`{"legal":"terms","last_updated":1}`)))
	assert.False(t, isLegal([]byte(`{"id":"1","legal":"weird"}`)), "a job with a legal field is still a job")
	assert.False(t, isLegal([]byte(`{"id":"1"}`)))
	assert.False(t, isLegal([]byte(`not json`)))
}
