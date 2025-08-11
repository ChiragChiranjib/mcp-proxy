import { useEffect, useState } from 'react'
import { api, HubServer, Tool } from '../lib/api'
import { notifyError, notifySuccess } from '../components/ToastHost'
import { JSONViewer } from '../components/JSONViewer'

export function Hub() {
  const [items, setItems] = useState<HubServer[]>([])
  const [reloading, setReloading] = useState(false)
  const [refreshingHubId, setRefreshingHubId] = useState<string | null>(null)
  const [open, setOpen] = useState<Record<string, boolean>>({})
  const [toolsByHub, setToolsByHub] = useState<Record<string, Tool[]>>({})
  const [loadingHubTools, setLoadingHubTools] = useState<Record<string, boolean>>({})
  const [hubQuery, setHubQuery] = useState<Record<string, string>>({})
  const [role, setRole] = useState<string | undefined>(undefined)

  const load = () => {
    setReloading(true)
    api.listHubs()
      .then(r => setItems(r.items as any))
      .catch((e:any) => notifyError(e?.message || 'Failed to load hubs'))
      .finally(() => setReloading(false))
  }
  useEffect(() => { load(); api.me().then(m=>setRole((m as any).role)).catch(()=>{}) }, [])

  const toggleHub = async (hubId: string) => {
    setOpen(s => ({ ...s, [hubId]: !s[hubId] }))
    const willOpen = !open[hubId]
    if (willOpen && !toolsByHub[hubId]) {
      await loadToolsForHub(hubId)
    }
  }

  const loadToolsForHub = async (hubId: string) => {
    setLoadingHubTools(s => ({ ...s, [hubId]: true }))
    try {
      const qp = new URLSearchParams()
      qp.set('hub_server_id', hubId)
      const r = await api.listTools(qp)
      setToolsByHub(s => ({ ...s, [hubId]: r.items }))
    } catch (e: any) {
      notifyError(e?.message || 'Failed to load tools')
    } finally {
      setLoadingHubTools(s => ({ ...s, [hubId]: false }))
    }
  }

  const setToolStatus = async (t: Tool, hubId: string) => {
    const ns = t.status === 'ACTIVE' ? 'DEACTIVATED' : 'ACTIVE'
    try {
      await api.setToolStatus(t.id, ns)
      await loadToolsForHub(hubId)
    } catch (e: any) {
      notifyError(e?.message || 'Failed to update tool')
    }
  }

  // Delete option removed as per request; keeping handler out for now

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Your MCP Hub</h1>
        <div className="flex items-center gap-2">
          <button
            onClick={load}
            className="relative px-3 py-1.5 rounded-lg border border-white/10
            bg-transparent text-slate-300 active:scale-95 transition flex
            items-center gap-2 hover:bg-white/5 hover:border-white/20
            hover:text-slate-100 focus:outline-none focus:ring-1
            focus:ring-white/10 group"
          >
            <span
              aria-hidden
              className={`${reloading ? 'animate-spin'
                : 'group-hover:rotate-12'} inline-flex items-center
                justify-center w-5 h-5 text-lg transition-all`}
            >
              ⟳
            </span>
            <span>Refresh</span>
            <span className="relative inline-flex items-center ml-2 group/inf">
              <span
                className="inline-flex items-center justify-center w-4 h-4
                rounded-full border border-white/20 text-[10px] leading-none
                not-italic text-slate-200"
              >
                i
              </span>
              <span
                role="tooltip"
                className="absolute left-1/2 -translate-x-1/2 top-full mt-2
                whitespace-nowrap rounded-md border border-white/10
                bg-black/85 px-2.5 py-1 text-xs text-slate-200 shadow-xl
                opacity-0 pointer-events-none translate-y-1 transition
                group-hover/inf:opacity-100 group-hover/inf:translate-y-0"
              >
                Refreshes the hub list only. To refresh tools for a hub, use
                “Refresh Tools” on that hub card.
              </span>
            </span>
          </button>
        </div>
      </div>
      <div className="grid grid-cols-1 gap-4">
        {items.map(h => (
          <div key={h.id} className="relative group rounded-2xl border border-white/10 bg-white/[0.04] p-4 transition hover:border-white/20 hover:shadow-[0_8px_30px_rgba(0,0,0,0.35)]">
            <div className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition duration-300" style={{
              background: 'radial-gradient(1000px 300px at 10% -20%, rgba(59,130,246,0.12), transparent 60%), radial-gradient(1000px 300px at 110% 120%, rgba(16,185,129,0.12), transparent 60%)'
            }} />
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">{h.name || h.server_name || h.mcp_server_id}</div>
                { (h.url || h.server_url) && (
                  <a href={h.url || h.server_url} target="_blank" className="text-xs text-blue-500 break-all">{h.url || h.server_url}</a>
                )}
              </div>
              <div className="flex items-center gap-2">
                <button
                  disabled={role !== 'ADMIN'}
                  title={role==='ADMIN' ? 'Edit server (admin)' : 'Admin only'}
                  className={`text-xs px-2 py-1 rounded border border-white/10 ${role!=='ADMIN' ? 'text-slate-500 cursor-not-allowed' : 'hover:bg-white/10 hover:border-white/20'}`}
                >Edit</button>
                <span className="inline-flex items-center gap-1.5 text-xs px-2 py-1 rounded border border-white/10">
                  <span className={`inline-block w-2 h-2 rounded-full ${h.status==='ACTIVE'?'bg-emerald-400 shadow-[0_0_8px_rgba(16,185,129,0.8)]':'bg-slate-500'}`} />
                  <span className={`${h.status==='ACTIVE'?'text-emerald-300':'text-slate-300'}`}>{h.status}</span>
                </span>
                <button
                  aria-label="Toggle tools"
                  aria-expanded={!!open[(h as any).id || (h as any).mcp_hub_server_id || (h as any).mcp_server_id]}
                  onClick={()=>toggleHub((h as any).id || (h as any).mcp_hub_server_id || (h as any).mcp_server_id)}
                  title="Show tools"
                  className="text-sm px-3 py-1.5 rounded-lg border border-white/15 bg-white/5 hover:bg-white/10 hover:border-white/25 transition shadow hover:shadow-blue-900/10 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
                >
                  <span className="mr-1">Tools</span>
                  <span className={`inline-block transition-transform ${open[(h as any).id || (h as any).mcp_hub_server_id || (h as any).mcp_server_id] ? 'rotate-180' : ''}`}>▾</span>
                </button>
              </div>
            </div>
            {h.description && <p className="text-xs text-slate-400 mt-2">{h.description}</p>}
            {h.capabilities && (
              <details className="mt-2">
                <summary className="text-xs text-slate-400 cursor-pointer">Capabilities</summary>
                <pre className="text-xs bg-black/30 border border-white/10 rounded p-3 overflow-auto max-h-56">{JSON.stringify(h.capabilities, null, 2)}</pre>
              </details>
            )}
            <div className="mt-3 flex gap-2">
              <button onClick={async ()=>{
                const hubId = (h as any).id || (h as any).mcp_hub_server_id || (h as any).mcp_server_id
                if (!hubId) { notifyError('Hub id missing'); return }
                setRefreshingHubId(hubId)
                try {
                  await api.refreshHub(hubId)
                  notifySuccess('Refreshed')
                } catch (e:any) {
                  notifyError(e?.message || 'Refresh failed')
                } finally {
                  setRefreshingHubId(null)
                  load()
                  if (open[hubId]) { await loadToolsForHub(hubId) }
                }
              }} className="text-sm px-3 py-1.5 rounded bg-blue-600 text-white active:scale-95 transition flex items-center gap-2 shadow hover:shadow-blue-900/30 group focus:outline-none focus:ring-2 focus:ring-blue-500/30">
                <span aria-hidden className={`${refreshingHubId === ((h as any).id || (h as any).mcp_hub_server_id) ? 'animate-spin' : 'group-hover:rotate-12 group-hover:text-blue-200'} inline-block transition-all`}>⟳</span>
                <span>Refresh Tools</span>
              </button>
              <button onClick={()=>{ const hubId = (h as any).id || (h as any).mcp_hub_server_id || (h as any).mcp_server_id; if (!hubId) { notifyError('Hub id missing'); return } if(confirm('Delete this MCP server from hub?')) api.deleteHub(hubId).then(()=>{ notifySuccess('Deleted'); load() }).catch(e=>notifyError(e.message || 'Delete failed')) }} className="text-sm px-3 py-1.5 rounded border border-white/10 hover:border-white/20 transition">Delete</button>
            </div>

            {/* Expanded Tools Section */}
            {(() => {
              const hubId = (h as any).id || (h as any).mcp_hub_server_id || (h as any).mcp_server_id
              const isOpen = !!open[hubId]
              const q = hubQuery[hubId] || ''
              const tools = toolsByHub[hubId] || []
              const filtered = tools.filter(t => (
                (t.modified_name||'').toLowerCase().includes(q.toLowerCase()) ||
                (t.original_name||'').toLowerCase().includes(q.toLowerCase())
              ))
              return isOpen ? (
                <div className="mt-4 border-t border-white/10 pt-4">
                  <div className="flex items-center justify-between gap-3">
                    <div className="font-medium text-sm">Tools</div>
                    <input
                      value={q}
                      onChange={e=>setHubQuery(s=>({ ...s, [hubId]: e.target.value }))}
                      placeholder="Search tools..."
                      className="px-3 py-1.5 rounded border border-white/10 bg-white/5 text-sm placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
                    />
                  </div>
                  {loadingHubTools[hubId] && (
                    <div className="text-xs text-slate-500 mt-3">Loading tools...</div>
                  )}
                  {!loadingHubTools[hubId] && filtered.length === 0 && (
                    <div className="text-xs text-slate-500 mt-3">No tools found.</div>
                  )}
                  <div className="mt-3 max-h-[420px] overflow-y-auto pr-1 scroll-panel">
                    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
                      {filtered.map(t => (
                        <div key={t.id} className="relative group rounded-2xl border border-white/5 bg-white/[0.04] p-4 transition hover:border-white/15 hover:shadow-[0_8px_24px_rgba(0,0,0,0.25)]">
                        <div className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition duration-300" style={{
                          background: 'radial-gradient(1000px 300px at 10% -20%, rgba(59,130,246,0.12), transparent 60%), radial-gradient(1000px 300px at 110% 120%, rgba(16,185,129,0.12), transparent 60%)'
                        }} />
                        <div className="font-medium break-words">{t.modified_name}</div>
                        <div className="text-xs text-slate-500 break-words">Original: {t.original_name}</div>
                        <details className="mt-2">
                          <summary className="text-xs text-slate-400 cursor-pointer">View input schema</summary>
                          <JSONViewer value={t.input_schema || {}} />
                        </details>
                        <div className="mt-2 flex items-center justify-between">
                          <span className="inline-flex items-center gap-1.5 text-xs px-2 py-1 rounded border border-white/10">
                            <span className={`inline-block w-2 h-2 rounded-full ${t.status==='ACTIVE'?'bg-emerald-400 shadow-[0_0_8px_rgba(16,185,129,0.8)]':'bg-slate-500'}`} />
                            <span className={`${t.status==='ACTIVE'?'text-emerald-300':'text-slate-300'}`}>{t.status}</span>
                          </span>
                          <div className="flex gap-2">
                            <button onClick={()=>setToolStatus(t, hubId)} className="text-xs px-2.5 py-1.5 rounded border border-white/10 cursor-pointer hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-blue-500/30 active:scale-95 transition">{t.status==='ACTIVE'?'Deactivate':'Activate'}</button>
                          </div>
                        </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              ) : null
            })()}
          </div>
        ))}
      </div>
    </div>
  )
}

