import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { userApi, type UserStatsResponse } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function UserStatsPage() {
  const { data: stats, loading, error, reload } = useAsync(
    () => userApi.getStats(),
    {} as UserStatsResponse,
  )

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
        <StatCard title="统计周期" value="最近 7 天" hint="明细数据覆盖范围" loading={loading} />
      </div>
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
                <span className="font-medium">{formatCredits(item.credits ?? 0)}</span>
              </div>
            ))
          )}
        </CardContent>
      </Card>
    </>
  )
}
