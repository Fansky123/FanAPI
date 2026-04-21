import { useEffect, useState } from 'react'
import { ActivityIcon, BadgeDollarSignIcon, UsersIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { adminApi, type AdminStatsResponse } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

export function AdminDashboardPage() {
  const [stats, setStats] = useState<AdminStatsResponse>({})
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await adminApi.getStats()
        setStats(response ?? {})
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="平台运营看板"
        description="用更稳定的后台卡片与内容节奏替代旧版堆叠式 dashboard。"
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <div className="grid gap-4 xl:grid-cols-3">
        <StatCard
          title="总用户数"
          value={String(stats.total_users ?? stats.users ?? 0)}
          icon={<UsersIcon className="size-4" />}
        />
        <StatCard
          title="总请求数"
          value={String(stats.total_requests ?? stats.requests ?? 0)}
          icon={<ActivityIcon className="size-4" />}
        />
        <StatCard
          title="总收入"
          value={String(stats.total_revenue ?? stats.revenue ?? 0)}
          icon={<BadgeDollarSignIcon className="size-4" />}
        />
      </div>
      <Card>
        <CardHeader>
          <CardTitle>后台重构目标</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
          <p>1. 列表页、筛选栏、详情和弹窗全部统一。</p>
          <p>2. 信息密度高，但不靠更小的字和更乱的卡片实现。</p>
          <p>3. 后续 agent/vendor 直接继承同一套模式。</p>
        </CardContent>
      </Card>
    </>
  )
}
