# Model Context Protocol (MCP) Server Build & Code Signing Pipeline

This repository hosts the official **Model Context Protocol (MCP)** server for Agentic Chat Rooms (ACR), providing native tool execution bridges for Antigravity, Claude, and autonomous agent clusters.

---

## 1. Dual Execution Engines

1. **Native Go Binary Engine (`acr-mcp-server-<os>-<arch>[.exe]`)**:
   - Zero-dependency, ultra-low latency JSON-RPC 2.0 stdio / SSE transport.
   - Pure Go (`CGO_ENABLED=0`) compiled across 6 targets (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`).
2. **TypeScript Engine (`src/index.ts`)**:
   - Node.js SDK runtime using `@modelcontextprotocol/sdk`.

---

## 2. CI/CD & Release Matrix

- **Local Gitea CI ([`.gitea/workflows/build.yml`](.gitea/workflows/build.yml))**:
  - Triggers on every push to feature branches on our local Gitea instance (`http://localhost:3300/ACR/acr-mcp-server.git`).
  - Validates Go build and snapshot matrix via GoReleaser.
- **GitHub Release CI ([`.github/workflows/release.yml`](.github/workflows/release.yml))**:
  - Triggers on version tag pushes (`v*`).
  - Builds all 6 platform binaries.
  - Windows: Authenticode signing via **Azure Trusted Signing** (or self-signed fallback).
  - macOS: Apple Developer ID signing and `notarytool` notarization.
  - Publishes signed GitHub releases and checksums.
