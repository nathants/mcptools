package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/f/mcptools/pkg/proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ProxyCmd creates the proxy command.
func ProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Proxy MCP tool requests to shell scripts",
		Long: `Proxy MCP tool requests to shell scripts.

This command allows you to register shell scripts as MCP tools and proxy MCP requests to them.
The scripts will receive tool parameters as environment variables.

Examples:
  # Register a shell script as an MCP tool
  mcp proxy tool add_operation "Adds a and b" "a:int,b:int" ./add.sh

  # Register an inline command as an MCP tool
  mcp proxy tool add_operation "Adds a and b" "a:int,b:int" -e 'echo "total is $a + $b = $(($a+$b))"'

  # Start a proxy server with the registered tools
  mcp proxy start`,
	}

	cmd.AddCommand(ProxyToolCmd())
	cmd.AddCommand(ProxyStartCmd())

	return cmd
}

// ProxyToolCmd creates the proxy tool command.
func ProxyToolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool [name] [description] [parameters] [script]",
		Short: "Register a shell script as an MCP tool",
		Long: `Register a shell script or inline command as an MCP tool that can be called via the MCP protocol.
Parameters should be specified in the format "name:type,name:type,..." where type can be:
- string: Text values
- int: Integer numbers
- float: Floating-point numbers
- bool: Boolean values (true/false)

The script or command will receive parameters as environment variables.

You can either provide a script file path or use the -e flag to specify an inline command.
Example with script:
  mcp proxy tool add_operation "Adds a and b" "a:int,b:int" ./add.sh

Example with inline command:
  mcp proxy tool add_op "Adds given numbers" "a:int,b:int" -e "echo \"total is $a + $b = ${$a+$b}\""

To unregister a tool, use the --unregister flag:
  mcp proxy tool --unregister tool_name`,
		Args: func(cmd *cobra.Command, args []string) error {
			unregister, _ := cmd.Flags().GetBool("unregister")
			if unregister {
				if len(args) != 1 {
					return fmt.Errorf("unregister requires exactly one argument: the tool name")
				}
				return nil
			}
			return cobra.RangeArgs(3, 4)(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			unregister, _ := cmd.Flags().GetBool("unregister")
			if unregister {
				name := args[0]
				// Load existing config
				config, loadErr := LoadProxyConfig()
				if loadErr != nil {
					return fmt.Errorf("error loading config: %w", loadErr)
				}

				// Check if tool exists
				if _, exists := config[name]; !exists {
					return fmt.Errorf("tool %s not found", name)
				}

				// Remove the tool from config
				delete(config, name)

				// Save updated config
				if saveErr := SaveProxyConfig(config); saveErr != nil {
					return fmt.Errorf("error saving config: %w", saveErr)
				}

				fmt.Printf("Unregistered tool: %s\n", name)
				return nil
			}

			name := args[0]
			description := args[1]
			parameters := args[2]
			scriptPath := ""
			if len(args) > 3 {
				scriptPath = args[3]
			}

			// Get the inline command from the -e flag
			command, _ := cmd.Flags().GetString("execute")

			// Either script path or command must be provided
			if scriptPath == "" && command == "" {
				return fmt.Errorf("either script path or command (-e) must be provided")
			}

			// Load existing config
			config, loadErr := LoadProxyConfig()
			if loadErr != nil {
				return fmt.Errorf("error loading config: %w", loadErr)
			}

			// Add the new tool to config
			config[name] = map[string]string{
				"description": description,
				"parameters":  parameters,
				"script":      scriptPath,
				"command":     command,
			}

			// Save updated config
			if saveErr := SaveProxyConfig(config); saveErr != nil {
				return fmt.Errorf("error saving config: %w", saveErr)
			}

			fmt.Printf("Registered tool: %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringP("execute", "e", "", "Inline command to execute instead of a script file")
	cmd.Flags().Bool("unregister", false, "Unregister a tool")
	return cmd
}

// ProxyStartCmd creates the proxy start command.
func ProxyStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a proxy server with registered tools",
		Long: `Start a proxy server that forwards MCP tool requests to shell scripts.

The server reads tool configurations from $HOME/.mcpt/proxy_config.json.

Example:
  mcp proxy start`,
		Run: func(_ *cobra.Command, _ []string) {
			// Load tool configurations
			viper.SetConfigName("proxy_config")
			viper.SetConfigType("json")
			viper.AddConfigPath("$HOME/.mcpt")

			if err := viper.ReadInConfig(); err != nil {
				log.Fatalf("Error reading config: %v", err)
			}

			var config map[string]map[string]string
			if err := viper.Unmarshal(&config); err != nil {
				log.Fatalf("Error unmarshaling config: %v", err)
			}

			// Run proxy server
			fmt.Fprintln(os.Stderr, "Starting proxy server...")
			if err := proxy.RunProxyServer(config); err != nil {
				log.Fatalf("Error running proxy server: %v", err)
			}
		},
	}

	return cmd
}

// LoadProxyConfig loads the proxy configuration from the config file.
func LoadProxyConfig() (map[string]map[string]string, error) {
	// Initialize config
	viper.SetConfigName("proxy_config")
	viper.SetConfigType("json")
	viper.AddConfigPath("$HOME/.mcpt")

	// Create config directory if it doesn't exist
	configDir := os.ExpandEnv("$HOME/.mcpt")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return nil, fmt.Errorf("error creating config directory: %w", err)
	}

	// Load existing config if it exists
	var config map[string]map[string]string
	var configFileNotFound viper.ConfigFileNotFoundError
	err := viper.ReadInConfig()
	if err != nil {
		if errors.As(err, &configFileNotFound) {
			// Config file not found, create a new one
			config = make(map[string]map[string]string)
		} else {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	} else {
		// Config file found, unmarshal it
		config = make(map[string]map[string]string)
		unmarshalErr := viper.Unmarshal(&config)
		if unmarshalErr != nil {
			return nil, fmt.Errorf("error unmarshaling config: %w", unmarshalErr)
		}
	}

	return config, nil
}

// SaveProxyConfig saves the proxy configuration to the config file.
func SaveProxyConfig(config map[string]map[string]string) error {
	configPath := os.ExpandEnv("$HOME/.mcpt/proxy_config.json")
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	writeErr := os.WriteFile(configPath, configJSON, 0o600)
	if writeErr != nil {
		return fmt.Errorf("error writing config: %w", writeErr)
	}

	return nil
}
