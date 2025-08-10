import { Link, NavLink, Outlet } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import { setBasicCredentials, clearBasicCredentials, hydrateBasicCredentials, getBasicUsername } from '../lib/auth'
import { GoogleLogin } from './GoogleLogin'
import { ToastHost } from '../components/ToastHost'

// Theme toggle removed per design

export function AppLayout() {
  const [user, setUser] = useState<{email?: string; name?: string; userId?: string} | null>(null)
  const [authChecked, setAuthChecked] = useState(false)
  const [basicU, setBasicU] = useState('')
  const [basicP, setBasicP] = useState('')
  const [err, setErr] = useState('')

  useEffect(() => {
    // Check session on load
    hydrateBasicCredentials()
    api.me().then((m) => {
      // we don't get full profile; read cached email if present
      const cached = localStorage.getItem('user-email') || undefined
      setUser({ email: cached, userId: (m as any).user_id })
    }).catch(() => {
      setUser(null)
    }).finally(() => setAuthChecked(true))

    // Initialize Google One-tap button if script is available
    // Google button is rendered via GoogleLogin component
  }, [])

  const onGoogleSuccess = async (credential: string) => {
    try {
      const data = await api.loginWithGoogle(credential)
      setUser({ email: data.email, name: data.name, userId: (data as any).user_id })
      if (data.email) localStorage.setItem('user-email', data.email)
    } catch (e) {
      console.error('login failed', e)
    }
  }

  useEffect(() => {
    const onExpired = () => {
      setUser(null)
      localStorage.removeItem('user-email')
    }
    window.addEventListener('session:expired', onExpired as any)
    return () => window.removeEventListener('session:expired', onExpired as any)
  }, [])

  const logout = async () => {
    try { await api.logout() } catch {}
    setUser(null)
    localStorage.removeItem('user-email')
  }

  const submitBasic = async (e: React.FormEvent) => {
    e.preventDefault()
    setErr('')
    try {
      // Preload basic auth so middleware accepts the login route as well
      setBasicCredentials(basicU, basicP, true)
      const data = await api.loginWithBasic(basicU, basicP)
      setUser({ email: data.email, userId: (data as any).user_id })
      if (data.email) localStorage.setItem('user-email', data.email)
    } catch (e) {
      setErr('Invalid username or password')
      // clear preloaded creds on failure
      clearBasicCredentials()
    }
  }

  if (!authChecked) {
    return <div className="min-h-screen bg-black" />
  }

  if (!user) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-black text-slate-100">
        <header className="border-b border-white/10 sticky top-0 z-10 bg-black/40 backdrop-blur">
          <div className="container flex h-14 items-center justify-between">
            <Link to="/" className="font-semibold">MCP Proxy</Link>
            <div />
          </div>
        </header>
        <main className="container min-h-[calc(100vh-3.5rem)] flex items-center justify-center py-12">
          <div className="mx-auto max-w-3xl">
            <div className="relative rounded-2xl border border-white/10 bg-white/[0.04] p-8 shadow-2xl">
              <div className="pointer-events-none absolute inset-0 rounded-2xl" style={{
                background: 'radial-gradient(1200px 400px at 10% -20%, rgba(59,130,246,0.18), transparent 60%), radial-gradient(1200px 400px at 110% 120%, rgba(16,185,129,0.18), transparent 60%)'
              }} />
              <div className="relative">
                <h1 className="text-2xl font-semibold tracking-tight mb-6">Sign in to continue</h1>
                <div className="grid gap-8 md:grid-cols-2 items-start">
                  <div className="space-y-4">
                    <div className="text-sm text-slate-400">Continue with SSO</div>
                    <div className="rounded-lg bg-white/5 p-4 border border-white/10">
                      <GoogleLogin onSuccess={onGoogleSuccess} />
                    </div>
                  </div>
                  <div className="space-y-4">
                    <div className="text-sm text-slate-400">Or use your credentials</div>
                    <form onSubmit={submitBasic} className="flex flex-col gap-3">
                      <input className="px-3 py-2 w-full rounded-lg border border-white/10 bg-black/30 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/40" placeholder="username" value={basicU} onChange={e=>setBasicU(e.target.value)} />
                      <input className="px-3 py-2 w-full rounded-lg border border-white/10 bg-black/30 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/40" type="password" placeholder="password" value={basicP} onChange={e=>setBasicP(e.target.value)} />
                      <button className="px-4 py-2 rounded-lg border border-white/10 bg-white/10 hover:bg-white/15 transition w-full" type="submit">Login</button>
                    </form>
                    {err && <div className="text-red-400 text-xs">{err}</div>}
                    <p className="text-xs text-slate-500">By continuing you agree to the internal acceptable use policy.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-black text-slate-100">
      <header className="border-b border-white/10 sticky top-0 z-10 bg-black/40 backdrop-blur">
        <div className="container flex h-14 items-center justify-between">
          <Link to="/" className="font-semibold">MCP Proxy</Link>
          <div className="flex items-center gap-6 text-sm">
            <nav className="flex items-center gap-6 text-sm">
              <NavLink to="/" className={({isActive}) => isActive ? 'text-blue-400' : 'text-slate-400 hover:text-slate-200 transition'}>Catalogue</NavLink>
              <NavLink to="/hub" className={({isActive}) => isActive ? 'text-blue-400' : 'text-slate-400 hover:text-slate-200 transition'}>Hub</NavLink>
              <NavLink to="/virtual-servers" className={({isActive}) => isActive ? 'text-blue-400' : 'text-slate-400 hover:text-slate-200 transition'}>Virtual Servers</NavLink>
            </nav>
            <UserMenu email={user.email || ''} onLogout={logout} />
          </div>
        </div>
      </header>
      <main className="container py-8">
        <Outlet />
      </main>
      <ToastHost />
    </div>
  )
}

function UserMenu({ email, onLogout }: { email: string; onLogout: ()=>void }) {
  const [open, setOpen] = useState(false)
  const username = getBasicUsername()
  return (
    <div className="relative">
      <button onClick={()=>setOpen(v=>!v)} className="px-2 py-1 rounded border hover:bg-white/5 hover:border-white/20 transition focus:outline-none focus:ring-2 focus:ring-white/20">Login Details</button>
      {open && (
        <div className="absolute right-0 mt-2 w-64 rounded-xl border border-white/15 bg-black/90 backdrop-blur p-3 shadow-2xl">
          <div className="text-xs text-slate-400 mb-1">Signed in as</div>
          <div className="text-sm break-all">{email || username || (typeof window !== 'undefined' ? (window as any).USER_ID : '') || 'unknown'}</div>
          <div className="mt-3 flex justify-end">
            <button
              className="px-2 py-1 rounded border border-white/10 hover:bg-white/10 hover:border-white/20 focus:outline-none focus:ring-2 focus:ring-rose-500/30 active:scale-95 transition"
              onClick={() => { clearBasicCredentials(); onLogout() }}
              title="Sign out"
            >
              Logout
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

