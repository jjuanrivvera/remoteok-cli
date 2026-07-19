package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobsList_TableAndAttribution(t *testing.T) {
	e := newEnv(t, nil)
	out, errOut, err := e.run("jobs", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "Senior Go Engineer")
	assert.Contains(t, out, "Initech")
	// The legal notice is never rendered as a job row.
	assert.NotContains(t, out, "link back")
	// Attribution goes to stderr (stdout stays pipe-clean).
	assert.Contains(t, errOut, "Source: Remote OK")
	assert.NotContains(t, out, "Source: Remote OK")
}

func TestJobsList_JSONIsPipeClean(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "list", "-o", "json")
	require.NoError(t, err)
	var jobs []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &jobs))
	require.Len(t, jobs, 3)
	assert.Equal(t, "1001", jobs[0]["id"], "newest-first, id normalized to string")
}

func TestJobsList_TagFilterAND(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "list", "--tag", "golang", "--tag", "aws", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1003", strings.TrimSpace(out))
}

func TestJobsList_SearchCompanyMinSalaryLimit(t *testing.T) {
	e := newEnv(t, nil)

	out, _, err := e.run("jobs", "list", "--search", "kubernetes", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1001", strings.TrimSpace(out))

	out, _, err = e.run("jobs", "list", "--company", "initech", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1003", strings.TrimSpace(out))

	out, _, err = e.run("jobs", "list", "--min-salary", "190000", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1003", strings.TrimSpace(out))

	out, _, err = e.run("jobs", "list", "--limit", "2", "-o", "id")
	require.NoError(t, err)
	assert.Len(t, strings.Fields(out), 2)
}

func TestParseSince(t *testing.T) {
	// Fixed reference so the relative grammar is deterministic regardless of wall clock.
	now := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		in      string
		want    time.Time
		wantErr bool
	}{
		{"empty is no filter", "", time.Time{}, false},
		{"absolute date", "2026-07-12", time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC), false},
		{"relative days", "7d", now.AddDate(0, 0, -7), false},
		{"relative weeks", "2w", now.AddDate(0, 0, -14), false},
		{"one day", "1d", now.AddDate(0, 0, -1), false},
		{"whitespace trimmed", "  3d  ", now.AddDate(0, 0, -3), false},
		{"bad unit", "7m", time.Time{}, true},
		{"bad month", "2026-13-01", time.Time{}, true},
		{"not a date at all", "yesterday", time.Time{}, true},
		{"missing number", "d", time.Time{}, true},
		{"datetime not accepted", "2026-07-12T00:00:00Z", time.Time{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSinceAt(tc.in, now)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "--since")
				return
			}
			require.NoError(t, err)
			assert.True(t, tc.want.Equal(got), "want %s, got %s", tc.want, got)
		})
	}
}

func TestParseSince_UsesNow(t *testing.T) {
	// The exported wrapper anchors on time.Now(); a relative window must land in the past.
	got, err := parseSince("1d")
	require.NoError(t, err)
	assert.True(t, got.Before(time.Now()), "a relative --since is a past instant")
}

func TestJobsList_SinceAbsoluteDropsOld(t *testing.T) {
	// Fixture postings: 1001=07-18, 1002=07-17, 1003=07-16. An absolute date is
	// wall-clock independent, so this stays deterministic forever.
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "list", "--since", "2026-07-17", "-o", "id")
	require.NoError(t, err)
	ids := strings.Fields(out)
	assert.Equal(t, []string{"1001", "1002"}, ids, "07-16 listing is dropped, 07-17/07-18 kept")
}

func TestJobsList_SincePostedAfterAlias(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "list", "--posted-after", "2026-07-18", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1001", strings.TrimSpace(out), "--posted-after is an alias for --since")
}

func TestJobsList_SinceComposesWithTag(t *testing.T) {
	// golang jobs are 1001 (07-18) and 1003 (07-16); the date filter narrows to 1001.
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "list", "--tag", "golang", "--since", "2026-07-17", "-o", "id")
	require.NoError(t, err)
	assert.Equal(t, "1001", strings.TrimSpace(out))
}

func TestJobsList_SinceInvalid(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("jobs", "list", "--since", "last-tuesday")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--since")
	assert.Contains(t, err.Error(), "YYYY-MM-DD")
}

func TestJobsList_CSV(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "list", "-o", "csv", "--columns", "id,company")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.Equal(t, "id,company", lines[0])
	assert.Equal(t, "1001,Acme Corp", lines[1])
}

func TestJobsGet(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "get", "1002", "-o", "json", "--quiet")
	require.NoError(t, err)
	assert.Contains(t, out, "Frontend Developer")

	_, _, err = e.run("jobs", "get", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestJobsList_DryRun(t *testing.T) {
	e := newEnv(t, nil)
	d := e.deps()
	var curl bytes.Buffer
	d.out = &curl // dry-run curls are written here (not cmd stdout)
	_, _, err := runWithDeps(t, d, "jobs", "list", "--tag", "golang", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, curl.String(), "curl -X GET")
	assert.Contains(t, curl.String(), "tags=golang")
}

func TestJobsList_QuietSuppressesAttribution(t *testing.T) {
	e := newEnv(t, nil)
	_, errOut, err := e.run("jobs", "list", "--quiet")
	require.NoError(t, err)
	assert.NotContains(t, errOut, "Source: Remote OK")
}

func TestJobsList_ServerError(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("blocked"))
	})
	_, _, err := e.run("jobs", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "User-Agent")
}
