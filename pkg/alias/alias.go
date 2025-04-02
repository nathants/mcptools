/*
Package alias implements server alias functionality for MCP.
*/
package alias

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ServerAlias represents a single server command alias.
type ServerAlias struct {
	Command string `json:"command"`
}

// Aliases stores command aliases for MCP servers.
type Aliases map[string]ServerAlias

// GetConfigPath returns the path to the aliases configuration file.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".mcpt")
	mkdirErr := os.MkdirAll(configDir, 0o750)
	if mkdirErr != nil {
		return "", fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	return filepath.Join(configDir, "aliases.json"), nil
}

// Load loads server aliases from the configuration file.
func Load() (Aliases, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	aliases := make(Aliases)

	var statErr error
	if _, statErr = os.Stat(configPath); os.IsNotExist(statErr) {
		return aliases, nil
	}

	configFile, err := os.ReadFile(configPath) // #nosec G304 - configPath is generated internally by GetConfigPath
	if err != nil {
		return nil, fmt.Errorf("failed to read alias config file: %w", err)
	}

	if len(configFile) == 0 {
		return aliases, nil
	}

	if unmarshalErr := json.Unmarshal(configFile, &aliases); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse alias config file: %w", unmarshalErr)
	}

	return aliases, nil
}

// Save saves server aliases to the configuration file.
func Save(aliases Aliases) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	configJSON, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal alias config: %w", err)
	}

	writeErr := os.WriteFile(configPath, configJSON, 0o600) // #nosec G304 - configPath is generated internally by GetConfigPath
	if writeErr != nil {
		return fmt.Errorf("failed to write alias config file: %w", writeErr)
	}

	return nil
}

// GetServerCommand retrieves the server command for a given alias.
func GetServerCommand(aliasName string) (string, bool) {
	aliases, err := Load()
	if err != nil {
		return "", false
	}

	alias, exists := aliases[aliasName]
	if !exists {
		return "", false
	}

	return alias.Command, true
}
