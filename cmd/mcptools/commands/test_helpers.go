// Package commands implements individual commands for the MCP CLI.
package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// MockTransport implements the transport.Transport interface for testing.
type MockTransport struct {
	ExecuteFunc func(method string, params any) (map[string]any, error)
}

func (m *MockTransport) Start(ctx context.Context) error {
	return nil
}

// Execute calls the mock implementation.
func (m *MockTransport) SendRequest(ctx context.Context, request transport.JSONRPCRequest) (*transport.JSONRPCResponse, error) {
	if request.Method == "initialize" {
		return &transport.JSONRPCResponse{Result: json.RawMessage(`{}`)}, nil
	}
	response, err := m.ExecuteFunc(request.Method, request.Params)
	if err != nil {
		return nil, err
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}
	fmt.Println("Returning response:", string(responseBytes))
	return &transport.JSONRPCResponse{Result: json.RawMessage(responseBytes)}, nil
}

func (m *MockTransport) SendNotification(ctx context.Context, notification mcp.JSONRPCNotification) error {
	return nil
}

func (m *MockTransport) SetNotificationHandler(handler func(notification mcp.JSONRPCNotification)) {
}

func (m *MockTransport) Close() error {
	return nil
}

// setupMockClient creates a mock client with the given execute function and returns cleanup function.
func setupMockClient(executeFunc func(method string, _ any) (map[string]any, error)) func() {
	// Save original function and restore later
	originalFunc := CreateClientFuncNew

	mockTransport := &MockTransport{
		ExecuteFunc: executeFunc,
	}

	mockClient := client.NewClient(mockTransport)
	mockClient.Initialize(context.Background(), mcp.InitializeRequest{})

	// Override the function that creates clients
	CreateClientFuncNew = func(_ []string, _ ...client.ClientOption) (*client.Client, error) {
		return mockClient, nil
	}

	// Return a cleanup function
	return func() {
		CreateClientFuncNew = originalFunc
	}
}

// assertContains checks if the output contains the expected string.
func assertContains(t *testing.T, output string, expected string) {
	t.Helper()
	if !bytes.Contains([]byte(output), []byte(expected)) {
		t.Errorf("Expected output to contain %q, got: %s", expected, output)
	}
}

// assertEqual checks if two values are equal.
func assertEquals(t *testing.T, output string, expected string) {
	t.Helper()
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}
}
