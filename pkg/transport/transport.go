package transport

import (
	"encoding/json"
	"io"
)

// Transport defines the interface for communicating with MCP servers
type Transport interface {
	// Execute sends a request to the MCP server and returns the response
	Execute(method string, params interface{}) (map[string]interface{}, error)
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      int         `json:"id"`
	Params  interface{} `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ReadJSONRPCResponse reads and parses a JSON-RPC response from a reader
func ReadJSONRPCResponse(r io.Reader) (*JSONRPCResponse, error) {
	var response JSONRPCResponse
	if err := json.NewDecoder(r).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
} 