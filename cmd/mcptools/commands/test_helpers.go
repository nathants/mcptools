// Package commands implements individual commands for the MCP CLI.
package commands

import (
	"bytes"
	"testing"

	"github.com/f/mcptools/pkg/client"
)

// MockTransport implements the transport.Transport interface for testing.
type MockTransport struct {
	ExecuteFunc func(method string, params any) (map[string]any, error)
}

// Execute calls the mock implementation.
func (m *MockTransport) Execute(method string, params any) (map[string]any, error) {
	return m.ExecuteFunc(method, params)
}

// setupMockClient creates a mock client with the given execute function and returns cleanup function.
func setupMockClient(executeFunc func(method string, _ any) (map[string]any, error)) func() {
	// Save original function and restore later
	originalFunc := CreateClientFunc

	mockTransport := &MockTransport{
		ExecuteFunc: executeFunc,
	}

	mockClient := client.NewWithTransport(mockTransport)

	// Override the function that creates clients
	CreateClientFunc = func(_ []string) (*client.Client, error) {
		return mockClient, nil
	}

	// Return a cleanup function
	return func() {
		CreateClientFunc = originalFunc
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
