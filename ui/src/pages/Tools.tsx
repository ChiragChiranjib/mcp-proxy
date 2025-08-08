import { useEffect, useMemo, useState } from 'react'
import { api, Tool } from '../lib/api'

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

  const filtered = useMemo(() => items.filter(t => t.modified_name.toLowerCase().includes(q.toLowerCase())), [items, q])

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
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered.map(t => (
          <div key={t.id} className="border border-white/10 rounded p-4 bg-white/5">
            <div className="font-medium">{t.modified_name}</div>
            <div className="text-xs text-slate-500">Original: {t.original_name}</div>
            <div className="mt-2 flex items-center justify-between">
              <span className={`text-xs px-2 py-1 rounded border border-white/10 ${t.status==='ACTIVE'?'bg-emerald-500/20 text-emerald-300':'bg-slate-500/20 text-slate-300'}`}>{t.status}</span>
              <div className="flex gap-2">
                <button onClick={()=>toggle(t)} className="text-sm px-3 py-1.5 rounded border border-white/10">{t.status==='ACTIVE'?'Deactivate':'Activate'}</button>
                <button onClick={()=>{ if(confirm('Delete this tool?')) api.deleteTool(t.id).then(load) }} className="text-sm px-3 py-1.5 rounded border border-white/10">Delete</button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

