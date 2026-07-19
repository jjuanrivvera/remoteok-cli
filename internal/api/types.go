package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// Flexible JSON types insulate the CLI from Remote OK's loose typing: an id or salary can
// arrive as a JSON number OR a quoted string across the feed, and mixing them in one decode
// would otherwise break the struct. See GOAL.md §2 "Flexible JSON types".

// ID unmarshals from a JSON string OR number and always marshals back as a string, so large
// ids never lose precision above 2^53 and table/JSON rendering is consistent.
type ID string

// UnmarshalJSON accepts "123" or 123.
func (id *ID) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*id = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*id = ID(s)
		return nil
	}
	// Bare number: validate it is a well-formed JSON number, then keep the text as-is.
	if _, err := strconv.ParseFloat(string(b), 64); err != nil {
		return fmt.Errorf("invalid id %q: %w", string(b), err)
	}
	*id = ID(string(b))
	return nil
}

// MarshalJSON always emits a string.
func (id ID) MarshalJSON() ([]byte, error) { return json.Marshal(string(id)) }

// String returns the id as a plain string.
func (id ID) String() string { return string(id) }

// Int accepts a JSON number OR a numeric string and holds an int64. It decodes Int64 before
// Float64 to avoid >2^53 precision loss and rejects NaN/Inf via the JSON-number grammar.
type Int int64

// UnmarshalJSON accepts 42, "42", or "" / null (→ 0).
func (n *Int) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*n = 0
		return nil
	}
	s := string(b)
	if b[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		if str == "" {
			*n = 0
			return nil
		}
		s = str
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		*n = Int(i)
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid integer %q: %w", s, err)
	}
	*n = Int(int64(f))
	return nil
}

// Int64 returns the value as a plain int64.
func (n Int) Int64() int64 { return int64(n) }
