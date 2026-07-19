// Package api is the Remote OK client core: a browser-User-Agent HTTP client with
// idempotent-only retry (honoring Retry-After), a dry-run curl mode, and the typed `jobs`
// resource on top of one generic request path. Remote OK is an unauthenticated, read-only
// public API — a single `GET /api` endpoint that returns a JSON array whose FIRST element is
// a legal/attribution notice, not a job (see jobs.go).
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// DefaultBaseURL is the Remote OK site root; the JSON API lives at /api under it.
const DefaultBaseURL = "https://remoteok.com"

// DefaultUserAgent is a real browser User-Agent. This is REQUIRED, not cosmetic: Remote OK
// (behind Cloudflare) returns 403 for a missing/default/bot-like UA. Pinned per DECISIONS.md.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// apiPath is the single documented Remote OK endpoint.
const apiPath = "/api"

// Client is a Remote OK HTTP client. It carries no credentials — the API is public.
type Client struct {
	baseURL   string
	userAgent string
	httpc     *http.Client

	// DryRun prints the equivalent curl to DryRunOut instead of sending the request.
	DryRun    bool
	DryRunOut io.Writer

	Verbose    bool
	VerboseOut io.Writer

	maxRetries int
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the HTTP transport (tests point it at httptest servers).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpc = h } }

// WithUserAgent overrides the browser User-Agent sent on every request.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// WithDryRun enables curl-printing mode.
func WithDryRun(dry bool, out io.Writer) Option {
	return func(c *Client) { c.DryRun = dry; c.DryRunOut = out }
}

// WithMaxRetries overrides the retry budget (tests set 0 for speed).
func WithMaxRetries(n int) Option { return func(c *Client) { c.maxRetries = n } }

// New builds a Remote OK client. An empty baseURL means DefaultBaseURL.
func New(baseURL string, opts ...Option) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		userAgent:  DefaultUserAgent,
		httpc:      http.DefaultClient,
		DryRunOut:  os.Stdout,
		VerboseOut: os.Stderr,
		maxRetries: 3,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// BaseURL returns the resolved Remote OK base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// UserAgent returns the User-Agent sent on every request.
func (c *Client) UserAgent() string { return c.userAgent }

// GetJSON GETs path (relative to the base URL) and decodes the JSON response into out.
// A dry-run returns nil with out untouched.
func (c *Client) GetJSON(ctx context.Context, path string, q url.Values, out any) error {
	status, body, err := c.Do(ctx, http.MethodGet, path, q, nil)
	if err != nil {
		return err
	}
	if status == 0 { // dry-run: no response to decode
		return nil
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode Remote OK response: %w", err)
	}
	return nil
}

// Do sends one request and returns status and body. A dry-run returns status 0 with no
// error. Non-2xx statuses return an *APIError.
func (c *Client) Do(ctx context.Context, method, path string, q url.Values, body []byte) (int, []byte, error) {
	u := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	headers := map[string]string{
		"Accept":     "application/json",
		"User-Agent": c.userAgent,
	}

	if c.DryRun {
		c.printCurl(method, u, body, headers)
		return 0, nil, nil
	}

	send := func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, method, u, nil)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		if c.Verbose {
			fmt.Fprintf(c.VerboseOut, "> %s %s\n", method, u)
		}
		return c.httpc.Do(req)
	}

	resp, err := c.sendWithRetry(ctx, method, send)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return 0, nil, fmt.Errorf("read response: %w", err)
	}
	if c.Verbose {
		fmt.Fprintf(c.VerboseOut, "< HTTP %d (%d bytes)\n", resp.StatusCode, len(respBody))
	}
	if resp.StatusCode >= 400 {
		return resp.StatusCode, respBody, parseAPIError(resp.StatusCode, respBody, resp.Header)
	}
	return resp.StatusCode, respBody, nil
}

// printCurl emits a copy-pasteable curl equivalent. There is no secret to redact — the API
// is public — but the User-Agent is shown because it is load-bearing (a wrong UA 403s).
func (c *Client) printCurl(method, fullURL string, body []byte, headers map[string]string) {
	var b strings.Builder
	b.WriteString("curl -X " + method + " " + shellQuote(fullURL))
	for _, k := range sortedKeys(headers) {
		b.WriteString(" \\\n  -H " + shellQuote(k+": "+headers[k]))
	}
	if body != nil {
		b.WriteString(" \\\n  -d " + shellQuote(string(body)))
	}
	fmt.Fprintln(c.DryRunOut, b.String())
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Deterministic dry-run output — never map-iteration order.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
