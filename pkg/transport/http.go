package transport

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// HTTP implements the Transport interface by communicating with a MCP server over HTTP using JSON-RPC.
type HTTP struct {
	eventCh chan string
	address string
	debug   bool
	nextID  int
}

// NewHTTP creates a new Http transport that will execute the given command.
// It communicates with the command using JSON-RPC over HTTP.
// Currently Http transport is implements MCP's Final draft version 2024-11-05,
// https://spec.modelcontextprotocol.io/specification/2024-11-05/basic/transports/#http-with-sse
func NewHTTP(address string) (*HTTP, error) {
	debug := os.Getenv("MCP_DEBUG") == "1"

	_, uriErr := url.ParseRequestURI(address)
	if uriErr != nil {
		return nil, fmt.Errorf("invalid address: %w", uriErr)
	}

	resp, err := http.Get(address + "/sse")
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	eventCh := make(chan string, 1)

	go func() {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to close response body: %v\n", closeErr)
			}
		}()

		reader := bufio.NewReader(resp.Body)
		for {
			line, lineErr := reader.ReadString('\n')
			if lineErr != nil {
				fmt.Fprintf(os.Stderr, "SSE read error: %v\n", lineErr)
				return
			}
			line = strings.TrimSpace(line)
			if debug {
				fmt.Fprintf(os.Stderr, "DEBUG: Received SSE: %s\n", line)
			}
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(line[5:])
				select {
				case eventCh <- data:
				default:
				}
			}
		}
	}()

	// First event we receive from SSE is the message address. We will use this endpoint to keep
	// a session alive.
	var messageAddress string
	select {
	case msg := <-eventCh:
		messageAddress = msg
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for SSE response")
	}

	client := &HTTP{
		// Use the SSE message address as the base address for the HTTP transport
		address: address + "/sse" + messageAddress,
		nextID:  1,
		debug:   debug,
		eventCh: eventCh,
	}

	// Send initialize request
	_, err = client.Execute("initialize", map[string]any{
		"clientInfo": map[string]any{
			"name":    "mcp-client",
			"version": "0.1.0",
		},
		"capabilities":    map[string]any{},
		"protocolVersion": "2024-11-05",
	})
	if err != nil {
		return nil, fmt.Errorf("error sending initialize request: %w", err)
	}

	// Send intialized notification
	if err := client.send("notifications/initialized", nil); err != nil {
		return nil, fmt.Errorf("error sending initialized notification: %w", err)
	}

	return client, nil
}

// Execute implements the Transport via JSON-RPC over HTTP.
func (t *HTTP) Execute(method string, params any) (map[string]any, error) {
	if err := t.send(method, params); err != nil {
		return nil, err
	}

	// After sending the request, we listen the SSE channel for the response
	var response Response
	select {
	case msg := <-t.eventCh:
		if unmarshalErr := json.Unmarshal([]byte(msg), &response); unmarshalErr != nil {
			return nil, fmt.Errorf("error unmarshaling response: %w, response: %s", unmarshalErr, msg)
		}
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for SSE response")
	}

	if response.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Successfully parsed response\n")
	}

	return response.Result, nil
}

func (t *HTTP) send(method string, params any) error {
	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Connecting to server: %s\n", t.address)
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
		return fmt.Errorf("error marshaling request: %w", err)
	}

	requestJSON = append(requestJSON, '\n')

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Sending request: %s\n", string(requestJSON))
	}

	resp, err := http.Post(t.address, "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Sent request to server\n")
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to close response body: %v\n", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	if t.debug {
		fmt.Fprintf(os.Stderr, "DEBUG: Read from server: %s\n", string(body))
	}

	return nil
}
