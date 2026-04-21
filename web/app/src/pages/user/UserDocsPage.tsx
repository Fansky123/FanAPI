import { useEffect, useRef } from 'react'

const scalarConfiguration = JSON.stringify({
  theme: 'default',
  darkMode: false,
  layout: 'sidebar',
  hideDarkModeToggle: true,
})

export function UserDocsPage() {
  const scalarRootRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const root = scalarRootRef.current
    if (!root) {
      return undefined
    }

    const reference = document.createElement('div')
    reference.id = 'api-reference'
    reference.setAttribute('data-url', '/openapi.json')
    reference.setAttribute('data-configuration', scalarConfiguration)
    root.replaceChildren(reference)

    const script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/@scalar/api-reference'
    script.async = true
    document.head.appendChild(script)

    return () => {
      script.remove()
      root.replaceChildren()
    }
  }, [])

  return (
    <div className="overflow-hidden rounded-[28px] border border-border/70 bg-background shadow-sm">
      <div
        ref={scalarRootRef}
        className="min-h-[calc(100vh-8rem)]"
        data-testid="scalar-root"
      />
    </div>
  )
}
