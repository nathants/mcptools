package commands

import (
	"bytes"
	"testing"
)

func TestPromptsCmdRun_Help(t *testing.T) {
	// Test that the help flag displays help text
	cmd := PromptsCmd()
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

func TestPromptsCmdRun_Success(t *testing.T) {
	// Create a mock client that returns successful response
	mockResponse := map[string]any{
		"prompts": []any{
			map[string]any{
				"name":        "test-prompt",
				"description": "Test prompt description",
			},
		},
	}

	cleanup := setupMockClient(func(method string, _ any) (map[string]any, error) {
		if method != "prompts/list" {
			t.Errorf("Expected method 'prompts/list', got %q", method)
		}
		return mockResponse, nil
	})
	defer cleanup()

	// Set up command
	cmd := PromptsCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Execute command
	cmd.SetArgs([]string{"server", "arg"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}

	// Verify output contains expected content
	output := buf.String()
	assertContains(t, output, "test-prompt")
	assertContains(t, output, "Test prompt description")
}
