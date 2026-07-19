package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/remoteok-cli/internal/update"
)

func TestVersion_Check(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9"}`))
	}))
	t.Cleanup(srv.Close)
	orig := latestReleaseURL
	latestReleaseURL = srv.URL
	t.Cleanup(func() { latestReleaseURL = orig })

	e := newEnv(t, nil)
	out, _, err := e.run("version", "--check")
	require.NoError(t, err)
	assert.Contains(t, out, "9.9.9")
}

func TestUpdate_DevBuildNoOp(t *testing.T) {
	orig := newUpdater
	newUpdater = func() *update.Updater { return update.NewUpdater("dev") }
	t.Cleanup(func() { newUpdater = orig })
	e := newEnv(t, nil)
	out, _, err := e.run("update")
	require.NoError(t, err)
	assert.Contains(t, out, "latest")
}

func TestAPI_DryRunAndData(t *testing.T) {
	e := newEnv(t, nil)
	d := e.deps()
	var curl bytes.Buffer
	d.out = &curl
	_, _, err := runWithDeps(t, d, "api", "GET", "api", "-q", "tags=golang", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, curl.String(), "curl -X GET")
	assert.Contains(t, curl.String(), "tags=golang")

	// -q without '=' is rejected.
	_, _, err = e.run("api", "GET", "api", "-q", "bad")
	require.Error(t, err)
}

func TestGuard_WriteFiles(t *testing.T) {
	e := newEnv(t, nil)
	dir := t.TempDir()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	_, _, err = e.run("agent", "guard", "--host", "claude-code", "--write")
	require.NoError(t, err)
	hook := filepath.Join(dir, ".claude", "hooks", "remoteok-guard.sh")
	assert.FileExists(t, hook)
	assert.FileExists(t, filepath.Join(dir, ".claude", "remoteok-guard.settings.json"))

	// A second run must refuse to overwrite.
	_, _, err = e.run("agent", "guard", "--host", "claude-code", "--write")
	require.Error(t, err)
}

func TestGuard_OutFile(t *testing.T) {
	e := newEnv(t, nil)
	out := filepath.Join(t.TempDir(), "codex.toml")
	_, _, err := e.run("agent", "guard", "--host", "codex", "--out", out)
	require.NoError(t, err)
	b, err := os.ReadFile(out) //nolint:gosec
	require.NoError(t, err)
	assert.Contains(t, string(b), "read-only")
}

func TestDoctor_HumanOutput(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("doctor")
	require.NoError(t, err)
	assert.Contains(t, out, "feed")
	assert.Contains(t, out, "GET /api OK")
}

func TestJobsGet_DryRun(t *testing.T) {
	e := newEnv(t, nil)
	d := e.deps()
	var curl bytes.Buffer
	d.out = &curl
	_, _, err := runWithDeps(t, d, "jobs", "get", "1001", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, curl.String(), "curl -X GET")
}

func TestConfig_SetBaseURLValid(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("config", "set", "base_url", "https://remoteok.com")
	require.NoError(t, err)
	out, _, err := e.run("config", "view")
	require.NoError(t, err)
	assert.Contains(t, out, "remoteok.com")
}

func TestReportLatest_UpToDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3"}`))
	}))
	t.Cleanup(srv.Close)
	orig := latestReleaseURL
	latestReleaseURL = srv.URL
	t.Cleanup(func() { latestReleaseURL = orig })

	var out bytes.Buffer
	cmd := newRootCmd(newDeps())
	cmd.SetOut(&out)
	cmd.SetContext(t.Context())
	require.NoError(t, reportLatest(cmd, "1.2.3"))
	assert.Contains(t, out.String(), "up to date")
}

func TestRoot_HelpListsJobs(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("--help")
	require.NoError(t, err)
	assert.True(t, strings.Contains(out, "jobs"))
}
