package commands

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPExcludesSetupCommands(t *testing.T) {
	root := NewRootCmd()
	byName := map[string]*cobra.Command{}
	for _, c := range root.Commands() {
		byName[c.Name()] = c
	}
	for _, name := range []string{"agent", "config", "alias", "init", "doctor", "completion", "version", "api", "update", "mcp"} {
		c, ok := byName[name]
		require.True(t, ok, "command %s should exist", name)
		assert.True(t, mcpExcluded(c), "%s must be excluded from the MCP surface", name)
	}
	// The real resource IS exposed.
	assert.False(t, mcpExcluded(byName["jobs"]))
}

func TestClassifyAPICommands_JobsAreRead(t *testing.T) {
	cls := classifyAPICommands(false)
	// remoteok is read-only: no write or destructive API commands.
	assert.Empty(t, cls.Write)
	assert.Empty(t, cls.Destructive)
	var paths []string
	for _, c := range cls.Read {
		paths = append(paths, c.Path)
	}
	assert.Contains(t, paths, "jobs list")
	assert.Contains(t, paths, "jobs get")
}

func TestEveryAPICommandIsAnnotated(t *testing.T) {
	// classifyTree fails closed (destructive) on an unannotated API leaf; assert none slip
	// through as destructive, i.e. every jobs leaf carries a read annotation.
	for _, c := range classifyTree() {
		assert.Equal(t, kindRead, c.Kind, "%s must be annotated read", c.Path)
	}
}

func TestGuard_RendersHosts(t *testing.T) {
	e := newEnv(t, nil)
	for _, host := range []string{"claude-code", "codex", "opencode"} {
		out, _, err := e.run("agent", "guard", "--host", host)
		require.NoError(t, err, host)
		assert.NotEmpty(t, out)
	}
	_, _, err := e.run("agent", "guard", "--host", "bogus")
	require.Error(t, err)
}

func TestGuard_ClaudeSettingsBlocksWriteEscapes(t *testing.T) {
	cls := classifyAPICommands(false)
	settings, err := claudeSettingsJSON(cls)
	require.NoError(t, err)
	var parsed struct {
		Permissions struct {
			Deny  []string `json:"deny"`
			Allow []string `json:"allow"`
		} `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal([]byte(settings), &parsed))
	deny := strings.Join(parsed.Permissions.Deny, " ")
	// Even with no destructive job commands, the two write escapes are always blocked.
	assert.Contains(t, deny, "api DELETE")
	assert.Contains(t, deny, "api POST")
	assert.Contains(t, deny, "alias set")
	// Reads are allowed.
	allow := strings.Join(parsed.Permissions.Allow, " ")
	assert.Contains(t, allow, "jobs list")
}

func TestGuard_OpenCodeAndCodex(t *testing.T) {
	cls := classifyAPICommands(false)
	oc, err := renderOpenCode(cls)
	require.NoError(t, err)
	assert.Contains(t, oc, `"remoteok api DELETE*": "deny"`)
	assert.Contains(t, oc, `"remoteok alias set*": "deny"`)
	assert.Contains(t, oc, `"remoteok jobs list*": "allow"`)

	cx, err := renderCodex(cls)
	require.NoError(t, err)
	assert.Contains(t, cx, `sandbox_mode = "read-only"`)
}
