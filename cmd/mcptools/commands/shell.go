package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/f/mcptools/pkg/jsonutils"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

// ShellCmd creates the shell command.
func ShellCmd() *cobra.Command { //nolint:gocyclo
	return &cobra.Command{
		Use:                "shell [command args...]",
		Short:              "Start an interactive shell for MCP commands",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(thisCmd *cobra.Command, args []string) {
			if len(args) == 1 && (args[0] == FlagHelp || args[0] == FlagHelpShort) {
				_ = thisCmd.Help()
				return
			}

			cmdArgs := args
			parsedArgs := []string{}

			i := 0
			for i < len(cmdArgs) {
				switch {
				case (cmdArgs[i] == FlagFormat || cmdArgs[i] == FlagFormatShort) && i+1 < len(cmdArgs):
					FormatOption = cmdArgs[i+1]
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

			mcpClient, clientErr := CreateClientFunc(parsedArgs)
			if clientErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", clientErr)
				os.Exit(1)
			}

			_, listErr := mcpClient.ListTools()
			if listErr != nil {
				fmt.Fprintf(os.Stderr, "Error connecting to MCP server: %v\n", listErr)
				os.Exit(1)
			}

			fmt.Printf("mcp > MCP Tools Shell (%s)\n", Version)
			fmt.Println("mcp > Connected to Server:", strings.Join(parsedArgs, " "))
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

					output, formatErr := jsonutils.Format(resp, FormatOption)
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

					output, formatErr := jsonutils.Format(resp, FormatOption)
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

					output, formatErr := jsonutils.Format(resp, FormatOption)
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
					entityType := EntityTypeTool

					parts = strings.SplitN(entityName, ":", 2)
					if len(parts) == 2 {
						entityType = parts[0]
						entityName = parts[1]
					}

					params := map[string]any{}
					for ii := 1; ii < len(commandArgs); ii++ {
						if commandArgs[ii] == FlagParams || commandArgs[ii] == FlagParamsShort {
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
					case EntityTypeTool:
						resp, execErr = mcpClient.CallTool(entityName, params)
					case EntityTypeRes:
						resp, execErr = mcpClient.ReadResource(entityName)
					case EntityTypePrompt:
						resp, execErr = mcpClient.GetPrompt(entityName)
					default:
						fmt.Fprintf(os.Stderr, "Error: unsupported entity type: %s\n", entityType)
						continue
					}

					if execErr != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)
						continue
					}

					output, formatErr := jsonutils.Format(resp, FormatOption)
					if formatErr != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", formatErr)
						continue
					}

					fmt.Println(output)

				case "format":
					if len(commandArgs) < 1 {
						fmt.Printf("Current format: %s\n", FormatOption)
						continue
					}

					newFormat := commandArgs[0]
					if newFormat == "json" || newFormat == "j" ||
						newFormat == "pretty" || newFormat == "p" ||
						newFormat == "table" || newFormat == "t" {
						FormatOption = newFormat
						fmt.Printf("Format set to: %s\n", FormatOption)
					} else {
						fmt.Println("Invalid format. Use: table, json, or pretty")
					}

				default:
					entityName := command
					entityType := EntityTypeTool

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
								if commandArgs[iii] == FlagParams || commandArgs[iii] == FlagParamsShort {
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
					case EntityTypeTool:
						resp, execErr = mcpClient.CallTool(entityName, params)
					case EntityTypeRes:
						resp, execErr = mcpClient.ReadResource(entityName)
					case EntityTypePrompt:
						resp, execErr = mcpClient.GetPrompt(entityName)
					default:
						fmt.Printf("Unknown command: %s\nType '/h' for help\n", command)
						continue
					}

					if execErr != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)
						continue
					}

					output, formatErr := jsonutils.Format(resp, FormatOption)
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
