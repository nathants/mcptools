package formatter

import (
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	testCases := []struct {
		name         string
		data         interface{}
		format       string
		expectPretty bool
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
			name:         "format default",
			data:         map[string]string{"key": "value"},
			format:       "unknown",
			expectPretty: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := Format(tc.data, tc.format)
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