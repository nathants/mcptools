package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version information placeholder.
var Version = "dev"

// TemplatesPath information placeholder.
var TemplatesPath = os.Getenv("HOME") + "/.mcpt/templates"

// VersionCmd creates the version command.
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "MCP Tools version %s\n", Version)
		},
	}
}
