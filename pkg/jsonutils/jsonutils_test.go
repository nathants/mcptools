package jsonutils

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestWrapText tests the text wrapping functionality.
func TestWrapText(t *testing.T) {
	testCases := []struct {
		name     string   // String
		text     string   // String
		expected []string // Slice pointer (largest)
		width    int      // Integer (smallest)
	}{
		{
			name:     "empty text",
			text:     "",
			width:    10,
			expected: []string{},
		},
		{
			name:     "single word",
			text:     "hello",
			width:    10,
			expected: []string{"hello"},
		},
		{
			name:     "multiple words fitting in one line",
			text:     "hello world",
			width:    20,
			expected: []string{"hello world"},
		},
		{
			name:     "multiple words requiring wrapping",
			text:     "this is a longer text that needs to be wrapped",
			width:    15,
			expected: []string{"this is a", "longer text", "that needs to", "be wrapped"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := wrapText(tc.text, tc.width)

			if len(result) != len(tc.expected) {
				t.Fatalf("Expected %d lines, got %d", len(tc.expected), len(result))
			}

			for i, line := range result {
				if line != tc.expected[i] {
					t.Errorf("Line %d: expected '%s', got '%s'", i, tc.expected[i], line)
				}
			}
		})
	}
}

// TestGetTermWidth tests the terminal width detection.
func TestGetTermWidth(t *testing.T) {
	// Save original stdout and restore it after the test
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()

	// Non-terminal case should return default width
	r, w, _ := os.Pipe()
	os.Stdout = w

	width := getTermWidth()

	// Close pipe
	if err := w.Close(); err != nil {
		t.Errorf("Error closing pipe: %v", err)
	}
	os.Stdout = origStdout

	// Read and discard pipe content
	_, _ = r.Read(make([]byte, 1024))

	// Should return default width for non-terminal
	if width != 80 {
		t.Errorf("Expected default width 80 for non-terminal, got %d", width)
	}
}

func TestFormat(t *testing.T) {
	testCases := []struct {
		name         string
		data         any
		format       string
		expectPretty bool
		expectError  bool
	}{
		{
			name:         "format json",
			data:         map[string]string{"key": "value"},
			format:       "json",
			expectPretty: false,
		},
		{
			name:         "format pretty",
			data:         map[string]string{"key": "value"},
			format:       "pretty",
			expectPretty: true,
		},
		{
			name:         "format j",
			data:         map[string]string{"key": "value"},
			format:       "j",
			expectPretty: false,
		},
		{
			name:         "format p",
			data:         map[string]string{"key": "value"},
			format:       "p",
			expectPretty: true,
		},
		{
			name:         "format table",
			data:         map[string]string{"key": "value"},
			format:       "table",
			expectPretty: true,
		},
		{
			name:         "format t",
			data:         map[string]string{"key": "value"},
			format:       "t",
			expectPretty: true,
		},
		{
			name:         "format default",
			data:         map[string]string{"key": "value"},
			format:       "unknown",
			expectPretty: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := Format(tc.data, tc.format)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expectPretty {
				if !strings.Contains(output, "\n") {
					t.Errorf("expected pretty output with newlines, got: %s", output)
				}
			} else {
				if strings.Contains(output, "\n") {
					t.Errorf("expected compact JSON without newlines, got: %s", output)
				}
			}

			if !strings.Contains(output, "key") || !strings.Contains(output, "value") {
				t.Errorf("output doesn't contain expected keys and values: %s", output)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	testCases := []struct {
		expected OutputFormat
		input    string
	}{
		{FormatJSON, "json"},
		{FormatJSON, "J"},
		{FormatPretty, "pretty"},
		{FormatPretty, "P"},
		{FormatTable, "table"},
		{FormatTable, "T"},
		{FormatTable, "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseFormat(tc.input)
			if result != tc.expected {
				t.Errorf("ParseFormat(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestToolsListFormatting tests the man-like formatting for tools list.
func TestToolsListFormatting(t *testing.T) {
	tools := []any{
		map[string]any{
			"name":        "tool1",
			"description": "This is a short description",
		},
		map[string]any{
			"name":        "tool2",
			"description": "This is a longer description that should wrap across multiple lines in the table output when displayed to the user",
		},
	}

	// Convert to expected structure
	toolsData := map[string]any{
		"tools": tools,
	}

	output, err := formatTable(toolsData)
	if err != nil {
		t.Fatalf("Error formatting tools list: %v", err)
	}

	// Basic verification
	if !strings.Contains(output, "tool1") || !strings.Contains(output, "tool2") {
		t.Errorf("Missing tool names in output: %s", output)
	}

	if !strings.Contains(output, "This is a short description") {
		t.Errorf("Missing tool description in output: %s", output)
	}

	// Check for man-like format
	if !strings.Contains(output, "     This is") {
		t.Errorf("Missing indented description in output: %s", output)
	}
}

func TestToolListManFormat(t *testing.T) {
	// Create a mock tools list
	tools := []interface{}{
		map[string]interface{}{
			"name":        "get_file_info",
			"description": "Retrieve detailed metadata about a file or directory. Returns comprehensive information including size, creation time, last modified time, permissions, and type.",
			"inputSchema": map[string]interface{}{
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []interface{}{"path"},
			},
		},
		map[string]interface{}{
			"name":        "read_file",
			"description": "Read the contents of a file at the specified path.",
			"inputSchema": map[string]interface{}{
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type": "string",
					},
					"offset": map[string]interface{}{
						"type": "integer",
					},
					"limit": map[string]interface{}{
						"type": "integer",
					},
				},
				"required": []interface{}{"path"},
			},
		},
	}

	// Format the tools list
	result, err := formatToolsList(tools)
	if err != nil {
		t.Fatalf("Failed to format tools list: %v", err)
	}

	// Print the result for manual verification
	fmt.Println("Formatted Tools List:")
	fmt.Println(result)

	// Verify the expected format is present
	expectedSubstrings := []string{
		"get_file_info(path:string)",
		"     Retrieve detailed metadata",
		"read_file(path:string, [limit:integer], [offset:integer])",
		"     Read the contents",
	}

	for _, expected := range expectedSubstrings {
		if !containsSubstring(result, expected) {
			t.Errorf("Expected output to contain %q, but it didn't", expected)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
