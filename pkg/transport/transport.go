package transport

import (
	"encoding/json"
	"io"
)

// Transport defines the interface for communicating with MCP servers.
// Implementations should handle the specifics of communication protocols.
type Transport interface {
	// Execute sends a request to the MCP server and returns the response.
	// The method parameter specifies the RPC method to call, and params contains
	// the parameters to pass to that method.
	Execute(method string, params any) (map[string]any, error)
}

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int    `json:"id"`
	Params  any    `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Result  map[string]any `json:"result,omitempty"`
	Error   *Error         `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ParseResponse reads and parses a JSON-RPC response from a reader.
func ParseResponse(r io.Reader) (*Response, error) {
	var response Response
	if err := json.NewDecoder(r).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}
