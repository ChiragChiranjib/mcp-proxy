# Global MCP Gateway

Run locally:

```bash
make run
```

Config file: `config/dev_streamable-http.toml`.

MCP endpoint will be served at `/servers/{virtual_server_id}/mcp`.

## Admin API (WIP)

- GET `/api/virtual-servers` → list VS for current user (`X-User-ID`)
- POST `/api/virtual-servers` → create VS, returns `{id}`
- PUT `/api/virtual-servers/{id}/tools` → replace tool IDs (cap 50)
- PATCH `/api/virtual-servers/{id}/status` → set status
- DELETE `/api/virtual-servers/{id}` → delete VS

- GET `/api/virtual-servers/{id}/tools` → list tools for VS
- PATCH `/api/tools/{id}/status` → set tool status

- POST `/api/hub/servers` → create hub server (auth stored encrypted when key configured)
- POST `/api/hub/servers/{id}/refresh` → pull tools from upstream; returns `{updated}`
- POST `/api/hub/servers/{id}/ping` → check connectivity; returns `{ok}`
- PATCH `/api/hub/servers/{id}` → set hub status
- DELETE `/api/hub/servers/{id}`

Auth: send `X-User-ID` header for now.
