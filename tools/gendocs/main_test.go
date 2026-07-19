package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "commands")
	require.NoError(t, generate(out))
	// The root reference and the jobs subcommand must be generated.
	assert.FileExists(t, filepath.Join(out, "remoteok.md"))
	assert.FileExists(t, filepath.Join(out, "remoteok_jobs_list.md"))
	entries, err := os.ReadDir(out)
	require.NoError(t, err)
	assert.NotEmpty(t, entries)
}
