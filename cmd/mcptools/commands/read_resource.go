package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/cobra"
)

// ReadResourceCmd creates the read-resource command.
func ReadResourceCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "read-resource resource [command args...]",
		Short:              "Read a resource on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == FlagHelp || args[0] == FlagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: resource name is required")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp read-resource test://static/resource/1 npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			cmdArgs := args
			parsedArgs := []string{}
			resourceName := ""

			i := 0
			resourceExtracted := false

			for i < len(cmdArgs) {
				switch {
				case (cmdArgs[i] == FlagFormat || cmdArgs[i] == FlagFormatShort) && i+1 < len(cmdArgs):
					FormatOption = cmdArgs[i+1]
					i += 2
				case (cmdArgs[i] == FlagParams || cmdArgs[i] == FlagParamsShort) && i+1 < len(cmdArgs):
					ParamsString = cmdArgs[i+1]
					i += 2
				case !resourceExtracted:
					resourceName = cmdArgs[i]
					resourceExtracted = true
					i++
				default:
					parsedArgs = append(parsedArgs, cmdArgs[i])
					i++
				}
			}

			if resourceName == "" {
				fmt.Fprintln(os.Stderr, "Error: resource name is required")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp read-resource test://static/resource/1 npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			mcpClient, clientErr := CreateClientFunc(parsedArgs)
			if clientErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", clientErr)
				os.Exit(1)
			}

			request := mcp.ReadResourceRequest{}
			request.Params.URI = resourceName
			resp, execErr := mcpClient.ReadResource(context.Background(), request)

			var responseMap map[string]any
			if execErr == nil && resp != nil {
				responseMap = ConvertJSONToMap(resp)
			} else {
				responseMap = map[string]any{}
			}

			if formatErr := FormatAndPrintResponse(thisCmd, responseMap, execErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}
