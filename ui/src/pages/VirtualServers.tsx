import { useEffect, useMemo, useState, useEffect as ReactUseEffect } from 'react'
import { api, VirtualServer, Tool, HubServer, CatalogServer } from '../lib/api'
import { notifyError, notifySuccess } from '../components/ToastHost'

export function VirtualServers() {
  const [items, setItems] = useState<VirtualServer[]>([])
  const [tools, setTools] = useState<Tool[]>([])
  const [hubs, setHubs] = useState<HubServer[]>([])
  const [catalog, setCatalog] = useState<CatalogServer[]>([])
  const [selected, setSelected] = useState<string[]>([])
  const [q, setQ] = useState('')
  const [hubFilter, setHubFilter] = useState('')
  const [toolsByVS, setToolsByVS] = useState<Record<string, Tool[]>>({})
  const [vsToolsVisible, setVsToolsVisible] = useState<Record<string, boolean>>({})
  const [copied, setCopied] = useState<Record<string, boolean>>({})

  const load = () => { api.listVS().then(r=>setItems(r.items)) }
  useEffect(() => { load() }, [])

  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [createSelected, setCreateSelected] = useState<string[]>([])
  const [createQ, setCreateQ] = useState('')
  const [createGroups, setCreateGroups] = useState<Array<{server: string; tools: Tool[]}>>([])

  const create = async () => {
    setCreateOpen(true)
    setCreateSelected([])
    setCreateQ('')
    // Fetch all tools grouped by server for current user
    try {
      const qp = new URLSearchParams()
      const res = await api.listTools(qp)
      // group by server prefix from modified_name
      const m = new Map<string, Tool[]>()
      for (const t of res.items) {
        const name = t.modified_name || ''
        const idx = name.lastIndexOf('-')
        const server = idx > 0 ? name.substring(0, idx) : 'unknown'
        const arr = m.get(server) || []
        arr.push(t)
        m.set(server, arr)
      }
      const groups = Array
        .from(m.entries())
        .map(([server, tools])=>({
          server,
          tools: tools.sort((a,b)=>a.modified_name.localeCompare(b.modified_name)),
        }))
        .sort((a,b)=>a.server.localeCompare(b.server))
      setCreateGroups(groups)
    } catch (e: any) {
      notifyError(e?.message || 'Failed to load tools')
    }
  }

  // Lock body scroll when modal open and ensure overlay covers full page
  useEffect(() => {
    if (createOpen) {
      const prev = document.body.style.overflow
      document.body.style.overflow = 'hidden'
      return () => { document.body.style.overflow = prev }
    }
    return
  }, [createOpen])

  const openToolPicker = async (vs: VirtualServer) => {
    const qp = new URLSearchParams()
    const [toolsRes, hubsRes, catRes] = await Promise.all([
      api.listTools(qp), api.listHubs(), api.listCatalog()
    ])
    setTools(toolsRes.items)
    setHubs(hubsRes.items)
    setCatalog(catRes.items)
    setSelected([])
    // preselect current vs tools
    try {
      const cur = await api.listVSTools(vs.id)
      setSelected(cur.items.map(t=>t.id))
    } catch { setSelected([]) }
    ;(document.getElementById('picker-'+vs.id) as HTMLDialogElement).showModal()
  }

  const saveSelection = async (vs: VirtualServer) => {
    await api.replaceVSTools(vs.id, selected)
    ;(document.getElementById('picker-'+vs.id) as HTMLDialogElement).close()
    // refresh visible tools for this VS
    try {
      const res = await api.listVSTools(vs.id)
      setToolsByVS(s => ({ ...s, [vs.id]: res.items }))
    } catch {}
  }

  const ensureVSToolsLoaded = async (vsId: string) => {
    if (toolsByVS[vsId]) return
    try {
      const res = await api.listVSTools(vsId)
      setToolsByVS(s => ({ ...s, [vsId]: res.items }))
    } catch {}
  }

  const toggleShowTools = async (vsId: string) => {
    const visible = !!vsToolsVisible[vsId]
    if (!visible && !toolsByVS[vsId]) {
      await ensureVSToolsLoaded(vsId)
    }
    setVsToolsVisible(s => ({ ...s, [vsId]: !visible }))
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold">Virtual Servers</h1>
        <div className="flex items-center gap-2">
          <button onClick={create} className="px-3 py-1.5 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-500 text-white shadow-lg shadow-blue-900/30 hover:from-blue-400 hover:to-indigo-400 active:scale-95 transition">
            Create
          </button>
        </div>
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {items.map(vs => (
          <div key={vs.id} className="relative group rounded-2xl border border-white/10 bg-white/[0.04] p-4 transition hover:border-white/20 hover:shadow-[0_8px_30px_rgba(0,0,0,0.35)] shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]">
            <div className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition duration-300" style={{
              background: 'radial-gradient(1000px 300px at 10% -20%, rgba(59,130,246,0.12), transparent 60%), radial-gradient(1000px 300px at 110% 120%, rgba(16,185,129,0.12), transparent 60%)'
            }} />
            <div className="flex items-center justify-between">
              <div className="font-medium truncate max-w-[60%]">{vs.name || vs.id}</div>
              <div className="flex items-center gap-2">
                <button
                  onClick={()=>toggleShowTools(vs.id)}
                  className="text-xs px-2 py-1 rounded-lg border border-white/10 hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-blue-500/30 active:scale-95 transition"
                >
                  {vsToolsVisible[vs.id] ? 'Hide tools' : 'Show tools'}
                </button>
                <span className="inline-flex items-center gap-1.5 text-xs px-2 py-1 rounded border border-white/10">
                  <span className={`inline-block w-2 h-2 rounded-full ${vs.status==='ACTIVE'?'bg-emerald-400 shadow-[0_0_8px_rgba(16,185,129,0.8)]':'bg-slate-500'}`} />
                  <span className={`${vs.status==='ACTIVE'?'text-emerald-300':'text-slate-300'}`}>{vs.status}</span>
                </span>
              </div>
            </div>
            <div className="mt-2 text-xs opacity-80 flex items-center gap-2">
              <code className="bg-black/30 px-1 py-0.5 rounded border border-white/10">/servers/{vs.id}/mcp</code>
              <button
                onClick={async ()=>{
                  try {
                    await navigator.clipboard.writeText(`/servers/${vs.id}/mcp`)
                    setCopied(s => ({ ...s, [vs.id]: true }))
                    window.setTimeout(() => {
                      setCopied(s => { const n = { ...s }; delete n[vs.id]; return n })
                    }, 1200)
                  } catch {}
                }}
                className={`inline-flex items-center justify-center w-6 h-6 rounded border border-white/10 focus:outline-none focus:ring-2 focus:ring-blue-500/30 transition ${copied[vs.id] ? 'bg-emerald-600/20 border-emerald-400/30' : 'hover:bg-white/10 hover:border-white/20'}`}
                title={copied[vs.id] ? 'Copied!' : 'Copy endpoint'}
                aria-label={copied[vs.id] ? 'Copied!' : 'Copy endpoint'}
              >
                {copied[vs.id] ? (
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-3.5 h-3.5 text-emerald-300">
                    <path d="M20 6L9 17l-5-5" />
                  </svg>
                ) : (
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="1.5"
                    className="w-3.5 h-3.5 text-slate-200"
                  >
                    <rect x="9" y="9" width="11" height="11" rx="2" />
                    <rect x="4" y="4" width="11" height="11" rx="2" />
                  </svg>
                )}
              </button>
            </div>
            <div className="mt-3 flex gap-2">
              <button onClick={()=>openToolPicker(vs)} className="text-sm px-3 py-1.5 rounded-lg border border-white/10 hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-blue-500/30 active:scale-95 transition">Manage Tools</button>
              <button onClick={()=>api.setVSStatus(vs.id, vs.status==='ACTIVE'?'DEACTIVATED':'ACTIVE').then(load)} className="text-sm px-3 py-1.5 rounded-lg border border-white/10 hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-blue-500/30 active:scale-95 transition">{vs.status==='ACTIVE'?'Deactivate':'Activate'}</button>
              <button onClick={()=>api.deleteVS(vs.id).then(load)} className="text-sm px-3 py-1.5 rounded-lg border border-white/10 hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-rose-500/30 active:scale-95 transition">Delete</button>
            </div>

            {/* VS Tools chips */}
            {vsToolsVisible[vs.id] && (
            <div className="mt-3">
              <div className="flex flex-wrap gap-2">
                {(toolsByVS[vs.id] || []).map(t => (
                  <span
                    key={t.id}
                    className="inline-flex items-center gap-1.5 pl-3 pr-1 py-1 h-7 max-w-[220px] rounded-full text-xs border border-white/10 bg-white/[0.06] hover:bg-white/[0.12] hover:border-white/20 shadow-[inset_0_1px_0_rgba(255,255,255,0.08)] transition backdrop-blur-sm"
                    title={t.original_name}
                  >
                    <span className="truncate max-w-[170px] font-mono text-[12px] text-slate-200">
                      {t.modified_name}
                    </span>
                    <button
                      onClick={async ()=>{
                        await api.removeVSTool(vs.id, t.id)
                        const res = await api.listVSTools(vs.id)
                        setToolsByVS(s=>({ ...s, [vs.id]: res.items }))
                      }}
                      className="ml-1 grid place-items-center w-5 h-5 text-[13px] text-slate-400 hover:text-slate-100 focus:outline-none"
                      title="Remove from virtual server"
                    >×</button>
                  </span>
                ))}
                {(toolsByVS[vs.id] && (toolsByVS[vs.id] || []).length === 0) && (
                  <span className="text-xs text-slate-500">No tools assigned</span>
                )}
              </div>
            </div>
            )}
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
                <div className="h-72 overflow-y-auto pr-1 space-y-1 scroll-panel">
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
      {/* Create modal */}
      {createOpen && (
        <div className="fixed inset-0 z-[2000] bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="w-[min(860px,95vw)] max-h-[85vh] overflow-hidden rounded-2xl border border-white/10 bg-gradient-to-b from-blue-950/60 to-slate-900/80 shadow-2xl">
            <div className="p-5 md:p-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="text-lg font-semibold">Create Virtual Server</div>
                <button onClick={()=>setCreateOpen(false)} className="px-2 py-1 rounded border border-white/10 hover:bg-white/10">Close</button>
              </div>
              <div className="grid md:grid-cols-3 gap-3 items-center">
                <input value={newName} onChange={e=>setNewName(e.target.value)} placeholder="Name" className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/40 md:col-span-1" />
                <input value={createQ} onChange={e=>setCreateQ(e.target.value)} placeholder="Search tools..." className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/40 md:col-span-2" />
              </div>
              <div className="h-[48vh] overflow-y-auto pr-1 scroll-panel space-y-3">
                {createGroups.map(g => {
                  const tools = g.tools.filter(t => t.modified_name.toLowerCase().includes(createQ.toLowerCase()) || (t.original_name||'').toLowerCase().includes(createQ.toLowerCase()))
                  if (tools.length === 0) return null
                  const allSelected = tools.every(t => createSelected.includes(t.id))
                  return (
                    <div key={g.server} className="rounded-xl border border-white/10 bg-white/[0.03]">
                      <button
                        onClick={()=>{
                          setCreateSelected(s => {
                            const ids = tools.map(t=>t.id)
                            const set = new Set(s)
                            const every = ids.every(id=>set.has(id))
                            if (every) {
                              ids.forEach(id=>set.delete(id))
                            } else {
                              ids.forEach(id=>set.add(id))
                            }
                            return Array.from(set)
                          })
                        }}
                        className="w-full px-4 py-2 text-sm font-medium flex items-center justify-between text-left hover:bg-white/5 rounded-t-xl"
                        aria-pressed={allSelected}
                        title="Click to select/deselect all"
                      >
                        <span className="flex items-center gap-2">
                          <span className={`inline-block w-2.5 h-2.5 rounded-full ${allSelected ? 'bg-emerald-400' : 'bg-slate-500'}`} />
                          {g.server}
                        </span>
                        <span className="text-xs text-slate-500">{tools.length} tool{tools.length===1?'':'s'}</span>
                      </button>
                      <div className="border-t border-white/10 divide-y divide-white/10">
                        {tools.map(t => (
                          <label key={t.id} className="flex items-start gap-3 px-4 py-2 hover:bg-white/5 transition text-sm">
                            <input type="checkbox" className="mt-0.5 accent-blue-500" checked={createSelected.includes(t.id)} onChange={e=>setCreateSelected(s=>e.target.checked?[...s,t.id]:s.filter(x=>x!==t.id))} />
                            <div className="min-w-0">
                              <div title={t.modified_name} className="whitespace-normal break-words leading-snug text-slate-100 font-medium">
                                {t.modified_name}
                              </div>
                              <div className="text-xs text-slate-500 italic">
                                {t.original_name}
                              </div>
                            </div>
                          </label>
                        ))}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
            <div className="p-4 border-t border-white/10 bg-black/20 flex justify-end gap-2">
              <button onClick={()=>setCreateOpen(false)} className="px-3 py-1.5 rounded-lg border border-white/15 bg-white/5 text-slate-200 hover:bg-white/10 hover:border-white/25 transition">Cancel</button>
              <button
                onClick={async ()=>{
                  setCreating(true)
                  try {
                    const res = await api.createVS(newName || undefined, createSelected)
                    notifySuccess('Virtual server created')
                    setCreateOpen(false)
                    setCreateSelected([])
                    setNewName('')
                    await load()
                  } catch (e: any) {
                    notifyError(e?.message || 'Create failed')
                  } finally {
                    setCreating(false)
                  }
                }}
                disabled={creating}
                className={`px-4 py-1.5 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-500 text-white shadow-lg shadow-blue-900/30 hover:from-blue-400 hover:to-indigo-400 transition ${creating ? 'opacity-60 cursor-not-allowed' : ''}`}
              >
                {creating ? 'Creating...' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

