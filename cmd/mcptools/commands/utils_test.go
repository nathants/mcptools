package commands

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestProcessFlags(t *testing.T) {
	// Save original value to restore later
	originalFormat := FormatOption
	defer func() { FormatOption = originalFormat }()

	// Fix fieldalignment with //nolint directive
	tests := []struct { //nolint:govet
		name       string
		args       []string
		wantArgs   []string
		wantFormat string
	}{
		{
			name:       "no flags",
			args:       []string{"cmd", "arg1", "arg2"},
			wantArgs:   []string{"cmd", "arg1", "arg2"},
			wantFormat: "",
		},
		{
			name:       "with long format flag",
			args:       []string{"cmd", "--format", "json", "arg1"},
			wantArgs:   []string{"cmd", "arg1"},
			wantFormat: "json",
		},
		{
			name:       "with short format flag",
			args:       []string{"cmd", "-f", "pretty", "arg1"},
			wantArgs:   []string{"cmd", "arg1"},
			wantFormat: "pretty",
		},
		{
			name:       "with format flag at end",
			args:       []string{"cmd", "arg1", "--format", "table"},
			wantArgs:   []string{"cmd", "arg1"},
			wantFormat: "table",
		},
		{
			name:       "with invalid format option",
			args:       []string{"cmd", "--format", "invalid", "arg1"},
			wantArgs:   []string{"cmd", "arg1"},
			wantFormat: "invalid",
		},
		{
			name:       "with format flag without value",
			args:       []string{"cmd", "--format", "json"},
			wantArgs:   []string{"cmd"},
			wantFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset FormatOption for each test
			FormatOption = ""

			gotArgs := ProcessFlags(tt.args)

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ProcessFlags() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}

			if FormatOption != tt.wantFormat {
				t.Errorf("ProcessFlags() FormatOption = %v, want %v", FormatOption, tt.wantFormat)
			}
		})
	}
}

func TestFormatAndPrintResponse(t *testing.T) {
	// Save original value to restore later
	originalFormat := FormatOption
	defer func() { FormatOption = originalFormat }()

	// Setup test data
	testResp := map[string]any{
		"name": "test",
		"age":  30,
		"nested": map[string]any{
			"key": "value",
		},
	}
	respJSON := `{"age":30,"name":"test","nested":{"key":"value"}}`
	respPretty := `
{
  "age": 30,
  "name": "test",
  "nested": {
    "key": "value"
  }
}`[1:] // remove first newline
	respTable := `
KEY     VALUE
---     -----
age     30
name    test
nested  {"key":"value"}`[1:] // remove first newline

	// Fix fieldalignment with //nolint directive
	tests := []struct { //nolint:govet
		name      string
		resp      map[string]any
		err       error
		formatOpt string
		wantErr   bool
		expected  string
	}{
		{
			name:      "json format",
			resp:      testResp,
			err:       nil,
			formatOpt: "json",
			wantErr:   false,
			expected:  respJSON,
		},
		{
			name:      "pretty format",
			resp:      testResp,
			err:       nil,
			formatOpt: "pretty",
			wantErr:   false,
			expected:  respPretty,
		},
		{
			name:      "table format",
			resp:      testResp,
			err:       nil,
			formatOpt: "table",
			wantErr:   false,
			expected:  respTable,
		},
		{
			name:      "with error",
			resp:      nil,
			err:       fmt.Errorf("test error"),
			formatOpt: "json",
			wantErr:   true,
			expected:  "error: test error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a cobra command with a buffer
			buf := new(bytes.Buffer)
			cmd := &cobra.Command{}
			cmd.SetOut(buf)

			// Set format option
			FormatOption = tt.formatOpt

			err := FormatAndPrintResponse(cmd, tt.resp, tt.err)

			// Read the captured output
			output := buf.String()

			if (err != nil) != tt.wantErr {
				t.Errorf("FormatAndPrintResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assertEquals(t, strings.TrimSpace(output), tt.expected)
			}
		})
	}
}
