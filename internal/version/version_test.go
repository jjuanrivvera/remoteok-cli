package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert.True(t, strings.HasPrefix(String(), "remoteok "))
}

func TestGet(t *testing.T) {
	info := Get()
	assert.Equal(t, Version, info.Version)
	assert.Equal(t, Commit, info.Commit)
	assert.Equal(t, Date, info.Date)
}
