import { useEffect, useState } from 'react'
import { api, HubServer } from '../lib/api'

export function Hub() {
  const [items, setItems] = useState<HubServer[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.listHubs().then(r => setItems(r.items as any)).finally(()=>setLoading(false))
  }
  useEffect(() => { load() }, [])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Your MCP Hub</h1>
        <button onClick={load} className="px-3 py-1.5 rounded border">Refresh</button>
      </div>
      {loading && <div className="animate-pulse text-slate-500">Loading hub servers...</div>}
      <div className="grid md:grid-cols-2 gap-4">
        {items.map(h => (
          <div key={h.id} className="border border-white/10 rounded-lg p-4 bg-white/5">
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium">{h.server_name || h.mcp_server_id}</div>
                <div className="text-xs text-slate-400">{h.server_url}</div>
              </div>
              <span className={`text-xs px-2 py-1 rounded border border-white/10 ${h.status==='ACTIVE'?'bg-emerald-500/20 text-emerald-300':'bg-slate-500/20 text-slate-300'}`}>{h.status}</span>
            </div>
            <div className="mt-3 flex gap-2">
              <button onClick={()=>api.refreshHub(h.id).then(load)} className="text-sm px-3 py-1.5 rounded bg-blue-600 text-white">Refresh Tools</button>
              <button onClick={()=>{ if(confirm('Delete this MCP server from hub?')) api.deleteHub(h.id).then(load) }} className="text-sm px-3 py-1.5 rounded border border-white/10">Delete</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

