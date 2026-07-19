package api

import (
	"fmt"
	"net/http"
	"strings"
)

// APIError is a Remote OK HTTP error with an actionable hint keyed by status. Remote OK does
// not return a structured error envelope (the endpoint either serves the JSON array or an
// HTML/edge error page), so the body is kept as truncated text for context.
type APIError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func parseAPIError(status int, body []byte, _ http.Header) *APIError {
	msg := http.StatusText(status)
	if msg == "" {
		msg = "request failed"
	}
	return &APIError{StatusCode: status, Message: msg, Body: body}
}

func (e *APIError) Error() string {
	msg := fmt.Sprintf("Remote OK API error %d: %s", e.StatusCode, e.Message)
	if hint := e.hint(); hint != "" {
		msg += "\nHint: " + hint
	}
	if snippet := bodySnippet(e.Body); snippet != "" {
		msg += "\nResponse: " + snippet
	}
	return msg
}

// hint maps a status to the remedy a user actually needs. Remote OK's most common failure is
// a 403 from Cloudflare when the User-Agent is missing or looks like a bot/script.
func (e *APIError) hint() string {
	switch e.StatusCode {
	case http.StatusForbidden:
		return "Remote OK blocks requests without a browser-like User-Agent — set one with `remoteok config set user_agent \"Mozilla/5.0 …\"` (the CLI already sends a default browser UA; a proxy or firewall may be stripping it)"
	case http.StatusNotFound:
		return "endpoint not found — check --base-url (the API lives at https://remoteok.com/api)"
	case http.StatusTooManyRequests:
		return "rate limited — the CLI already honored Retry-After; slow down or cache the feed"
	}
	if e.StatusCode >= 500 {
		return "Remote OK server error — usually transient, retry shortly"
	}
	return ""
}

// bodySnippet returns a short, single-line preview of an error body for context without
// dumping a whole HTML error page into the terminal.
func bodySnippet(body []byte) string {
	s := strings.TrimSpace(string(body))
	if s == "" {
		return ""
	}
	s = strings.Join(strings.Fields(s), " ")
	const max = 200
	if len(s) > max {
		s = s[:max] + "…"
	}
	return s
}
