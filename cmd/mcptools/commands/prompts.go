package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// PromptsCmd creates the prompts command.
func PromptsCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "prompts [command args...]",
		Short:              "List available prompts on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == FlagHelp || args[0] == FlagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			parsedArgs := ProcessFlags(args)

			mcpClient, err := CreateClientFunc(parsedArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "Example: mcp prompts npx -y @modelcontextprotocol/server-filesystem ~\n")
				os.Exit(1)
			}

			resp, listErr := mcpClient.ListPrompts(context.Background(), mcp.ListPromptsRequest{})
			promptsMap := map[string]any{"prompts": ConvertJSONToSlice(resp.Prompts)}
			if formatErr := FormatAndPrintResponse(thisCmd, promptsMap, listErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}
