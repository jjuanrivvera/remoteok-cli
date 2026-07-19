package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/remoteok-cli/internal/api"
	"github.com/jjuanrivvera/remoteok-cli/internal/config"
	"github.com/jjuanrivvera/remoteok-cli/internal/version"
)

// doctorCheck is one diagnostic result.
type doctorCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		var jsonOut bool
		cmd := &cobra.Command{
			Use:   "doctor",
			Short: "Diagnose configuration and Remote OK connectivity",
			Long: `Run local and remote health checks: config file, resolved base URL and User-Agent,
and a live fetch of the Remote OK feed. Exits non-zero when any check fails, so it is
scriptable.`,
			Example: `  remoteok doctor
  remoteok doctor --json`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				checks := d.runDoctor(cmd)
				failed := false
				for _, c := range checks {
					if !c.OK {
						failed = true
					}
				}
				if jsonOut {
					b, err := json.MarshalIndent(checks, "", "  ")
					if err != nil {
						return err
					}
					fmt.Fprintln(cmd.OutOrStdout(), string(b))
				} else {
					for _, c := range checks {
						mark := "✓"
						if !c.OK {
							mark = "✗"
						}
						fmt.Fprintf(cmd.OutOrStdout(), "%s %-12s %s\n", mark, c.Name, c.Detail)
					}
				}
				if failed {
					return fmt.Errorf("doctor found problems")
				}
				return nil
			},
		}
		cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
		return cmd
	})
}

func (d *deps) runDoctor(cmd *cobra.Command) []doctorCheck {
	var checks []doctorCheck
	add := func(name string, ok bool, detail string) {
		checks = append(checks, doctorCheck{Name: name, OK: ok, Detail: detail})
	}

	add("version", true, version.String())

	cfgPath, err := config.Path()
	if err != nil {
		add("config", false, err.Error())
		return checks
	}
	add("config", true, cfgPath)

	c, _, err := d.getAPIClient()
	if err != nil {
		add("base-url", false, err.Error())
		return checks
	}
	add("base-url", true, c.BaseURL())
	add("user-agent", true, uaSummary(c.UserAgent()))

	jobs, _, _, err := c.Jobs(cmd.Context(), api.JobListOptions{Limit: 1})
	if err != nil {
		add("feed", false, err.Error())
	} else {
		add("feed", true, fmt.Sprintf("GET /api OK (%d job fetched)", len(jobs)))
	}
	return checks
}

// uaSummary shortens a long User-Agent for the doctor line.
func uaSummary(ua string) string {
	const max = 48
	if len(ua) > max {
		return ua[:max] + "…"
	}
	return ua
}
