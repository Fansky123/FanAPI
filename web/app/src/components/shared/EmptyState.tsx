import type { ReactNode } from 'react'

import { Card, CardContent } from '@/components/ui/card'

export function EmptyState({
  title,
  description,
  icon,
}: {
  title: string
  description: string
  icon?: ReactNode
}) {
  return (
    <Card className="border-dashed">
      <CardContent className="flex min-h-48 flex-col items-center justify-center gap-3 text-center">
        {icon}
        <div className="space-y-1">
          <h2 className="text-base font-medium">{title}</h2>
          <p className="max-w-md text-sm text-muted-foreground">{description}</p>
        </div>
      </CardContent>
    </Card>
  )
}
