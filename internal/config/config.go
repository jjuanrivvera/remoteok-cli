// Package config resolves remoteok configuration with a manual flag > env > file > default
// precedence (no Viper, per the cliwright house pattern). Remote OK is a single, fixed,
// unauthenticated public endpoint, so there are no profiles and no secrets: the config file
// holds only non-secret overrides (base URL, User-Agent) and user-defined aliases.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the on-disk configuration. Remote OK needs no credentials, so nothing here is
// ever secret — it is safe to print in full (see `remoteok config view`).
type Config struct {
	BaseURL   string            `yaml:"base_url,omitempty"`   // API base URL override (default https://remoteok.com)
	UserAgent string            `yaml:"user_agent,omitempty"` // User-Agent override (Remote OK blocks default/empty UAs)
	Aliases   map[string]string `yaml:"aliases,omitempty"`    // name -> expansion (e.g. "go" -> "jobs list --tag golang")

	path string `yaml:"-"`
}

// Dir returns the configuration directory: $XDG_CONFIG_HOME/remoteok if set, else
// ~/.remoteok-cli.
func Dir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "remoteok"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".remoteok-cli"), nil
}

// Path returns the config file path.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the config file, returning an empty (but valid) Config if it does not exist.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	return LoadFrom(p)
}

// LoadFrom reads a config from an explicit path (used in tests).
func LoadFrom(p string) (*Config, error) {
	c := &Config{path: p}
	data, err := os.ReadFile(p) //nolint:gosec // G304: p is the user's own config path
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", p, err)
	}
	c.path = p
	return c, nil
}

// Save writes the config atomically: a temp file in the same directory then a rename, with
// dir 0700 and file 0600 so a hand-edited config is never torn on a crash.
func (c *Config) Save() error {
	if c.path == "" {
		p, err := Path()
		if err != nil {
			return err
		}
		c.path = p
	}
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".config-*.yaml.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, c.path)
}

// FilePath returns where this config reads/writes.
func (c *Config) FilePath() string { return c.path }

// FirstNonEmpty returns the first non-empty string — the manual precedence helper used
// across remoteok instead of a config framework.
func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// invalidNameChars are path/traversal-dangerous characters disallowed in an alias name.
const invalidNameChars = `/\:*?"<>|#`

// ValidateAliasName rejects empty names and traversal-dangerous characters, so an alias
// name can never be used to escape the config namespace.
func ValidateAliasName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("alias name cannot be empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("invalid alias name %q", name)
	}
	if strings.ContainsAny(name, invalidNameChars) {
		return fmt.Errorf("alias name %q contains an invalid character (one of %s)", name, invalidNameChars)
	}
	return nil
}

// ValidateBaseURL requires an http/https URL with a host and rejects cleartext http:// for a
// non-loopback host. Remote OK carries no token, but a plain-http base URL still risks a
// downgrade/MITM of the fetched listings, so the same guard applies.
func ValidateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid base URL %q: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("base URL must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("base URL %q has no host", raw)
	}
	if u.Scheme == "http" && !isLoopback(u.Hostname()) {
		return fmt.Errorf("refusing cleartext http:// for non-loopback host %q; use https", u.Hostname())
	}
	return nil
}

func isLoopback(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	}
	return strings.HasPrefix(host, "127.")
}
