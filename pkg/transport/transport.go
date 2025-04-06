// Package transport contains implementatations for different transport options for MCP.
package transport

import (
	"encoding/json"
	"io"
)

const (
	protocolVersion = "2024-11-05"
)

// Transport defines the interface for communicating with MCP servers.
// Implementations should handle the specifics of communication protocols.
type Transport interface {
	Execute(method string, params any) (map[string]any, error)
}

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	Params  any    `json:"params,omitempty"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	ID      int    `json:"id,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	Result  map[string]any `json:"result,omitempty"`
	Error   *Error         `json:"error,omitempty"`
	JSONRPC string         `json:"jsonrpc"`
	ID      int            `json:"id"`
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ParseResponse reads and parses a JSON-RPC response from a reader.
func ParseResponse(r io.Reader) (*Response, error) {
	var response Response
	if err := json.NewDecoder(r).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}
