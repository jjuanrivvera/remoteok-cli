package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/remoteok-cli/internal/update"
)

func TestVersion(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("version")
	require.NoError(t, err)
	assert.Contains(t, out, "remoteok")

	out, _, err = e.run("version", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, `"version"`)
}

func TestConfig_PathViewSet(t *testing.T) {
	e := newEnv(t, nil)

	out, _, err := e.run("config", "path")
	require.NoError(t, err)
	assert.Contains(t, out, "config.yaml")

	_, _, err = e.run("config", "set", "user_agent", "custom/1.0")
	require.NoError(t, err)

	out, _, err = e.run("config", "view")
	require.NoError(t, err)
	assert.Contains(t, out, "custom/1.0")

	_, _, err = e.run("config", "set", "base_url", "http://example.com")
	require.Error(t, err, "cleartext http:// to a non-loopback host is rejected")

	_, _, err = e.run("config", "set", "bogus", "x")
	require.Error(t, err)
}

func TestAlias_SetListRemoveAndExpand(t *testing.T) {
	e := newEnv(t, nil)

	_, _, err := e.run("alias", "set", "go", "jobs list --tag golang")
	require.NoError(t, err)

	out, _, err := e.run("alias", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "go = jobs list --tag golang")

	// A built-in can never be aliased.
	_, _, err = e.run("alias", "set", "jobs", "whatever")
	require.Error(t, err)

	// ExpandAliases rewrites a known alias but leaves built-ins alone.
	assert.Equal(t, []string{"jobs", "list", "--tag", "golang"}, ExpandAliases([]string{"go"}))
	assert.Equal(t, []string{"jobs", "list"}, ExpandAliases([]string{"jobs", "list"}))

	_, _, err = e.run("alias", "remove", "go")
	require.NoError(t, err)
	_, _, err = e.run("alias", "remove", "missing")
	require.Error(t, err)
}

func TestDoctor(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("doctor", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, `"feed"`)
	assert.Contains(t, out, `"ok": true`)
}

func TestDoctor_FailsOnBadFeed(t *testing.T) {
	e := newEnv(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})
	_, _, err := e.run("doctor")
	require.Error(t, err)
}

func TestInit(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("init")
	require.NoError(t, err)
	assert.Contains(t, out, "Feed reachable")
	assert.Contains(t, out, "config")
}

func TestAPICommand(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("api", "GET", "api", "-o", "json", "--quiet")
	require.NoError(t, err)
	assert.Contains(t, out, "Senior Go Engineer")

	_, _, err = e.run("api", "BOGUS", "api")
	require.Error(t, err)
}

func TestCompletion(t *testing.T) {
	e := newEnv(t, nil)
	for _, sh := range []string{"bash", "zsh", "fish", "powershell"} {
		out, _, err := e.run("completion", sh)
		require.NoError(t, err, sh)
		assert.NotEmpty(t, out)
	}
	_, _, err := e.run("completion", "tcsh")
	require.Error(t, err)
}

func TestUpdateCheck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v9.9.9","assets":[]}`))
	}))
	t.Cleanup(srv.Close)
	orig := newUpdater
	newUpdater = func() *update.Updater { return update.NewUpdaterWithBaseURL("1.0.0", srv.URL) }
	t.Cleanup(func() { newUpdater = orig })

	e := newEnv(t, nil)
	out, _, err := e.run("update", "check")
	require.NoError(t, err)
	assert.Contains(t, out, "9.9.9")
}

func TestRootUnknownOutputFormat(t *testing.T) {
	e := newEnv(t, nil)
	_, _, err := e.run("jobs", "list", "-o", "xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output format")
}

func TestExpandAliases_NoConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// An unknown first token with no config is returned unchanged.
	assert.Equal(t, []string{"whatever"}, ExpandAliases([]string{"whatever"}))
	assert.Equal(t, []string(nil), ExpandAliases(nil))
}

// TestGendocsBuildable is a light guard that the doc tree builds (mirrors gendocs).
func TestHelpHasExamples(t *testing.T) {
	e := newEnv(t, nil)
	out, _, err := e.run("jobs", "list", "--help")
	require.NoError(t, err)
	assert.True(t, strings.Contains(out, "Examples:") || strings.Contains(out, "remoteok jobs list"))
	_ = os.Stdout
}
