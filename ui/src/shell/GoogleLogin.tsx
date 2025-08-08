import { useEffect, useRef } from 'react'

type Props = { onSuccess: (credential: string) => void }

export function GoogleLogin({ onSuccess }: Props) {
  const btnRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const clientId = document
      .querySelector('meta[name="google-client-id"]')
      ?.getAttribute('content') || ''
    const win = window as any
    const gis = win.google?.accounts?.id
    if (!clientId || !btnRef.current || !gis) return

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
  }, [onSuccess])

  return <div ref={btnRef} />
}

