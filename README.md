# MCP Proxy

A concise proxy and UI for composing “Virtual MCP Servers” from tools
discovered across your MCP hubs. The backend is Go (GORM + goose, slog), the
frontend is React + Vite + Tailwind.

## What it does

- Connect to upstream MCP servers ("Hubs"), fetch capabilities and tools
- Persist tools and create user-owned Virtual MCP Servers (VS)
- Expose a stateless streamable MCP endpoint per VS:
  `/servers/{virtual_server_id}/mcp`
- Simple Admin UI with four panels: Catalogue, Hub, Tools-in-Hub, Virtual
  Servers

## Quick start

Prereqs: Go 1.22+, Node 20+ (for UI), a local MySQL-compatible DB configured in
`config/dev_streamable-http.toml`.

Backend (dev):

```bash
make setup        # migrate + seed + run
# or
make run          # only run the server
```

Frontend (dev):

```bash
cd ui
npm install
npm run dev
```

Auth: Sign in with Google SSO or Basic credentials (dev). The UI uses the
session cookie set by the backend.

## Key endpoints (admin)

- `GET /api/virtual-servers` — list VS for current user
- `POST /api/virtual-servers` — create VS → `{ id }`
- `PUT /api/virtual-servers/{id}/tools` — replace tool IDs (cap 50)
- `PATCH /api/virtual-servers/{id}/status` — set status
- `DELETE /api/virtual-servers/{id}` — delete VS
- `GET /api/virtual-servers/{id}/tools` — list tools for VS
- `POST /api/hub/servers` — add hub (stores auth encrypted when configured)
- `POST /api/hub/servers/{id}/refresh` — pull tools from upstream

## Cursor config snippet: Add a Virtual MCP Server

Create a virtual MCP Server on the UI and add the below config in our cursor IDE to start using the MCP tools

```json
{
  "mcpServers": {
    "virtual-mcp-server": {
        "transport": "http",
        "url": "http://localhost:8080/servers/<virtual-server-id>/mcp"
    }
  }
}
```
