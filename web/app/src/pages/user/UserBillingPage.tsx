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
import { userApi, type UserTransaction } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function UserBillingPage() {
  const { data: rows, loading, error, reload } = useAsync(async () => {
    const response = await userApi.getTransactions()
    return Array.isArray(response) ? response : response.items ?? response.transactions ?? []
  }, [] as UserTransaction[])

  return (
    <>
      <PageHeader
        eyebrow="Finance"
        title="账单与流水"
        description="查看所有积分充值与消耗流水记录。"
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
                    <TableCell>{row.created_at ?? row.time ?? '-'}</TableCell>
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
