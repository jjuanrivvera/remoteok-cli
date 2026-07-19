package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleJobs = `[
  {"id":"1001","position":"Senior Go Engineer","company":"Acme","location":"Worldwide","tags":["golang"],"salary_max":180000,"date":"2026-07-18"},
  {"id":"1003","position":"Platform SRE","company":"Initech","location":"US only","tags":["golang","aws"],"salary_max":200000,"date":"2026-07-16"}
]`

func render(t *testing.T, data string, opts Options) (string, string) {
	t.Helper()
	var out, errB bytes.Buffer
	opts.Out = &out
	opts.Err = &errB
	require.NoError(t, Render(json.RawMessage(data), opts))
	return out.String(), errB.String()
}

func TestRender_Table(t *testing.T) {
	out, _ := render(t, sampleJobs, Options{Format: FormatTable, NoColor: true})
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "POSITION")
	assert.Contains(t, out, "Senior Go Engineer")
	// Preferred order: id leads position leads company.
	assert.Less(t, strings.Index(out, "POSITION"), strings.Index(out, "COMPANY"))
}

func TestRender_JSON(t *testing.T) {
	out, _ := render(t, sampleJobs, Options{Format: FormatJSON})
	assert.Contains(t, out, `"position": "Senior Go Engineer"`)
}

func TestRender_YAML(t *testing.T) {
	out, _ := render(t, sampleJobs, Options{Format: FormatYAML})
	assert.Contains(t, out, "position: Senior Go Engineer")
}

func TestRender_CSV(t *testing.T) {
	out, _ := render(t, sampleJobs, Options{Format: FormatCSV, Columns: []string{"id", "company"}})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.Equal(t, "id,company", lines[0])
	assert.Equal(t, "1001,Acme", lines[1])
}

func TestRender_CSVFormulaInjection(t *testing.T) {
	data := `[{"id":"=cmd()","company":"+evil"}]`
	out, _ := render(t, data, Options{Format: FormatCSV, Columns: []string{"id", "company"}})
	assert.Contains(t, out, "'=cmd()")
	assert.Contains(t, out, "'+evil")
}

func TestRender_ID(t *testing.T) {
	out, _ := render(t, sampleJobs, Options{Format: FormatID})
	assert.Equal(t, "1001\n1003\n", out)
}

func TestRender_JQ(t *testing.T) {
	out, _ := render(t, sampleJobs, Options{Format: FormatJSON, JQ: ".[].company"})
	assert.Contains(t, out, "Acme")
	assert.Contains(t, out, "Initech")
}

func TestRender_JQInvalid(t *testing.T) {
	var out, errB bytes.Buffer
	err := Render(json.RawMessage(sampleJobs), Options{Format: FormatJSON, JQ: ".[", Out: &out, Err: &errB})
	assert.Error(t, err)
}

func TestFormat_Valid(t *testing.T) {
	for _, f := range []Format{FormatTable, FormatJSON, FormatYAML, FormatCSV, FormatID} {
		assert.True(t, f.Valid())
	}
	assert.False(t, Format("xml").Valid())
}

func TestSanitizeTerminal(t *testing.T) {
	assert.Equal(t, "plain", SanitizeTerminal("plain"))
	// A full OSC title-set sequence (payload included) is stripped entirely.
	assert.Equal(t, "", SanitizeTerminal("\x1b]0;pwned\a"))
	// A CSI color sequence is removed but its wrapped text survives.
	assert.Equal(t, "red", SanitizeTerminal("\x1b[31mred\x1b[0m"))
}
