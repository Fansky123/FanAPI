import { ActivityIcon, BadgeDollarSignIcon, UsersIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { adminApi, type AdminStatsResponse } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

export function AdminDashboardPage() {
  const { data: stats, loading, error, reload } = useAsync(
    () => adminApi.getStats(),
    {} as AdminStatsResponse,
  )

  return (
    <>
      <PageHeader
        eyebrow="Operations"
        title="平台运营看板"
        description="平台核心运营指标：用户数、请求量与收入概览。"
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
      <div className="grid gap-4 xl:grid-cols-3">
        <StatCard
          title="总用户数"
          value={String(stats.total_users ?? stats.users ?? 0)}
          icon={<UsersIcon className="size-4" />}
          loading={loading}
        />
        <StatCard
          title="总请求数"
          value={String(stats.total_requests ?? stats.requests ?? 0)}
          icon={<ActivityIcon className="size-4" />}
          loading={loading}
        />
        <StatCard
          title="总收入"
          value={String(stats.total_revenue ?? stats.revenue ?? 0)}
          icon={<BadgeDollarSignIcon className="size-4" />}
          loading={loading}
        />
      </div>
    </>
  )
}
