import { useEffect, useMemo, useState } from 'react'
import { api, Tool } from '../lib/api'
import { JSONViewer } from '../components/JSONViewer'

export function Tools() {
  const [items, setItems] = useState<Tool[]>([])
  const [q, setQ] = useState('')
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    const qp = new URLSearchParams()
    if (q) qp.set('q', q)
    api.listTools(qp).then(r => setItems(r.items)).finally(()=>setLoading(false))
  }
  useEffect(() => { load() }, [])

  const filtered = useMemo(() => items.filter(t => (
    t.modified_name.toLowerCase().includes(q.toLowerCase()) ||
    t.original_name.toLowerCase().includes(q.toLowerCase())
  )), [items, q])

  // Group tools by MCP server name derived from modified_name prefix
  const groups = useMemo(() => {
    const m = new Map<string, Tool[]>()
    for (const t of filtered) {
      const name = t.modified_name || ''
      // server_name is everything before the hyphen we added when composing
      // modified_name = server_name + '-' + original_name
      const idx = name.lastIndexOf('-')
      const server = idx > 0 ? name.substring(0, idx) : 'unknown'
      const arr = m.get(server) || []
      arr.push(t)
      m.set(server, arr)
    }
    return Array.from(m.entries()).sort((a,b)=>a[0].localeCompare(b[0]))
  }, [filtered])

  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({})
  const toggleGroup = (g: string) => setOpenGroups(s => ({ ...s, [g]: !s[g] }))

  const toggle = (t: Tool) => {
    const ns = t.status === 'ACTIVE' ? 'DEACTIVATED' : 'ACTIVE'
    api.setToolStatus(t.id, ns).then(load)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Tools</h1>
        <div className="flex gap-2">
          <input value={q} onChange={e=>setQ(e.target.value)} placeholder="Search..." className="border rounded px-3 py-2 bg-transparent" />
          <button onClick={load} className="px-3 py-1.5 rounded border">Filter</button>
        </div>
      </div>
      {loading && <div className="animate-pulse text-slate-500">Loading tools...</div>}

      <div className="space-y-4">
        {groups.map(([server, tools]) => (
          <div key={server} className="rounded-2xl border border-white/10 bg-white/[0.03] overflow-hidden shadow-[inset_0_1px_0_rgba(255,255,255,0.04)]">
            <button
              onClick={()=>toggleGroup(server)}
              aria-expanded={!!openGroups[server]}
              className="w-full flex items-center justify-between px-4 py-3 rounded-2xl text-left hover:bg-white/5 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-blue-500/30"
            >
              <div className="flex items-center gap-3">
                <span className={`inline-block transition-transform ${openGroups[server] ? 'rotate-90' : ''}`}>▸</span>
                <span className="font-medium">{server}</span>
                <span className="text-xs text-slate-400">{tools.length} tool{tools.length===1?'':'s'}</span>
              </div>
            </button>
            {openGroups[server] && (
              <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4 p-4 border-t border-white/10">
                {tools.map(t => (
                  <div key={t.id} className="relative group rounded-2xl border border-white/5 bg-white/[0.04] p-4 transition hover:border-white/15 hover:shadow-[0_8px_24px_rgba(0,0,0,0.25)]">
                    <div className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition duration-300" style={{
                      background: 'radial-gradient(1000px 300px at 10% -20%, rgba(59,130,246,0.12), transparent 60%), radial-gradient(1000px 300px at 110% 120%, rgba(16,185,129,0.12), transparent 60%)'
                    }} />
                    <div className="font-medium">{t.modified_name}</div>
                    <div className="text-xs text-slate-500">Original: {t.original_name}</div>
                    <details className="mt-2">
                      <summary className="text-xs text-slate-400 cursor-pointer">View input schema</summary>
                      <JSONViewer value={t.input_schema || {}} />
                    </details>
                    <div className="mt-2 flex items-center justify-between">
                      <span className={`text-xs px-2 py-1 rounded border border-white/10 ${t.status==='ACTIVE'?'bg-emerald-500/20 text-emerald-300':'bg-slate-500/20 text-slate-300'}`}>{t.status}</span>
                      <div className="flex gap-2">
                        <button onClick={()=>toggle(t)} className="text-sm px-3 py-1.5 rounded border border-white/10 cursor-pointer hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-blue-500/30 active:scale-95 transition">{t.status==='ACTIVE'?'Deactivate':'Activate'}</button>
                        <button onClick={()=>{ if(confirm('Delete this tool?')) api.deleteTool(t.id).then(load) }} className="text-sm px-3 py-1.5 rounded border border-white/10 cursor-pointer hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-rose-500/30 active:scale-95 transition">Delete</button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

