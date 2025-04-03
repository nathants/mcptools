import { z } from "zod";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";

// Define a prompt template
export default (server: McpServer) => {
  server.prompt(
    "PROMPT_NAME",
    {
      // Define the parameters for your prompt using Zod
      name: z.string({
        description: "The name to use in the greeting"
      }),
      time_of_day: z.enum(["morning", "afternoon", "evening", "night"], {
        description: "The time of day for the greeting"
      })
    },
    (params) => ({
      messages: [{
        role: "user",
        content: {
          type: "text",
          text: `Hello ${params.name}! Good ${params.time_of_day}. How are you today?`
        }
      }]
    })
  ); 
};