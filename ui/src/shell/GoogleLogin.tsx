import { useEffect, useRef } from 'react'

type Props = { onSuccess: (credential: string) => void }

export function GoogleLogin({ onSuccess }: Props) {
  const btnRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    let cancelled = false
    let retries = 0

    const init = () => {
      if (cancelled) return
      const clientId = document
        .querySelector('meta[name="google-client-id"]')
        ?.getAttribute('content') || ''
      const win = window as any
      const gis = win.google?.accounts?.id
      if (!clientId || !btnRef.current || !gis) {
        // Retry a few times while the GIS script loads
        if (retries < 50) {
          retries += 1
          window.setTimeout(init, 150)
        }
        return
      }

      try {
        gis.initialize({
          client_id: clientId,
          callback: (resp: any) => {
            if (resp?.credential) onSuccess(resp.credential)
          },
        })
        gis.renderButton(btnRef.current, {
          type: 'standard',
          theme: 'outline',
          size: 'large',
          text: 'signin_with',
          shape: 'rectangular',
          logo_alignment: 'left',
        })
      } catch {
        // ignore
      }
    }

    init()
    return () => { cancelled = true }
  }, [onSuccess])

  return <div ref={btnRef} />
}

