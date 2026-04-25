import type { ComponentType } from 'react'
import { InboxIcon } from 'lucide-react'

import { TableCell, TableRow } from '@/components/ui/table'

export function TableEmpty({
  cols,
  title = '暂无数据',
  description,
  Icon = InboxIcon,
}: {
  cols: number
  title?: string
  description?: string
  Icon?: ComponentType<{ className?: string }>
}) {
  return (
    <TableRow>
      <TableCell colSpan={cols} className="py-12">
        <div className="flex flex-col items-center justify-center gap-2 text-center">
          <Icon className="size-10 text-muted-foreground/40" />
          <p className="text-sm font-medium text-foreground">{title}</p>
          {description ? (
            <p className="max-w-sm text-xs text-muted-foreground">{description}</p>
          ) : null}
        </div>
      </TableCell>
    </TableRow>
  )
}
