export type CatalogServer = { id: string; name: string; url: string; description: string }
export type HubServer = {
  id?: string
  user_id: string
  mcp_server_id: string
  status: string
  transport: string
  capabilities?: any
  auth_type?: string
  auth_value?: string
  server_url?: string
  server_name?: string
  name?: string
  url?: string
  description?: string
}
export type Tool = { id: string; user_id: string; original_name: string; modified_name: string; mcp_hub_server_id?: string; hub_server_id?: string; status: string; description?: string; input_schema?: any }
export type VirtualServer = { id: string; user_id: string; status: string }

const USER_ID = localStorage.getItem('x-user-id') || ''

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
    headers: { 'Content-Type': 'application/json', ...(USER_ID ? {'X-User-ID': USER_ID} : {}), ...(init.headers || {}) },
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
  loginWithGoogle: (credential: string) => http<{user_id: string; email: string; name?: string}>('/api/auth/google', { method: 'POST', body: JSON.stringify({ credential }) }),
  logout: () => http<void>('/api/auth/logout', { method: 'POST' }),
  loginWithBasic: (username: string, password: string) => http<{user_id: string; email: string}>('/api/auth/basic', { method: 'POST', body: JSON.stringify({ username, password }) }),
  me: () => http<{user_id: string}>('/api/auth/me'),
  listCatalog: () => http<{items: CatalogServer[]}>('/api/catalog/servers'),
  listHubs: () => http<{items: HubServer[]}>('/api/hub/servers'),
  addHub: (body: any) => http<{id: string}>('/api/hub/servers', { method: 'POST', body: JSON.stringify(body) }),
  deleteHub: (id: string) => http<{ok: string}>(`/api/hub/servers/${id}`, { method: 'DELETE' }),
  refreshHub: (id: string) => http<{ok: string}>(`/api/hub/servers/${id}/refresh`, { method: 'POST' }),
  listTools: (q: URLSearchParams) => http<{items: Tool[]}>(`/api/tools?${q.toString()}`),
  setToolStatus: (id: string, status: string) => http<{ok: string}>(`/api/tools/${id}/status`, { method: 'PATCH', body: JSON.stringify({status}) }),
  deleteTool: (id: string) => http<{ok: string}>(`/api/tools/${id}`, { method: 'DELETE' }),
  createVS: () => http<{id: string}>(`/api/virtual-servers`, { method: 'POST' }),
  listVS: () => http<{items: VirtualServer[]}>(`/api/virtual-servers`),
  replaceVSTools: (id: string, tool_ids: string[]) => http<{ok: string}>(`/api/virtual-servers/${id}/tools`, { method: 'PUT', body: JSON.stringify({tool_ids}) }),
  setVSStatus: (id: string, status: string) => http<{ok: string}>(`/api/virtual-servers/${id}/status`, { method: 'PATCH', body: JSON.stringify({status}) }),
  deleteVS: (id: string) => http<{ok: string}>(`/api/virtual-servers/${id}`, { method: 'DELETE' }),
}

