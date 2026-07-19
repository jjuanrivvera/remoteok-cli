package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

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
