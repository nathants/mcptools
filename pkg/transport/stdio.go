package transport

import (
	"bufio"
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
	stdin, stdout, cmd, stderrBuf, err := t.setupCommand()
	if err != nil {
		return nil, err
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Starting initialization\n")
	}

	if initErr := t.initialize(stdin, stdout); initErr != nil {
		if t.debug {
			fmt.Fprintf(os.Stderr, "DEBUG: Initialization failed: %v\n", initErr)
			if stderrBuf.Len() > 0 {
				fmt.Fprintf(os.Stderr, "DEBUG: stderr during init: %s\n", stderrBuf.String())
			}
		}
		return nil, initErr
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Initialization successful, sending method request\n")
	}

	request := Request{
		JSONRPC: "2.0",
		Method:  method,
		ID:      t.nextID,
		Params:  params,
	}
	t.nextID++

	if sendErr := t.sendRequest(stdin, request); sendErr != nil {
		return nil, sendErr
	}
	_ = stdin.Close()

	response, err := t.readResponse(stdout)
	if err != nil {
		return nil, err
	}

	waitErr := cmd.Wait()

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Command completed with err: %v\n", waitErr)
		if stderrBuf.Len() > 0 {
			fmt.Fprintf(os.Stderr, "DEBUG: stderr output:\n%s\n", stderrBuf.String())
		}
	}

	if waitErr != nil && stderrBuf.Len() > 0 {
		return nil, fmt.Errorf("command error: %w, stderr: %s", waitErr, stderrBuf.String())
	}

	return response.Result, nil
}

// setupCommand prepares and starts the command, returning the stdin/stdout pipes and any error.
func (t *Stdio) setupCommand() (stdin io.WriteCloser, stdout io.ReadCloser, cmd *exec.Cmd, stderrBuf *bytes.Buffer, err error) {
	if len(t.command) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("no command specified for stdio transport")
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Executing command: %v\n", t.command)
	}

	cmd = exec.Command(t.command[0], t.command[1:]...) // #nosec G204

	stdin, err = cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error getting stdin pipe: %w", err)
	}

	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error getting stdout pipe: %w", err)
	}

	stderrBuf = &bytes.Buffer{}
	cmd.Stderr = stderrBuf

	if err = cmd.Start(); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error starting command: %w", err)
	}

	return stdin, stdout, cmd, stderrBuf, nil
}

// initialize sends the initialization request and waits for response and then sends the initialized
// notification.
func (t *Stdio) initialize(stdin io.WriteCloser, stdout io.ReadCloser) error {
	initRequest := Request{
		JSONRPC: "2.0",
		Method:  "initialize",
		ID:      t.nextID,
		Params: map[string]any{
			"clientInfo": map[string]any{
				"name":    "f/mcptools",
				"version": "beta",
			},
			"protocolVersion": protocolVersion,
			"capabilities":    map[string]any{},
		},
	}
	t.nextID++

	if err := t.sendRequest(stdin, initRequest); err != nil {
		return fmt.Errorf("init request failed: %w", err)
	}

	_, err := t.readResponse(stdout)
	if err != nil {
		return fmt.Errorf("init response failed: %w", err)
	}

	initNotification := Request{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	if sendErr := t.sendRequest(stdin, initNotification); sendErr != nil {
		return fmt.Errorf("init notification failed: %w", sendErr)
	}

	return nil
}

// sendRequest sends a JSON-RPC request and returns the marshaled request.
func (t *Stdio) sendRequest(stdin io.WriteCloser, request Request) error {
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}
	requestJSON = append(requestJSON, '\n')

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Preparing to send request: %s\n", string(requestJSON))
	}

	writer := bufio.NewWriter(stdin)
	n, err := writer.Write(requestJSON)
	if err != nil {
		return fmt.Errorf("error writing bytes to stdin: %w", err)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Wrote %d bytes\n", n)
	}

	if flushErr := writer.Flush(); flushErr != nil {
		return fmt.Errorf("error flushing bytes to stdin: %w", flushErr)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Successfully flushed bytes\n")
	}

	return nil
}

// readResponse reads and parses a JSON-RPC response.
func (t *Stdio) readResponse(stdout io.ReadCloser) (*Response, error) {
	reader := bufio.NewReader(stdout)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading from stdout: %w", err)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Read from stdout: %s", string(line))
	}

	if len(line) == 0 {
		return nil, fmt.Errorf("no response from command")
	}

	var response Response
	if unmarshalErr := json.Unmarshal(line, &response); unmarshalErr != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w, response: %s", unmarshalErr, string(line))
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Successfully parsed response\n")
	}

	return &response, nil
}
