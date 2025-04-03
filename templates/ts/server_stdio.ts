import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";

// Initialize server
const server = new McpServer({
  name: "PROJECT_NAME",
  version: "1.0.0"
});

// === Start server with stdio transport ===
const transport = new StdioServerTransport();
await server.connect(transport);
