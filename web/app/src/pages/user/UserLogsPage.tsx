import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type UserLog } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'

export function UserLogsPage() {
  const [rows, setRows] = useState<UserLog[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await userApi.listLogs()
        setRows(Array.isArray(response) ? response : response.items ?? response.logs ?? [])
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="Observability"
        title="调用日志"
        description="日志页先用清晰的可扫描表格表达关键信息，避免旧版一页塞太多细节。"
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>相关 ID</TableHead>
              <TableHead>消耗</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.id ?? '-'}</TableCell>
                <TableCell>{row.model ?? '-'}</TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {row.corr_id ?? '-'}
                </TableCell>
                <TableCell>{formatCredits(row.cost_credits ?? 0)}</TableCell>
                <TableCell>{row.status ?? '-'}</TableCell>
                <TableCell>{row.created_at ?? '-'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </>
  )
}
