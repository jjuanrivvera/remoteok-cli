// Command gendocs generates the Markdown command reference under docs/commands from the
// live cobra tree, so the published docs never drift from the actual CLI surface.
package main

import (
	"log"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/jjuanrivvera/remoteok-cli/commands"
)

// generate writes the Markdown command tree into out. Split out from main so it is testable.
func generate(out string) error {
	if err := os.MkdirAll(out, 0o750); err != nil {
		return err
	}
	root := commands.NewRootCmd()
	root.DisableAutoGenTag = true // omit the timestamp so output is reproducible (no CI drift)
	return doc.GenMarkdownTree(root, out)
}

func main() {
	if err := generate("docs/commands"); err != nil {
		log.Fatalf("generate docs: %v", err)
	}
}
