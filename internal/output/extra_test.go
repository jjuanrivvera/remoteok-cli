package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncCell(t *testing.T) {
	s, cut := truncCell("hello", 10)
	assert.False(t, cut)
	assert.Equal(t, "hello", s)
	s, cut = truncCell("hello world", 5)
	assert.True(t, cut)
	assert.Equal(t, "hell…", s)
}

func TestIsNumeric(t *testing.T) {
	assert.True(t, isNumeric("-123"))
	assert.True(t, isNumeric("12.5"))
	assert.False(t, isNumeric("12x"))
	assert.False(t, isNumeric(""))
}

func TestSanitizeCSV_NegativeNumberKept(t *testing.T) {
	assert.Equal(t, "-5", sanitizeCSV("-5"))      // a real negative is left alone
	assert.Equal(t, "'-cmd", sanitizeCSV("-cmd")) // a leading-dash non-number is quoted
	assert.Equal(t, "plain", sanitizeCSV("plain"))
}

func TestRenderID_FallbackAndNoRows(t *testing.T) {
	var out, errB bytes.Buffer
	// No id-like column → falls back to the first column.
	_ = Render(json.RawMessage(`[{"name":"a"},{"name":"b"}]`), Options{Format: FormatID, Out: &out, Err: &errB})
	assert.Equal(t, "a\nb\n", out.String())

	out.Reset()
	_ = Render(json.RawMessage(`[]`), Options{Format: FormatID, Out: &out, Err: &errB})
	assert.Empty(t, out.String())
}

func TestRenderTable_CappedColumnsNote(t *testing.T) {
	var out, errB bytes.Buffer
	data := `[{"a":"1","b":"2","c":"3"}]`
	_ = Render(json.RawMessage(data), Options{Format: FormatTable, MaxCols: 2, NoColor: true, Out: &out, Err: &errB})
	assert.Contains(t, errB.String(), "columns")
}

func TestRenderTable_ScalarValue(t *testing.T) {
	var out, errB bytes.Buffer
	_ = Render(json.RawMessage(`true`), Options{Format: FormatTable, Out: &out, Err: &errB})
	assert.Equal(t, "true", strings.TrimSpace(out.String()))
}

func TestYamlNormalizeFloat(t *testing.T) {
	var out, errB bytes.Buffer
	_ = Render(json.RawMessage(`{"x":1.5,"y":2}`), Options{Format: FormatYAML, Out: &out, Err: &errB})
	assert.Contains(t, out.String(), "1.5")
}

func TestRender_NilAndEmpty(t *testing.T) {
	var out, errB bytes.Buffer
	assert.NoError(t, Render(nil, Options{Format: FormatJSON, Out: &out, Err: &errB}))
	assert.NoError(t, Render(json.RawMessage(`null`), Options{Format: FormatTable, Out: &out, Err: &errB}))
}
