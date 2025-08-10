import { useEffect, useState } from 'react'
import { api, HubServer } from '../lib/api'
import { notifyError, notifySuccess } from '../components/ToastHost'

export function Hub() {
  const [items, setItems] = useState<HubServer[]>([])
  const [reloading, setReloading] = useState(false)
  const [refreshingHubId, setRefreshingHubId] = useState<string | null>(null)

  const load = () => {
    setReloading(true)
    api.listHubs()
      .then(r => setItems(r.items as any))
      .catch((e:any) => notifyError(e?.message || 'Failed to load hubs'))
      .finally(() => setReloading(false))
  }
  useEffect(() => { load() }, [])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Your MCP Hub</h1>
        <button onClick={load} className="px-3 py-1.5 rounded border active:scale-95 transition flex items-center gap-2 hover:bg-white/5 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-blue-500/30 group">
          <span aria-hidden className={`${reloading ? 'animate-spin' : 'group-hover:rotate-12 group-hover:text-blue-300'} inline-block transition-all`}>⟳</span>
          <span>Refresh</span>
        </button>
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
              <span className={`text-xs px-2 py-1 rounded border border-white/10 ${h.status==='ACTIVE'?'bg-emerald-500/20 text-emerald-300':'bg-slate-500/20 text-slate-300'}`}>{h.status}</span>
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
                }
              }} className="text-sm px-3 py-1.5 rounded bg-blue-600 text-white active:scale-95 transition flex items-center gap-2 shadow hover:shadow-blue-900/30 group focus:outline-none focus:ring-2 focus:ring-blue-500/30">
                <span aria-hidden className={`${refreshingHubId === ((h as any).id || (h as any).mcp_hub_server_id) ? 'animate-spin' : 'group-hover:rotate-12 group-hover:text-blue-200'} inline-block transition-all`}>⟳</span>
                <span>Refresh Tools</span>
              </button>
              <button onClick={()=>{ const hubId = (h as any).id || (h as any).mcp_hub_server_id || (h as any).mcp_server_id; if (!hubId) { notifyError('Hub id missing'); return } if(confirm('Delete this MCP server from hub?')) api.deleteHub(hubId).then(()=>{ notifySuccess('Deleted'); load() }).catch(e=>notifyError(e.message || 'Delete failed')) }} className="text-sm px-3 py-1.5 rounded border border-white/10 hover:border-white/20 transition">Delete</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

