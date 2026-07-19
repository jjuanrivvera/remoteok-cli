// Package commands wires the cobra command tree. root.go owns the global flags, the shared
// Remote OK client factory, and the single render() path used by every command. The tree is
// built fresh per NewRootCmd() call so tests never leak flag state across cases.
package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/remoteok-cli/internal/api"
	"github.com/jjuanrivvera/remoteok-cli/internal/config"
	"github.com/jjuanrivvera/remoteok-cli/internal/output"
)

// globalFlags holds the persistent flag values for one command tree.
type globalFlags struct {
	outputFormat string
	baseURL      string
	userAgent    string
	dryRun       bool
	verbose      bool
	noColor      bool
	columns      []string
	quiet        bool
	jq           string

	// list flags (read by list commands)
	limit int
}

// deps carries the per-tree state into every command builder.
type deps struct {
	gf *globalFlags

	// overridable in tests
	loadConfig func() (*config.Config, error)
	// clientOpts are extra api.Options injected by tests (e.g. WithMaxRetries(0)).
	clientOpts []api.Option
	// out overrides where dry-run curls go (tests capture it; default os.Stdout).
	out io.Writer
}

func newDeps() *deps {
	return &deps{
		gf:         &globalFlags{},
		loadConfig: config.Load,
	}
}

// NewRootCmd assembles the full command tree. main.go calls
// NewRootCmd().ExecuteContext(ctx) with a signal.NotifyContext so Ctrl-C cancels
// in-flight work.
func NewRootCmd() *cobra.Command { return newRootCmd(newDeps()) }

// registrars build the resource commands; resource files append from init().
var registrars []func(d *deps) *cobra.Command

// metaRegistrars register the non-resource commands (config, doctor, alias, …).
var metaRegistrars []func(d *deps) *cobra.Command

// newRootCmd is the deps-injected assembly used by tests (temp config, mock server).
func newRootCmd(d *deps) *cobra.Command {
	root := &cobra.Command{
		Use:   "remoteok",
		Short: "Discover remote jobs from the Remote OK public API",
		Long: `remoteok is a fast, scriptable, read-only client for the Remote OK jobs API
(https://remoteok.com/api). It is built for machine consumption — JSON/YAML/CSV
output, an -o id mode for piping, a --jq filter, and an MCP server — so an AI
assistant or a shell pipeline can discover remote work programmatically.

Remote OK is a free public API and needs no account or token. Its Terms of Service
require a follow backlink to https://remoteok.com when you display their data — the
CLI prints a Source attribution on stderr for that reason.

Examples:
  remoteok jobs list --tag golang --limit 20
  remoteok jobs list --search kubernetes -o json
  remoteok jobs list --tag golang --tag remote --min-salary 100000
  remoteok jobs list --company stripe -o csv
  remoteok jobs get 1135010
  remoteok jobs list -o id | head`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if d.gf.outputFormat != "" && !output.Format(d.gf.outputFormat).Valid() {
				return fmt.Errorf("unknown output format %q (want table|json|yaml|csv|id)", d.gf.outputFormat)
			}
			return nil
		},
	}
	registerGlobalFlags(root, d.gf)

	for _, build := range registrars {
		root.AddCommand(build(d))
	}
	for _, build := range metaRegistrars {
		root.AddCommand(build(d))
	}
	return root
}

func registerGlobalFlags(root *cobra.Command, gf *globalFlags) {
	pf := root.PersistentFlags()
	pf.StringVarP(&gf.outputFormat, "output", "o", "", "output format: table|json|yaml|csv|id")
	pf.StringVar(&gf.baseURL, "base-url", "", "Remote OK base URL override (default https://remoteok.com)")
	pf.StringVar(&gf.userAgent, "user-agent", "", "override the browser User-Agent (Remote OK 403s a default/bot UA)")
	pf.BoolVar(&gf.dryRun, "dry-run", false, "print the equivalent curl and make no request")
	pf.BoolVarP(&gf.verbose, "verbose", "v", false, "verbose request logging (stderr)")
	pf.BoolVar(&gf.noColor, "no-color", false, "disable colored output")
	pf.StringSliceVar(&gf.columns, "columns", nil, "comma-separated columns to show")
	pf.BoolVar(&gf.quiet, "quiet", false, "suppress non-essential chatter (incl. the source attribution)")
	pf.StringVar(&gf.jq, "jq", "", "gojq expression applied to the response before rendering")

	pf.IntVar(&gf.limit, "limit", 0, "max jobs to return (list commands)")
}

// getAPIClient builds a Remote OK client, honoring flag > env > config > default for the
// base URL and User-Agent.
func (d *deps) getAPIClient() (*api.Client, *config.Config, error) {
	cfg, err := d.loadConfig()
	if err != nil {
		return nil, nil, err
	}
	baseURL := config.FirstNonEmpty(d.gf.baseURL, os.Getenv("REMOTEOK_BASE_URL"), cfg.BaseURL, api.DefaultBaseURL)
	if err := config.ValidateBaseURL(baseURL); err != nil {
		return nil, nil, err
	}
	ua := config.FirstNonEmpty(d.gf.userAgent, os.Getenv("REMOTEOK_USER_AGENT"), cfg.UserAgent, api.DefaultUserAgent)

	opts := []api.Option{
		api.WithUserAgent(ua),
		api.WithDryRun(d.gf.dryRun, d.stdout()),
	}
	opts = append(opts, d.clientOpts...)
	c := api.New(baseURL, opts...)
	c.Verbose = d.gf.verbose
	c.VerboseOut = os.Stderr
	return c, cfg, nil
}

func (d *deps) stdout() io.Writer {
	if d.out != nil {
		return d.out
	}
	return os.Stdout
}

// render is the single output path for every command: normalize v to JSON, then hand it to
// the shared renderer with the resolved global flags.
func (d *deps) render(cmd *cobra.Command, v any, defaultColumns []string) error {
	raw, ok := v.(json.RawMessage)
	if !ok {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		raw = b
	}
	format := output.Format(config.FirstNonEmpty(d.gf.outputFormat, string(output.FormatTable)))
	cols := normalizeColumns(d.gf.columns)
	if len(cols) == 0 && format != output.FormatID {
		cols = defaultColumns
	}
	return output.Render(raw, output.Options{
		Format:  format,
		Columns: cols,
		NoColor: d.gf.noColor,
		Quiet:   d.gf.quiet,
		JQ:      d.gf.jq,
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
	})
}

func normalizeColumns(cols []string) []string {
	var out []string
	for _, c := range cols {
		if c = strings.TrimSpace(c); c != "" {
			out = append(out, c)
		}
	}
	return out
}
