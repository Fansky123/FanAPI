import { ActivityIcon, BadgeDollarSignIcon, UsersIcon, ZapIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { adminApi, type AdminStatsResponse } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

function fmtCredits(v: number | undefined) {
  if (v == null) return '--'
  return (v / 1_000_000).toFixed(4)
}

function profitColor(v: number | undefined) {
  if (v == null) return ''
  return v >= 0 ? 'text-emerald-600' : 'text-red-500'
}

export function AdminDashboardPage() {
  const { data: stats, loading, error, reload } = useAsync(
    () => adminApi.getStats(),
    {} as AdminStatsResponse,
  )

  const marginPct = (() => {
    const r = stats.total?.revenue
    const p = stats.total?.profit
    if (!r) return null
    return ((p ?? 0) / r * 100).toFixed(2)
  })()

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="平台运营看板"
        description="平台核心运营指标：渠道、用户、今日收入与累计利润。"
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

      {/* 第一行：核心指标 */}
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          title="渠道数量"
          value={stats.active_channels != null
            ? `${stats.active_channels} / ${stats.channels ?? '--'}`
            : String(stats.channels ?? '--')}
          hint="活跃 / 全部"
          icon={<ZapIcon className="size-4" />}
          loading={loading}
        />
        <StatCard
          title="用户数量"
          value={String(stats.total_users ?? stats.users ?? '--')}
          hint="普通用户数"
          icon={<UsersIcon className="size-4" />}
          loading={loading}
        />
        <StatCard
          title="今日收入"
          value={`¥${fmtCredits(stats.today?.revenue)}`}
          hint={`今日结算 ${stats.today?.count ?? 0} 笔`}
          icon={<BadgeDollarSignIcon className="size-4" />}
          loading={loading}
        />
        <StatCard
          title="今日利润"
          value={`¥${fmtCredits(stats.today?.profit)}`}
          hint="收入 - 上游成本"
          icon={<ActivityIcon className="size-4" />}
          loading={loading}
        />
      </div>

      {/* 第二行：累计数据 */}
      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">累计营收</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">¥{fmtCredits(stats.total?.revenue)}</p>
            <p className="mt-1 text-xs text-muted-foreground">历史全部结算</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">累计成本</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">¥{fmtCredits(stats.total?.cost)}</p>
            <p className="mt-1 text-xs text-muted-foreground">上游 API 消耗</p>
          </CardContent>
        </Card>
        <Card className={stats.total?.profit != null && stats.total.profit >= 0 ? 'border-l-4 border-l-emerald-500' : stats.total?.profit != null ? 'border-l-4 border-l-red-500' : ''}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">累计利润</CardTitle>
          </CardHeader>
          <CardContent>
            <p className={`text-2xl font-bold ${profitColor(stats.total?.profit)}`}>
              ¥{fmtCredits(stats.total?.profit)}
            </p>
            <p className="mt-1 text-xs text-muted-foreground">历史净利润（含今日）</p>
          </CardContent>
        </Card>
      </div>

      {/* 利润率 */}
      {marginPct !== null ? (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">综合利润率</CardTitle>
          </CardHeader>
          <CardContent>
            <p className={`text-3xl font-bold ${profitColor(parseFloat(marginPct))}`}>
              {marginPct}%
            </p>
            <p className="mt-1 text-xs text-muted-foreground">累计利润 ÷ 累计营收</p>
            <div className="mt-3 h-2 w-full rounded-full bg-muted overflow-hidden">
              <div
                className={`h-full rounded-full ${parseFloat(marginPct) >= 0 ? 'bg-emerald-500' : 'bg-red-500'}`}
                style={{ width: `${Math.min(Math.max(parseFloat(marginPct), 0), 100)}%` }}
              />
            </div>
          </CardContent>
        </Card>
      ) : null}
    </>
  )
}

