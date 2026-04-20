import { useEffect, useState } from 'react'
import { MoonIcon, SunIcon } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { isDarkModeEnabled, setDarkMode } from '@/lib/auth/storage'

export function ThemeToggle() {
  const [isDark, setIsDark] = useState(isDarkModeEnabled())

  useEffect(() => {
    document.documentElement.classList.toggle('dark', isDark)
  }, [isDark])

  return (
    <Button
      variant="outline"
      size="icon-sm"
      onClick={() => {
        const next = !isDark
        setIsDark(next)
        setDarkMode(next)
      }}
    >
      {isDark ? <SunIcon /> : <MoonIcon />}
      <span className="sr-only">Toggle theme</span>
    </Button>
  )
}
