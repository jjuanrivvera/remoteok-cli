package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_SaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	c := &Config{BaseURL: "https://remoteok.com", UserAgent: "ua/1", Aliases: map[string]string{"go": "jobs list --tag golang"}}
	c.path = p
	require.NoError(t, c.Save())

	// File perms are 0600. Windows has no POSIX permission bits (Stat reports 0666),
	// so the assertion only holds on Unix.
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(p)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
	}

	got, err := LoadFrom(p)
	require.NoError(t, err)
	assert.Equal(t, "https://remoteok.com", got.BaseURL)
	assert.Equal(t, "ua/1", got.UserAgent)
	assert.Equal(t, "jobs list --tag golang", got.Aliases["go"])
	assert.Equal(t, p, got.FilePath())
}

func TestLoadFrom_MissingIsEmpty(t *testing.T) {
	c, err := LoadFrom(filepath.Join(t.TempDir(), "nope.yaml"))
	require.NoError(t, err)
	assert.Empty(t, c.BaseURL)
}

func TestDirAndPath_XDG(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	dir, err := Dir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(xdg, "remoteok"), dir)
	p, err := Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(xdg, "remoteok", "config.yaml"), p)
}

func TestFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "b", FirstNonEmpty("", "b", "c"))
	assert.Equal(t, "", FirstNonEmpty("", ""))
}

func TestValidateAliasName(t *testing.T) {
	require.NoError(t, ValidateAliasName("go"))
	for _, bad := range []string{"", " ", ".", "..", "a/b", "a:b", "a#b"} {
		assert.Error(t, ValidateAliasName(bad), bad)
	}
}

func TestValidateBaseURL(t *testing.T) {
	require.NoError(t, ValidateBaseURL("https://remoteok.com"))
	require.NoError(t, ValidateBaseURL("http://localhost:8080"))
	require.NoError(t, ValidateBaseURL("http://127.0.0.1"))
	for _, bad := range []string{"ftp://x", "https://", "http://example.com", "://bad"} {
		assert.Error(t, ValidateBaseURL(bad), bad)
	}
}

func TestLoadAndSaveDefaultPath_XDG(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	// Load() with no file returns an empty config at the default path.
	c, err := Load()
	require.NoError(t, err)
	c.UserAgent = "x/1"
	require.NoError(t, c.Save())

	got, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "x/1", got.UserAgent)
}

func TestDir_HomeDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // os.UserHomeDir reads USERPROFILE on Windows, not HOME
	dir, err := Dir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".remoteok-cli"), dir)
}

func TestLoadFrom_MalformedYAML(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.yaml")
	require.NoError(t, os.WriteFile(p, []byte("::: not yaml :::"), 0o600))
	_, err := LoadFrom(p)
	require.Error(t, err)
}

func TestSave_CreatesDir(t *testing.T) {
	p := filepath.Join(t.TempDir(), "nested", "deep", "config.yaml")
	c := &Config{BaseURL: "https://remoteok.com"}
	c.path = p
	require.NoError(t, c.Save())
	assert.FileExists(t, p)
}
