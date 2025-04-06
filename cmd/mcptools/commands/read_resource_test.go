package commands

import (
	"bytes"
	"testing"
)

func TestResourceReadCmd_RunHelp(t *testing.T) {
	// Given: the read resource command is run
	cmd := ReadResourceCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// When: the command is run with the help flag
	cmd.SetArgs([]string{"--help"})

	// Then: no error is returned.
	err := cmd.Execute()
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}

	// Then: the help output is not empty.
	if buf.String() == "" {
		t.Error("Expected help output to not be empty.")
	}
}

func TestReadResourceCmdRun_Success(t *testing.T) {
	t.Setenv("COLS", "120")
	// Given: the read resource command is run
	cmd := ReadResourceCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	// NOTE: we currently truncate output in tabular output.
	cmd.SetArgs([]string{"server", "-f", "json", "arg"})

	// Given: a mock client that returns a successful read resource response
	mockResponse := map[string]any{
		"result": map[string]any{
			"contents": []any{
				map[string]any{
					"uri":      "test://foo",
					"mimeType": "text/plain",
					"text":     "bar",
				},
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

	// When: the command is executed
	err := cmd.Execute()

	// Then: no error is returned
	if err != nil {
		t.Errorf("cmd.Execute() error = %v", err)
	}

	// Then: the expected content is returned
	output := buf.String()
	assertContains(t, output, "test://foo")
	assertContains(t, output, "text/plain")
	assertContains(t, output, "bar")
}
