package transport

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"bytes"
	"io"
	"os"
)

// StdioTransport provides Stdio transport for MCP servers
type StdioTransport struct {
	Command []string
	NextID  int
	Debug   bool
}

// NewStdio creates a new Stdio transport
func NewStdio(command []string) *StdioTransport {
	debug := os.Getenv("MCP_DEBUG") == "1"
	return &StdioTransport{
		Command: command,
		NextID:  1,
		Debug:   debug,
	}
}

// Execute sends a request to the MCP server and returns the response
func (t *StdioTransport) Execute(method string, params interface{}) (map[string]interface{}, error) {
	if len(t.Command) == 0 {
		return nil, fmt.Errorf("no command specified for stdio transport")
	}

	if t.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Executing command: %v\n", t.Command)
	}

	// Create the JSON-RPC request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		ID:      t.NextID,
		Params:  params,
	}
	t.NextID++

	// Marshal the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}
	
	// Add newline to the request
	requestJSON = append(requestJSON, '\n')

	if t.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Sending request: %s\n", string(requestJSON))
	}

	// Create the command
	cmd := exec.Command(t.Command[0], t.Command[1:]...)
	
	// Create stdin and stdout pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("error getting stdin pipe: %w", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error getting stdout pipe: %w", err)
	}
	
	// Capture stderr for debugging
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	
	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error starting command: %w", err)
	}

	// Write request to stdin and close it
	if _, err := stdin.Write(requestJSON); err != nil {
		return nil, fmt.Errorf("error writing to stdin: %w", err)
	}
	stdin.Close()
	
	if t.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Wrote request to stdin\n")
	}
	
	// Read response from stdout
	var respBytes bytes.Buffer
	if _, err := io.Copy(&respBytes, stdout); err != nil {
		return nil, fmt.Errorf("error reading from stdout: %w", err)
	}
	
	if t.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Read from stdout: %s\n", respBytes.String())
	}
	
	// Wait for the command to complete
	err = cmd.Wait()
	
	if t.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Command completed with err: %v\n", err)
		if stderrBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: stderr output: %s\n", stderrBuf.String())
		}
	}
	
	// If we have stderr output and an error, include it in the error message
	if err != nil && stderrBuf.Len() > 0 {
		return nil, fmt.Errorf("command error: %w, stderr: %s", err, stderrBuf.String())
	}
	
	// If we didn't get any response, this might be a command error
	if respBytes.Len() == 0 {
		if stderrBuf.Len() > 0 {
			return nil, fmt.Errorf("no response from command, stderr: %s", stderrBuf.String())
		}
		return nil, fmt.Errorf("no response from command")
	}
	
	// Parse the response
	var response JSONRPCResponse
	if err := json.Unmarshal(respBytes.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w, response: %s", err, respBytes.String())
	}

	// Check for error in response
	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	if t.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Successfully parsed response\n")
	}

	return response.Result, nil
} 