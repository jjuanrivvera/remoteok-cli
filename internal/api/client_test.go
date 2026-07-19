package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServer starts an httptest server with handler and returns its URL.
func newTestServer(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}

// newTestClient starts an httptest server with handler and returns a Client pointed at it
// with retries disabled for speed.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	return New(newTestServer(t, handler), WithMaxRetries(0))
}

// feedFixture loads the recorded Remote OK feed (with the leading attribution element).
func feedFixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "feed.json"))
	require.NoError(t, err)
	return b
}

func TestNew_Defaults(t *testing.T) {
	c := New("")
	assert.Equal(t, DefaultBaseURL, c.BaseURL())
	assert.Equal(t, DefaultUserAgent, c.UserAgent())
}

func TestClient_SendsBrowserUserAgent(t *testing.T) {
	var gotUA string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte("[]"))
	})
	_, _, err := c.Do(t.Context(), http.MethodGet, apiPath, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, DefaultUserAgent, gotUA, "a browser UA is required — Remote OK 403s the default")
}

func TestClient_WithUserAgentOverride(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(srv.Close)
	c := New(srv.URL, WithMaxRetries(0), WithUserAgent("custom-ua/1.0"))
	_, _, err := c.Do(t.Context(), http.MethodGet, apiPath, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "custom-ua/1.0", gotUA)
	// An empty override keeps the previous UA.
	WithUserAgent("")(c)
	assert.Equal(t, "custom-ua/1.0", c.UserAgent())
}

func TestClient_DryRunPrintsCurlAndSkipsRequest(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hit = true }))
	t.Cleanup(srv.Close)
	var buf bytes.Buffer
	c := New(srv.URL, WithMaxRetries(0), WithDryRun(true, &buf))
	status, body, err := c.Do(t.Context(), http.MethodGet, apiPath, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, status)
	assert.Nil(t, body)
	assert.False(t, hit, "dry-run must not hit the network")
	out := buf.String()
	assert.Contains(t, out, "curl -X GET")
	assert.Contains(t, out, "User-Agent:")
}

func TestClient_APIErrorHints(t *testing.T) {
	cases := []struct {
		status int
		hint   string
	}{
		{http.StatusForbidden, "User-Agent"},
		{http.StatusNotFound, "base-url"},
		{http.StatusTooManyRequests, "rate limited"},
		{http.StatusInternalServerError, "server error"},
	}
	for _, tc := range cases {
		c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte("<html>blocked</html>"))
		})
		_, _, err := c.Do(t.Context(), http.MethodGet, apiPath, nil, nil)
		require.Error(t, err)
		var apiErr *APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, tc.status, apiErr.StatusCode)
		assert.Contains(t, err.Error(), tc.hint)
	}
}

func TestClient_GetJSONDecodes(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1"}]`))
	})
	var out []map[string]any
	require.NoError(t, c.GetJSON(t.Context(), apiPath, nil, &out))
	require.Len(t, out, 1)
}
