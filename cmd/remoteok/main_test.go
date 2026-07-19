package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_Version(t *testing.T) {
	var errB bytes.Buffer
	code := run(context.Background(), []string{"version"}, &errB)
	assert.Equal(t, 0, code)
	assert.Empty(t, errB.String())
}

func TestRun_UnknownCommand(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	var errB bytes.Buffer
	code := run(context.Background(), []string{"definitely-not-a-command"}, &errB)
	assert.Equal(t, 1, code)
	assert.Contains(t, errB.String(), "Error:")
}
