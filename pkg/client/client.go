package client

import (
	"strings"

	"github.com/f/mcptools/pkg/transport"
)

// Client represents an MCP client
type Client struct {
	Transport transport.Transport
}

// New creates a new MCP client with HTTP transport
func New(baseURL string) *Client {
	return &Client{
		Transport: transport.NewHTTP(baseURL),
	}
}

// NewWithTransport creates a new MCP client with the given transport
func NewWithTransport(t transport.Transport) *Client {
	return &Client{
		Transport: t,
	}
}

// NewStdio creates a new MCP client with Stdio transport
func NewStdio(command []string) *Client {
	return &Client{
		Transport: transport.NewStdio(command),
	}
}

// ListTools lists all available tools on the MCP server
func (c *Client) ListTools() (map[string]interface{}, error) {
	return c.Transport.Execute("tools/list", nil)
}

// ListResources lists all available resources on the MCP server
func (c *Client) ListResources() (map[string]interface{}, error) {
	return c.Transport.Execute("resources/list", nil)
}

// ListPrompts lists all available prompts on the MCP server
func (c *Client) ListPrompts() (map[string]interface{}, error) {
	return c.Transport.Execute("prompts/list", nil)
}

// CallTool calls a specific tool on the MCP server
func (c *Client) CallTool(toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}
	return c.Transport.Execute("tools/call", params)
}

// GetPrompt gets a specific prompt from the MCP server
func (c *Client) GetPrompt(promptName string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"name": promptName,
	}
	return c.Transport.Execute("prompts/get", params)
}

// ReadResource reads a specific resource from the MCP server
func (c *Client) ReadResource(uri string) (map[string]interface{}, error) {
	params := map[string]interface{}{
		"uri": uri,
	}
	return c.Transport.Execute("resources/read", params)
}

// ParseCommandString parses a command string into a slice of command arguments
func ParseCommandString(cmdStr string) []string {
	if cmdStr == "" {
		return nil
	}
	
	// Simple split by space - in a real implementation, you'd handle quotes and escapes better
	return strings.Fields(cmdStr)
} 