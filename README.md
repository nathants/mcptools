<p align="center">
  <img src="./.github/resources/logo.png" alt="MCP Tools" height="300">
</p>

# MCP Tools CLI

A command-line interface for interacting with MCP (Model Context Protocol) servers.

[![Blog Post](https://img.shields.io/badge/Blog-Read%20about%20MCP%20Tools-blue)](https://blog.fka.dev/blog/2025-03-26-introducing-mcp-tools-cli/)

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
  - [Using Homebrew](#using-homebrew)
  - [From Source](#from-source)
- [Getting Started](#getting-started)
- [Features](#features)
  - [Transport Options](#transport-options)
  - [Output Formats](#output-formats)
  - [Commands](#commands)
  - [Interactive Shell](#interactive-shell)
- [Server Modes](#server-modes)
  - [Mock Server Mode](#mock-server-mode)
  - [Proxy Mode](#proxy-mode)
- [Examples](#examples)
  - [Basic Usage](#basic-usage)
  - [Script Integration](#script-integration)
  - [Debugging](#debugging)
- [Contributing](#contributing)
- [Roadmap](#roadmap)
- [License](#license)

## Overview

MCP Tools provides a versatile CLI for working with Model Context Protocol (MCP) servers. It enables you to:

- Discover and call tools provided by MCP servers
- Access and utilize resources exposed by MCP servers
- Create mock servers for testing client applications
- Proxy MCP requests to shell scripts for easy extensibility
- Create interactive shells for exploring and using MCP servers
- Format output in various styles (JSON, pretty-printed, table)
- Support all transport methods (HTTP, stdio)

<p align="center">
  <img src=".github/resources/screenshot.png" alt="MCP Tools Screenshot" width="700">
</p>

## Installation

### Using Homebrew

```bash
brew tap f/mcptools
brew install mcp
```

> ❕ The binary is installed as `mcp` but can also be accessed as `mcpt` to avoid conflicts with other tools that might use the `mcp` command name.

### From Source

```bash
go install github.com/f/mcptools/cmd/mcptools@latest
```

The binary will be installed as `mcptools` but can be aliased to `mcpt` for convenience.

## Getting Started

The simplest way to start using MCP Tools is to connect to an MCP server and list available tools:

```bash
# List all available tools from a filesystem server
mcp tools npx -y @modelcontextprotocol/server-filesystem ~

# Call a specific tool
mcp call read_file --params '{"path": "README.md"}' npx -y @modelcontextprotocol/server-filesystem ~

# Open an interactive shell
mcp shell npx -y @modelcontextprotocol/server-filesystem ~
```

## Features

MCP Tools supports a wide range of features for interacting with MCP servers:

```
Usage:
  mcp [command]
  
Available Commands:
  call           Call a tool, resource, or prompt on the MCP server
  help           Help about any command
  mock           Create a mock MCP server with tools, prompts, and resources
  proxy          Proxy MCP tool requests to shell scripts
  prompts        List available prompts on the MCP server
  resources      List available resources on the MCP server
  shell          Start an interactive shell for MCP commands
  tools          List available tools on the MCP server
  version        Print the version information
  
Flags:
  -f, --format string   Output format (table, json, pretty) (default "table")
  -h, --help            Help for mcp
  -p, --params string   JSON string of parameters to pass to the tool (default "{}")
```

### Transport Options

MCP Tools supports multiple transport methods for communicating with MCP servers:

#### Stdio Transport

Uses stdin/stdout to communicate with an MCP server via JSON-RPC 2.0. This is useful for command-line tools that implement the MCP protocol.

```bash
mcp tools npx -y @modelcontextprotocol/server-filesystem ~
```

#### HTTP SSE Transport

Uses HTTP and Server-Sent Events (SSE) to communicate with an MCP server via JSON-RPC 2.0. This is useful for connecting to remote servers that implement the MCP protocol.

```bash
mcp tools http://127.0.0.1:3001

# Example: Use the everything sample server
# docker run -p 3001:3001 --rm -it tzolov/mcp-everything-server:v1
```

_Note:_ HTTP SSE currently supports only MCP protocol version 2024-11-05.

### Output Formats

MCP Tools supports three output formats to accommodate different needs:

#### Table Format (Default)

```bash
mcp tools npx -y @modelcontextprotocol/server-filesystem ~
```

Output:
```
NAME        DESCRIPTION
----        -----------
read_file   Reads a file from the filesystem
list_dir    Lists directory contents
...
```

#### JSON Format (Compact)

```bash
mcp tools --format json npx -y @modelcontextprotocol/server-filesystem ~
```

#### Pretty JSON Format (Indented)

```bash
mcp tools --format pretty npx -y @modelcontextprotocol/server-filesystem ~
```

### Commands

MCP Tools includes several core commands for interacting with MCP servers:

#### List Available Tools

```bash
mcp tools npx -y @modelcontextprotocol/server-filesystem ~
```

#### List Available Resources

```bash
mcp resources npx -y @modelcontextprotocol/server-filesystem ~
```

#### List Available Prompts

```bash
mcp prompts npx -y @modelcontextprotocol/server-filesystem ~
```

#### Call a Tool

```bash
mcp call read_file --params '{"path":"/path/to/file"}' npx -y @modelcontextprotocol/server-filesystem ~
```

#### Call a Resource

```bash
mcp call resource:my-resource npx -y @modelcontextprotocol/server-filesystem ~
```

#### Call a Prompt

```bash
mcp call prompt:my-prompt --params '{"name":"John"}' npx -y @modelcontextprotocol/server-filesystem ~
```

### Interactive Shell

The interactive shell mode allows you to run multiple MCP commands in a single session:

```bash
mcp shell npx -y @modelcontextprotocol/server-filesystem ~
```

This opens an interactive shell with the following capabilities:

```
mcp tools shell
connected to: npx -y @modelcontextprotocol/server-filesystem /Users/fka

mcp > Type '/h' for help or '/q' to quit
mcp > tools
NAME        DESCRIPTION
----        -----------
read_file   Reads a file from the filesystem
...

# Direct tool calling is supported
mcp > read_file {"path": "README.md"}
...content of README.md...

# Get help
mcp > /h
MCP Shell Commands:
  tools                      List available tools
  resources                  List available resources
  prompts                    List available prompts
  call <entity> [--params '{...}']  Call a tool, resource, or prompt
  format [json|pretty|table] Get or set output format
Special Commands:
  /h, /help                  Show this help
  /q, /quit, exit            Exit the shell
```

## Server Modes

MCP Tools can operate as both a client and a server, with two server modes available:

### Mock Server Mode

The mock server mode creates a simulated MCP server for testing clients without implementing a full server:

```bash
# Create a mock server with a simple tool
mcp mock tool hello_world "A simple greeting tool"

# Create a mock server with multiple entity types
mcp mock tool hello_world "A greeting tool" \
       prompt welcome "A welcome prompt" "Hello {{name}}, welcome to {{location}}!" \
       resource docs://readme "Documentation" "Mock MCP Server\nThis is a mock server"
```

Features of the mock server:

- Full initialization handshake
- Tool listing with standardized schema
- Tool calling with simple responses
- Resource listing and reading
- Prompt listing and retrieval with argument substitution
- Detailed request/response logging to `~/.mcpt/logs/mock.log`

#### Using Prompt Templates

For prompts, any text in `{{double_braces}}` is automatically detected as an argument:

```bash
# Create a prompt with name and location arguments
mcp mock prompt greeting "Greeting template" "Hello {{name}}! Welcome to {{location}}."
```

When a client requests the prompt, it can provide values for these arguments which will be substituted in the response.

### Proxy Mode

The proxy mode allows you to register shell scripts or inline commands as MCP tools, making it easy to extend MCP functionality without writing code:

```bash
# Register a shell script as an MCP tool
mcp proxy tool add_operation "Adds a and b" "a:int,b:int" ./examples/add.sh

# Register an inline command as an MCP tool
mcp proxy tool add_operation "Adds a and b" "a:int,b:int" -e 'echo "total is $a + $b = $(($a+$b))"'

# Unregister a tool
mcp proxy tool --unregister add_operation

# Start the proxy server
mcp proxy start
```

#### How It Works

1. Register a shell script or inline command with a tool name, description, and parameter specification
2. Start the proxy server, which implements the MCP protocol
3. When a tool is called, parameters are passed as environment variables to the script/command
4. The script/command's output is returned as the tool response

#### Parameter Types

Parameters are specified in the format `name:type,name:type,...` where `type` can be:
- `string`: Text values
- `int`: Integer numbers
- `float`: Floating-point numbers
- `bool`: Boolean values (true/false)

#### Example Scripts and Commands

**Adding Numbers (add.sh):**

```bash
#!/bin/bash
# Get values from environment variables
if [ -z "$a" ] || [ -z "$b" ]; then
  echo "Error: Missing required parameters 'a' or 'b'"
  exit 1
fi

# Perform the addition
result=$(($a + $b))
echo "The sum of $a and $b is $result"
```

**Inline Command Example:**

```bash
# Simple addition
mcp proxy tool add_op "Adds given numbers" "a:int,b:int" -e 'echo "total is $a + $b = $(($a+$b))"'

# Customized greeting
mcp proxy tool greet "Greets a user" "name:string,greeting:string,formal:bool" -e '
if [ "$formal" = "true" ]; then
  title="Mr./Ms."
  echo "${greeting:-Hello}, ${title} ${name}. How may I assist you today?"
else
  echo "${greeting:-Hello}, ${name}! Nice to meet you!"
fi
'

# File operations
mcp proxy tool count_lines "Counts lines in a file" "file:string" -e "wc -l < \"$file\""
```

#### Configuration and Logging

- Tools are registered in `~/.mcpt/proxy_config.json`
- The proxy server logs all requests and responses to `~/.mcpt/logs/proxy.log`
- Use `--unregister` to remove a tool from the configuration

## Examples

### Basic Usage

List tools from a filesystem server:

```bash
mcp tools npx -y @modelcontextprotocol/server-filesystem ~
```

Call a tool with pretty JSON output:

```bash
mcp call read_file --params '{"path": "README.md"}' --format pretty npx -y @modelcontextprotocol/server-filesystem ~
```

### Script Integration

Using the proxy mode with a simple shell script:

```bash
# 1. Create a simple shell script for addition
cat > add.sh << 'EOF'
#!/bin/bash
# Get values from environment variables
if [ -z "$a" ] || [ -z "$b" ]; then
  echo "Error: Missing required parameters 'a' or 'b'"
  exit 1
fi
result=$(($a + $b))
echo "The sum of $a and $b is $result"
EOF

# 2. Make it executable
chmod +x add.sh

# 3. Register it as an MCP tool
mcp proxy tool add_numbers "Adds two numbers" "a:int,b:int" ./add.sh

# 4. In one terminal, start the proxy server
mcp proxy start

# 5. In another terminal, you can call it as an MCP tool
mcp call add_numbers --params '{"a":5,"b":3}' --format pretty
```

### Debugging

Tailing the logs to debug your proxy or mock server:

```bash
# For the mock server logs
tail -f ~/.mcpt/logs/mock.log

# For the proxy server logs
tail -f ~/.mcpt/logs/proxy.log

# To watch all logs in real-time (on macOS/Linux)
find ~/.mcpt/logs -name "*.log" -exec tail -f {} \;
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on how to submit pull requests, report issues, and contribute to the project.

## Roadmap

The following features are planned for future releases:

- Authentication: Support for secure authentication mechanisms

## License

This project is licensed under the MIT License.
