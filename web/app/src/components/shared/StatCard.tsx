import type { ReactNode } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

const iconWrapStyles = cva(
  'flex size-11 shrink-0 items-center justify-center rounded-xl [&>svg]:size-5',
  {
    variants: {
      variant: {
        primary: 'bg-primary/10 text-primary',
        success: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
        warning: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
        info: 'bg-sky-500/10 text-sky-600 dark:text-sky-400',
        rose: 'bg-rose-500/10 text-rose-600 dark:text-rose-400',
      },
    },
    defaultVariants: { variant: 'primary' },
  },
)

export type StatCardVariant = NonNullable<VariantProps<typeof iconWrapStyles>['variant']>

export function StatCard({
  title,
  value,
  icon,
  hint,
  loading,
  variant,
}: {
  title: string
  value: string
  icon?: ReactNode
  hint?: string
  loading?: boolean
  variant?: StatCardVariant
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
          {hint ? <p className="text-xs text-muted-foreground/80">{hint}</p> : null}
        </div>
        {icon ? (
          <div className={cn(iconWrapStyles({ variant }))}>
            <span>{icon}</span>
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}
