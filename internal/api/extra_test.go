package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := New("", WithHTTPClient(custom))
	assert.Same(t, custom, c.httpc)
}

func TestGetJSON_DryRunSkips(t *testing.T) {
	c := New("https://remoteok.com", WithDryRun(true, discard{}))
	var out []map[string]any
	require.NoError(t, c.GetJSON(t.Context(), apiPath, nil, &out))
	assert.Nil(t, out)
}

func TestDo_NonIdempotentNotRetried(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	c.maxRetries = 3
	_, _, err := c.Do(t.Context(), http.MethodPost, apiPath, nil, []byte(`{}`))
	require.Error(t, err)
	assert.Equal(t, 1, calls, "POST is never auto-retried")
}

func TestBodySnippetTruncates(t *testing.T) {
	long := make([]byte, 500)
	for i := range long {
		long[i] = 'a'
	}
	s := bodySnippet(long)
	assert.LessOrEqual(t, len(s), 201+len("…"))
	assert.Empty(t, bodySnippet([]byte("   ")))
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
