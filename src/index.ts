#!/usr/bin/env node
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { CallToolRequestSchema, ListToolsRequestSchema } from "@modelcontextprotocol/sdk/types.js";

const server = new Server(
  { name: "acr-mcp-server", version: "0.8.2" },
  { capabilities: { tools: {} } }
);

server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [
    {
      name: "chat_send_message",
      description: "Send message into an ACR deliberation room with optional file attachment",
      inputSchema: {
        type: "object",
        properties: {
          roomId: { type: "string" },
          content: { type: "string" },
          attachmentId: { type: "string" }
        },
        required: ["roomId", "content"]
      }
    },
    {
      name: "chat_cast_vote",
      description: "Vote on a consensus proposal. If choice is DISSENT, rationale is strictly required.",
      inputSchema: {
        type: "object",
        properties: {
          proposalId: { type: "string" },
          choice: { type: "string", enum: ["APPROVE", "REJECT", "DISSENT"] },
          rationale: { type: "string" }
        },
        required: ["proposalId", "choice"]
      }
    },
    {
      name: "mcp_ast_diff_verify",
      description: "Verify AST invariant composition safety before requesting pull request merge",
      inputSchema: {
        type: "object",
        properties: {
          astJson: { type: "string" }
        },
        required: ["astJson"]
      }
    }
  ]
}));

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch(console.error);
