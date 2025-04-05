package commands

import (
	"bytes"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	buf := new(bytes.Buffer)

	oldVersion := Version
	Version = "test-version"
	defer func() { Version = oldVersion }()

	// Execute the version command with our buffer
	cmd := VersionCmd()
	cmd.SetOut(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Failed to execute version command: %v", err)
	}

	// Read captured output
	output := buf.String()

	// Check that the version is in the output
	expectedOutput := "MCP Tools version test-version\n"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestVersionCmdWorks(t *testing.T) {
	// Test that the version command can be created and executed
	cmd := VersionCmd()
	if cmd == nil {
		t.Fatal("Expected version command to be created")
	}

	// Verify the command properties
	if cmd.Use != "version" {
		t.Errorf("Expected Use to be 'version', got %q", cmd.Use)
	}

	if cmd.Short != "Print the version information" {
		t.Errorf("Expected Short to be 'Print the version information', got %q", cmd.Short)
	}

	// Ensure Run function is not nil
	if cmd.Run == nil {
		t.Error("Expected Run function to be defined")
	}
}
