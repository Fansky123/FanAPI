import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { agentApi, type AgentUser } from '@/lib/api/agent'
import { useAsync } from '@/hooks/use-async'

export function AgentDashboardPage() {
  const { data: rows, loading, error, reload } = useAsync(async () => {
    const response = await agentApi.listUsers()
    return Array.isArray(response) ? response : response.items ?? response.users ?? []
  }, [] as AgentUser[])

  return (
    <>
      <PageHeader
        eyebrow="Agent"
        title="Agent 工作台"
        description="查看并管理您名下的用户。"
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
        <CardHeader>
          <CardTitle>可管理用户</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>用户名</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead>余额</TableHead>
              </TableRow>
            </TableHeader>
            {loading ? (
              <TableSkeleton cols={4} />
            ) : (
              <TableBody>
                {rows.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={4} className="py-10 text-center text-muted-foreground">
                      暂无用户数据
                    </TableCell>
                  </TableRow>
                ) : (
                  rows.map((row, index) => (
                    <TableRow key={row.id ?? index}>
                      <TableCell>{row.id ?? '-'}</TableCell>
                      <TableCell>{row.username ?? '-'}</TableCell>
                      <TableCell>{row.email ?? '-'}</TableCell>
                      <TableCell>{row.balance_credits ?? '-'}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            )}
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
