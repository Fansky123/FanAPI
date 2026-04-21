import type { ReactNode } from 'react'

import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

export function StatCard({
  title,
  value,
  icon,
  loading,
}: {
  title: string
  value: string
  icon?: ReactNode
  hint?: string
  loading?: boolean
}) {
  return (
    <Card className="border-border/60">
      <CardContent className="flex items-center justify-between pt-6">
        <div className="flex flex-col gap-1">
          {loading ? (
            <Skeleton className="h-8 w-28" />
          ) : (
            <p className="text-2xl font-semibold tracking-tight">{value}</p>
          )}
          <p className="text-sm text-muted-foreground">{title}</p>
        </div>
        {icon ? (
          <div className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <span className="[&>svg]:size-5">{icon}</span>
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}
