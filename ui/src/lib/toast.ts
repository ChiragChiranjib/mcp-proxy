export type Toast = {
  id: string
  message: string
  variant?: 'error' | 'success' | 'info'
  durationMs?: number
  showProgress?: boolean
}

type Listener = (t: Toast) => void
const listeners: Listener[] = []

export function subscribeToast(fn: Listener) {
  listeners.push(fn)
  return () => {
    const i = listeners.indexOf(fn)
    if (i >= 0) listeners.splice(i, 1)
  }
}

export function showToast(partial: Omit<Toast, 'id'>) {
  const t: Toast = {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    durationMs: 4500,
    variant: 'info',
    showProgress: true,
    ...partial,
  }
  for (const l of listeners) l(t)
}


