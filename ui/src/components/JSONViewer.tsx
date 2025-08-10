export function JSONViewer({ value }: { value: any }) {
  let text = ''
  try { text = JSON.stringify(value, null, 2) } catch { text = String(value || '') }
  return (
    <pre className="text-xs bg-black/30 border border-white/10 rounded p-3 overflow-auto max-h-64">
      {text}
    </pre>
  )
}


