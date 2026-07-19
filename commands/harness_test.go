package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jjuanrivvera/remoteok-cli/internal/api"
)

// feedJSON is a recorded Remote OK feed: the leading legal/attribution element, then three
// jobs with varied tags/salary/company so command tests can exercise every filter and prove
// the attribution element is skipped.
const feedJSON = `[
  {"last_updated":1784422736,"legal":"Please link back (with follow!) to Remote OK as a source."},
  {"id":"1001","epoch":1784341832,"date":"2026-07-18","company":"Acme Corp","position":"Senior Go Engineer","tags":["golang","backend","remote","kubernetes"],"description":"Distributed systems in Go.","location":"Worldwide","salary_min":120000,"salary_max":180000,"url":"https://remoteOK.com/l/1001","apply_url":"https://remoteOK.com/l/1001"},
  {"id":1002,"epoch":1784241832,"date":"2026-07-17","company":"Globex","position":"Frontend Developer","tags":["javascript","react","remote"],"description":"React product.","location":"Europe","salary_min":"70000","salary_max":"95000","url":"https://remoteOK.com/l/1002","apply_url":"https://remoteOK.com/l/1002"},
  {"id":"1003","epoch":1784141832,"date":"2026-07-16","company":"Initech","position":"Platform SRE","tags":["golang","devops","remote","aws"],"description":"Go tooling and Terraform.","location":"US only","salary_min":140000,"salary_max":200000,"url":"https://remoteOK.com/l/1003","apply_url":"https://remoteOK.com/l/1003"}
]`

// env wires one test invocation: an httptest Remote OK, an isolated config dir, and a
// retry-free client.
type env struct {
	t   *testing.T
	srv *httptest.Server
}

// newEnv starts a mock Remote OK server (serving feedJSON unless handler overrides) and
// isolates all state under t.TempDir().
func newEnv(t *testing.T, handler http.HandlerFunc) *env {
	t.Helper()
	if handler == nil {
		handler = func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(feedJSON))
		}
	}
	e := &env{t: t, srv: httptest.NewServer(handler)}
	t.Cleanup(e.srv.Close)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("REMOTEOK_BASE_URL", e.srv.URL)
	t.Setenv("REMOTEOK_USER_AGENT", "")
	t.Setenv("NO_COLOR", "1")
	return e
}

func (e *env) deps() *deps {
	d := newDeps()
	d.clientOpts = []api.Option{api.WithMaxRetries(0)}
	return d
}

// run executes the real command tree with captured output.
func (e *env) run(args ...string) (string, string, error) {
	e.t.Helper()
	return runWithDeps(e.t, e.deps(), args...)
}

// runWithDeps builds a fresh tree from d and runs it with captured output.
func runWithDeps(t *testing.T, d *deps, args ...string) (string, string, error) {
	t.Helper()
	root := newRootCmd(d)
	root.SetArgs(args)
	var out, errB bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errB)
	err := root.ExecuteContext(t.Context())
	return out.String(), errB.String(), err
}
