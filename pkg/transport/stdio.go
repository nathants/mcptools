package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Stdio implements the Transport interface by executing a command
// and communicating with it via stdin/stdout using JSON-RPC.
type Stdio struct {
	command []string
	nextID  int
	debug   bool
}

// NewStdio creates a new Stdio transport that will execute the given command.
// It communicates with the command using JSON-RPC over stdin/stdout.
func NewStdio(command []string) *Stdio {
	debug := os.Getenv("MCP_DEBUG") == "1"
	return &Stdio{
		command: command,
		nextID:  1,
		debug:   debug,
	}
}

// Execute implements the Transport interface by spawning a subprocess
// and communicating with it via JSON-RPC over stdin/stdout.
func (t *Stdio) Execute(method string, params any) (map[string]any, error) {
	if len(t.command) == 0 {
		return nil, fmt.Errorf("no command specified for stdio transport")
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Executing command: %v\n", t.command)
	}

	request := Request{
		JSONRPC: "2.0",
		Method:  method,
		ID:      t.nextID,
		Params:  params,
	}
	t.nextID++

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	requestJSON = append(requestJSON, '\n')

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Sending request: %s\n", string(requestJSON))
	}

	cmd := exec.Command(t.command[0], t.command[1:]...) // #nosec G204

	stdin, stdinErr := cmd.StdinPipe()
	if stdinErr != nil {
		return nil, fmt.Errorf("error getting stdin pipe: %w", stdinErr)
	}

	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		return nil, fmt.Errorf("error getting stdout pipe: %w", stdoutErr)
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if startErr := cmd.Start(); startErr != nil {
		return nil, fmt.Errorf("error starting command: %w", startErr)
	}

	if _, writeErr := stdin.Write(requestJSON); writeErr != nil {
		return nil, fmt.Errorf("error writing to stdin: %w", writeErr)
	}
	_ = stdin.Close()

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Wrote request to stdin\n")
	}

	var respBytes bytes.Buffer
	if _, copyErr := io.Copy(&respBytes, stdout); copyErr != nil {
		return nil, fmt.Errorf("error reading from stdout: %w", copyErr)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Read from stdout: %s\n", respBytes.String())
	}

	waitErr := cmd.Wait()

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Command completed with err: %v\n", waitErr)
		if stderrBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: stderr output: %s\n", stderrBuf.String())
		}
	}

	if waitErr != nil && stderrBuf.Len() > 0 {
		return nil, fmt.Errorf("command error: %w, stderr: %s", waitErr, stderrBuf.String())
	}

	if respBytes.Len() == 0 {
		if stderrBuf.Len() > 0 {
			return nil, fmt.Errorf("no response from command, stderr: %s", stderrBuf.String())
		}
		return nil, fmt.Errorf("no response from command")
	}

	var response Response
	if unmarshalErr := json.Unmarshal(respBytes.Bytes(), &response); unmarshalErr != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w, response: %s", unmarshalErr, respBytes.String())
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Successfully parsed response\n")
	}

	return response.Result, nil
}
