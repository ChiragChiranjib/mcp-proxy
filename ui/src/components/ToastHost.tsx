import { useEffect, useRef, useState } from 'react'
import { showToast, subscribeToast, Toast } from '../lib/toast'

export function ToastHost() {
  const [items, setItems] = useState<Toast[]>([])
  useEffect(() => {
    const unsub = subscribeToast(t => {
      setItems(s => [...s, t])
      window.setTimeout(() => dismiss(t.id), t.durationMs || 4500)
    })
    return unsub
  }, [])

  const dismiss = (id: string) => setItems(s => s.filter(i => i.id !== id))

  return (
    <div className="fixed bottom-4 right-4 z-[100] space-y-3">
      {items.map(t => (
        <ToastItem key={t.id} toast={t} onDismiss={() => dismiss(t.id)} />
      ))}
    </div>
  )
}

function ToastItem({ toast, onDismiss }: { toast: Toast; onDismiss: () => void }) {
  const [progress, setProgress] = useState(100)
  const start = useRef(Date.now())
  const timer = useRef<number | null>(null)
  const duration = toast.durationMs || 4500
  useEffect(() => {
    if (!toast.showProgress) {
      const id = window.setTimeout(onDismiss, duration)
      return () => window.clearTimeout(id)
    }
    const tick = () => {
      const elapsed = Date.now() - start.current
      const p = Math.max(0, 100 - (elapsed / duration) * 100)
      setProgress(p)
      if (p === 0) onDismiss()
      else timer.current = window.setTimeout(tick, 50) as unknown as number
    }
    tick()
    return () => { if (timer.current) window.clearTimeout(timer.current) }
  }, [duration, onDismiss, toast.showProgress])

  const color = toast.variant === 'error' ? 'from-rose-500 to-red-500' : toast.variant === 'success' ? 'from-emerald-500 to-green-500' : 'from-blue-500 to-indigo-500'

  return (
    <div className={`w-80 rounded-2xl border border-white/10 bg-gradient-to-b from-slate-900/95 to-black/95 backdrop-blur shadow-[0_8px_30px_rgba(0,0,0,0.35)] overflow-hidden`}> 
      <div className={`px-4 py-3 text-sm text-slate-100`}>{toast.message}</div>
      {toast.showProgress && (
        <div className={`h-1 bg-gradient-to-r ${color} transition-[width] duration-50`} style={{ width: `${progress}%` }} />
      )}
    </div>
  )
}

// Convenience API for other modules
export function notifyError(msg: string) { showToast({ message: msg, variant: 'error' }) }
export function notifySuccess(msg: string) { showToast({ message: msg, variant: 'success' }) }


