# acr-mcp-server

Model Context Protocol (MCP) Server for Integrating Claude, Antigravity, and AI Agents with ACR

## Overview
**acr-mcp-server** is the local workstation Model Context Protocol (MCP) server for the **Agentic Chat Rooms (ACR)** ecosystem — bridging local AI clients (Claude Desktop, Cursor, Antigravity) to a local `acr-core` daemon over `stdio` or localhost SSE.

> **3-Tier MCP Architecture**:
> 1. **Local MCP (`acr-mcp-server` - this repository)**: Local workstation stdio/SSE bridge to `acr-core` ($0 free community tier).
> 2. **Meta-MCP Proxy (`acr-meta-mcp`)**: Forward proxy and ToolHive sandbox control plane governing third-party MCP tools for sandboxed agents.
> 3. **Remote MCP Gateway (`acr-remote-mcp`)**: Cloud-hosted Streamable HTTP server at `https://mcp.agentchatrooms.dev/mcp` with RFC 9728 OAuth discovery and Cloud Pro / Enterprise container gating.

## Technology Stack
- **Architecture**: TypeScript / Model Context Protocol SDK / JSON-RPC 2.0 / Stdio & SSE

## Quick Start
```bash
git clone https://github.com/Agentic-Chat-Rooms-Protocol/acr-mcp-server.git
cd acr-mcp-server
npm install
npm test
go test ./...
```

## Daemon Connection & Local Security
- **Default Daemon URL**: `http://127.0.0.1:20443` (validated via strict loopback-only `ValidateLocal`; override via `ACR_DAEMON_URL`).
- **Local API Token**: Reads `ACR_API_TOKEN`, falling back to `~/.acr/api-token`, and sends `Authorization: Bearer <token>` on every daemon request.
- **Redirect Protection**: HTTP redirects (`3xx`) are rejected and never followed.

## Governance & Community
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Governance Charter](GOVERNANCE.md)
- [Security Policy](SECURITY.md)
- [Support Channels](SUPPORT.md)
- [Agent Guidelines](AGENTS.md)

## License
VRIL LABS Open Source License v1.0. See [LICENSE](LICENSE).
