package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID_UnmarshalMarshal(t *testing.T) {
	cases := []struct {
		in   string
		want ID
	}{
		{`"abc"`, "abc"},
		{`123`, "123"},
		{`"123"`, "123"},
		{`9007199254740993`, "9007199254740993"}, // > 2^53 — no float rounding
		{`null`, ""},
		{`""`, ""},
	}
	for _, tc := range cases {
		var id ID
		require.NoError(t, json.Unmarshal([]byte(tc.in), &id), tc.in)
		assert.Equal(t, tc.want, id, tc.in)
		// Always marshals back as a string.
		b, err := json.Marshal(id)
		require.NoError(t, err)
		assert.Equal(t, `"`+string(tc.want)+`"`, string(b))
	}

	var bad ID
	assert.Error(t, json.Unmarshal([]byte(`{}`), &bad))
	assert.Error(t, json.Unmarshal([]byte(`1.2.3`), &bad))
}

func TestInt_Unmarshal(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{`42`, 42},
		{`"42"`, 42},
		{`0`, 0},
		{`""`, 0},
		{`null`, 0},
		{`180000.0`, 180000},
	}
	for _, tc := range cases {
		var n Int
		require.NoError(t, json.Unmarshal([]byte(tc.in), &n), tc.in)
		assert.Equal(t, tc.want, n.Int64(), tc.in)
	}
	var bad Int
	assert.Error(t, json.Unmarshal([]byte(`"nope"`), &bad))
}

// FuzzID ensures the flexible ID decoder never panics on arbitrary JSON tokens.
func FuzzID(f *testing.F) {
	for _, s := range []string{`"x"`, `1`, `null`, `""`, `123456789012345`, `{}`, `[]`, `1e3`} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		var id ID
		if err := json.Unmarshal([]byte(s), &id); err == nil {
			// A successful decode must round-trip to a JSON string.
			b, err := json.Marshal(id)
			require.NoError(t, err)
			assert.True(t, len(b) >= 2 && b[0] == '"')
		}
	})
}

// FuzzInt ensures the flexible Int decoder never panics.
func FuzzInt(f *testing.F) {
	for _, s := range []string{`42`, `"42"`, `null`, `""`, `1.5`, `{}`, `-3`} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		var n Int
		_ = json.Unmarshal([]byte(s), &n) // must not panic
	})
}
