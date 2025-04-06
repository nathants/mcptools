package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/f/mcptools/pkg/jsonutils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ScanCmd creates the scan command.
func ScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan for available MCP servers in various configurations",
		Long: `Scan for available MCP servers in various configuration files of these Applications on macOS:
VS Code, VS Code Insiders, Windsurf, Cursor, Claude Desktop`,
		Run: func(cmd *cobra.Command, _ []string) {
			servers, err := scanForServers()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error scanning for servers: %v\n", err)
				return
			}

			// Table format (default) now uses the colored grouped display
			if strings.ToLower(FormatOption) == "table" || strings.ToLower(FormatOption) == "pretty" {
				output := formatColoredGroupedServers(servers)
				fmt.Fprintln(cmd.OutOrStdout(), output)
				return
			}

			// For JSON format, use the grouped display
			if strings.ToLower(FormatOption) == "json" {
				output := formatSourceGroupedJSON(servers)
				fmt.Fprintln(cmd.OutOrStdout(), output)
				return
			}

			// For other formats, use the full server data
			output, err := jsonutils.Format(servers, FormatOption)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error formatting output: %v\n", err)
				return
			}

			fmt.Fprintln(cmd.OutOrStdout(), output)
		},
	}
}

// formatSourceGroupedJSON formats servers grouped by source with raw JSON.
func formatSourceGroupedJSON(servers []ServerConfig) string {
	if len(servers) == 0 {
		return "No MCP servers found"
	}

	// Group servers by source
	serversBySource := make(map[string][]ServerConfig)
	var sourceOrder []string

	for _, server := range servers {
		if _, exists := serversBySource[server.Source]; !exists {
			sourceOrder = append(sourceOrder, server.Source)
		}
		serversBySource[server.Source] = append(serversBySource[server.Source], server)
	}

	// Sort sources
	sort.Strings(sourceOrder)

	var buf bytes.Buffer
	for _, source := range sourceOrder {
		// Print source header
		fmt.Fprintf(&buf, "%s:\n\n", source)

		// Reconstruct the original JSON structure
		sourceServers := serversBySource[source]

		if strings.Contains(source, "VS Code") {
			// VS Code format with mcp.servers
			vsCodeJSON := map[string]interface{}{
				"mcp": map[string]interface{}{
					"servers": make(map[string]interface{}),
				},
			}

			serversMap := vsCodeJSON["mcp"].(map[string]interface{})["servers"].(map[string]interface{})
			for _, server := range sourceServers {
				serversMap[server.Name] = server.Config
			}

			jsonData, _ := json.MarshalIndent(vsCodeJSON, "", "  ")
			fmt.Fprintf(&buf, "%s\n\n", string(jsonData))
		} else {
			// Other formats with mcpServers
			otherJSON := map[string]interface{}{
				"mcpServers": make(map[string]interface{}),
			}

			serversMap := otherJSON["mcpServers"].(map[string]interface{})
			for _, server := range sourceServers {
				serversMap[server.Name] = server.Config
			}

			jsonData, _ := json.MarshalIndent(otherJSON, "", "  ")
			fmt.Fprintf(&buf, "%s\n\n", string(jsonData))
		}
	}

	return buf.String()
}

// formatColoredGroupedServers formats servers in a colored, grouped display by source.
func formatColoredGroupedServers(servers []ServerConfig) string {
	if len(servers) == 0 {
		return "No MCP servers found"
	}

	// Group servers by source
	serversBySource := make(map[string][]ServerConfig)
	var sourceOrder []string

	for _, server := range servers {
		if _, exists := serversBySource[server.Source]; !exists {
			sourceOrder = append(sourceOrder, server.Source)
		}
		serversBySource[server.Source] = append(serversBySource[server.Source], server)
	}

	// Sort sources
	sort.Strings(sourceOrder)

	var buf bytes.Buffer
	// Check if we're outputting to a terminal (for colors)
	useColors := term.IsTerminal(int(os.Stdout.Fd()))

	for _, source := range sourceOrder {
		// Print source header with bold blue
		if useColors {
			fmt.Fprintf(&buf, "%s%s%s\n", jsonutils.ColorBold+jsonutils.ColorBlue, source, jsonutils.ColorReset)
		} else {
			fmt.Fprintf(&buf, "%s\n", source)
		}

		servers := serversBySource[source]
		// Sort servers by name
		sort.Slice(servers, func(i, j int) bool {
			return servers[i].Name < servers[j].Name
		})

		for _, server := range servers {
			// Determine server type
			serverType := server.Type
			if serverType == "" {
				if server.URL != "" {
					serverType = "sse" //nolint
				} else {
					serverType = "stdio"
				}
			}

			// Print server name and type
			if useColors {
				fmt.Fprintf(&buf, "  %s%s%s", jsonutils.ColorBold+jsonutils.ColorPurple, server.Name, jsonutils.ColorReset)
				fmt.Fprintf(&buf, " %s(%s)%s:", jsonutils.ColorBold+jsonutils.ColorCyan, serverType, jsonutils.ColorReset)
			} else {
				fmt.Fprintf(&buf, "  %s (%s):", server.Name, serverType)
			}

			// Print command or URL
			if serverType == "sse" {
				if useColors {
					fmt.Fprintf(&buf, "\n    %s%s%s\n", jsonutils.ColorGreen, server.URL, jsonutils.ColorReset)
				} else {
					fmt.Fprintf(&buf, "\n    %s\n", server.URL)
				}

				// Print headers for SSE servers
				if len(server.Headers) > 0 {
					// Get sorted header keys
					var headerKeys []string
					for k := range server.Headers {
						headerKeys = append(headerKeys, k)
					}
					sort.Strings(headerKeys)

					// Print each header
					for _, k := range headerKeys {
						if useColors {
							fmt.Fprintf(&buf, "      %s%s%s: %s\n", jsonutils.ColorYellow, k, jsonutils.ColorReset, server.Headers[k])
						} else {
							fmt.Fprintf(&buf, "      %s: %s\n", k, server.Headers[k])
						}
					}
				}
			} else {
				// Print command and args
				commandStr := server.Command

				// Add args with quotes for ones containing spaces
				if len(server.Args) > 0 {
					commandStr += " "
					quotedArgs := make([]string, len(server.Args))
					for i, arg := range server.Args {
						if strings.Contains(arg, " ") && !strings.HasPrefix(arg, "\"") && !strings.HasSuffix(arg, "\"") {
							quotedArgs[i] = "\"" + arg + "\""
						} else {
							quotedArgs[i] = arg
						}
					}
					commandStr += strings.Join(quotedArgs, " ")
				}

				if useColors {
					fmt.Fprintf(&buf, "\n    %s%s%s\n", jsonutils.ColorGreen, commandStr, jsonutils.ColorReset)
				} else {
					fmt.Fprintf(&buf, "\n    %s\n", commandStr)
				}

				// Print env vars
				if len(server.Env) > 0 {
					// Get sorted env keys
					var envKeys []string
					for k := range server.Env {
						envKeys = append(envKeys, k)
					}
					sort.Strings(envKeys)

					// Print each env var
					for _, k := range envKeys {
						if useColors {
							fmt.Fprintf(&buf, "      %s%s%s: %s\n", jsonutils.ColorYellow, k, jsonutils.ColorReset, server.Env[k])
						} else {
							fmt.Fprintf(&buf, "      %s: %s\n", k, server.Env[k])
						}
					}
				}
			}

			// Add a newline after each server for readability
			fmt.Fprintln(&buf)
		}
	}

	return buf.String()
}

// ServerConfig represents a configuration for a server.
type ServerConfig struct {
	Headers     map[string]string      `json:"headers,omitempty"`
	Env         map[string]string      `json:"env,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Source      string                 `json:"source"`
	Type        string                 `json:"type,omitempty"`
	Command     string                 `json:"command,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Path        string                 `json:"path,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Args        []string               `json:"args,omitempty"`
}

// scanForServers scans various configuration files for MCP servers.
func scanForServers() ([]ServerConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	var servers []ServerConfig

	// Scan VS Code Insiders
	vscodeInsidersPath := filepath.Join(homeDir, "Library", "Application Support", "Code - Insiders", "User", "settings.json")
	vscodeServers, err := scanVSCodeConfig(vscodeInsidersPath, "VS Code Insiders")
	if err == nil {
		servers = append(servers, vscodeServers...)
	}

	// Scan VS Code
	vscodePath := filepath.Join(homeDir, "Library", "Application Support", "Code", "User", "settings.json")
	vscodeServers, err = scanVSCodeConfig(vscodePath, "VS Code")
	if err == nil {
		servers = append(servers, vscodeServers...)
	}

	// Scan Windsurf
	windsurfPath := filepath.Join(homeDir, ".codeium", "windsurf", "mcp_config.json")
	windsurfServers, err := scanMCPServersConfig(windsurfPath, "Windsurf")
	if err == nil {
		servers = append(servers, windsurfServers...)
	}

	// Scan Cursor
	cursorPath := filepath.Join(homeDir, ".cursor", "mcp.json")
	cursorServers, err := scanMCPServersConfig(cursorPath, "Cursor")
	if err == nil {
		servers = append(servers, cursorServers...)
	}

	// Scan Claude Desktop
	claudeDesktopPath := filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	claudeServers, err := scanMCPServersConfig(claudeDesktopPath, "Claude Desktop")
	if err == nil {
		servers = append(servers, claudeServers...)
	}

	return servers, nil
}

// scanVSCodeConfig scans a VS Code settings.json file for MCP servers.
func scanVSCodeConfig(path, source string) ([]ServerConfig, error) {
	data, err := os.ReadFile(path) //nolint
	if err != nil {
		return nil, fmt.Errorf("failed to read %s config: %w", source, err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse %s settings.json: %w", source, err)
	}

	mcpObject, ok := settings["mcp"]
	if !ok {
		return nil, fmt.Errorf("no mcp configuration found in %s settings", source)
	}

	mcpMap, ok := mcpObject.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid mcp format in %s settings", source)
	}

	mcpServers, ok := mcpMap["servers"]
	if !ok {
		return nil, fmt.Errorf("no mcp.servers found in %s settings", source)
	}

	servers, ok := mcpServers.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid mcp.servers format in %s settings", source)
	}

	var result []ServerConfig
	for name, config := range servers {
		serverConfig, ok := config.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract common properties
		serverType, _ := serverConfig["type"].(string)
		command, _ := serverConfig["command"].(string)
		description, _ := serverConfig["description"].(string)
		url, _ := serverConfig["url"].(string)

		// Extract args if available
		var args []string
		if argsInterface, ok := serverConfig["args"].([]interface{}); ok {
			for _, arg := range argsInterface {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}

		// Extract headers if available
		headers := make(map[string]string)
		if headersInterface, ok := serverConfig["headers"].(map[string]interface{}); ok {
			for k, v := range headersInterface {
				if valStr, ok := v.(string); ok {
					headers[k] = valStr
				}
			}
		}

		// Extract env if available
		env := make(map[string]string)
		if envInterface, ok := serverConfig["env"].(map[string]interface{}); ok {
			for k, v := range envInterface {
				if valStr, ok := v.(string); ok {
					env[k] = valStr
				}
			}
		}

		result = append(result, ServerConfig{
			Source:      source,
			Type:        serverType,
			Command:     command,
			Args:        args,
			URL:         url,
			Headers:     headers,
			Env:         env,
			Name:        name,
			Config:      serverConfig,
			Description: description,
		})
	}

	return result, nil
}

// scanMCPServersConfig scans a config file with mcpServers as the top-level key.
func scanMCPServersConfig(path, source string) ([]ServerConfig, error) {
	data, err := os.ReadFile(path) //nolint
	if err != nil {
		return nil, fmt.Errorf("failed to read %s config: %w", source, err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s config: %w", source, err)
	}

	mcpServers, ok := config["mcpServers"]
	if !ok {
		return nil, fmt.Errorf("no mcpServers key found in %s config", source)
	}

	servers, ok := mcpServers.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid mcpServers format in %s config", source)
	}

	var result []ServerConfig
	for name, serverData := range servers {
		serverConfig, ok := serverData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract common properties
		command, _ := serverConfig["command"].(string)
		url, _ := serverConfig["url"].(string)

		// Extract args if available
		var args []string
		if argsInterface, ok := serverConfig["args"].([]interface{}); ok {
			for _, arg := range argsInterface {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}

		// Extract headers if available
		headers := make(map[string]string)
		if headersInterface, ok := serverConfig["headers"].(map[string]interface{}); ok {
			for k, v := range headersInterface {
				if valStr, ok := v.(string); ok {
					headers[k] = valStr
				}
			}
		}

		// Extract env if available
		env := make(map[string]string)
		if envInterface, ok := serverConfig["env"].(map[string]interface{}); ok {
			for k, v := range envInterface {
				if valStr, ok := v.(string); ok {
					env[k] = valStr
				}
			}
		}

		// Determine type based on whether URL is present
		serverType := ""
		if url != "" {
			serverType = "sse"
		}

		result = append(result, ServerConfig{
			Source:  source,
			Type:    serverType,
			Command: command,
			Args:    args,
			URL:     url,
			Headers: headers,
			Env:     env,
			Name:    name,
			Config:  serverConfig,
		})
	}

	return result, nil
}
