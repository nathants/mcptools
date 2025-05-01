package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/f/mcptools/pkg/alias"
	"github.com/f/mcptools/pkg/client"
	"github.com/f/mcptools/pkg/jsonutils"
	sdkclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/spf13/cobra"
)

// sentinel errors.
var (
	ErrCommandRequired = fmt.Errorf("command to execute is required when using stdio transport")
)

// IsHTTP returns true if the string is a valid HTTP URL.
func IsHTTP(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// CreateClientFunc is the function used to create MCP clients.
// This can be replaced in tests to use a mock transport.
var CreateClientFunc = func(args []string, opts ...client.Option) (*client.Client, error) {
	if len(args) == 0 {
		return nil, ErrCommandRequired
	}

	opts = append(opts, client.SetShowServerLogs(ShowServerLogs))

	// Check if the first argument is an alias
	if len(args) == 1 {
		server, found := alias.GetServerCommand(args[0])
		if found {
			if IsHTTP(server) {
				return client.NewHTTP(server), nil
			}
			cmdParts := client.ParseCommandString(server)
			c := client.NewStdio(cmdParts, opts...)
			return c, nil
		}
	}

	if len(args) == 1 && IsHTTP(args[0]) {
		return client.NewHTTP(args[0]), nil
	}

	c := client.NewStdio(args, opts...)

	return c, nil
}

// CreateClientFuncNew is the function used to create MCP clients.
// This can be replaced in tests to use a mock transport.
var CreateClientFuncNew = func(args []string, _ ...sdkclient.ClientOption) (*sdkclient.Client, error) {
	if len(args) == 0 {
		return nil, ErrCommandRequired
	}

	// opts = append(opts, client.SetShowServerLogs(ShowServerLogs))

	// Check if the first argument is an alias
	if len(args) == 1 {
		server, found := alias.GetServerCommand(args[0])
		if found {
			args = ParseCommandString(server)
		}
	}

	var c *sdkclient.Client
	var err error

	if len(args) == 1 && IsHTTP(args[0]) {
		c, err = sdkclient.NewSSEMCPClient(args[0])
	} else {
		c, err = sdkclient.NewStdioMCPClient(args[0], nil, args[1:]...)
	}

	if err != nil {
		return nil, err
	}

	_, initErr := c.Initialize(context.Background(), mcp.InitializeRequest{})
	if initErr != nil {
		return nil, initErr
	}

	return c, nil
}

// ProcessFlags processes command line flags, sets the format option, and returns the remaining
// arguments. Supported format options: json, pretty, and table.
//
// For example, if the input arguments are ["tools", "--format", "pretty", "npx", "-y",
// "@modelcontextprotocol/server-filesystem", "~"], it would return ["npx", "-y",
// "@modelcontextprotocol/server-filesystem", "~"] and set the format option to "pretty".
func ProcessFlags(args []string) []string {
	parsedArgs := []string{}

	i := 0
	for i < len(args) {
		switch {
		case (args[i] == FlagFormat || args[i] == FlagFormatShort) && i+1 < len(args):
			FormatOption = args[i+1]
			i += 2
		case args[i] == FlagServerLogs:
			ShowServerLogs = true
			i++
		default:
			parsedArgs = append(parsedArgs, args[i])
			i++
		}
	}

	return parsedArgs
}

// FormatAndPrintResponse formats and prints an MCP response in the format specified by
// FormatOption.
func FormatAndPrintResponse(cmd *cobra.Command, resp any, err error) error {
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	output, err := jsonutils.Format(resp, FormatOption)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), output)
	return nil
}

// IsValidFormat returns true if the format is valid.
func IsValidFormat(format string) bool {
	return format == "json" || format == "j" ||
		format == "pretty" || format == "p" ||
		format == "table" || format == "t"
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

// ConvertJSONToSlice converts a JSON serialized object to a slice of any type.
func ConvertJSONToSlice(jsonData any) []any {
	if jsonData == nil {
		return nil
	}
	var toolsSlice []any
	data, _ := json.Marshal(jsonData)
	_ = json.Unmarshal(data, &toolsSlice)
	return toolsSlice
}

// ConvertJSONToMap converts a JSON serialized object to a map of strings to any type.
func ConvertJSONToMap(jsonData any) map[string]any {
	if jsonData == nil {
		return nil
	}
	var promptMap map[string]any
	data, _ := json.Marshal(jsonData)
	_ = json.Unmarshal(data, &promptMap)
	return promptMap
}
