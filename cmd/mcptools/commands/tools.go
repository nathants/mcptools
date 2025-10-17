package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// ToolsCmd creates the tools command.
func ToolsCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "tools [command args...]",
		Short:              "List available tools on the MCP server",
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
				fmt.Fprintf(os.Stderr, "Example: mcp tools npx -y @modelcontextprotocol/server-filesystem ~\n")
				os.Exit(1)
			}
			defer mcpClient.Close()

			resp, listErr := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})

			var tools []any
			if listErr == nil && resp != nil {
				tools = ConvertJSONToSlice(resp.Tools)
			}

			toolsMap := map[string]any{"tools": tools}
			if formatErr := FormatAndPrintResponse(thisCmd, toolsMap, listErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}
