package api

import (
	"net/http"
	"testing"

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
	jobs, legal, err := c.Jobs(t.Context(), JobListOptions{})
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
	jobs, _, err := c.Jobs(t.Context(), JobListOptions{Tags: []string{"golang"}})
	require.NoError(t, err)
	assert.Equal(t, "tags=golang", q)
	require.Len(t, jobs, 2)
	for _, j := range jobs {
		assert.True(t, hasTag(j.Tags, "golang"))
	}

	// Two tags → AND semantics, no single server tag.
	q = ""
	jobs, _, err = c.Jobs(t.Context(), JobListOptions{Tags: []string{"golang", "aws"}})
	require.NoError(t, err)
	assert.Empty(t, q, "multi-tag requests are not narrowed server-side")
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1003"), jobs[0].ID)
}

func TestJobs_SearchCompanySalaryLimit(t *testing.T) {
	c := newTestClient(t, serveFeed(t, nil))

	// Search matches description/position/tags.
	jobs, _, err := c.Jobs(t.Context(), JobListOptions{Search: "kubernetes"})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1001"), jobs[0].ID)

	// Company substring (case-insensitive).
	jobs, _, err = c.Jobs(t.Context(), JobListOptions{Company: "initech"})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1003"), jobs[0].ID)

	// Min salary keeps only jobs whose salary_max clears the bar.
	jobs, _, err = c.Jobs(t.Context(), JobListOptions{MinSalary: 190000})
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, ID("1003"), jobs[0].ID)

	// Limit caps the result set.
	jobs, _, err = c.Jobs(t.Context(), JobListOptions{Limit: 2})
	require.NoError(t, err)
	require.Len(t, jobs, 2)
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
	jobs, _, err := c.Jobs(t.Context(), JobListOptions{})
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
