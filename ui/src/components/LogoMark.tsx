import React from 'react'

type Props = { className?: string }

export function LogoMark({ className = 'w-6 h-6' }: Props) {
  const client = { x: 6, y: 18, r: 3 }
  const proxy = { x: 20, y: 18, r: 6 }
  const srvA = { x: 32, y: 11, r: 3 }
  const srvB = { x: 32, y: 25, r: 3 }

  const pointTowards = (ax: number, ay: number, tx: number, ty: number, d: number) => {
    const dx = tx - ax
    const dy = ty - ay
    const len = Math.hypot(dx, dy) || 1
    return { x: ax + (dx / len) * d, y: ay + (dy / len) * d }
  }

  // No overlap: lines end exactly at circle edges
  const trunkStart = pointTowards(client.x, client.y, proxy.x, proxy.y, client.r)
  const trunkEnd = pointTowards(proxy.x, proxy.y, client.x, client.y, proxy.r)

  const aStart = pointTowards(proxy.x, proxy.y, srvA.x, srvA.y, proxy.r)
  const aEnd = pointTowards(srvA.x, srvA.y, proxy.x, proxy.y, srvA.r)

  const bStart = pointTowards(proxy.x, proxy.y, srvB.x, srvB.y, proxy.r)
  const bEnd = pointTowards(srvB.x, srvB.y, proxy.x, proxy.y, srvB.r)

  return (
    <svg viewBox="0 0 36 36" className={className} aria-hidden>
      <defs>
        <linearGradient id="lm_g" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#60a5fa" />
          <stop offset="100%" stopColor="#34d399" />
        </linearGradient>
      </defs>
      {/* Connections underneath; no overlap onto circles */}
      <path
        d={`M ${trunkStart.x} ${trunkStart.y} L ${trunkEnd.x} ${trunkEnd.y}`}
        fill="none"
        stroke="#7dd3fc"
        strokeWidth="3.0"
        strokeLinecap="butt"
        strokeLinejoin="round"
        strokeOpacity="0.95"
      />
      <path
        d={`M ${aStart.x} ${aStart.y} L ${aEnd.x} ${aEnd.y}`}
        fill="none"
        stroke="url(#lm_g)"
        strokeWidth="2.6"
        strokeLinecap="butt"
        strokeLinejoin="round"
      />
      <path
        d={`M ${bStart.x} ${bStart.y} L ${bEnd.x} ${bEnd.y}`}
        fill="none"
        stroke="url(#lm_g)"
        strokeWidth="2.6"
        strokeLinecap="butt"
        strokeLinejoin="round"
      />

      {/* Nodes on top */}
      <circle
        cx={proxy.x}
        cy={proxy.y}
        r={proxy.r}
        fill="#2563eb"
        stroke="#60a5fa"
        strokeOpacity=".35"
      />
      <circle cx={client.x} cy={client.y} r={client.r} fill="#93c5fd" />
      <circle cx={srvA.x} cy={srvA.y} r={srvA.r} fill="#22c55e" />
      <circle cx={srvB.x} cy={srvB.y} r={srvB.r} fill="#22c55e" />
    </svg>
  )
}


