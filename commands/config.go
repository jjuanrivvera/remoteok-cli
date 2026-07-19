package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/jjuanrivvera/remoteok-cli/internal/config"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		cfgCmd := &cobra.Command{
			Use:   "config",
			Short: "Inspect and edit remoteok configuration",
			Long: `The config file holds only non-secret settings (base URL, User-Agent, aliases).
Remote OK is an unauthenticated public API, so nothing here is ever a secret.`,
		}
		cfgCmd.AddCommand(
			newConfigPathCmd(),
			newConfigViewCmd(d),
			newConfigSetCmd(d),
		)
		return cfgCmd
	})
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
}

func newConfigViewCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Show the resolved configuration",
		Long:  "Print the config as YAML. No secrets are stored in the config, so nothing needs redacting.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			b, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(b)
			return err
		},
	}
}

// configSetKeys are the keys `config set` accepts.
var configSetKeys = map[string]func(*config.Config, string){
	"base_url":   func(c *config.Config, v string) { c.BaseURL = v },
	"user_agent": func(c *config.Config, v string) { c.UserAgent = v },
}

func newConfigSetCmd(d *deps) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a non-secret option (base_url, user_agent)",
		Long: `Set a non-secret option in the config file.
Keys: base_url (API endpoint override), user_agent (browser UA sent on every request —
Remote OK 403s a default/bot UA, so override only with a real browser string).`,
		Example: `  remoteok config set user_agent "Mozilla/5.0 (X11; Linux x86_64) …"
  remoteok config set base_url https://remoteok.com`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			setter, ok := configSetKeys[args[0]]
			if !ok {
				return fmt.Errorf("unknown key %q (want base_url|user_agent)", args[0])
			}
			if args[0] == "base_url" && args[1] != "" {
				if err := config.ValidateBaseURL(args[1]); err != nil {
					return err
				}
			}
			cfg, err := d.loadConfig()
			if err != nil {
				return err
			}
			setter(cfg, args[1])
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s = %q\n", args[0], args[1])
			return nil
		},
	}
}
