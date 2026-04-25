import { useState } from 'react'
import { BotIcon } from 'lucide-react'

import { cn } from '@/lib/utils'

export function AppLogo({
  siteName = 'FanAPI',
  logoUrl,
  label,
  className,
}: {
  siteName?: string
  logoUrl?: string
  label: string
  className?: string
}) {
  const [erroredUrl, setErroredUrl] = useState<string | undefined>(undefined)
  const initial = (siteName || label || '?').slice(0, 1).toUpperCase()
  const errored = Boolean(logoUrl) && erroredUrl === logoUrl
  const showImage = Boolean(logoUrl) && !errored

  return (
    <div className={cn('flex items-center gap-3', className)}>
      {showImage ? (
        <div className="flex size-10 items-center justify-center overflow-hidden rounded-xl border border-border/70 bg-card shadow-sm">
          <img
            className="size-full object-contain"
            src={logoUrl}
            alt={`${siteName} logo`}
            onError={() => setErroredUrl(logoUrl)}
          />
        </div>
      ) : logoUrl && errored ? (
        <div className="flex size-10 items-center justify-center rounded-xl bg-primary text-base font-semibold text-primary-foreground shadow-sm">
          {initial}
        </div>
      ) : (
        <div className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
          <BotIcon className="size-5" />
        </div>
      )}
      <div className="min-w-0">
        <p className="truncate text-sm font-medium uppercase tracking-[0.18em] text-muted-foreground">
          {siteName}
        </p>
        <p className="truncate text-base font-semibold text-foreground">{label}</p>
      </div>
    </div>
  )
}
