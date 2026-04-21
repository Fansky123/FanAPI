import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type UserTask } from '@/lib/api/user'

export function UserTasksPage() {
  const [rows, setRows] = useState<UserTask[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await userApi.listTasks()
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
        eyebrow="Jobs"
        title="任务中心"
        description="异步任务页已经接入真实列表数据，后续继续补筛选、详情和重试动作。"
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
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
