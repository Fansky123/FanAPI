import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type UserStatsResponse } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'

export function UserStatsPage() {
  const [stats, setStats] = useState<UserStatsResponse>({})
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await userApi.getStats()
        setStats(response)
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="Metrics"
        title="使用统计"
        description="用统一卡片和明细区块承接统计视图，替代旧版自绘图表的碎片化结构。"
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <div className="grid gap-4 md:grid-cols-3">
        <StatCard title="累计消耗积分" value={formatCredits(stats.total_consumed ?? 0)} />
        <StatCard title="今日消耗积分" value={formatCredits(stats.today_consumed ?? 0)} />
        <StatCard title="统计说明" value="7 天" hint="当前明细覆盖最近 7 天" />
      </div>
      <Card>
        <CardHeader>
          <CardTitle>最近 7 天积分消耗</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {(stats.daily_credits ?? []).length === 0 ? (
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
