package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	metaRegistrars = append(metaRegistrars, func(d *deps) *cobra.Command {
		var data string
		var query []string
		cmd := &cobra.Command{
			Use:   "api <METHOD> <PATH> [-d body] [-q key=value ...]",
			Short: "Send a raw request to the Remote OK API (escape hatch)",
			Long: `Call the Remote OK site directly. The path is relative to the configured base URL
(default https://remoteok.com), e.g. api.

This is the documented escape hatch for anything remoteok does not wrap as a
first-class command. It honors --dry-run, -o/--output, and --jq like every other
command. The public API is read-only; in practice only GET /api serves data.`,
			Example: `  remoteok api GET api
  remoteok api GET api -q tags=golang
  remoteok api GET api --dry-run`,
			Args: cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				method := strings.ToUpper(args[0])
				valid := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}
				if !slices.Contains(valid, method) {
					return fmt.Errorf("invalid method %q (want one of %s)", args[0], strings.Join(valid, "|"))
				}
				path := strings.TrimLeft(args[1], "/")

				q := url.Values{}
				for _, kv := range query {
					k, v, ok := strings.Cut(kv, "=")
					if !ok {
						return fmt.Errorf("invalid -q %q (want key=value)", kv)
					}
					q.Add(k, v)
				}

				var body []byte
				if data != "" {
					raw, err := readDataArg(cmd, data)
					if err != nil {
						return err
					}
					body = raw
				}

				c, _, err := d.getAPIClient()
				if err != nil {
					return err
				}
				status, respBody, err := c.Do(cmd.Context(), method, path, q, body)
				if err != nil {
					return err
				}
				if status == 0 { // dry-run
					return nil
				}
				if len(respBody) == 0 {
					if !d.gf.quiet {
						fmt.Fprintf(cmd.OutOrStdout(), "HTTP %d (empty body)\n", status)
					}
					return nil
				}
				if json.Valid(respBody) {
					return d.render(cmd, json.RawMessage(respBody), nil)
				}
				_, err = cmd.OutOrStdout().Write(respBody)
				return err
			},
		}
		cmd.Flags().StringVarP(&data, "data", "d", "", "request body: inline, @file, or - for stdin")
		cmd.Flags().StringArrayVarP(&query, "query", "q", nil, "query parameter key=value (repeatable)")
		return annotate(cmd, kindWrite) // raw calls could mutate on some APIs; the guard gates by METHOD (§3b.6)
	})
}
