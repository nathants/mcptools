import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { SSEServerTransport } from "@modelcontextprotocol/sdk/server/sse.js";
import express from "express";
import { z } from "zod";

// Initialize server
const server = new McpServer({
  name: "PROJECT_NAME",
  version: "1.0.0"
});

// Setup Express app
const app = express();

// Create an SSE endpoint that clients can connect to
app.get("/sse", async (req, res) => {
  const transport = new SSEServerTransport("/messages", res);
  await server.connect(transport);
});

// Create an endpoint to receive messages from clients 
app.post("/messages", express.json(), async (req, res) => {
  // Handle the message and send response
  res.json({ success: true });
});

// Start HTTP server
const port = 3000;
app.listen(port, () => {
  console.log(`MCP server running on http://localhost:${port}/sse`);
}); 