package commands

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	// Save original stdout and restore it at the end
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set a test version
	oldVersion := Version
	Version = "test-version"
	defer func() { Version = oldVersion }()

	// Execute the version command
	cmd := NewVersionCmd()
	cmd.Execute()

	// Close the write end of the pipe to read from it
	w.Close()

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check that the version is in the output
	expectedOutput := "MCP Tools version test-version\n"
	if output != expectedOutput {
		t.Errorf("Expected output %q, got %q", expectedOutput, output)
	}
}

func TestVersionCmdWorks(t *testing.T) {
	// Test that the version command can be created and executed
	cmd := NewVersionCmd()
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