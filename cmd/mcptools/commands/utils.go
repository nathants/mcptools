package commands

import (
	"fmt"
	"strings"

	"github.com/f/mcptools/pkg/alias"
	"github.com/f/mcptools/pkg/client"
	"github.com/f/mcptools/pkg/jsonutils"
)

// sentinel errors.
var (
	ErrCommandRequired = fmt.Errorf("command to execute is required when using stdio transport")
)

// CreateClientFunc is the function used to create MCP clients.
// This can be replaced in tests to use a mock transport.
var CreateClientFunc = func(args []string) (*client.Client, error) {
	if len(args) == 0 {
		return nil, ErrCommandRequired
	}

	isHTTP := func(str string) bool {
		return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
	}

	// Check if the first argument is an alias
	if len(args) == 1 {
		server, found := alias.GetServerCommand(args[0])
		if found {
			if isHTTP(server) {
				return client.NewHTTP(server), nil
			}
			cmdParts := client.ParseCommandString(server)
			return client.NewStdio(cmdParts), nil
		}
	}

	if len(args) == 1 && isHTTP(args[0]) {
		return client.NewHTTP(args[0]), nil
	}

	return client.NewStdio(args), nil
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
		default:
			parsedArgs = append(parsedArgs, args[i])
			i++
		}
	}

	return parsedArgs
}

// FormatAndPrintResponse formats and prints an MCP response in the format specified by
// FormatOption.
func FormatAndPrintResponse(resp map[string]any, err error) error {
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	output, err := jsonutils.Format(resp, FormatOption)
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	fmt.Println(output)
	return nil
}
