package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/f/mcptools/pkg/transport"
)

const entityTypeValue = "tool"

type MockTransport struct {
	Responses map[string]map[string]interface{}
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		Responses: map[string]map[string]interface{}{},
	}
}

func (t *MockTransport) Execute(method string, params interface{}) (map[string]interface{}, error) {
	if resp, ok := t.Responses[method]; ok {
		return resp, nil
	}

	if method == "tools/list" {
		return map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "test_tool",
					"description": "A test tool",
				},
				{
					"name":        "another_tool",
					"description": "Another test tool",
				},
			},
		}, nil
	}

	if method == "tools/call" {
		paramsMap := params.(map[string]interface{})
		toolName := paramsMap["name"].(string)
		return map[string]interface{}{
			"result": fmt.Sprintf("Called tool: %s", toolName),
		}, nil
	}

	if method == "resources/list" {
		return map[string]interface{}{
			"resources": []map[string]interface{}{
				{
					"uri":         "test_resource",
					"description": "A test resource",
				},
			},
		}, nil
	}

	if method == "resources/read" {
		paramsMap := params.(map[string]interface{})
		uri := paramsMap["uri"].(string)
		return map[string]interface{}{
			"content": fmt.Sprintf("Content of resource: %s", uri),
		}, nil
	}

	if method == "prompts/list" {
		return map[string]interface{}{
			"prompts": []map[string]interface{}{
				{
					"name":        "test_prompt",
					"description": "A test prompt",
				},
			},
		}, nil
	}

	if method == "prompts/get" {
		paramsMap := params.(map[string]interface{})
		promptName := paramsMap["name"].(string)
		return map[string]interface{}{
			"content": fmt.Sprintf("Content of prompt: %s", promptName),
		}, nil
	}

	return map[string]interface{}{}, fmt.Errorf("unknown method: %s", method)
}

// func setupTestCommand() (*cobra.Command, *bytes.Buffer) {
// 	outBuf := &bytes.Buffer{}
//
// 	rootCmd := &cobra.Command{
// 		Use:   "mcp",
// 		Short: "MCP CLI",
// 	}
//
// 	shellCmd := &cobra.Command{
// 		Use:   "shell",
// 		Short: "Interactive shell",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			mockTransport := NewMockTransport()
//
// 			shell := &Shell{
// 				Transport: mockTransport,
// 				Format:    "table",
// 				Reader: strings.NewReader(
// 					"tools\ncall test_tool --params '{\"foo\":\"bar\"}'\ntest_tool {\"foo\":\"bar\"}\nresource:test_resource\nprompt:test_prompt\n/q\n",
// 				),
// 				Writer: outBuf,
// 			}
//
// 			shell.Run()
// 		},
// 	}
//
// 	rootCmd.AddCommand(shellCmd)
//
// 	return rootCmd, outBuf
// }

type Shell struct {
	Transport transport.Transport
	Reader    io.Reader
	Writer    io.Writer
	Format    string
}

func (s *Shell) Run() {
	scanner := bufio.NewScanner(s.Reader)

	for scanner.Scan() {
		input := scanner.Text()

		if input == "/q" || input == "/quit" || input == "exit" {
			fmt.Fprintln(s.Writer, "Exiting MCP shell")
			break
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]
		args := parts[1:]

		switch command {
		case "tools":
			resp, _ := s.Transport.Execute("tools/list", nil)
			fmt.Fprintln(s.Writer, "Tools:", resp)

		case "resources":
			resp, _ := s.Transport.Execute("resources/list", nil)
			fmt.Fprintln(s.Writer, "Resources:", resp)

		case "prompts":
			resp, _ := s.Transport.Execute("prompts/list", nil)
			fmt.Fprintln(s.Writer, "Prompts:", resp)

		case "call":
			if len(args) < 1 {
				fmt.Fprintln(s.Writer, "Usage: call <entity> [--params '{...}']")
				continue
			}

			entityName := args[0]
			entityType := entityTypeValue

			parts = strings.SplitN(entityName, ":", 2)
			if len(parts) == 2 {
				entityType = parts[0]
				entityName = parts[1]
			}

			params := map[string]interface{}{}

			for i := 1; i < len(args); i++ {
				if args[i] == "--params" || args[i] == "-p" {
					if i+1 < len(args) {
						_ = json.Unmarshal([]byte(args[i+1]), &params)
						break
					}
				}
			}

			var resp map[string]interface{}

			switch entityType {
			case "tool":
				resp, _ = s.Transport.Execute("tools/call", map[string]interface{}{
					"name":      entityName,
					"arguments": params,
				})
			case "resource":
				resp, _ = s.Transport.Execute("resources/read", map[string]interface{}{
					"uri": entityName,
				})
			case "prompt":
				resp, _ = s.Transport.Execute("prompts/get", map[string]interface{}{
					"name": entityName,
				})
			}

			fmt.Fprintln(s.Writer, "Call result:", resp)

		default:
			entityName := command
			entityType := entityTypeValue

			parts = strings.SplitN(entityName, ":", 2)
			if len(parts) == 2 {
				entityType = parts[0]
				entityName = parts[1]
			}

			params := map[string]interface{}{}

			if len(args) > 0 {
				firstArg := args[0]
				if strings.HasPrefix(firstArg, "{") && strings.HasSuffix(firstArg, "}") {
					_ = json.Unmarshal([]byte(firstArg), &params)
				} else {
					for i := 0; i < len(args); i++ {
						if args[i] == "--params" || args[i] == "-p" {
							if i+1 < len(args) {
								_ = json.Unmarshal([]byte(args[i+1]), &params)
								break
							}
						}
					}
				}
			}

			var resp map[string]interface{}

			switch entityType {
			case "tool":
				resp, _ = s.Transport.Execute("tools/call", map[string]interface{}{
					"name":      entityName,
					"arguments": params,
				})
				fmt.Fprintln(s.Writer, "Direct tool call result:", resp)
			case "resource":
				resp, _ = s.Transport.Execute("resources/read", map[string]interface{}{
					"uri": entityName,
				})
				fmt.Fprintln(s.Writer, "Direct resource read result:", resp)
			case "prompt":
				resp, _ = s.Transport.Execute("prompts/get", map[string]interface{}{
					"name": entityName,
				})
				fmt.Fprintln(s.Writer, "Direct prompt get result:", resp)
			default:
				fmt.Fprintln(s.Writer, "Unknown command:", command)
			}
		}
	}
}

func TestDirectToolCalling(t *testing.T) {
	testCases := []struct {
		input          string
		expectedOutput string
	}{
		{
			input:          "test_tool {\"param\": \"value\"}",
			expectedOutput: "Called tool: test_tool",
		},
		{
			input:          "resource:test_resource",
			expectedOutput: "Content of resource: test_resource",
		},
		{
			input:          "prompt:test_prompt",
			expectedOutput: "Content of prompt: test_prompt",
		},
	}

	mockTransport := NewMockTransport()

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			outBuf := &bytes.Buffer{}

			shell := &Shell{
				Transport: mockTransport,
				Format:    "table",
				Reader:    strings.NewReader(tc.input + "\n/q\n"),
				Writer:    outBuf,
			}

			shell.Run()

			if !strings.Contains(outBuf.String(), tc.expectedOutput) {
				t.Errorf("Expected output to contain %q, got: %s", tc.expectedOutput, outBuf.String())
			}
		})
	}
}

func TestExecuteShell(t *testing.T) {
	mockTransport := NewMockTransport()

	inputs := []string{
		"tools",
		"resources",
		"prompts",
		"call test_tool --params '{\"foo\":\"bar\"}'",
		"test_tool {\"foo\":\"bar\"}",
		"resource:test_resource",
		"prompt:test_prompt",
		"/q",
	}

	expectedOutputs := []string{
		"A test tool",                        // tools command
		"A test resource",                    // resources command
		"A test prompt",                      // prompts command
		"Called tool: test_tool",             // call command
		"Called tool: test_tool",             // direct tool call
		"Content of resource: test_resource", // direct resource read
		"Content of prompt: test_prompt",     // direct prompt get
		"Exiting MCP shell",                  // quit command
	}

	outBuf := &bytes.Buffer{}

	shell := &Shell{
		Transport: mockTransport,
		Format:    "table",
		Reader:    strings.NewReader(strings.Join(inputs, "\n") + "\n"),
		Writer:    outBuf,
	}

	shell.Run()

	output := outBuf.String()
	for _, expected := range expectedOutputs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it doesn't.\nFull output: %s", expected, output)
		}
	}
}
