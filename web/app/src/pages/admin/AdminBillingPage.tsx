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
import { adminApi, type AdminTransaction } from '@/lib/api/admin'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function AdminBillingPage() {
  const { data: rows, loading, error, reload } = useAsync(async () => {
    const response = await adminApi.listTransactions()
    return Array.isArray(response) ? response : response.items ?? response.transactions ?? []
  }, [] as AdminTransaction[])

  return (
    <>
      <PageHeader
        eyebrow="Finance"
        title="后台账单流水"
        description="查看平台全量账单与积分流水记录。"
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
              <TableHead>时间</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>金额</TableHead>
              <TableHead>说明</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={4} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="py-10 text-center text-muted-foreground">
                    暂无账单记录
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell>{row.created_at ?? '-'}</TableCell>
                    <TableCell>{row.type ?? '-'}</TableCell>
                    <TableCell>{formatCredits(row.amount ?? row.credits ?? 0)}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {row.remark ?? row.description ?? '-'}
                    </TableCell>
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
