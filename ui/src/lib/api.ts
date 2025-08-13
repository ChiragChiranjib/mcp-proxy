export type CatalogServer = { 
  id: string
  name: string
  url: string
  description: string
  access_type?: 'public' | 'private'
  transport?: string
  capabilities?: any
}

export type HubServer = {
  id?: string
  user_id: string
  mcp_server_id: string
  status: string
  auth_type?: string
  auth_value?: string
  // Server details from join
  name?: string
  url?: string
  description?: string
  capabilities?: any
  transport?: string
  access_type?: string
}

export type Tool = { 
  id: string
  user_id?: string  // Nullable for global tools
  mcp_server_id: string  // Server reference
  mcp_hub_server_id?: string  // Hub server reference (for private tools)
  original_name: string
  modified_name: string
  status: string
  description?: string
  input_schema?: any
  annotations?: any
}

export type VirtualServer = { 
  id: string
  user_id: string
  name?: string
  status: string
}

class ApiError extends Error {
  status: number
  requestId?: string
  constructor(message: string, status: number, requestId?: string) {
    super(message)
    this.status = status
    this.requestId = requestId
  }
}

async function http<T>(input: RequestInfo, init: RequestInit = {}): Promise<T> {
  const res = await fetch(input, {
    ...init,
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(init.headers || {}) },
  })
  const requestId = res.headers.get('x-request-id') || undefined
  if (!res.ok) {
    // try to extract server error body
    let msg = res.statusText || `HTTP ${res.status}`
    try {
      const text = await res.text()
      if (text) {
        try {
          const j = JSON.parse(text)
          msg = j.error || j.message || j.msg || text
        } catch {
          msg = text
        }
      }
    } catch {}
    if (res.status === 401) {
      try { window.dispatchEvent(new CustomEvent('session:expired')) } catch {}
    }
    throw new ApiError(msg, res.status, requestId)
  }
  if (res.status === 204) {
    return undefined as unknown as T
  }
  return res.json()
}

export const api = {
  // Auth endpoints
  loginWithGoogle: (credential: string) => http<{user_id: string; email: string; name?: string}>('/api/auth/google', { method: 'POST', body: JSON.stringify({ credential }) }),
  logout: () => http<void>('/api/auth/logout', { method: 'POST' }),
  loginWithBasic: (username: string, password: string) => http<{user_id: string; email: string}>('/api/auth/basic', { method: 'POST', body: JSON.stringify({ username, password }) }),
  me: () => http<{user_id: string; email?: string; name?: string; role?: string}>('/api/auth/me'),
  
  // Catalog endpoints
  listCatalog: () => http<{items: CatalogServer[]}>('/api/catalog/servers'),
  addCatalog: (body: { name: string; url: string; description?: string; access_type?: string; transport?: string }) => 
    http<{id: string}>('/api/catalog/servers', { method: 'POST', body: JSON.stringify(body) }),
  updateCatalog: (id: string, body: { url?: string; description?: string }) =>
    http<{ok: boolean}>(`/api/catalog/servers/${id}`, { method: 'PATCH', body: JSON.stringify(body) }),
  refreshCatalog: (id: string) => http<{ok: boolean; added: Tool[]; deleted: Tool[]; total_added: number; total_deleted: number}>(`/api/catalog/servers/${id}/refresh`, { method: 'POST' }),
  getCatalogTools: (id: string) => http<{items: Tool[]}>(`/api/catalog/servers/${id}/tools`),
  
  // Hub endpoints  
  listHubs: () => http<{items: HubServer[]}>('/api/hub/servers'),
  addHub: (body: any) => http<{id: string}>('/api/hub/servers', { method: 'POST', body: JSON.stringify(body) }),
  deleteHub: (id: string) => http<{ok: string}>(`/api/hub/servers/${id}`, { method: 'DELETE' }),
  refreshHub: (id: string) => http<{ok: boolean; added: Tool[]; deleted: Tool[]; total_added: number; total_deleted: number}>(`/api/hub/servers/${id}/refresh`, { method: 'POST' }),
  
  // Tools endpoints (UPDATED: server_id instead of hub_server_id)
  listTools: (q: URLSearchParams) => http<{items: Tool[]}>(`/api/tools?${q.toString()}`),
  setToolStatus: (id: string, status: string) => http<{ok: string}>(`/api/tools/${id}/status`, { method: 'PATCH', body: JSON.stringify({status}) }),
  deleteTool: (id: string) => http<{ok: string}>(`/api/tools/${id}`, { method: 'DELETE' }),
  
  // Virtual Server endpoints
  createVS: (name?: string, tool_ids?: string[]) => http<{id: string}>(`/api/virtual-servers`, { method: 'POST', body: JSON.stringify({ name, tool_ids }) }),
  listVS: () => http<{items: VirtualServer[]}>(`/api/virtual-servers`),
  updateVS: (id: string, name: string) => http<{ok: string}>(`/api/virtual-servers/${id}`, { method: 'PATCH', body: JSON.stringify({name}) }),
  replaceVSTools: (id: string, tool_ids: string[]) => http<{ok: string}>(`/api/virtual-servers/${id}/tools`, { method: 'PUT', body: JSON.stringify({tool_ids}) }),
  removeVSTool: (id: string, tool_id: string) => http<{ok: string}>(`/api/virtual-servers/${id}/tools/${tool_id}`, { method: 'DELETE' }),
  setVSStatus: (id: string, status: string) => http<{ok: string}>(`/api/virtual-servers/${id}/status`, { method: 'PATCH', body: JSON.stringify({status}) }),
  deleteVS: (id: string) => http<{ok: string}>(`/api/virtual-servers/${id}`, { method: 'DELETE' }),
  listVSTools: (id: string) => http<{items: Tool[]}>(`/api/virtual-servers/${id}/tools`),
}

