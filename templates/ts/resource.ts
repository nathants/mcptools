import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";

// Define a resource
export default (server: McpServer) => {
  server.resource(
    "RESOURCE_NAME",
    "RESOURCE_URI",
    async (uri) => ({
    contents: [{
      uri: uri.href,
        text: "This is a sample resource content. Replace with your actual content."
      }]
    })
  );
};
