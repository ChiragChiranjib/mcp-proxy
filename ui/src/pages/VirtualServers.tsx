import { useEffect, useState } from 'react'
import { api, VirtualServer, Tool } from '../lib/api'

export function VirtualServers() {
  const [items, setItems] = useState<VirtualServer[]>([])
  const [tools, setTools] = useState<Tool[]>([])
  const [selected, setSelected] = useState<string[]>([])

  const load = () => { api.listVS().then(r=>setItems(r.items)) }
  useEffect(() => { load() }, [])

  const create = async () => {
    const r = await api.createVS()
    await load()
    // optionally open editor for r.id
  }

  const openToolPicker = async (vs: VirtualServer) => {
    const r = await api.listTools(new URLSearchParams())
    setTools(r.items)
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
          <div key={vs.id} className="border border-white/10 rounded p-4 bg-white/5">
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
            <dialog id={`picker-${vs.id}`} className="rounded p-0">
              <div className="p-4 space-y-3">
                <div className="font-medium">Select Tools</div>
                <div className="h-64 overflow-y-auto space-y-1">
                  {tools.map(t => (
                    <label key={t.id} className="flex items-center gap-2 text-sm">
                      <input type="checkbox" checked={selected.includes(t.id)} onChange={e=>setSelected(s=>e.target.checked?[...s,t.id]:s.filter(x=>x!==t.id))} />
                      <span>{t.modified_name}</span>
                    </label>
                  ))}
                </div>
                <div className="flex gap-2 justify-end">
                  <button onClick={()=> (document.getElementById('picker-'+vs.id) as HTMLDialogElement).close()} className="px-3 py-1.5 rounded border">Cancel</button>
                  <button onClick={()=>saveSelection(vs)} className="px-3 py-1.5 rounded bg-blue-600 text-white">Save</button>
                </div>
              </div>
            </dialog>
          </div>
        ))}
      </div>
    </div>
  )
}

