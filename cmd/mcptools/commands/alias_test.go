package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/f/mcptools/pkg/alias"
)

func TestAliasCommands(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()
	if err := os.Setenv("HOME", tmpDir); err != nil {
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}

	// Test alias add command
	t.Run("add", func(t *testing.T) {
		cmd := aliasAddCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		// Run add command
		cmd.SetArgs([]string{"myalias", "echo", "hello"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("add command failed: %v", err)
		}

		// Check output
		output := buf.String()
		if output == "" {
			t.Error("Expected output from add command, got empty string")
		}

		// Verify alias was saved
		aliases, err := alias.Load()
		if err != nil {
			t.Errorf("Failed to load aliases: %v", err)
		}

		serverAlias, exists := aliases["myalias"]
		if !exists {
			t.Error("Alias 'myalias' was not saved")
		} else if serverAlias.Command != "echo hello" {
			t.Errorf("Incorrect command stored. Expected 'echo hello', got '%s'", serverAlias.Command)
		}
	})

	// Test alias list command
	t.Run("list", func(t *testing.T) {
		// First verify the alias exists
		aliases, err := alias.Load()
		if err != nil {
			t.Errorf("Failed to load aliases: %v", err)
		}

		if _, exists := aliases["myalias"]; !exists {
			t.Log("Alias 'myalias' not found, adding it for list test")
			// Add it back if missing
			aliases["myalias"] = alias.ServerAlias{Command: "echo hello"}
			err = alias.Save(aliases)
			if err != nil {
				t.Fatalf("Failed to save alias for test: %v", err)
			}
		}

		// Setup cmd
		cmd := aliasListCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		// Run list command
		err = cmd.Execute()
		if err != nil {
			t.Errorf("list command failed: %v", err)
		}

		// Check output
		output := buf.String()
		if !bytes.Contains(buf.Bytes(), []byte("myalias")) {
			t.Errorf("Expected output to contain 'myalias', got: %s", output)
		}
		if !bytes.Contains(buf.Bytes(), []byte("echo hello")) {
			t.Errorf("Expected output to contain 'echo hello', got: %s", output)
		}
	})

	// Test alias remove command
	t.Run("remove", func(t *testing.T) {
		// First verify the alias exists
		aliases, err := alias.Load()
		if err != nil {
			t.Errorf("Failed to load aliases: %v", err)
		}

		if _, exists := aliases["myalias"]; !exists {
			t.Log("Alias 'myalias' not found, adding it for remove test")
			// Add it back if missing
			aliases["myalias"] = alias.ServerAlias{Command: "echo hello"}
			err = alias.Save(aliases)
			if err != nil {
				t.Fatalf("Failed to save alias for test: %v", err)
			}
		}

		// Setup cmd
		cmd := aliasRemoveCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		// Run remove command
		cmd.SetArgs([]string{"myalias"})
		err = cmd.Execute()
		if err != nil {
			t.Errorf("remove command failed: %v", err)
		}

		// Check output
		output := buf.String()
		if !bytes.Contains(buf.Bytes(), []byte("removed")) {
			t.Errorf("Expected output to contain 'removed', got: %s", output)
		}

		// Verify alias was removed
		updatedAliases, err := alias.Load()
		if err != nil {
			t.Errorf("Failed to load aliases: %v", err)
		}

		if _, exists := updatedAliases["myalias"]; exists {
			t.Error("Alias 'myalias' was not removed")
		}
	})

	// Test remove non-existent alias
	t.Run("remove_nonexistent", func(t *testing.T) {
		// Setup cmd
		cmd := aliasRemoveCmd()
		cmd.SetArgs([]string{"nonexistent"})
		err := cmd.Execute()

		// Should get an error
		if err == nil {
			t.Error("Expected error when removing non-existent alias, got nil")
		}
	})

	// Test main alias command
	t.Run("main_command_help", func(t *testing.T) {
		cmd := AliasCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		cmd.SetArgs([]string{"--help"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("main alias help command failed: %v", err)
		}

		output := buf.String()
		if output == "" {
			t.Error("Expected help output for main alias command, got empty string")
		}
	})
}
