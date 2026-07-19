package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHookScript_BashExecution exercises the generated hook with real bash. remoteok is a
// read-only client, so the two blocked paths are the raw `api` escape hatch (write METHOD)
// and `alias set` (which could mint a shorthand for a future blocked command). It verifies
// obfuscation, path-invoked binaries, command-position anchoring, and the benign lookalikes
// that must stay allowed.
func TestHookScript_BashExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "remoteok-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]any{"command": command}})
		return string(b)
	}
	mcpPayload := func(toolName string) string {
		b, _ := json.Marshal(map[string]any{"tool_name": toolName, "tool_input": map[string]any{}})
		return string(b)
	}

	runHook := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}
	isDenied := func(output string) bool { return strings.Contains(output, `"permissionDecision":"deny"`) }

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		// --- alias minting ---
		{"alias_set_denied", bashPayload(`remoteok alias set kill "api DELETE x"`), true},
		{"quote_split_denied", bashPayload(`remoteok alias s""et kill "x"`), true},
		{"single_quote_split_denied", bashPayload(`remoteok alias s''et kill "x"`), true},
		{"backslash_denied", bashPayload(`remoteok alias s\et kill "x"`), true},
		{"newline_continuation_denied", bashPayload("remoteok alias \\\nset kill x"), true},
		// --- command position after separators ---
		{"after_semicolon_denied", bashPayload("true; remoteok alias set kill x"), true},
		{"after_pipe_denied", bashPayload("echo hi | remoteok alias set kill x"), true},
		{"after_and_denied", bashPayload("true && remoteok alias set kill x"), true},
		{"trailing_separator_denied", bashPayload("remoteok alias set kill x;true"), true},
		{"env_prefix_denied", bashPayload("env REMOTEOK_BASE_URL=x remoteok alias set kill x"), true},
		// --- path-invoked binaries ---
		{"relative_path_binary_denied", bashPayload("./bin/remoteok alias set kill x"), true},
		{"absolute_path_binary_denied", bashPayload("/usr/local/bin/remoteok alias set kill x"), true},
		{"absolute_path_api_denied", bashPayload("/usr/local/bin/remoteok api DELETE api"), true},
		// --- raw api escape hatch (METHOD position; only GET/HEAD/OPTIONS pass) ---
		{"api_delete_denied", bashPayload("remoteok api DELETE api"), true},
		{"api_lowercase_delete_denied", bashPayload("remoteok api delete api"), true},
		{"api_post_denied", bashPayload("remoteok api POST api -d '{}'"), true},
		{"api_patch_denied", bashPayload("remoteok api PATCH api -d '{}'"), true},
		{"api_put_denied", bashPayload("remoteok api PUT api -d '{}'"), true},
		{"api_flag_before_method_denied", bashPayload("remoteok api -q x=1 DELETE api"), true},
		{"api_compound_get_then_delete_denied", bashPayload("remoteok api GET api;remoteok api DELETE api"), true},
		// --- raw api reads stay allowed ---
		{"api_get_allowed", bashPayload("remoteok api GET api"), false},
		{"api_get_lowercase_allowed", bashPayload("remoteok api get api"), false},
		{"api_head_allowed", bashPayload("remoteok api HEAD api"), false},
		{"api_get_delete_in_path_allowed", bashPayload("remoteok api GET api?tags=delete"), false},
		// --- benign lookalikes that must stay allowed ---
		{"jobs_list_allowed", bashPayload("remoteok jobs list --tag golang"), false},
		{"jobs_get_allowed", bashPayload("remoteok jobs get 1001"), false},
		{"search_with_delete_in_arg_allowed", bashPayload(`remoteok jobs list --search "how to delete a job"`), false},
		{"alias_list_allowed", bashPayload("remoteok alias list"), false},
		{"alias_remove_allowed", bashPayload("remoteok alias remove go"), false},
		{"quoted_blocked_cmd_in_arg_denied_conservatively", bashPayload(`rg "remoteok alias set" docs/`), true},
		{"cat_file_allowed", bashPayload("cat jobs_delete.go"), false},
		{"other_binary_allowed", bashPayload("myremoteok alias set kill x"), false},
		{"other_binary_api_allowed", bashPayload("myremoteok api DELETE api"), false},
		// --- MCP branch: read tools allowed (there are no destructive MCP tools) ---
		{"mcp_jobs_list_allowed", mcpPayload("mcp__remoteok__remoteok_jobs_list"), false},
		{"mcp_jobs_get_allowed", mcpPayload("mcp__remoteok__remoteok_jobs_get"), false},
		{"mcp_near_miss_allowed", mcpPayload("mcp__remoteok__remoteok_jobs_list2"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHook(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}

// TestHookScript_BashExecutionNoJq exercises the no-jq fallback with a STRICT PATH holding
// only the POSIX tools the hook needs, so jq is genuinely unreachable (GOAL.md §3b #3).
func TestHookScript_BashExecutionNoJq(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "remoteok-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	binDir := filepath.Join(tmpDir, "nojq-bin")
	if err := os.Mkdir(binDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, tool := range []string{"cat", "tr", "grep", "sed", "printf", "env"} {
		p, lerr := exec.LookPath(tool)
		if lerr != nil {
			continue
		}
		if serr := os.Symlink(p, filepath.Join(binDir, tool)); serr != nil {
			t.Fatalf("symlink %s: %v", tool, serr)
		}
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{"tool_name": "Bash", "tool_input": map[string]any{"command": command}})
		return string(b)
	}
	runHookNoJq := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		env := make([]string, 0, len(os.Environ()))
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "PATH=") {
				env = append(env, e)
			}
		}
		cmd.Env = append(env, "PATH="+binDir)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}
	isDenied := func(output string) bool { return strings.Contains(output, `"permissionDecision":"deny"`) }

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		{"nojq_alias_set_denied", bashPayload("remoteok alias set kill x"), true},
		{"nojq_obfuscated_alias_set_denied", bashPayload(`remoteok alias s""et kill x`), true},
		{"nojq_path_binary_denied", bashPayload("./bin/remoteok alias set kill x"), true},
		{"nojq_api_delete_denied", bashPayload("remoteok api DELETE api"), true},
		{"nojq_cat_file_allowed", bashPayload("cat jobs_delete.go"), false},
		{"nojq_jobs_list_allowed", bashPayload("remoteok jobs list --tag golang"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHookNoJq(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}
