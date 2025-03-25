# MCP

A command-line interface for interacting with MCP (Model Context Protocol) servers using both stdio and HTTP transport.

## Installation

### Using Homebrew

```bash
brew tap f/mcptools
brew install mcp
```

### From Source

```bash
go install github.com/f/mcptools/cmd/mcptools@latest
```

## Usage

```
MCP is a command line interface for interacting with MCP servers.
It allows you to discover and call tools, list resources, and interact with MCP-compatible services.

Usage:
  mcp [command]

Available Commands:
  call           Call a tool, resource, or prompt on the MCP server
  help           Help about any command
  prompts        List available prompts on the MCP server
  resources      List available resources on the MCP server
  tools          List available tools on the MCP server
  version        Print the version information

Flags:
  -f, --format string   Output format (table, json, pretty) (default "table")
  -h, --help            Help for mcp
  -H, --http            Use HTTP transport instead of stdio
  -p, --params string   JSON string of parameters to pass to the tool (default "{}")
  -s, --server string   MCP server URL (when using HTTP transport) (default "http://localhost:8080")
```

## Transport Options

MCP supports two transport methods for communicating with MCP servers:

### Stdio Transport (Default)

Uses stdin/stdout to communicate with an MCP server via JSON-RPC 2.0. This is useful for command-line tools that implement the MCP protocol.

```bash
mcp tools npx -y @modelcontextprotocol/server-filesystem ~/Code
```

### HTTP Transport

Uses HTTP protocol to communicate with an MCP server. Use the `--http` flag for HTTP transport.

```bash
mcp --http tools --server "http://mcp.example.com:8080"
```

## Output Formats

MCP supports three output formats:

### Table Format (Default)

Displays the output in a table-like view for better readability.

```bash
mcp tools npx -y @modelcontextprotocol/server-filesystem ~/Code
```

### JSON Format

Displays the output as compact JSON.

```bash
mcp tools --format json npx -y @modelcontextprotocol/server-filesystem ~/Code
```

### Pretty Format

Displays the output as indented JSON.

```bash
mcp tools --format pretty npx -y @modelcontextprotocol/server-filesystem ~/Code
```

## Commands

### List Available Tools

```bash
# Using stdio transport (default)
mcp tools npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcp --http tools
```

### List Available Resources

```bash
# Using stdio transport
mcp resources npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcp --http resources
```

### List Available Prompts

```bash
# Using stdio transport
mcp prompts npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcp --http prompts
```

### Call a Tool

```bash
# Using stdio transport
mcp call read_file --params '{"path": "/path/to/file"}' npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcp --http call read_file --params '{"path": "/path/to/file"}'
```

### Call a Resource

```bash
# Using stdio transport
mcp call resource:my-resource npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcp --http call resource:my-resource
```

### Call a Prompt

```bash
# Using stdio transport
mcp call prompt:my-prompt npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcp --http call prompt:my-prompt
```

## Examples

List tools from a filesystem server:

```bash
mcp tools npx -y @modelcontextprotocol/server-filesystem ~/Code
```

Call the read_file tool with pretty JSON output:

```bash
mcp call read_file --params '{"path": "README.md"}' --format pretty npx -y @modelcontextprotocol/server-filesystem ~/Code
```

Using HTTP transport with a remote server:

```bash
mcp --http --server "http://mcp.example.com:8080" tools
```

## License

MIT 