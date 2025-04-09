/*
Package client implements mcp client functionality.
*/
package client

import (
	"fmt"
	"os"
	"strings"

	"github.com/f/mcptools/pkg/transport"
)

// Client provides an interface to interact with MCP servers.
// It abstracts away the transport mechanism so callers don't need
// to worry about the details of HTTP, stdio, etc.
type Client struct {
	transport transport.Transport
}

// Option provides a way for passing options to the Client to change its
// configuration.
type Option func(*Client)

// CloseTransportAfterExecute allows keeping a transport alive if supported by
// the transport.
func CloseTransportAfterExecute(closeTransport bool) Option {
	return func(c *Client) {
		t, ok := c.transport.(interface{ SetCloseAfterExecute(bool) })
		if ok {
			t.SetCloseAfterExecute(closeTransport)
		}
	}
}

// NewWithTransport creates a new MCP client using the provided transport.
// This allows callers to provide a custom transport implementation.
func NewWithTransport(t transport.Transport) *Client {
	return &Client{
		transport: t,
	}
}

// NewStdio creates a new MCP client that communicates with a command
// via stdin/stdout using JSON-RPC.
func NewStdio(command []string) *Client {
	return &Client{
		transport: transport.NewStdio(command),
	}
}

// NewHTTP creates a MCP client that communicates with a server via HTTP using JSON-RPC.
func NewHTTP(address string) *Client {
	transport, err := transport.NewHTTP(address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating HTTP transport: %s\n", err)
		os.Exit(1)
	}

	return &Client{
		transport: transport,
	}
}

// ListTools retrieves the list of available tools from the MCP server.
func (c *Client) ListTools() (map[string]any, error) {
	return c.transport.Execute("tools/list", nil)
}

// ListResources retrieves the list of available resources from the MCP server.
func (c *Client) ListResources() (map[string]any, error) {
	return c.transport.Execute("resources/list", nil)
}

// ListPrompts retrieves the list of available prompts from the MCP server.
func (c *Client) ListPrompts() (map[string]any, error) {
	return c.transport.Execute("prompts/list", nil)
}

// CallTool calls a specific tool on the MCP server with the given arguments.
func (c *Client) CallTool(toolName string, args map[string]any) (map[string]any, error) {
	params := map[string]any{
		"name":      toolName,
		"arguments": args,
	}
	return c.transport.Execute("tools/call", params)
}

// GetPrompt retrieves a specific prompt from the MCP server.
func (c *Client) GetPrompt(promptName string) (map[string]any, error) {
	params := map[string]any{
		"name": promptName,
	}
	return c.transport.Execute("prompts/get", params)
}

// ReadResource reads the content of a specific resource from the MCP server.
func (c *Client) ReadResource(uri string) (map[string]any, error) {
	params := map[string]any{
		"uri": uri,
	}
	return c.transport.Execute("resources/read", params)
}

// ParseCommandString splits a command string into separate arguments,
// respecting spaces as argument separators.
// Note: This is a simple implementation that doesn't handle quotes or escapes.
func ParseCommandString(cmdStr string) []string {
	if cmdStr == "" {
		return nil
	}

	return strings.Fields(cmdStr)
}
