import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminTask } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

export function AdminTasksPage() {
  const [rows, setRows] = useState<AdminTask[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await adminApi.listTasks()
        setRows(Array.isArray(response) ? response : response.items ?? response.tasks ?? [])
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="任务中心"
        description="异步任务页已接入真实列表，后续继续补筛选、详情和管理动作。"
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
              <TableHead>类型</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>上游任务 ID</TableHead>
              <TableHead>创建时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.id ?? '-'}</TableCell>
                <TableCell>{row.type ?? '-'}</TableCell>
                <TableCell>
                  <Badge variant="secondary">{row.status ?? '-'}</Badge>
                </TableCell>
                <TableCell className="font-mono text-xs text-muted-foreground">
                  {row.upstream_task_id ?? '-'}
                </TableCell>
                <TableCell>{row.created_at ?? '-'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </>
  )
}
