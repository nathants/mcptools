package commands

import (
	"bytes"
	"testing"
)

func TestToolsCmdRun_Help(t *testing.T) {
	// Test that the help flag displays help text
	cmd := ToolsCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Execute with help flag
	cmd.SetArgs([]string{"--help"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}

	// Check that help output is not empty
	if buf.String() == "" {
		t.Error("Expected help output, got empty string")
	}
}

func TestToolsCmdRun(t *testing.T) {
	// Save original format option
	origFormatOption := FormatOption
	defer func() { FormatOption = origFormatOption }()

	// Create a mock client that returns successful response
	mockResponse := map[string]any{
		"tools": []any{
			map[string]any{
				"name":        "test-tool",
				"description": "A test tool",
			},
		},
	}

	cleanup := setupMockClient(func(method string, _ any) (map[string]any, error) {
		if method != "tools/list" {
			t.Errorf("Expected method 'tools/list', got %q", method)
		}
		return mockResponse, nil
	})
	defer cleanup()

	// Set up command
	cmd := ToolsCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"server", "args"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}
	output := buf.String()
	assertContains(t, output, "test-tool")
	assertContains(t, output, "A test tool")
}
