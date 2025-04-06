package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestCallCmdRun_Help(t *testing.T) {
	// Test that the help flag displays help text
	cmd := CallCmd()
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

func TestCallCmdRun_Tool(t *testing.T) {
	// Create a mock client that returns successful response
	mockResponse := map[string]any{
		"content": []any{
			map[string]any{
				"type": "text",
				"text": "Tool executed successfully",
			},
		},
	}

	cleanup := setupMockClient(func(method string, _ any) (map[string]any, error) {
		if method != "tools/call" {
			t.Errorf("Expected method 'tools/call', got %q", method)
		}
		return mockResponse, nil
	})
	defer cleanup()

	// Set up command
	cmd := CallCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Execute command with tool
	cmd.SetArgs([]string{"test-tool", "--params", `{"key":"value"}`, "server", "arg"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}

	// Verify output contains expected content
	output := strings.TrimSpace(buf.String())
	expectedOutput := "Tool executed successfully"
	assertEquals(t, output, expectedOutput)
}

func TestCallCmdRun_Resource(t *testing.T) {
	// Create a mock client that returns successful response
	mockResponse := map[string]any{
		"contents": []any{
			map[string]any{
				"uri":      "test://foo",
				"mimeType": "text/plain",
				"text":     "bar",
			},
		},
	}

	cleanup := setupMockClient(func(method string, _ any) (map[string]any, error) {
		if method != "resources/read" {
			t.Errorf("Expected method 'resources/read', got %q", method)
		}
		return mockResponse, nil
	})
	defer cleanup()

	// Set up command
	cmd := CallCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Execute command with resource
	cmd.SetArgs([]string{"resource:test-resource", "-f", "json", "server", "arg"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}

	// Verify output contains expected content
	output := buf.String()
	expectedOutput := `{"contents":[{"mimeType":"text/plain","text":"bar","uri":"test://foo"}]}`
	assertContains(t, output, expectedOutput)
}
