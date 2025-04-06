package commands

import (
	"bytes"
	"testing"
)

func TestResourcesCmdRun_Help(t *testing.T) {
	// Test that the help flag displays help text
	cmd := ResourcesCmd()
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

func TestResourcesCmdRun_Success(t *testing.T) {
	// Create a mock client that returns successful response
	mockResponse := map[string]any{
		"resources": []any{
			map[string]any{
				"uri":         "test://resource",
				"mimeType":    "text/plain",
				"name":        "TestResource",
				"description": "Test resource description",
			},
		},
	}

	cleanup := setupMockClient(func(method string, _ any) (map[string]any, error) {
		if method != "resources/list" {
			t.Errorf("Expected method 'resources/list', got %q", method)
		}
		return mockResponse, nil
	})
	defer cleanup()

	// Set up command
	cmd := ResourcesCmd()
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
	assertContains(t, output, "TestResource")
	assertContains(t, output, "test://resource")
	assertContains(t, output, "text/plain")
	assertContains(t, output, "Test resource description")
}
