import { useEffect, useMemo, useState } from 'react'
import { api, VirtualServer, Tool, HubServer, CatalogServer } from '../lib/api'

export function VirtualServers() {
  const [items, setItems] = useState<VirtualServer[]>([])
  const [tools, setTools] = useState<Tool[]>([])
  const [hubs, setHubs] = useState<HubServer[]>([])
  const [catalog, setCatalog] = useState<CatalogServer[]>([])
  const [selected, setSelected] = useState<string[]>([])
  const [q, setQ] = useState('')
  const [hubFilter, setHubFilter] = useState('')

  const load = () => { api.listVS().then(r=>setItems(r.items)) }
  useEffect(() => { load() }, [])

  const create = async () => {
    const r = await api.createVS()
    await load()
    // optionally open editor for r.id
  }

  const openToolPicker = async (vs: VirtualServer) => {
    const qp = new URLSearchParams()
    qp.set('status', 'ACTIVE')
    const [toolsRes, hubsRes, catRes] = await Promise.all([
      api.listTools(qp), api.listHubs(), api.listCatalog()
    ])
    setTools(toolsRes.items)
    setHubs(hubsRes.items)
    setCatalog(catRes.items)
    setSelected([])
    ;(document.getElementById('picker-'+vs.id) as HTMLDialogElement).showModal()
  }

  const saveSelection = async (vs: VirtualServer) => {
    await api.replaceVSTools(vs.id, selected)
    ;(document.getElementById('picker-'+vs.id) as HTMLDialogElement).close()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Virtual Servers</h1>
        <button onClick={create} className="px-3 py-1.5 rounded bg-blue-600 text-white">Create</button>
      </div>
      <div className="grid md:grid-cols-2 gap-4">
        {items.map(vs => (
          <div key={vs.id} className="relative group rounded-2xl border border-white/10 bg-white/[0.04] p-4 transition hover:border-white/20 hover:shadow-[0_8px_30px_rgba(0,0,0,0.35)]">
            <div className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition duration-300" style={{
              background: 'radial-gradient(1000px 300px at 10% -20%, rgba(59,130,246,0.12), transparent 60%), radial-gradient(1000px 300px at 110% 120%, rgba(16,185,129,0.12), transparent 60%)'
            }} />
            <div className="flex items-center justify-between">
              <div className="font-medium">{vs.id}</div>
              <span className={`text-xs px-2 py-1 rounded border border-white/10 ${vs.status==='ACTIVE'?'bg-emerald-500/20 text-emerald-300':'bg-slate-500/20 text-slate-300'}`}>{vs.status}</span>
            </div>
            <div className="mt-2 text-xs">Endpoint: <code>/servers/{vs.id}/mcp</code></div>
            <div className="mt-3 flex gap-2">
              <button onClick={()=>openToolPicker(vs)} className="text-sm px-3 py-1.5 rounded border border-white/10">Manage Tools</button>
              <button onClick={()=>api.setVSStatus(vs.id, vs.status==='ACTIVE'?'DEACTIVATED':'ACTIVE').then(load)} className="text-sm px-3 py-1.5 rounded border border-white/10">{vs.status==='ACTIVE'?'Deactivate':'Activate'}</button>
              <button onClick={()=>api.deleteVS(vs.id).then(load)} className="text-sm px-3 py-1.5 rounded border border-white/10">Delete</button>
            </div>
            <dialog
              id={`picker-${vs.id}`}
              className="w-[min(720px,90vw)] rounded-2xl p-0 border border-white/10 shadow-2xl bg-gradient-to-b from-blue-950/60 to-slate-900/80 backdrop:backdrop-blur-sm backdrop:bg-black/60"
            >
              <div className="p-5 md:p-6 space-y-4">
                <div className="text-lg font-semibold tracking-tight">Select Tools</div>
                <div className="flex items-center gap-3">
                  <input
                    value={q}
                    onChange={e=>setQ(e.target.value)}
                    placeholder="Search..."
                    className="flex-1 px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
                  />
                  <select
                    value={hubFilter}
                    onChange={e=>setHubFilter(e.target.value)}
                    className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
                  >
                    <option value="">All hubs</option>
                    {hubs.map(h => (
                      <option key={h.id} value={h.id}>
                        {catalog.find(c=>c.id===h.mcp_server_id)?.name || h.server_name || h.id}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="h-72 overflow-y-auto pr-1 space-y-1">
                  {tools
                    .filter(t => (hubFilter ? (t.mcp_hub_server_id || t.hub_server_id) === hubFilter : true))
                    .filter(t => t.modified_name.toLowerCase().includes(q.toLowerCase()) || (t.original_name||'').toLowerCase().includes(q.toLowerCase()))
                    .map(t => (
                      <label
                        key={t.id}
                        className="flex items-center gap-3 px-2 py-2 rounded-lg hover:bg-white/5 transition text-sm"
                      >
                        <input
                          type="checkbox"
                          className="accent-blue-500"
                          checked={selected.includes(t.id)}
                          onChange={e=>setSelected(s=>e.target.checked?[...s,t.id]:s.filter(x=>x!==t.id))}
                        />
                        <span className="truncate font-mono text-slate-200">{t.modified_name}</span>
                        <span className="text-xs text-slate-500 italic">({t.original_name})</span>
                      </label>
                  ))}
                </div>
              </div>
              <div className="p-4 border-t border-white/10 flex justify-end gap-2 bg-black/20">
                <button
                  onClick={()=> (document.getElementById('picker-'+vs.id) as HTMLDialogElement).close()}
                  className="px-3 py-1.5 rounded-lg border border-white/15 bg-white/5 text-slate-200 hover:bg-white/10 hover:border-white/25 transition"
                >
                  Cancel
                </button>
                <button
                  onClick={()=>saveSelection(vs)}
                  className="px-4 py-1.5 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-500 text-white shadow-lg shadow-blue-900/30 hover:from-blue-400 hover:to-indigo-400 transition"
                >
                  Save
                </button>
              </div>
            </dialog>
          </div>
        ))}
      </div>
    </div>
  )
}

