// Command remoteok is a read-only CLI for the Remote OK jobs API.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/jjuanrivvera/remoteok-cli/commands"
	"github.com/jjuanrivvera/remoteok-cli/internal/output"
	"github.com/jjuanrivvera/remoteok-cli/internal/version"
)

// run builds and executes the command tree for args, writing any error to stderr and
// returning the process exit code. Split from main so it is testable without os.Exit.
func run(ctx context.Context, args []string, stderr io.Writer) int {
	root := commands.NewRootCmd()
	root.Version = version.Get().Version
	root.SetVersionTemplate(version.String() + "\n")

	// Expand user-defined aliases BEFORE cobra parses, so an alias can map to any command
	// without shadowing a built-in.
	root.SetArgs(commands.ExpandAliases(args))

	if err := root.ExecuteContext(ctx); err != nil {
		// Error text can carry API-returned free text (a company name, a job title); strip
		// terminal escapes before printing so a crafted value can't hijack the terminal.
		fmt.Fprintln(stderr, "Error:", output.SanitizeTerminal(err.Error()))
		return 1
	}
	return 0
}

func main() {
	// signal.NotifyContext makes Ctrl-C (SIGINT/SIGTERM) cancel in-flight work: the feed
	// fetch and retry backoff all observe this context.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	os.Exit(run(ctx, os.Args[1:], os.Stderr))
}
