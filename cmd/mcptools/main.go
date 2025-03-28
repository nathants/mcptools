/*
Package main implements mcp functionality.
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/f/mcptools/pkg/client"
	"github.com/f/mcptools/pkg/jsonutils"
	"github.com/f/mcptools/pkg/mock"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

// version information placeholders.
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// flags.
const (
	flagFormat       = "--format"
	flagFormatShort  = "-f"
	flagParams       = "--params"
	flagParamsShort  = "-p"
	flagHelp         = "--help"
	flagHelpShort    = "-h"
	entityTypeTool   = "tool"
	entityTypePrompt = "prompt"
	entityTypeRes    = "resource"
)

var (
	formatOption string
	paramsString string
)

// sentinel errors.
var (
	errCommandRequired = fmt.Errorf("command to execute is required when using stdio transport")
)

func main() {
	cobra.EnableCommandSorting = false

	rootCmd := newRootCmd()
	rootCmd.AddCommand(
		newVersionCmd(),
		newToolsCmd(),
		newResourcesCmd(),
		newPromptsCmd(),
		newCallCmd(),
		newGetPromptCmd(),
		newReadResourceCmd(),
		newShellCmd(),
		newMockCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP is a command line interface for interacting with MCP servers",
		Long: `MCP is a command line interface for interacting with Model Context Protocol (MCP) servers.
It allows you to discover and call tools, list resources, and interact with MCP-compatible services.`,
	}

	cmd.PersistentFlags().StringVarP(&formatOption, "format", "f", "table", "Output format (table, json, pretty)")
	cmd.PersistentFlags().
		StringVarP(&paramsString, "params", "p", "{}", "JSON string of parameters to pass to the tool (for call command)")

	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("MCP version %s (built at %s)\n", Version, BuildTime)
		},
	}
}

func createClient(args []string) (*client.Client, error) {
	if len(args) == 0 {
		return nil, errCommandRequired
	}

	if len(args) == 1 && (strings.HasPrefix(args[0], "http://") || strings.HasPrefix(args[0], "https://")) {
		return client.NewHTTP(args[0]), nil
	}

	return client.NewStdio(args), nil
}

func processFlags(args []string) []string {
	parsedArgs := []string{}

	i := 0
	for i < len(args) {
		switch {
		case (args[i] == flagFormat || args[i] == flagFormatShort) && i+1 < len(args):
			formatOption = args[i+1]
			i += 2
		default:
			parsedArgs = append(parsedArgs, args[i])
			i++
		}
	}

	return parsedArgs
}

func formatAndPrintResponse(resp map[string]any, err error) error {
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	output, err := jsonutils.Format(resp, formatOption)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	fmt.Println(output)
	return nil
}

func newToolsCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "tools [command args...]",
		Short:              "List available tools on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == flagHelp || args[0] == flagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			parsedArgs := processFlags(args)

			mcpClient, err := createClient(parsedArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "Example: mcp tools npx -y @modelcontextprotocol/server-filesystem ~\n")
				os.Exit(1)
			}

			resp, listErr := mcpClient.ListTools()
			if formatErr := formatAndPrintResponse(resp, listErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}

func newResourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "resources [command args...]",
		Short:              "List available resources on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == flagHelp || args[0] == flagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			parsedArgs := processFlags(args)

			mcpClient, err := createClient(parsedArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "Example: mcp resources npx -y @modelcontextprotocol/server-filesystem ~\n")
				os.Exit(1)
			}

			resp, listErr := mcpClient.ListResources()
			if formatErr := formatAndPrintResponse(resp, listErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}

func newPromptsCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "prompts [command args...]",
		Short:              "List available prompts on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == flagHelp || args[0] == flagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			parsedArgs := processFlags(args)

			mcpClient, err := createClient(parsedArgs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "Example: mcp prompts npx -y @modelcontextprotocol/server-filesystem ~\n")
				os.Exit(1)
			}

			resp, listErr := mcpClient.ListPrompts()
			if formatErr := formatAndPrintResponse(resp, listErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}

func newCallCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "call entity [command args...]",
		Short:              "Call a tool, resource, or prompt on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == flagHelp || args[0] == flagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: entity name is required")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp call read_file npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			cmdArgs := args
			parsedArgs := []string{}
			entityName := ""

			i := 0
			entityExtracted := false

			for i < len(cmdArgs) {
				switch {
				case (cmdArgs[i] == flagFormat || cmdArgs[i] == flagFormatShort) && i+1 < len(cmdArgs):
					formatOption = cmdArgs[i+1]
					i += 2
				case (cmdArgs[i] == flagParams || cmdArgs[i] == flagParamsShort) && i+1 < len(cmdArgs):
					paramsString = cmdArgs[i+1]
					i += 2
				case !entityExtracted:
					entityName = cmdArgs[i]
					entityExtracted = true
					i++
				default:
					parsedArgs = append(parsedArgs, cmdArgs[i])
					i++
				}
			}

			if entityName == "" {
				fmt.Fprintln(os.Stderr, "Error: entity name is required")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp call read_file npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			entityType := entityTypeTool

			parts := strings.SplitN(entityName, ":", 2)
			if len(parts) == 2 {
				entityType = parts[0]
				entityName = parts[1]
			}

			if len(parsedArgs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: command to execute is required when using stdio transport")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp call read_file npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			var params map[string]any
			if paramsString != "" {
				if jsonErr := json.Unmarshal([]byte(paramsString), &params); jsonErr != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", jsonErr)
					os.Exit(1)
				}
			}

			mcpClient, clientErr := createClient(parsedArgs)
			if clientErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", clientErr)
				os.Exit(1)
			}

			var resp map[string]any
			var execErr error

			switch entityType {
			case entityTypeTool:
				resp, execErr = mcpClient.CallTool(entityName, params)
			case entityTypeRes:
				resp, execErr = mcpClient.ReadResource(entityName)
			case entityTypePrompt:
				resp, execErr = mcpClient.GetPrompt(entityName)
			default:
				fmt.Fprintf(os.Stderr, "Error: unsupported entity type: %s\n", entityType)
				os.Exit(1)
			}

			if formatErr := formatAndPrintResponse(resp, execErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}

func newGetPromptCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "get-prompt prompt [command args...]",
		Short:              "Get a prompt on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == flagHelp || args[0] == flagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: prompt name is required")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp get-prompt read_file npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			cmdArgs := args
			parsedArgs := []string{}
			promptName := ""

			i := 0
			promptExtracted := false

			for i < len(cmdArgs) {
				switch {
				case (cmdArgs[i] == flagFormat || cmdArgs[i] == flagFormatShort) && i+1 < len(cmdArgs):
					formatOption = cmdArgs[i+1]
					i += 2
				case (cmdArgs[i] == flagParams || cmdArgs[i] == flagParamsShort) && i+1 < len(cmdArgs):
					paramsString = cmdArgs[i+1]
					i += 2
				case !promptExtracted:
					promptName = cmdArgs[i]
					promptExtracted = true
					i++
				default:
					parsedArgs = append(parsedArgs, cmdArgs[i])
					i++
				}
			}

			if promptName == "" {
				fmt.Fprintln(os.Stderr, "Error: prompt name is required")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp get-prompt read_file npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			var params map[string]any
			if paramsString != "" {
				if jsonErr := json.Unmarshal([]byte(paramsString), &params); jsonErr != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", jsonErr)
					os.Exit(1)
				}
			}

			mcpClient, clientErr := createClient(parsedArgs)
			if clientErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", clientErr)
				os.Exit(1)
			}

			resp, execErr := mcpClient.GetPrompt(promptName)
			if formatErr := formatAndPrintResponse(resp, execErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}

func newReadResourceCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "read-resource resource [command args...]",
		Short:              "Read a resource on the MCP server",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == flagHelp || args[0] == flagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: resource name is required")
				fmt.Fprintln(
					os.Stderr,
					"Example: mcp read-resource npx -y @modelcontextprotocol/server-filesystem ~",
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
				case (cmdArgs[i] == flagFormat || cmdArgs[i] == flagFormatShort) && i+1 < len(cmdArgs):
					formatOption = cmdArgs[i+1]
					i += 2
				case (cmdArgs[i] == flagParams || cmdArgs[i] == flagParamsShort) && i+1 < len(cmdArgs):
					paramsString = cmdArgs[i+1]
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
					"Example: mcp read-resource npx -y @modelcontextprotocol/server-filesystem ~",
				)
				os.Exit(1)
			}

			var params map[string]any
			if len(parsedArgs) > 0 {
				if jsonErr := json.Unmarshal([]byte(strings.Join(parsedArgs, " ")), &params); jsonErr != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", jsonErr)
					os.Exit(1)
				}
			}

			mcpClient, clientErr := createClient(parsedArgs)
			if clientErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", clientErr)
				os.Exit(1)
			}

			resp, execErr := mcpClient.ReadResource(resourceName)
			if formatErr := formatAndPrintResponse(resp, execErr); formatErr != nil {
				fmt.Fprintf(os.Stderr, "%v\n", formatErr)
				os.Exit(1)
			}
		},
	}
}

func newShellCmd() *cobra.Command { //nolint:gocyclo
	return &cobra.Command{
		Use:                "shell [command args...]",
		Short:              "Start an interactive shell for MCP commands",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == flagHelp || args[0] == flagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			cmdArgs := args
			parsedArgs := []string{}

			i := 0
			for i < len(cmdArgs) {
				switch {
				case (cmdArgs[i] == flagFormat || cmdArgs[i] == flagFormatShort) && i+1 < len(cmdArgs):
					formatOption = cmdArgs[i+1]
					i += 2
				default:
					parsedArgs = append(parsedArgs, cmdArgs[i])
					i++
				}
			}

			if len(parsedArgs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: command to execute is required when using the shell")
				fmt.Fprintln(os.Stderr, "Example: mcp shell npx -y @modelcontextprotocol/server-filesystem ~")
				os.Exit(1)
			}

			mcpClient, clientErr := createClient(parsedArgs)
			if clientErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", clientErr)
				os.Exit(1)
			}

			_, listErr := mcpClient.ListTools()
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "Error connecting to MCP server: %v\n", listErr)
				os.Exit(1)
			}

			fmt.Println("mcp tools shell")
			fmt.Println("connected to:", strings.Join(parsedArgs, " "))
			fmt.Println("\nmcp > Type '/h' for help or '/q' to quit")

			line := liner.NewLiner()
			defer func() { _ = line.Close() }()

			historyFile := filepath.Join(os.Getenv("HOME"), ".mcp_history")
			if f, err := os.Open(filepath.Clean(historyFile)); err == nil {
				_, _ = line.ReadHistory(f)
				_ = f.Close()
			}

			defer func() {
				if f, err := os.Create(historyFile); err == nil {
					_, _ = line.WriteHistory(f)
					_ = f.Close()
				}
			}()

			line.SetCompleter(func(line string) (c []string) {
				commands := []string{
					"tools",
					"resources",
					"prompts",
					"call",
					"format",
					"help",
					"exit",
					"/h",
					"/q",
					"/help",
					"/quit",
				}
				for _, cmd := range commands {
					if strings.HasPrefix(cmd, line) {
						c = append(c, cmd)
					}
				}
				return
			})

			for {
				input, err := line.Prompt("mcp > ")
				if err != nil {
					if errors.Is(err, liner.ErrPromptAborted) {
						fmt.Println("Exiting MCP shell")
						break
					}
					fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
					break
				}

				if input == "" {
					continue
				}

				line.AppendHistory(input)

				if input == "/q" || input == "/quit" || input == "exit" {
					fmt.Println("Exiting MCP shell")
					break
				}

				if input == "/h" || input == "/help" || input == "help" {
					printShellHelp()
					continue
				}

				parts := strings.Fields(input)
				if len(parts) == 0 {
					continue
				}

				command := parts[0]
				commandArgs := parts[1:]

				var resp map[string]any
				var respErr error

				switch command {
				case "tools":
					resp, respErr = mcpClient.ListTools()
					if respErr != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", respErr)

						continue
					}

					output, formatErr := jsonutils.Format(resp, formatOption)
					if formatErr != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", formatErr)

						continue
					}

					fmt.Println(output)

				case "resources":
					resp, respErr = mcpClient.ListResources()
					if respErr != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", respErr)

						continue
					}

					output, formatErr := jsonutils.Format(resp, formatOption)
					if formatErr != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", formatErr)

						continue
					}

					fmt.Println(output)

				case "prompts":
					resp, respErr = mcpClient.ListPrompts()
					if respErr != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", respErr)

						continue
					}

					output, formatErr := jsonutils.Format(resp, formatOption)
					if formatErr != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", formatErr)

						continue
					}

					fmt.Println(output)

				case "call":
					if len(commandArgs) < 1 {
						fmt.Println("Usage: call <entity> [--params '{...}']")

						continue
					}

					entityName := commandArgs[0]
					entityType := entityTypeTool

					parts = strings.SplitN(entityName, ":", 2)
					if len(parts) == 2 {
						entityType = parts[0]
						entityName = parts[1]
					}

					params := map[string]any{}
					for ii := 1; ii < len(commandArgs); ii++ {
						if commandArgs[ii] == flagParams || commandArgs[ii] == flagParamsShort {
							if ii+1 < len(commandArgs) {
								if jsonErr := json.Unmarshal([]byte(commandArgs[ii+1]), &params); jsonErr != nil {
									fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", jsonErr)

									continue
								}
								break
							}
						}
					}

					var execErr error

					switch entityType {
					case entityTypeTool:
						resp, execErr = mcpClient.CallTool(entityName, params)
					case entityTypeRes:
						resp, execErr = mcpClient.ReadResource(entityName)
					case entityTypePrompt:
						resp, execErr = mcpClient.GetPrompt(entityName)
					default:
						fmt.Fprintf(os.Stderr, "Error: unsupported entity type: %s\n", entityType)
						continue
					}

					if execErr != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)
						continue
					}

					output, formatErr := jsonutils.Format(resp, formatOption)
					if formatErr != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", formatErr)
						continue
					}

					fmt.Println(output)

				case "format":
					if len(commandArgs) < 1 {
						fmt.Printf("Current format: %s\n", formatOption)
						continue
					}

					newFormat := commandArgs[0]
					if newFormat == "json" || newFormat == "j" ||
						newFormat == "pretty" || newFormat == "p" ||
						newFormat == "table" || newFormat == "t" {
						formatOption = newFormat
						fmt.Printf("Format set to: %s\n", formatOption)
					} else {
						fmt.Println("Invalid format. Use: table, json, or pretty")
					}

				default:
					entityName := command
					entityType := entityTypeTool

					parts = strings.SplitN(entityName, ":", 2)
					if len(parts) == 2 {
						entityType = parts[0]
						entityName = parts[1]
					}

					params := map[string]any{}

					if len(commandArgs) > 0 {
						firstArg := commandArgs[0]
						if strings.HasPrefix(firstArg, "{") && strings.HasSuffix(firstArg, "}") {
							if jsonErr := json.Unmarshal([]byte(firstArg), &params); jsonErr != nil {
								fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", jsonErr)
								continue
							}
						} else {
							for iii := 0; iii < len(commandArgs); iii++ {
								if commandArgs[iii] == flagParams || commandArgs[iii] == flagParamsShort {
									if iii+1 < len(commandArgs) {
										if jsonErr := json.Unmarshal([]byte(commandArgs[iii+1]), &params); jsonErr != nil {
											fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", jsonErr)
											continue
										}
										break
									}
								}
							}
						}
					}

					var execErr error

					switch entityType {
					case entityTypeTool:
						resp, execErr = mcpClient.CallTool(entityName, params)
					case entityTypeRes:
						resp, execErr = mcpClient.ReadResource(entityName)
					case entityTypePrompt:
						resp, execErr = mcpClient.GetPrompt(entityName)
					default:
						fmt.Printf("Unknown command: %s\nType '/h' for help\n", command)
						continue
					}

					if execErr != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)
						continue
					}

					output, formatErr := jsonutils.Format(resp, formatOption)
					if formatErr != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", formatErr)
						continue
					}

					fmt.Println(output)
				}
			}
		},
	}
}

func printShellHelp() {
	fmt.Println("MCP Shell Commands:")
	fmt.Println("  tools                      List available tools")
	fmt.Println("  resources                  List available resources")
	fmt.Println("  prompts                    List available prompts")
	fmt.Println("  call <entity> [--params '{...}']  Call a tool, resource, or prompt")
	fmt.Println("  format [json|pretty|table] Get or set output format")
	fmt.Println("Direct Tool Calling:")
	fmt.Println("  <tool_name> {\"param\": \"value\"}  Call a tool directly with JSON parameters")
	fmt.Println("  resource:<name>            Read a resource directly")
	fmt.Println("  prompt:<name>              Get a prompt directly")
	fmt.Println("Special Commands:")
	fmt.Println("  /h, /help                  Show this help")
	fmt.Println("  /q, /quit, exit            Exit the shell")
}

func newMockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock [type] [name] [description] [content]...",
		Short: "Create a mock MCP server with tools, prompts, and resources",
		Long: `Create a mock MCP server with tools, prompts, and resources.
This is useful for testing MCP clients without implementing a full server.

The mock server implements the MCP protocol with:
- Full initialization handshake (initialize method)
- Support for notifications/initialized notification
- Tool listing with standardized schema format
- Tool calling with simple responses
- Resource listing and reading with proper format
- Prompt listing and retrieving with proper format
- Standard error codes (-32601 for method not found)

Available types:
- tool <name> <description>
- prompt <name> <description> <template>
- resource <uri> <description> <content>

Example: 
  mcp mock tool hello_world "when user says hello world, run this tool"
  mcp mock tool hello_world "A greeting tool" \
         prompt welcome "A welcome prompt" "Hello {{name}}, welcome to the system!" \
         resource docs:readme "Documentation" "# Mock MCP Server\nThis is a mock server"`,
		Args: cobra.MinimumNArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			tools := make(map[string]string)
			prompts := make(map[string]map[string]string)
			resources := make(map[string]map[string]string)

			i := 0
			for i < len(args) {
				entityType := args[i]
				i++

				switch entityType {
				case entityTypeTool:
					if i+1 >= len(args) {
						fmt.Fprintln(os.Stderr, "Error: each tool must have both a name and description")
						fmt.Fprintln(os.Stderr, "Example: mcp mock tool hello_world \"when user says hello world, run this tool\"")
						os.Exit(1)
					}

					toolName := args[i]
					toolDescription := args[i+1]
					tools[toolName] = toolDescription
					fmt.Fprintf(os.Stderr, "Added tool: %s - %s\n", toolName, toolDescription)
					i += 2

				case entityTypePrompt:
					if i+2 >= len(args) {
						fmt.Fprintln(os.Stderr, "Error: each prompt must have a name, description, and template")
						fmt.Fprintln(os.Stderr, "Example: mcp mock prompt welcome \"Welcome message\" \"Hello {{name}}!\"")
						os.Exit(1)
					}

					promptName := args[i]
					promptDescription := args[i+1]
					promptTemplate := args[i+2]

					prompts[promptName] = map[string]string{
						"description": promptDescription,
						"template":    promptTemplate,
					}

					fmt.Fprintf(os.Stderr, "Added prompt: %s - %s\n", promptName, promptDescription)
					i += 3

				case entityTypeRes:
					if i+2 >= len(args) {
						fmt.Fprintln(os.Stderr, "Error: each resource must have a URI, description, and content")
						fmt.Fprintln(os.Stderr, "Example: mcp mock resource docs:readme \"Documentation\" \"# README\"")
						os.Exit(1)
					}

					resourceURI := args[i]
					resourceDescription := args[i+1]
					resourceContent := args[i+2]

					resources[resourceURI] = map[string]string{
						"description": resourceDescription,
						"content":     resourceContent,
					}

					fmt.Fprintf(os.Stderr, "Added resource: %s - %s\n", resourceURI, resourceDescription)
					i += 3

				default:
					fmt.Fprintf(os.Stderr, "Error: unknown entity type: %s\n", entityType)
					fmt.Fprintln(os.Stderr, "Available types: tool, prompt, resource")
					os.Exit(1)
				}
			}

			if len(tools) == 0 && len(prompts) == 0 && len(resources) == 0 {
				fmt.Fprintln(os.Stderr, "Error: at least one tool, prompt, or resource must be specified")
				os.Exit(1)
			}

			fmt.Fprintf(os.Stderr, "Starting mock MCP server with %d tool(s), %d prompt(s), and %d resource(s)\n",
				len(tools), len(prompts), len(resources))
			fmt.Fprintf(os.Stderr, "Use Ctrl+C to exit\n")

			if err := mock.RunMockServer(tools, prompts, resources); err != nil {
				fmt.Fprintf(os.Stderr, "Error running mock server: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
