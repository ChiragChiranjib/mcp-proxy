import { useEffect, useMemo, useState } from 'react'
import { notifyError, notifySuccess } from '../components/ToastHost'
import { api, CatalogServer } from '../lib/api'
import { useNavigate } from 'react-router-dom'

export function Catalogue() {
  const [data, setData] = useState<CatalogServer[]>([])
  const [q, setQ] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    api.listCatalog().then(r => setData(r.items)).catch(e => setError(String(e))).finally(() => setLoading(false))
  }, [])

  const [hubIds, setHubIds] = useState<Set<string>>(new Set())
  useEffect(() => {
    api.listHubs().then(r => {
      const ids = new Set<string>((r.items||[]).map((h:any)=>h.mcp_server_id))
      setHubIds(ids)
    })
  }, [])

  const filtered = useMemo(() => data.filter(s => s.name.toLowerCase().includes(q.toLowerCase())), [data, q])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Catalogue</h1>
        <input value={q} onChange={e=>setQ(e.target.value)} placeholder="Search..." className="border border-white/10 rounded px-3 py-2 bg-black/30 text-slate-200 placeholder:text-slate-500" />
      </div>
      {loading && <div className="animate-pulse text-slate-500">Loading servers...</div>}
      {error && <div className="text-red-500">{error}</div>}
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered.map(s => (
          <div key={s.id} className="relative group rounded-2xl border border-white/10 bg-white/[0.04] p-4 transition hover:border-white/20 hover:shadow-[0_8px_30px_rgba(0,0,0,0.35)]">
            <div className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition duration-300" style={{
              background: 'radial-gradient(1000px 300px at 10% -20%, rgba(59,130,246,0.12), transparent 60%), radial-gradient(1000px 300px at 110% 120%, rgba(16,185,129,0.12), transparent 60%)'
            }} />
            <div className="flex items-center justify-between">
              <div className="font-medium">{s.name}</div>
              <a href={s.url} target="_blank" className="text-xs text-blue-600">Open</a>
            </div>
            <p className="text-sm text-slate-400 mt-1 line-clamp-3">{s.description || '—'}</p>
            <details className="mt-2">
              <summary className="text-xs text-slate-400 cursor-pointer">Details</summary>
              <div className="text-xs mt-2 space-y-1">
                <div><span className="text-slate-500">URL:</span> <a href={s.url} target="_blank" className="text-blue-500 break-all">{s.url}</a></div>
                <div><span className="text-slate-500">ID:</span> <code className="break-all">{s.id}</code></div>
              </div>
            </details>
            <AddToHubButton serverId={s.id} added={hubIds.has(s.id)} onAdded={()=>setHubIds(new Set([...Array.from(hubIds), s.id]))} />
          </div>
        ))}
      </div>
    </div>
  )
}

function AddToHubButton({ serverId, added, onAdded }: { serverId: string, added?: boolean, onAdded?: ()=>void }) {
  const [open, setOpen] = useState(false)
  const [authType, setAuthType] = useState<'none' | 'bearer' | 'custom_headers'>('none')
  const [authValue, setAuthValue] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const submit = async () => {
    setSaving(true)
    setError(null)
    try {
      let val: any = null
      if (authType === 'bearer') val = authValue
      if (authType === 'custom_headers') val = JSON.parse(authValue || '{}')
      await api.addHub({ mcp_server_id: serverId, transport: 'streamable-http', capabilities: null, auth_type: authType, auth_value: val })
      setOpen(false)
      notifySuccess('Added to hub')
    } catch (e:any) {
      setError(e.message || String(e))
      notifyError(e?.message || 'Failed to add to hub')
    } finally { setSaving(false) }
  }

  if (added) {
    return <div className="mt-3"><span className="text-xs px-2 py-1 rounded border border-white/10 bg-emerald-500/20 text-emerald-300">Added</span></div>
  }

  return (
    <div className="mt-3">
      {!open ? (
        <button
          onClick={()=>setOpen(true)}
          className="inline-flex items-center justify-center text-sm font-medium px-4 py-2 rounded-xl bg-gradient-to-r from-emerald-500 to-green-400 text-white shadow-lg shadow-emerald-900/30 hover:from-emerald-400 hover:to-teal-400 hover:shadow-emerald-800/40 active:scale-[0.98] transition-all duration-300 cursor-pointer focus:outline-none focus:ring-2 focus:ring-emerald-500/30"
        >
          Add to Hub
        </button>
      ) : (
          <div className="border border-white/10 rounded p-3 space-y-2 bg-black/30">
          <div className="text-sm font-medium">Add to Hub</div>
            <label className="text-xs block mb-1">Auth Type</label>
          <select value={authType} onChange={e=>setAuthType(e.target.value as any)} className="border border-white/10 rounded px-2 py-1 bg-black/30 focus:outline-none focus:ring-2 focus:ring-blue-500/30">
            <option value="none">None</option>
            <option value="bearer">Bearer Token</option>
            <option value="custom_headers">Custom Headers (JSON)</option>
          </select>
          {(authType === 'bearer' || authType === 'custom_headers') && (
            <textarea value={authValue} onChange={e=>setAuthValue(e.target.value)} placeholder={authType==='bearer' ? 'token' : '{"X-Api-Key":"..."}'} className="border border-white/10 rounded p-2 w-full bg-black/30" rows={3} />
          )}
          {error && <div className="text-xs text-red-500">{error}</div>}
          <div className="flex gap-3">
            <button
              onClick={async ()=>{ await submit(); onAdded && onAdded(); }}
              disabled={saving}
              className="inline-flex items-center justify-center text-sm font-medium px-4 py-2 rounded-xl bg-gradient-to-r from-emerald-500 to-green-400 text-white shadow-lg shadow-emerald-900/30 hover:from-emerald-400 hover:to-teal-400 hover:shadow-emerald-800/40 active:scale-[0.98] transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer focus:outline-none focus:ring-2 focus:ring-emerald-500/30"
            >
              {saving? 'Adding…' : 'Add'}
            </button>
            <button
              onClick={()=>setOpen(false)}
              className="inline-flex items-center justify-center text-sm font-medium px-4 py-2 rounded-xl border border-white/15 bg-white/5 text-slate-200 hover:bg-white/10 hover:border-white/25 active:scale-[0.98] transition-all duration-300 cursor-pointer focus:outline-none focus:ring-2 focus:ring-white/20"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

