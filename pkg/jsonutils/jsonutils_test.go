package jsonutils

import (
	"strings"
	"testing"
)

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
		input    string
		expected OutputFormat
	}{
		{"json", FormatJSON},
		{"J", FormatJSON},
		{"pretty", FormatPretty},
		{"P", FormatPretty},
		{"table", FormatTable},
		{"T", FormatTable},
		{"unknown", FormatTable}, // Default is table
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
