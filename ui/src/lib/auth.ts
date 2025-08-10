// Basic auth credential store (in-memory + sessionStorage opt-in)

let basicUsername: string | null = null
let basicPassword: string | null = null

export function setBasicCredentials(u: string, p: string, remember = true) {
  basicUsername = u
  basicPassword = p
  if (remember) {
    sessionStorage.setItem('basic_u', u)
    sessionStorage.setItem('basic_p', p)
  }
}

export function clearBasicCredentials() {
  basicUsername = null
  basicPassword = null
  sessionStorage.removeItem('basic_u')
  sessionStorage.removeItem('basic_p')
}

export function hydrateBasicCredentials() {
  if (!basicUsername || !basicPassword) {
    const u = sessionStorage.getItem('basic_u')
    const p = sessionStorage.getItem('basic_p')
    if (u && p) {
      basicUsername = u
      basicPassword = p
    }
  }
}

export function getBasicAuthHeader(): Record<string, string> {
  hydrateBasicCredentials()
  if (!basicUsername || !basicPassword) return {}
  const token = btoa(`${basicUsername}:${basicPassword}`)
  return { Authorization: `Basic ${token}` }
}


