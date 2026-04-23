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

type MetricPoint = {
  label: string
  value: number
}

function TrendChart({
  title,
  points,
  color,
  formatValue,
}: {
  title: string
  points: MetricPoint[]
  color: string
  formatValue: (value: number) => string
}) {
  const values = points.map((point) => point.value)
  const max = Math.max(...values, 1)
  const width = 100
  const height = 44
  const step = points.length > 1 ? width / (points.length - 1) : width
  const path = points.map((point, index) => {
    const x = index * step
    const y = height - (point.value / max) * height
    return `${index === 0 ? 'M' : 'L'} ${x.toFixed(2)} ${y.toFixed(2)}`
  }).join(' ')

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="h-48 rounded-xl border border-border/70 bg-muted/20 p-4">
          <svg viewBox={`0 0 ${width} ${height}`} className="h-full w-full overflow-visible" preserveAspectRatio="none">
            {Array.from({ length: 4 }).map((_, index) => {
              const y = (height / 3) * index
              return (
                <line
                  key={index}
                  x1="0"
                  y1={y}
                  x2={width}
                  y2={y}
                  stroke="currentColor"
                  className="text-border/70"
                  strokeDasharray="2 2"
                />
              )
            })}
            <path d={path} fill="none" stroke={color} strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
            {points.map((point, index) => {
              const x = index * step
              const y = height - (point.value / max) * height
              return <circle key={point.label} cx={x} cy={y} r="1.8" fill={color} />
            })}
          </svg>
        </div>
        <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
          {points.map((point) => (
            <div key={point.label} className="rounded-lg border border-border/70 bg-background px-3 py-2">
              <div className="text-xs text-muted-foreground">{point.label}</div>
              <div className="mt-1 text-sm font-medium">{formatValue(point.value)}</div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

function DualTrendChart({
  title,
  points,
}: {
  title: string
  points: Array<{ label: string; success: number; failed: number }>
}) {
  const max = Math.max(...points.flatMap((point) => [point.success, point.failed]), 1)
  const width = 100
  const height = 44
  const step = points.length > 1 ? width / (points.length - 1) : width

  const buildPath = (selector: (point: { success: number; failed: number }) => number) => (
    points.map((point, index) => {
      const x = index * step
      const value = selector(point)
      const y = height - (value / max) * height
      return `${index === 0 ? 'M' : 'L'} ${x.toFixed(2)} ${y.toFixed(2)}`
    }).join(' ')
  )

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <div className="flex items-center gap-2"><span className="size-2 rounded-full bg-blue-500" />成功请求</div>
          <div className="flex items-center gap-2"><span className="size-2 rounded-full bg-orange-500" />失败请求</div>
        </div>
        <div className="h-48 rounded-xl border border-border/70 bg-muted/20 p-4">
          <svg viewBox={`0 0 ${width} ${height}`} className="h-full w-full overflow-visible" preserveAspectRatio="none">
            {Array.from({ length: 4 }).map((_, index) => {
              const y = (height / 3) * index
              return (
                <line
                  key={index}
                  x1="0"
                  y1={y}
                  x2={width}
                  y2={y}
                  stroke="currentColor"
                  className="text-border/70"
                  strokeDasharray="2 2"
                />
              )
            })}
            <path d={buildPath((point) => point.success)} fill="none" stroke="#3b82f6" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
            <path d={buildPath((point) => point.failed)} fill="none" stroke="#f97316" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
        <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
          {points.map((point) => (
            <div key={point.label} className="rounded-lg border border-border/70 bg-background px-3 py-2 text-sm">
              <div className="text-xs text-muted-foreground">{point.label}</div>
              <div className="mt-1 flex items-center gap-3">
                <span className="text-blue-600">成功 {point.success}</span>
                <span className="text-orange-600">失败 {point.failed}</span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

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
  const creditsTrend = daily.map((row) => ({ label: row.label, value: row.credits / 1e6 }))
  const requestTrend = daily.map((row) => ({ label: row.label, success: row.success, failed: row.failed }))

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

      {loading ? (
        <div className="grid gap-4 xl:grid-cols-2">
          <Card><CardContent className="p-6"><Skeleton className="h-64 w-full" /></CardContent></Card>
          <Card><CardContent className="p-6"><Skeleton className="h-64 w-full" /></CardContent></Card>
        </div>
      ) : (stats.daily_credits ?? []).length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            暂无最近 7 天统计数据。
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 xl:grid-cols-2">
          <TrendChart
            title="积分消耗趋势（最近 7 天）"
            points={creditsTrend}
            color="#2563eb"
            formatValue={(value) => `${value.toFixed(2)} 积分`}
          />
          <DualTrendChart title="请求次数统计（最近 7 天）" points={requestTrend} />
        </div>
      )}

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

