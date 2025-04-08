package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// ConfigFileOption stores the path to the configuration file
var ConfigFileOption string

// HeadersOption stores the headers for URL-based servers
var HeadersOption string

// EnvOption stores the environment variables
var EnvOption string

// URLOption stores the URL for URL-based servers
var URLOption string

// ConfigAlias represents a configuration alias
type ConfigAlias struct {
	Path     string `json:"path"`
	JSONPath string `json:"jsonPath"`
	Source   string `json:"source,omitempty"`
}

// ConfigsFile represents the structure of the configs file
type ConfigsFile struct {
	Aliases map[string]ConfigAlias `json:"aliases"`
}

// getConfigsFilePath returns the path to the configs file
func getConfigsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".mcpt")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "configs.json"), nil
}

// loadConfigsFile loads the configs file, creating it if it doesn't exist
func loadConfigsFile() (*ConfigsFile, error) {
	configsPath, err := getConfigsFilePath()
	if err != nil {
		return nil, err
	}

	// Create default config if file doesn't exist
	if _, err := os.Stat(configsPath); os.IsNotExist(err) {
		defaultConfig := &ConfigsFile{
			Aliases: map[string]ConfigAlias{
				"vscode": {
					Path:     "~/Library/Application Support/Code/User/settings.json",
					JSONPath: "$.mcp.servers",
					Source:   "VS Code",
				},
				"vscode-insiders": {
					Path:     "~/Library/Application Support/Code - Insiders/User/settings.json",
					JSONPath: "$.mcp.servers",
					Source:   "VS Code Insiders",
				},
				"windsurf": {
					Path:     "~/.codeium/windsurf/mcp_config.json",
					JSONPath: "$.mcpServers",
					Source:   "Windsurf",
				},
				"cursor": {
					Path:     "~/.cursor/mcp.json",
					JSONPath: "$.mcpServers",
					Source:   "Cursor",
				},
				"claude-desktop": {
					Path:     "~/Library/Application Support/Claude/claude_desktop_config.json",
					JSONPath: "$.mcpServers",
					Source:   "Claude Desktop",
				},
				"claude-code": {
					Path:     "~/.claude.json",
					JSONPath: "$.mcpServers",
					Source:   "Claude Code",
				},
			},
		}

		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal default config: %w", err)
		}

		if err := os.WriteFile(configsPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}

		return defaultConfig, nil
	}

	// Read existing config
	data, err := os.ReadFile(configsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ConfigsFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize map if nil
	if config.Aliases == nil {
		config.Aliases = make(map[string]ConfigAlias)
	}

	return &config, nil
}

// saveConfigsFile saves the configs file
func saveConfigsFile(config *ConfigsFile) error {
	configsPath, err := getConfigsFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configsPath, data, 0644)
}

// expandPath expands the ~ in the path
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(homeDir, path[1:])
}

// parseKeyValueOption parses a comma-separated list of key=value pairs
func parseKeyValueOption(option string) (map[string]string, error) {
	result := make(map[string]string)
	if option == "" {
		return result, nil
	}

	pairs := strings.Split(option, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid key-value pair: %s", pair)
		}
		result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	return result, nil
}

// getConfigFileAndPath gets the config file path and json path from an alias or direct file path
func getConfigFileAndPath(configs *ConfigsFile, aliasName, configFile string) (string, string, error) {
	var jsonPath string
	if configFile == "" {
		// Check if the name is an alias
		aliasConfig, ok := configs.Aliases[strings.ToLower(aliasName)]
		if !ok {
			return "", "", fmt.Errorf("alias '%s' not found and no config file specified", aliasName)
		}
		configFile = aliasConfig.Path
		jsonPath = aliasConfig.JSONPath
	} else {
		// Default JSON path if using direct file path
		jsonPath = "$.mcpServers"
	}

	// Expand the path if needed
	configFile = expandPath(configFile)
	return configFile, jsonPath, nil
}

// readConfigFile reads and parses a config file
func readConfigFile(configFile string) (map[string]interface{}, error) {
	var configData map[string]interface{}
	if _, err := os.Stat(configFile); err == nil {
		// File exists, read and parse it
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, &configData); err != nil {
			// Create new empty config if file exists but isn't valid JSON
			configData = make(map[string]interface{})
		}
	} else {
		// Create new empty config if file doesn't exist
		configData = make(map[string]interface{})
	}
	return configData, nil
}

// addServerToConfig adds a server configuration to the config data
func addServerToConfig(configData map[string]interface{}, jsonPath, serverName string, serverConfig map[string]interface{}) {
	if strings.Contains(jsonPath, "mcp.servers") {
		// VS Code format
		if _, ok := configData["mcp"]; !ok {
			configData["mcp"] = map[string]interface{}{}
		}

		mcpMap, ok := configData["mcp"].(map[string]interface{})
		if !ok {
			mcpMap = map[string]interface{}{}
			configData["mcp"] = mcpMap
		}

		if _, ok := mcpMap["servers"]; !ok {
			mcpMap["servers"] = map[string]interface{}{}
		}

		serversMap, ok := mcpMap["servers"].(map[string]interface{})
		if !ok {
			serversMap = map[string]interface{}{}
			mcpMap["servers"] = serversMap
		}

		serversMap[serverName] = serverConfig
	} else {
		// Other formats with mcpServers
		if _, ok := configData["mcpServers"]; !ok {
			configData["mcpServers"] = map[string]interface{}{}
		}

		serversMap, ok := configData["mcpServers"].(map[string]interface{})
		if !ok {
			serversMap = map[string]interface{}{}
			configData["mcpServers"] = serversMap
		}

		serversMap[serverName] = serverConfig
	}
}

// getServerFromConfig gets a server configuration from the config data
func getServerFromConfig(configData map[string]interface{}, jsonPath, serverName string) (map[string]interface{}, bool) {
	if strings.Contains(jsonPath, "mcp.servers") {
		// VS Code format
		mcpMap, ok := configData["mcp"].(map[string]interface{})
		if !ok {
			return nil, false
		}

		serversMap, ok := mcpMap["servers"].(map[string]interface{})
		if !ok {
			return nil, false
		}

		server, ok := serversMap[serverName].(map[string]interface{})
		return server, ok
	} else {
		// Other formats with mcpServers
		serversMap, ok := configData["mcpServers"].(map[string]interface{})
		if !ok {
			return nil, false
		}

		server, ok := serversMap[serverName].(map[string]interface{})
		return server, ok
	}
}

// removeServerFromConfig removes a server configuration from the config data
func removeServerFromConfig(configData map[string]interface{}, jsonPath, serverName string) bool {
	if strings.Contains(jsonPath, "mcp.servers") {
		// VS Code format
		mcpMap, ok := configData["mcp"].(map[string]interface{})
		if !ok {
			return false
		}

		serversMap, ok := mcpMap["servers"].(map[string]interface{})
		if !ok {
			return false
		}

		if _, exists := serversMap[serverName]; !exists {
			return false
		}

		delete(serversMap, serverName)
		return true
	} else {
		// Other formats with mcpServers
		serversMap, ok := configData["mcpServers"].(map[string]interface{})
		if !ok {
			return false
		}

		if _, exists := serversMap[serverName]; !exists {
			return false
		}

		delete(serversMap, serverName)
		return true
	}
}

// ConfigsCmd creates the configs command
func ConfigsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configs",
		Short: "Manage MCP server configurations",
		Long:  `Manage MCP server configurations including scanning, adding, and aliasing.`,
	}

	// Add the view subcommand with AllOption flag
	var AllOption bool
	viewCmd := &cobra.Command{
		Use:   "view [alias or path]",
		Short: "View MCP servers in configurations",
		Long:  `View MCP servers in a specific configuration file or all configured aliases with --all flag.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Load configs
			configs, err := loadConfigsFile()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error loading configs: %v\n", err)
				return
			}

			var servers []ServerConfig

			// All mode - scan all aliases (same as previous scan command)
			if AllOption {
				if len(args) > 0 {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: ignoring specified alias/path when using --all flag\n")
				}

				// Scan all configured aliases
				for alias, config := range configs.Aliases {
					// Skip if path is empty
					if config.Path == "" {
						continue
					}

					// Determine the scanner function based on the JSONPath
					expandedPath := expandPath(config.Path)
					var configServers []ServerConfig
					var scanErr error

					source := config.Source
					if source == "" {
						source = strings.Title(alias) // Use capitalized alias name if source not provided
					}

					if strings.Contains(config.JSONPath, "mcp.servers") {
						configServers, scanErr = scanVSCodeConfig(expandedPath, source)
					} else if strings.Contains(config.JSONPath, "mcpServers") {
						configServers, scanErr = scanMCPServersConfig(expandedPath, source)
					}

					if scanErr == nil {
						servers = append(servers, configServers...)
					}
				}
			} else {
				// Single config mode - require an argument
				if len(args) == 0 {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: must specify an alias or path, use --all flag, or use `configs ls` command\n")
					return
				}

				target := args[0]
				var configPath string
				var source string
				var jsonPath string

				// Check if the argument is an alias
				if aliasConfig, ok := configs.Aliases[strings.ToLower(target)]; ok {
					configPath = aliasConfig.Path
					source = aliasConfig.Source
					if source == "" {
						source = strings.Title(target)
					}
					jsonPath = aliasConfig.JSONPath
				} else {
					// Assume it's a direct path
					configPath = target
					source = filepath.Base(target)
					jsonPath = "$.mcpServers" // Default JSON path
				}

				// Expand the path if needed
				expandedPath := expandPath(configPath)

				// Scan the config file
				var configServers []ServerConfig
				var scanErr error

				if strings.Contains(jsonPath, "mcp.servers") {
					configServers, scanErr = scanVSCodeConfig(expandedPath, source)
				} else {
					configServers, scanErr = scanMCPServersConfig(expandedPath, source)
				}

				if scanErr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error scanning configuration: %v\n", scanErr)
					return
				}

				servers = configServers
			}

			// Handle empty results
			if len(servers) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No MCP servers found")
				return
			}

			// Output based on format
			if strings.ToLower(FormatOption) == "table" || strings.ToLower(FormatOption) == "pretty" {
				output := formatColoredGroupedServers(servers)
				fmt.Fprintln(cmd.OutOrStdout(), output)
				return
			}

			if strings.ToLower(FormatOption) == "json" {
				output := formatSourceGroupedJSON(servers)
				fmt.Fprintln(cmd.OutOrStdout(), output)
				return
			}

			// For other formats, use the full server data
			output, err := json.MarshalIndent(servers, "", "  ")
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error formatting output: %v\n", err)
				return
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(output))
		},
	}

	// Add --all flag to view command
	viewCmd.Flags().BoolVar(&AllOption, "all", false, "View all configured aliases")

	// Add ls command as an alias for view --all
	lsCmd := &cobra.Command{
		Use:   "ls [alias or path]",
		Short: "List all MCP servers in configurations (alias for view --all)",
		Long:  `List all MCP servers in configurations. If a path is specified, only show that configuration.`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Set AllOption to true to make this default to --all
			AllOption = true

			// Run the view command logic
			viewCmd.Run(cmd, args)
		},
	}

	// Add --all flag to ls command (though it's true by default)
	lsCmd.Flags().BoolVar(&AllOption, "all", false, "View all configured aliases (default: false)")

	// Add the add subcommand
	addCmd := &cobra.Command{
		Use:   "add [alias] [server] [command/url] [args...]",
		Short: "Add a new MCP server configuration",
		Long:  `Add a new MCP server configuration using either an alias or direct file path. For URL-based servers, use --url flag.`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Load configs
			configs, err := loadConfigsFile()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error loading configs: %v\n", err)
				return
			}

			// Get the alias/config file and server name
			aliasName := args[0]
			serverName := args[1]

			// Get config file and JSON path from alias or direct path
			configFile, jsonPath, err := getConfigFileAndPath(configs, aliasName, ConfigFileOption)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return
			}

			// Read the target config file
			configData, err := readConfigFile(configFile)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return
			}

			// Create server config
			serverConfig := make(map[string]interface{})

			// Determine if this is a URL-based or command-based server
			if URLOption != "" {
				// URL-based server
				serverConfig["url"] = URLOption

				// Parse headers
				if HeadersOption != "" {
					headers, err := parseKeyValueOption(HeadersOption)
					if err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Error parsing headers: %v\n", err)
						return
					}
					if len(headers) > 0 {
						serverConfig["headers"] = headers
					}
				}
			} else if len(args) > 2 {
				// Command-based server
				command := args[2]
				serverConfig["command"] = command

				// Add command args if provided
				if len(args) > 3 {
					serverConfig["args"] = args[3:]
				}
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: either command or --url must be provided\n")
				return
			}

			// Parse environment variables for both URL and command servers
			if EnvOption != "" {
				env, err := parseKeyValueOption(EnvOption)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error parsing environment variables: %v\n", err)
					return
				}
				if len(env) > 0 {
					serverConfig["env"] = env
				}
			}

			// Add the server to the config
			addServerToConfig(configData, jsonPath, serverName, serverConfig)

			// Write the updated config back to the file
			data, err := json.MarshalIndent(configData, "", "  ")
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error marshaling config: %v\n", err)
				return
			}

			if err := os.WriteFile(configFile, data, 0644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing config file: %v\n", err)
				return
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Server '%s' added to %s\n", serverName, configFile)
		},
	}

	// Add the update subcommand
	updateCmd := &cobra.Command{
		Use:   "update [alias] [server] [command/url] [args...]",
		Short: "Update an existing MCP server configuration",
		Long:  `Update an existing MCP server configuration. For URL-based servers, use --url flag.`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Load configs
			configs, err := loadConfigsFile()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error loading configs: %v\n", err)
				return
			}

			// Get the alias/config file and server name
			aliasName := args[0]
			serverName := args[1]

			// Get config file and JSON path from alias or direct path
			configFile, jsonPath, err := getConfigFileAndPath(configs, aliasName, ConfigFileOption)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return
			}

			// Read the target config file
			configData, err := readConfigFile(configFile)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return
			}

			// Check if the server exists
			existingServer, exists := getServerFromConfig(configData, jsonPath, serverName)
			if !exists {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: server '%s' not found in %s\n", serverName, configFile)
				return
			}

			// Create server config starting with existing values
			serverConfig := existingServer
			if serverConfig == nil {
				serverConfig = make(map[string]interface{})
			}

			// Determine if this is a URL-based or command-based server update
			if URLOption != "" {
				// URL-based server - remove command and args if they exist
				delete(serverConfig, "command")
				delete(serverConfig, "args")

				// Set the URL
				serverConfig["url"] = URLOption

				// Parse headers
				if HeadersOption != "" {
					headers, err := parseKeyValueOption(HeadersOption)
					if err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Error parsing headers: %v\n", err)
						return
					}
					if len(headers) > 0 {
						serverConfig["headers"] = headers
					}
				}
			} else if len(args) > 2 {
				// Command-based server - remove url and headers if they exist
				delete(serverConfig, "url")
				delete(serverConfig, "headers")

				// Set the command
				command := args[2]
				serverConfig["command"] = command

				// Add command args if provided
				if len(args) > 3 {
					serverConfig["args"] = args[3:]
				} else {
					delete(serverConfig, "args")
				}
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: either command or --url must be provided\n")
				return
			}

			// Parse environment variables
			if EnvOption != "" {
				env, err := parseKeyValueOption(EnvOption)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error parsing environment variables: %v\n", err)
					return
				}
				if len(env) > 0 {
					serverConfig["env"] = env
				}
			}

			// Update the server in the config
			addServerToConfig(configData, jsonPath, serverName, serverConfig)

			// Write the updated config back to the file
			data, err := json.MarshalIndent(configData, "", "  ")
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error marshaling config: %v\n", err)
				return
			}

			if err := os.WriteFile(configFile, data, 0644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing config file: %v\n", err)
				return
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Server '%s' updated in %s\n", serverName, configFile)
		},
	}

	// Add the remove subcommand
	removeCmd := &cobra.Command{
		Use:   "remove [alias] [server]",
		Short: "Remove an MCP server configuration",
		Long:  `Remove an MCP server configuration from a config file.`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			// Load configs
			configs, err := loadConfigsFile()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error loading configs: %v\n", err)
				return
			}

			// Get the alias/config file and server name
			aliasName := args[0]
			serverName := args[1]

			// Get config file and JSON path from alias or direct path
			configFile, jsonPath, err := getConfigFileAndPath(configs, aliasName, ConfigFileOption)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return
			}

			// Read the target config file
			configData, err := readConfigFile(configFile)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
				return
			}

			// Remove the server
			removed := removeServerFromConfig(configData, jsonPath, serverName)
			if !removed {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error: server '%s' not found in %s\n", serverName, configFile)
				return
			}

			// Write the updated config back to the file
			data, err := json.MarshalIndent(configData, "", "  ")
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error marshaling config: %v\n", err)
				return
			}

			if err := os.WriteFile(configFile, data, 0644); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing config file: %v\n", err)
				return
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Server '%s' removed from %s\n", serverName, configFile)
		},
	}

	// Add the alias subcommand
	aliasCmd := &cobra.Command{
		Use:   "alias [name] [path] [jsonPath]",
		Short: "Create an alias for a config file",
		Long:  `Create an alias for a configuration file with the specified JSONPath (defaults to "$.mcpServers").`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			path := args[1]
			jsonPath := "$.mcpServers" // Default value

			// Use custom jsonPath if provided
			if len(args) > 2 {
				jsonPath = args[2]
			}

			// Load configs
			configs, err := loadConfigsFile()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error loading configs: %v\n", err)
				return
			}

			// Add or update the alias
			configs.Aliases[strings.ToLower(name)] = ConfigAlias{
				Path:     path,
				JSONPath: jsonPath,
				Source:   name, // Use the provided name as the source
			}

			// Save the updated configs
			if err := saveConfigsFile(configs); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error saving configs: %v\n", err)
				return
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Alias '%s' created for %s with JSONPath '%s'\n", name, path, jsonPath)
		},
	}

	// Add flags to the add and update commands
	addCmd.Flags().StringVar(&ConfigFileOption, "config", "", "Path to the configuration file")
	addCmd.Flags().StringVar(&URLOption, "url", "", "URL for SSE-based servers")
	addCmd.Flags().StringVar(&HeadersOption, "headers", "", "Headers for URL-based servers (comma-separated key=value pairs)")
	addCmd.Flags().StringVar(&EnvOption, "env", "", "Environment variables (comma-separated key=value pairs)")

	updateCmd.Flags().StringVar(&ConfigFileOption, "config", "", "Path to the configuration file")
	updateCmd.Flags().StringVar(&URLOption, "url", "", "URL for SSE-based servers")
	updateCmd.Flags().StringVar(&HeadersOption, "headers", "", "Headers for URL-based servers (comma-separated key=value pairs)")
	updateCmd.Flags().StringVar(&EnvOption, "env", "", "Environment variables (comma-separated key=value pairs)")

	removeCmd.Flags().StringVar(&ConfigFileOption, "config", "", "Path to the configuration file")

	// Add subcommands to the configs command
	cmd.AddCommand(addCmd, updateCmd, removeCmd, aliasCmd, viewCmd, lsCmd)

	return cmd
}
