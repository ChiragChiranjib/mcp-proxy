import React, { useEffect, useMemo, useState } from 'react'
import { notifyError, notifySuccess } from '../components/ToastHost'
import { api, CatalogServer } from '../lib/api'
import { useNavigate } from 'react-router-dom'

export function Catalogue() {
  const [data, setData] = useState<CatalogServer[]>([])
  const [q, setQ] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [role, setRole] = useState<string | undefined>(undefined)
  const [addOpen, setAddOpen] = useState(false)
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [description, setDescription] = useState('')
  const [accessType, setAccessType] = useState<'public' | 'private'>('public')
  const [saving, setSaving] = useState(false)
  const [editOpen, setEditOpen] = useState<null | CatalogServer>(null)
  const [editUrl, setEditUrl] = useState('')
  const [editDesc, setEditDesc] = useState('')
  const [refreshingId, setRefreshingId] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    api.listCatalog().then(r => setData(r.items)).catch(e => setError(String(e))).finally(() => setLoading(false))
    api.me().then(m=>setRole((m as any).role)).catch(()=>{})
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
      <div className="flex items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold">Catalogue</h1>
        <div className="flex items-center gap-2">
          <input
            value={q}
            onChange={e=>setQ(e.target.value)}
            placeholder="Search..."
            className="h-10 border border-white/10 rounded-lg px-3 bg-black/30 text-slate-200 placeholder:text-slate-500"
          />
          <button
            disabled={role !== 'ADMIN'}
            onClick={()=>setAddOpen(true)}
            className={`h-10 px-4 rounded-lg border border-white/10 text-sm ${role!=='ADMIN' ? 'text-slate-500 cursor-not-allowed' : 'hover:bg-white/10 hover:border-white/20'}`}
            title={role==='ADMIN' ? 'Add MCP Server' : 'Admin only'}
          >
            Add
          </button>
        </div>
      </div>
      {loading ? (
        <div className="animate-pulse text-slate-500">Loading servers...</div>
      ) : error ? (
        <div className="text-red-500">{error}</div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filtered.map(s => (
          <div key={s.id} className="relative group rounded-2xl border border-white/10 bg-white/[0.04] p-4 transition hover:border-white/20 hover:shadow-[0_8px_30px_rgba(0,0,0,0.35)]">
            <div className="pointer-events-none absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100 transition duration-300" style={{
              background: 'radial-gradient(1000px 300px at 10% -20%, rgba(59,130,246,0.12), transparent 60%), radial-gradient(1000px 300px at 110% 120%, rgba(16,185,129,0.12), transparent 60%)'
            }} />
            <div className="flex items-center justify-between">
              <div className="font-medium">{s.name}</div>
              <div className="flex items-center gap-2">
                <button
                  disabled={role !== 'ADMIN' || refreshingId === s.id}
                  onClick={async () => {
                    if (role !== 'ADMIN') return
                    setRefreshingId(s.id)
                    try {
                      const result = await api.refreshCatalog(s.id)
                      notifySuccess(`Refreshed: ${result.total_added} added, ${result.total_deleted} deleted`)
                    } catch (e: any) {
                      notifyError(e?.message || 'Refresh failed')
                    } finally {
                      setRefreshingId(null)
                    }
                  }}
                  className={`text-xs px-2 py-1 rounded border border-white/10 ${role!=='ADMIN' ? 'text-slate-500 cursor-not-allowed' : 'hover:bg-white/10 hover:border-white/20'} flex items-center gap-1`}
                  title={role==='ADMIN' ? 'Refresh server tools' : 'Admin only'}
                >
                  <span className={`${refreshingId === s.id ? 'animate-spin' : ''}`}>⟳</span>
                  Refresh
                </button>
                <button
                  disabled={role !== 'ADMIN'}
                  onClick={()=>{ setEditOpen(s); setEditUrl(s.url); setEditDesc(s.description||'') }}
                  className={`text-xs px-2 py-1 rounded border border-white/10 ${role!=='ADMIN' ? 'text-slate-500 cursor-not-allowed' : 'hover:bg-white/10 hover:border-white/20'}`}
                  title={role==='ADMIN' ? 'Edit' : 'Admin only'}
                >Edit</button>
              </div>
            </div>
            <p className="text-sm text-slate-400 mt-1 line-clamp-3">{s.description || '—'}</p>
            <details className="mt-2">
              <summary className="text-xs text-slate-400 cursor-pointer">Details</summary>
              <div className="text-xs mt-2 space-y-1">
                <div><span className="text-slate-500">URL:</span> <a href={s.url} target="_blank" className="text-blue-500 break-all">{s.url}</a></div>
                <div><span className="text-slate-500">ID:</span> <code className="break-all">{s.id}</code></div>
              </div>
            </details>
            <AddToHubButton serverId={s.id} serverAccessType={s.access_type} added={hubIds.has(s.id)} onAdded={()=>setHubIds(new Set([...Array.from(hubIds), s.id]))} />
          </div>
        ))}
        </div>
      )}
      {/* Add catalog modal */}
      {addOpen && (
        <div className="fixed inset-0 z-[2000] bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="w-[min(560px,95vw)] rounded-2xl border border-white/10 bg-gradient-to-b from-blue-950/60 to-slate-900/80 shadow-2xl">
            <div className="p-5 md:p-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="text-lg font-semibold">Add MCP Server</div>
                <button onClick={()=>{setAddOpen(false); setName(''); setUrl(''); setDescription(''); setAccessType('public')}} className="px-2 py-1 rounded border border-white/10 hover:bg-white/10">Close</button>
              </div>
              <div className="grid gap-3">
                <div className="grid gap-1">
                  <label className="text-xs text-slate-400">Name</label>
                  <input value={name} onChange={e=>setName(e.target.value)} className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40" placeholder="Name" />
                </div>
                <div className="grid gap-1">
                  <label className="text-xs text-slate-400">URL</label>
                  <input
                    value={url}
                    onChange={e=>setUrl(e.target.value)}
                    className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
                    placeholder="Provide the streamable MCP Server endpoint"
                  />
                </div>
                <div className="grid gap-1">
                  <label className="text-xs text-slate-400">Description</label>
                  <input value={description} onChange={e=>setDescription(e.target.value)} className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40" placeholder="What this server provides" />
                </div>
                <div className="grid gap-1">
                  <label className="text-xs text-slate-400">Access Type</label>
                  <select 
                    value={accessType} 
                    onChange={e=>setAccessType(e.target.value as 'public' | 'private')} 
                    className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40"
                  >
                    <option value="public">Public - Auto-fetch tools for all users</option>
                    <option value="private">Private - Users add with their own auth</option>
                  </select>
                </div>
              </div>
            </div>
            <div className="p-4 border-t border-white/10 bg-black/20 flex justify-end gap-2">
              <button onClick={()=>{setAddOpen(false); setName(''); setUrl(''); setDescription(''); setAccessType('public')}} className="px-3 py-1.5 rounded-lg border border-white/15 bg-white/5 text-slate-200 hover:bg-white/10 hover:border-white/25 transition">Cancel</button>
              <button
                disabled={saving || !name || !url || !description}
                onClick={async()=>{
                  setSaving(true)
                  try{
                    await api.addCatalog({ name, url: url.trim(), description, access_type: accessType })
                    setAddOpen(false)
                    setName(''); setUrl(''); setDescription(''); setAccessType('public')
                    // Refetch the catalog without showing loading state
                    const r = await api.listCatalog();
                    setData(r.items)
                  }catch(e:any){ notifyError(e?.message||'Failed to add server') } finally { setSaving(false) }
                }}
                className={`px-4 py-1.5 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-500 text-white ${saving || !name || !url || !description ? 'opacity-60 cursor-not-allowed' : 'hover:from-blue-400 hover:to-indigo-400'} transition`}
              >Save</button>
            </div>
          </div>
        </div>
      )}

      {/* Edit catalog modal */}
      {editOpen && (
        <div className="fixed inset-0 z-[2000] bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="w-[min(560px,95vw)] rounded-2xl border border-white/10 bg-gradient-to-b from-blue-950/60 to-slate-900/80 shadow-2xl">
            <div className="p-5 md:p-6 space-y-4">
              <div className="flex items-center justify-between">
                <div className="text-lg font-semibold">Edit {editOpen.name}</div>
                <button onClick={()=>setEditOpen(null)} className="px-2 py-1 rounded border border-white/10 hover:bg-white/10">Close</button>
              </div>
              <div className="grid gap-3">
                <div className="grid gap-1">
                  <label className="text-xs text-slate-400">URL</label>
                  <input value={editUrl} onChange={e=>setEditUrl(e.target.value)} className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40" placeholder="URL" />
                </div>
                <div className="grid gap-1">
                  <label className="text-xs text-slate-400">Description</label>
                  <input value={editDesc} onChange={e=>setEditDesc(e.target.value)} className="px-3 py-2 rounded-lg border border-white/10 bg-white/5 text-slate-200 focus:outline-none focus:ring-2 focus:ring-blue-500/40" placeholder="Description" />
                </div>
              </div>
            </div>
            <div className="p-4 border-t border-white/10 bg-black/20 flex justify-end gap-2">
              <button onClick={()=>setEditOpen(null)} className="px-3 py-1.5 rounded-lg border border-white/15 bg-white/5 text-slate-200 hover:bg-white/10 hover:border-white/25 transition">Cancel</button>
              <button
                onClick={async()=>{
                  if (!editOpen) return
                  try {
                    await api.updateCatalog(editOpen.id, { url: editUrl.trim() || undefined, description: editDesc })
                    notifySuccess('Updated')
                    setEditOpen(null)
                    const r = await api.listCatalog();
                    setData(r.items)
                  } catch(e:any) { notifyError(e?.message || 'Failed to update') }
                }}
                className="px-4 py-1.5 rounded-lg bg-gradient-to-r from-blue-500 to-indigo-500 text-white hover:from-blue-400 hover:to-indigo-400 transition"
              >Save</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// URL validation removed as per latest requirement

function AddCatalogButton({ disabled, onAdded }: { disabled?: boolean; onAdded?: ()=>void }) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [description, setDescription] = useState('')
  const [accessType, setAccessType] = useState<'public' | 'private'>('public')
  const [saving, setSaving] = useState(false)
  if (disabled) {
    return <button disabled className="text-xs px-3 py-1.5 rounded border border-white/10 text-slate-500 cursor-not-allowed">Add</button>
  }
  return (
    <div>
      {!open ? (
        <button onClick={()=>setOpen(true)} className="text-xs px-3 py-1.5 rounded border border-white/10 hover:bg-white/10 hover:border-white/20">Add</button>
      ) : (
        <div className="absolute right-6 mt-2 z-20 w-[min(420px,95vw)] rounded-xl border border-white/10 bg-black/80 backdrop-blur p-3 shadow-xl">
          <div className="text-sm font-medium mb-2">Add MCP Server</div>
          <div className="grid gap-2">
            <input className="px-2 py-1 rounded border border-white/10 bg-white/5" placeholder="Name" value={name} onChange={e=>setName(e.target.value)} />
            <input className="px-2 py-1 rounded border border-white/10 bg-white/5" placeholder="URL" value={url} onChange={e=>setUrl(e.target.value)} />
            <input className="px-2 py-1 rounded border border-white/10 bg-white/5" placeholder="Description (optional)" value={description} onChange={e=>setDescription(e.target.value)} />
            <select 
              value={accessType} 
              onChange={e=>setAccessType(e.target.value as 'public' | 'private')} 
              className="px-2 py-1 rounded border border-white/10 bg-white/5 text-xs"
            >
              <option value="public">Public</option>
              <option value="private">Private</option>
            </select>
          </div>
          <div className="mt-3 flex gap-2 justify-end">
            <button onClick={()=>{setOpen(false); setName(''); setUrl(''); setDescription(''); setAccessType('public')}} className="text-xs px-3 py-1.5 rounded border border-white/10 hover:bg-white/10 hover:border-white/20">Cancel</button>
            <button
              disabled={saving || !name || !url}
              onClick={async()=>{
                setSaving(true)
                try{
                  await api.addCatalog({ name, url, description, access_type: accessType })
                  setOpen(false)
                  setName(''); setUrl(''); setDescription(''); setAccessType('public')
                  onAdded && onAdded()
                }catch(e:any){ notifyError(e?.message||'Failed to add server') } finally { setSaving(false) }
              }}
              className="text-xs px-3 py-1.5 rounded bg-blue-600 text-white disabled:opacity-50"
            >Save</button>
          </div>
        </div>
      )}
    </div>
  )
}

function AddToHubButton({ serverId, serverAccessType, added, onAdded }: { serverId: string, serverAccessType?: string, added?: boolean, onAdded?: ()=>void }) {
  const [open, setOpen] = useState(false)
  const [authType, setAuthType] = useState<'none' | 'bearer' | 'custom_headers'>(serverAccessType === 'private' ? 'bearer' : 'none')
  const [authValue, setAuthValue] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  const isPublic = serverAccessType === 'public'

  const submit = async () => {
    setSaving(true)
    setError(null)
    try {
      let val: any = null
      if (authType === 'bearer') val = authValue
      if (authType === 'custom_headers') val = JSON.parse(authValue || '{}')
      await api.addHub({ mcp_server_id: serverId, auth_type: authType, auth_value: val })
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
          onClick={async () => {
            if (isPublic) {
              // For public servers, directly add with auth_type: 'none'
              setSaving(true)
              setError(null)
              try {
                await api.addHub({ mcp_server_id: serverId, auth_type: 'none', auth_value: null })
                notifySuccess('Added to hub')
                onAdded && onAdded()
              } catch (e: any) {
                notifyError(e?.message || 'Failed to add to hub')
              } finally {
                setSaving(false)
              }
            } else {
              // For private servers, show the auth dialog
              setOpen(true)
            }
          }}
          disabled={saving}
          className="inline-flex items-center justify-center text-sm font-medium px-4 py-2 rounded-xl bg-gradient-to-r from-emerald-500 to-green-400 text-white shadow-lg shadow-emerald-900/30 hover:from-emerald-400 hover:to-teal-400 hover:shadow-emerald-800/40 active:scale-[0.98] transition-all duration-300 cursor-pointer focus:outline-none focus:ring-2 focus:ring-emerald-500/30 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saving ? 'Adding...' : 'Add to Hub'}
        </button>
      ) : (
          <div className="border border-white/10 rounded p-3 space-y-2 bg-black/30">
          <div className="text-sm font-medium">Add to Hub</div>
            <label className="text-xs block mb-1">Auth Type</label>
          <select value={authType} onChange={e=>setAuthType(e.target.value as any)} className="border border-white/10 rounded px-2 py-1 bg-black/30 focus:outline-none focus:ring-2 focus:ring-blue-500/30">
            <option value="bearer">Bearer Token</option>
            <option value="custom_headers">Custom Headers (JSON)</option>
          </select>
          <textarea value={authValue} onChange={e=>setAuthValue(e.target.value)} placeholder={authType==='bearer' ? 'token' : '{"X-Api-Key":"..."}'} className="border border-white/10 rounded p-2 w-full bg-black/30" rows={3} />
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

