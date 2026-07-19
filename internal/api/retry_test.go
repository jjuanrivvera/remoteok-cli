package api

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry_RetriesOn503ThenSucceeds(t *testing.T) {
	retryBase = time.Millisecond // keep the test fast
	var calls int32
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("[]"))
	})
	c := New(srv, WithMaxRetries(3))
	status, _, err := c.Do(t.Context(), http.MethodGet, apiPath, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.EqualValues(t, 2, atomic.LoadInt32(&calls))
}

func TestRetry_HonorsRetryAfter(t *testing.T) {
	var calls int32
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte("[]"))
	})
	c := New(srv, WithMaxRetries(2))
	_, _, err := c.Do(t.Context(), http.MethodGet, apiPath, nil, nil)
	require.NoError(t, err)
	assert.EqualValues(t, 2, atomic.LoadInt32(&calls))
}

func TestRetryAfter(t *testing.T) {
	h := http.Header{}
	assert.Zero(t, retryAfter(h))
	h.Set("Retry-After", "2")
	assert.Equal(t, 2*time.Second, retryAfter(h))
	h.Set("Retry-After", "garbage")
	assert.Zero(t, retryAfter(h))
}

func TestIsTransient(t *testing.T) {
	assert.False(t, isTransient(context.Canceled))
	assert.False(t, isTransient(context.DeadlineExceeded))
}
