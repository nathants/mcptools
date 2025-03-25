package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/f/mcptools/pkg/client"
	"github.com/f/mcptools/pkg/formatter"
	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

var (
	// Version is set during build
	Version = "dev"
	// BuildTime is set during build
	BuildTime = "unknown"
)

var (
	serverURL   string
	format      string
	httpMode    bool
	paramsString string
)

func main() {
	cobra.EnableCommandSorting = false
	
	var rootCmd = &cobra.Command{
		Use:   "mcp",
		Short: "MCP is a command line interface for interacting with MCP servers",
		Long: `MCP is a command line interface for interacting with Model Context Protocol (MCP) servers.
It allows you to discover and call tools, list resources, and interact with MCP-compatible services.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://localhost:8080", "MCP server URL (when using HTTP transport)")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "table", "Output format (table, json, pretty)")
	rootCmd.PersistentFlags().BoolVarP(&httpMode, "http", "H", false, "Use HTTP transport instead of stdio")
	rootCmd.PersistentFlags().StringVarP(&paramsString, "params", "p", "{}", "JSON string of parameters to pass to the tool (for call command)")

	// Version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("MCP version %s (built at %s)\n", Version, BuildTime)
		},
	}

	// List tools command
	var toolsCmd = &cobra.Command{
		Use:                   "tools [command args...]",
		Short:                 "List available tools on the MCP server",
		DisableFlagParsing:    true,  // Important: Don't parse flags for this command
		SilenceUsage:          true,
		Run: func(cmd *cobra.Command, args []string) {
			// Special handling for --help flag
			if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
				cmd.Help()
				return
			}
			
			// For other flags like --format, --http, etc, we need to handle them manually
			// since DisableFlagParsing is true
			cmdArgs := args
			parsedArgs := []string{}
			
			// Process global flags and remove them from args
			i := 0
			for i < len(cmdArgs) {
				if cmdArgs[i] == "--format" || cmdArgs[i] == "-f" {
					if i+1 < len(cmdArgs) {
						format = cmdArgs[i+1]
						i += 2
						continue
					}
				} else if cmdArgs[i] == "--http" || cmdArgs[i] == "-H" {
					httpMode = true
					i++
					continue
				} else if cmdArgs[i] == "--server" || cmdArgs[i] == "-s" {
					if i+1 < len(cmdArgs) {
						serverURL = cmdArgs[i+1]
						i += 2
						continue
					}
				}
				
				parsedArgs = append(parsedArgs, cmdArgs[i])
				i++
			}
			
			// Now parsedArgs contains only the command to execute
			if !httpMode && len(parsedArgs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: command to execute is required when using stdio transport")
				fmt.Fprintln(os.Stderr, "Example: mcp tools npx -y @modelcontextprotocol/server-filesystem ~/Code")
				os.Exit(1)
			}

			var mcpClient *client.Client
			if httpMode {
				mcpClient = client.New(serverURL)
			} else {
				mcpClient = client.NewStdio(parsedArgs)
			}

			resp, err := mcpClient.ListTools()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			output, err := formatter.Format(resp, format)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(output)
		},
	}

	// List resources command
	var resourcesCmd = &cobra.Command{
		Use:                   "resources [command args...]",
		Short:                 "List available resources on the MCP server",
		DisableFlagParsing:    true,  // Important: Don't parse flags for this command
		SilenceUsage:          true,
		Run: func(cmd *cobra.Command, args []string) {
			// Special handling for --help flag
			if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
				cmd.Help()
				return
			}
			
			// For other flags like --format, --http, etc, we need to handle them manually
			// since DisableFlagParsing is true
			cmdArgs := args
			parsedArgs := []string{}
			
			// Process global flags and remove them from args
			i := 0
			for i < len(cmdArgs) {
				if cmdArgs[i] == "--format" || cmdArgs[i] == "-f" {
					if i+1 < len(cmdArgs) {
						format = cmdArgs[i+1]
						i += 2
						continue
					}
				} else if cmdArgs[i] == "--http" || cmdArgs[i] == "-H" {
					httpMode = true
					i++
					continue
				} else if cmdArgs[i] == "--server" || cmdArgs[i] == "-s" {
					if i+1 < len(cmdArgs) {
						serverURL = cmdArgs[i+1]
						i += 2
						continue
					}
				}
				
				parsedArgs = append(parsedArgs, cmdArgs[i])
				i++
			}
			
			if !httpMode && len(parsedArgs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: command to execute is required when using stdio transport")
				fmt.Fprintln(os.Stderr, "Example: mcp resources npx -y @modelcontextprotocol/server-filesystem ~/Code")
				os.Exit(1)
			}

			var mcpClient *client.Client
			if httpMode {
				mcpClient = client.New(serverURL)
			} else {
				mcpClient = client.NewStdio(parsedArgs)
			}

			resp, err := mcpClient.ListResources()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			output, err := formatter.Format(resp, format)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(output)
		},
	}

	// List prompts command
	var promptsCmd = &cobra.Command{
		Use:                   "prompts [command args...]",
		Short:                 "List available prompts on the MCP server",
		DisableFlagParsing:    true,  // Important: Don't parse flags for this command
		SilenceUsage:          true,
		Run: func(cmd *cobra.Command, args []string) {
			// Special handling for --help flag
			if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
				cmd.Help()
				return
			}
			
			// For other flags like --format, --http, etc, we need to handle them manually
			// since DisableFlagParsing is true
			cmdArgs := args
			parsedArgs := []string{}
			
			// Process global flags and remove them from args
			i := 0
			for i < len(cmdArgs) {
				if cmdArgs[i] == "--format" || cmdArgs[i] == "-f" {
					if i+1 < len(cmdArgs) {
						format = cmdArgs[i+1]
						i += 2
						continue
					}
				} else if cmdArgs[i] == "--http" || cmdArgs[i] == "-H" {
					httpMode = true
					i++
					continue
				} else if cmdArgs[i] == "--server" || cmdArgs[i] == "-s" {
					if i+1 < len(cmdArgs) {
						serverURL = cmdArgs[i+1]
						i += 2
						continue
					}
				}
				
				parsedArgs = append(parsedArgs, cmdArgs[i])
				i++
			}
			
			if !httpMode && len(parsedArgs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: command to execute is required when using stdio transport")
				fmt.Fprintln(os.Stderr, "Example: mcp prompts npx -y @modelcontextprotocol/server-filesystem ~/Code")
				os.Exit(1)
			}

			var mcpClient *client.Client
			if httpMode {
				mcpClient = client.New(serverURL)
			} else {
				mcpClient = client.NewStdio(parsedArgs)
			}

			resp, err := mcpClient.ListPrompts()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			output, err := formatter.Format(resp, format)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(output)
		},
	}

	// Call command
	var callCmd = &cobra.Command{
		Use:                   "call entity [command args...]",
		Short:                 "Call a tool, resource, or prompt on the MCP server",
		DisableFlagParsing:    true,  // Important: Don't parse flags for this command
		SilenceUsage:          true,
		Run: func(cmd *cobra.Command, args []string) {
			// Special handling for --help flag
			if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
				cmd.Help()
				return
			}
			
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "Error: entity name is required")
				fmt.Fprintln(os.Stderr, "Example: mcp call read_file npx -y @modelcontextprotocol/server-filesystem ~/Code")
				os.Exit(1)
			}
			
			// Process our flags manually since DisableFlagParsing is true
			cmdArgs := args
			parsedArgs := []string{}
			entityName := ""
			
			// Process global flags and remove them from args
			i := 0
			entityExtracted := false
			
			for i < len(cmdArgs) {
				if cmdArgs[i] == "--format" || cmdArgs[i] == "-f" {
					if i+1 < len(cmdArgs) {
						format = cmdArgs[i+1]
						i += 2
						continue
					}
				} else if cmdArgs[i] == "--http" || cmdArgs[i] == "-H" {
					httpMode = true
					i++
					continue
				} else if cmdArgs[i] == "--server" || cmdArgs[i] == "-s" {
					if i+1 < len(cmdArgs) {
						serverURL = cmdArgs[i+1]
						i += 2
						continue
					}
				} else if cmdArgs[i] == "--params" || cmdArgs[i] == "-p" {
					if i+1 < len(cmdArgs) {
						paramsString = cmdArgs[i+1]
						i += 2
						continue
					}
				} else if !entityExtracted {
					// The first non-flag argument is the entity name
					entityName = cmdArgs[i]
					entityExtracted = true
					i++
					continue
				}
				
				// Any other arguments get passed to the command
				parsedArgs = append(parsedArgs, cmdArgs[i])
				i++
			}
			
			if entityName == "" {
				fmt.Fprintln(os.Stderr, "Error: entity name is required")
				fmt.Fprintln(os.Stderr, "Example: mcp call read_file npx -y @modelcontextprotocol/server-filesystem ~/Code")
				os.Exit(1)
			}

			entityType := "tool" // Default to tool
			
			// Check if entityName contains a type prefix
			parts := strings.SplitN(entityName, ":", 2)
			if len(parts) == 2 {
				entityType = parts[0]
				entityName = parts[1]
			}
			
			if !httpMode && len(parsedArgs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: command to execute is required when using stdio transport")
				fmt.Fprintln(os.Stderr, "Example: mcp call read_file npx -y @modelcontextprotocol/server-filesystem ~/Code")
				os.Exit(1)
			}

			// Parse parameters
			var params map[string]interface{}
			if paramsString != "" {
				if err := json.Unmarshal([]byte(paramsString), &params); err != nil {
					fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", err)
					os.Exit(1)
				}
			}
			
			var mcpClient *client.Client
			
			if httpMode {
				mcpClient = client.New(serverURL)
			} else {
				mcpClient = client.NewStdio(parsedArgs)
			}
			
			var resp map[string]interface{}
			var err error
			
			switch entityType {
			case "tool":
				resp, err = mcpClient.CallTool(entityName, params)
			case "resource":
				resp, err = mcpClient.ReadResource(entityName)
			case "prompt":
				resp, err = mcpClient.GetPrompt(entityName)
			default:
				fmt.Fprintf(os.Stderr, "Error: unsupported entity type: %s\n", entityType)
				os.Exit(1)
			}
			
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			output, err := formatter.Format(resp, format)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
				os.Exit(1)
			}

			fmt.Println(output)
		},
	}

	// Shell command
	var shellCmd = &cobra.Command{
		Use:                "shell [command args...]",
		Short:              "Start an interactive shell for MCP commands",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		Run: func(cmd *cobra.Command, args []string) {
			// Special handling for --help flag
			if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
				cmd.Help()
				return
			}
			
			// For other flags like --format, we need to handle them manually
			cmdArgs := args
			parsedArgs := []string{}
			
			// Process global flags and remove them from args
			i := 0
			for i < len(cmdArgs) {
				if cmdArgs[i] == "--format" || cmdArgs[i] == "-f" {
					if i+1 < len(cmdArgs) {
						format = cmdArgs[i+1]
						i += 2
						continue
					}
				} else if cmdArgs[i] == "--server" || cmdArgs[i] == "-s" {
					if i+1 < len(cmdArgs) {
						serverURL = cmdArgs[i+1]
						i += 2
						continue
					}
				}
				
				parsedArgs = append(parsedArgs, cmdArgs[i])
				i++
			}
			
			if len(parsedArgs) == 0 {
				fmt.Fprintln(os.Stderr, "Error: command to execute is required when using the shell")
				fmt.Fprintln(os.Stderr, "Example: mcp shell npx -y @modelcontextprotocol/server-filesystem ~/Code")
				os.Exit(1)
			}

			// Create the stdio client that will be reused for all commands
			mcpClient := client.NewStdio(parsedArgs)
			
			// Try to connect and get server info
			_, err := mcpClient.ListTools()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error connecting to MCP server: %v\n", err)
				os.Exit(1)
			}
			
			// Start the interactive shell
			fmt.Println("mcp > connected to MCP server over stdio")
			fmt.Println("mcp > Type '/h' for help or '/q' to quit")
			
			// Create a new line state with history capability
			line := liner.NewLiner()
			defer line.Close()
			
			// Load command history from ~/.mcp_history if it exists
			historyFile := filepath.Join(os.Getenv("HOME"), ".mcp_history")
			if f, err := os.Open(historyFile); err == nil {
				line.ReadHistory(f)
				f.Close()
			}
			
			// Save history on exit
			defer func() {
				if f, err := os.Create(historyFile); err == nil {
					line.WriteHistory(f)
					f.Close()
				}
			}()
			
			// Set completion handler for commands
			line.SetCompleter(func(line string) (c []string) {
				commands := []string{"tools", "resources", "prompts", "call", "format", "help", "exit", "/h", "/q", "/help", "/quit"}
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
					if err == liner.ErrPromptAborted {
						fmt.Println("Exiting MCP shell")
						break
					}
					fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
					break
				}
				
				if input == "" {
					continue
				}
				
				// Add the command to history
				line.AppendHistory(input)
				
				// Handle special commands
				if input == "/q" || input == "/quit" || input == "exit" {
					fmt.Println("Exiting MCP shell")
					break
				}
				
				if input == "/h" || input == "/help" || input == "help" {
					printShellHelp()
					continue
				}
				
				// Parse the input as a command
				parts := strings.Fields(input)
				if len(parts) == 0 {
					continue
				}
				
				command := parts[0]
				commandArgs := parts[1:]
				
				// Process the command
				switch command {
				case "tools":
					resp, err := mcpClient.ListTools()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						continue
					}
					
					output, err := formatter.Format(resp, format)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
						continue
					}
					
					fmt.Println(output)
					
				case "resources":
					resp, err := mcpClient.ListResources()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						continue
					}
					
					output, err := formatter.Format(resp, format)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
						continue
					}
					
					fmt.Println(output)
					
				case "prompts":
					resp, err := mcpClient.ListPrompts()
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						continue
					}
					
					output, err := formatter.Format(resp, format)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
						continue
					}
					
					fmt.Println(output)
					
				case "call":
					if len(commandArgs) < 1 {
						fmt.Println("Usage: call <entity> [--params '{...}']")
						continue
					}
					
					entityName := commandArgs[0]
					entityType := "tool" // Default to tool
					
					// Check if entityName contains a type prefix
					parts := strings.SplitN(entityName, ":", 2)
					if len(parts) == 2 {
						entityType = parts[0]
						entityName = parts[1]
					}
					
					// Parse parameters if provided
					params := map[string]interface{}{}
					for i := 1; i < len(commandArgs); i++ {
						if commandArgs[i] == "--params" || commandArgs[i] == "-p" {
							if i+1 < len(commandArgs) {
								if err := json.Unmarshal([]byte(commandArgs[i+1]), &params); err != nil {
									fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", err)
									continue
								}
								break
							}
						}
					}
					
					var resp map[string]interface{}
					var err error
					
					switch entityType {
					case "tool":
						resp, err = mcpClient.CallTool(entityName, params)
					case "resource":
						resp, err = mcpClient.ReadResource(entityName)
					case "prompt":
						resp, err = mcpClient.GetPrompt(entityName)
					default:
						fmt.Fprintf(os.Stderr, "Error: unsupported entity type: %s\n", entityType)
						continue
					}
					
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						continue
					}
					
					output, err := formatter.Format(resp, format)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
						continue
					}
					
					fmt.Println(output)
					
				case "format":
					if len(commandArgs) < 1 {
						fmt.Printf("Current format: %s\n", format)
						continue
					}
					
					newFormat := commandArgs[0]
					if newFormat == "json" || newFormat == "j" ||
					   newFormat == "pretty" || newFormat == "p" ||
					   newFormat == "table" || newFormat == "t" {
						format = newFormat
						fmt.Printf("Format set to: %s\n", format)
					} else {
						fmt.Println("Invalid format. Use: table, json, or pretty")
					}
					
				default:
					// Try to interpret the command as a tool call
					entityName := command
					entityType := "tool" // Default to tool
					
					// Check if entityName contains a type prefix
					parts := strings.SplitN(entityName, ":", 2)
					if len(parts) == 2 {
						entityType = parts[0]
						entityName = parts[1]
					}
					
					// Parse parameters if provided
					params := map[string]interface{}{}
					
					// Check if the first argument is a JSON object
					if len(commandArgs) > 0 {
						firstArg := commandArgs[0]
						if strings.HasPrefix(firstArg, "{") && strings.HasSuffix(firstArg, "}") {
							if err := json.Unmarshal([]byte(firstArg), &params); err != nil {
								fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", err)
								continue
							}
						} else {
							// Process parameters with --params or -p flag
							for i := 0; i < len(commandArgs); i++ {
								if commandArgs[i] == "--params" || commandArgs[i] == "-p" {
									if i+1 < len(commandArgs) {
										if err := json.Unmarshal([]byte(commandArgs[i+1]), &params); err != nil {
											fmt.Fprintf(os.Stderr, "Error: invalid JSON for params: %v\n", err)
											continue
										}
										break
									}
								}
							}
						}
					}
					
					var resp map[string]interface{}
					var err error
					
					switch entityType {
					case "tool":
						resp, err = mcpClient.CallTool(entityName, params)
					case "resource":
						resp, err = mcpClient.ReadResource(entityName)
					case "prompt":
						resp, err = mcpClient.GetPrompt(entityName)
					default:
						fmt.Printf("Unknown command: %s\nType '/h' for help\n", command)
						continue
					}
					
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						continue
					}
					
					output, err := formatter.Format(resp, format)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
						continue
					}
					
					fmt.Println(output)
				}
			}
		},
	}

	// Add commands to root
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(toolsCmd)
	rootCmd.AddCommand(resourcesCmd)
	rootCmd.AddCommand(promptsCmd)
	rootCmd.AddCommand(callCmd)
	rootCmd.AddCommand(shellCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Helper function to print shell help
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