package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/remoteok-cli/internal/api"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		cmd := &cobra.Command{
			Use:     "init",
			Aliases: []string{"setup"},
			Short:   "First-run setup: write a config file and smoke-test the feed",
			Long: `Remote OK is a free public API — there is no account or token to set up. This
wizard writes a config file (creating the config directory) and runs a live
connectivity check so you know the feed is reachable from your network.

Pass --user-agent to pin a custom browser User-Agent in the config (Remote OK 403s
a default/bot UA; the CLI already ships a working default).`,
			Example: `  remoteok init
  remoteok init --user-agent "Mozilla/5.0 (X11; Linux x86_64) …"`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				cfg, err := d.loadConfig()
				if err != nil {
					return err
				}
				if d.gf.userAgent != "" {
					cfg.UserAgent = d.gf.userAgent
				}
				if err := cfg.Save(); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Wrote config to %s\n", cfg.FilePath())

				c, _, err := d.getAPIClient()
				if err != nil {
					return err
				}
				jobs, _, _, err := c.Jobs(cmd.Context(), api.JobListOptions{Limit: 1})
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "connectivity check failed: %v\n", err)
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Feed reachable (%d job fetched).\n\nTry:\n  remoteok jobs list --tag golang --limit 20\n", len(jobs))
				return nil
			},
		}
		return cmd
	})
}
