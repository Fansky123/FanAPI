import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { userApi, type UserStatsResponse } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

function buildDailyTable(stats: UserStatsResponse) {
  const days: string[] = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    const label = `${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    days.push(label)
  }
  return days.map((label) => {
    const creditsEntry = (stats.daily_credits ?? []).find((r) => r.day === label)
    const reqEntry = (stats.daily_requests ?? []).find((r) => r.day === label)
    const success = reqEntry?.success ?? 0
    const failed = reqEntry?.failed ?? 0
    const total = success + failed
    const rate = total > 0 ? Math.round((success / total) * 100) : 100
    return { label, credits: creditsEntry?.credits ?? 0, success, failed, total, rate }
  })
}

export function UserStatsPage() {
  const { data: stats, loading, error, reload } = useAsync(
    () => userApi.getStats(),
    {} as UserStatsResponse,
  )

  const daily = buildDailyTable(stats)
  const totalRequests = daily.reduce((s, r) => s + r.total, 0)

  return (
    <>
      <PageHeader
        eyebrow="Metrics"
        title="使用统计"
        description="查看积分消耗趋势与最近 7 天的调用统计。"
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
      <div className="grid gap-4 md:grid-cols-3">
        <StatCard
          title="累计消耗积分"
          value={formatCredits(stats.total_consumed ?? 0)}
          loading={loading}
        />
        <StatCard
          title="今日消耗积分"
          value={formatCredits(stats.today_consumed ?? 0)}
          loading={loading}
        />
        <StatCard
          title="累计请求次数"
          value={String(totalRequests)}
          hint="最近 7 天"
          loading={loading}
        />
      </div>

      {/* 积分消耗列表 */}
      <Card>
        <CardHeader>
          <CardTitle>最近 7 天积分消耗</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {loading ? (
            Array.from({ length: 7 }).map((_, index) => (
              <Skeleton key={index} className="h-12 w-full" />
            ))
          ) : (stats.daily_credits ?? []).length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无统计数据。</p>
          ) : (
            (stats.daily_credits ?? []).map((item, index: number) => (
              <div
                key={`${item.day ?? index}`}
                className="flex items-center justify-between rounded-xl border border-border/70 bg-muted/20 px-4 py-3 text-sm"
              >
                <span>{item.day ?? '-'}</span>
                <span className="font-medium text-blue-600">{formatCredits(item.credits ?? 0)}</span>
              </div>
            ))
          )}
        </CardContent>
      </Card>

      {/* 每日明细表 */}
      <Card>
        <CardHeader>
          <CardTitle>每日请求明细</CardTitle>
        </CardHeader>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>日期</TableHead>
              <TableHead className="text-right">消耗积分</TableHead>
              <TableHead className="text-right">成功请求</TableHead>
              <TableHead className="text-right">失败请求</TableHead>
              <TableHead>成功率</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 7 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 5 }).map((_, j) => (
                    <TableCell key={j}><Skeleton className="h-4 w-20" /></TableCell>
                  ))}
                </TableRow>
              ))
            ) : daily.map((row) => (
              <TableRow key={row.label}>
                <TableCell>{row.label}</TableCell>
                <TableCell className="text-right font-semibold text-blue-600">
                  {formatCredits(row.credits)}
                </TableCell>
                <TableCell className="text-right">
                  {row.success > 0 ? (
                    <Badge className="bg-emerald-600 hover:bg-emerald-600 text-white">{row.success}</Badge>
                  ) : <span className="text-muted-foreground">0</span>}
                </TableCell>
                <TableCell className="text-right">
                  {row.failed > 0 ? (
                    <Badge variant="destructive">{row.failed}</Badge>
                  ) : <span className="text-muted-foreground">0</span>}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <div className="h-2 w-24 rounded-full bg-muted overflow-hidden">
                      <div
                        className={`h-full rounded-full ${row.rate >= 90 ? 'bg-emerald-500' : row.rate >= 70 ? 'bg-yellow-500' : 'bg-red-500'}`}
                        style={{ width: `${row.rate}%` }}
                      />
                    </div>
                    <span className="text-xs text-muted-foreground">{row.rate}%</span>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </>
  )
}

