# MCPTools

A command-line interface for interacting with MCP (Model Context Protocol) servers using both stdio and HTTP transport.

## Installation

### Using Homebrew

```bash
brew tap f/mcptools
brew install mcptools
```

### From Source

```bash
go install github.com/f/mcptools/cmd/mcptools@latest
```

## Usage

```
MCPTools is a command line interface for interacting with MCP servers.
It allows you to discover and call tools, list resources, and interact with MCP-compatible services.

Usage:
  mcptools [command]

Available Commands:
  call           Call a tool, resource, or prompt on the MCP server
  help           Help about any command
  prompts        List available prompts on the MCP server
  resources      List available resources on the MCP server
  tools          List available tools on the MCP server
  version        Print the version information

Flags:
  -f, --format string   Output format (json, pretty) (default "pretty")
  -h, --help            Help for mcptools
  -H, --http            Use HTTP transport instead of stdio
  -p, --params string   JSON string of parameters to pass to the tool (default "{}")
  -s, --server string   MCP server URL (when using HTTP transport) (default "http://localhost:8080")
```

## Transport Options

MCPTools supports two transport methods for communicating with MCP servers:

### Stdio Transport (Default)

Uses stdin/stdout to communicate with an MCP server via JSON-RPC 2.0. This is useful for command-line tools that implement the MCP protocol.

```bash
mcptools tools npx -y @modelcontextprotocol/server-filesystem ~/Code
```

### HTTP Transport

Uses HTTP protocol to communicate with an MCP server. Use the `--http` flag for HTTP transport.

```bash
mcptools --http tools --server "http://mcp.example.com:8080"
```

## Commands

### List Available Tools

```bash
# Using stdio transport (default)
mcptools tools npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcptools --http tools
```

### List Available Resources

```bash
# Using stdio transport
mcptools resources npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcptools --http resources
```

### List Available Prompts

```bash
# Using stdio transport
mcptools prompts npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcptools --http prompts
```

### Call a Tool

```bash
# Using stdio transport
mcptools call read_file --params '{"path": "/path/to/file"}' npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcptools --http call read_file --params '{"path": "/path/to/file"}'
```

### Call a Resource

```bash
# Using stdio transport
mcptools call resource:my-resource npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcptools --http call resource:my-resource
```

### Call a Prompt

```bash
# Using stdio transport
mcptools call prompt:my-prompt npx -y @modelcontextprotocol/server-filesystem ~/Code

# Using HTTP transport
mcptools --http call prompt:my-prompt
```

## Examples

List tools from a filesystem server:

```bash
mcptools tools npx -y @modelcontextprotocol/server-filesystem ~/Code
```

Call the read_file tool with JSON output:

```bash
mcptools call read_file --params '{"path": "README.md"}' --format json npx -y @modelcontextprotocol/server-filesystem ~/Code
```

Using HTTP transport with a remote server:

```bash
mcptools --http --server "http://mcp.example.com:8080" tools
```

## License

MIT 