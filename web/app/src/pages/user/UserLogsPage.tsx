import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { userApi, type UserLog } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function UserLogsPage() {
  const { data: rows, loading, error, reload } = useAsync(async () => {
    const response = await userApi.listLogs()
    return Array.isArray(response) ? response : response.items ?? response.logs ?? []
  }, [] as UserLog[])

  return (
    <>
      <PageHeader
        eyebrow="Observability"
        title="调用日志"
        description="查看所有 API 调用记录，包含模型、消耗和状态信息。"
        actions={
          error ? (
            <Button size="sm" variant="outline" onClick={reload}>
              重试
            </Button>
          ) : null
        }
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
              <TableHead>模型</TableHead>
              <TableHead>相关 ID</TableHead>
              <TableHead>消耗</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>时间</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={6} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="py-10 text-center text-muted-foreground">
                    暂无调用日志
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
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
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>
    </>
  )
}
