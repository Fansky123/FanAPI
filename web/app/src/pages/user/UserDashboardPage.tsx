import { BotIcon, CreditCardIcon, SparklesIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { userApi } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'
import { useSiteSettings } from '@/hooks/use-site-settings'

export function UserDashboardPage() {
  const { data, loading, error, reload } = useAsync(async () => {
    const [balance, stats] = await Promise.all([userApi.getBalance(), userApi.getStats()])
    return {
      balance: balance.balance_credits ?? 0,
      totalConsumed: stats.total_consumed ?? 0,
      todayConsumed: stats.today_consumed ?? 0,
    }
  }, { balance: 0, totalConsumed: 0, todayConsumed: 0 })

  const { settings } = useSiteSettings()

  return (
    <>
      <PageHeader
        eyebrow="Overview"
        title="用户数据看板"
        description="账户余额与消费概览，以及平台快速操作入口。"
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
      {settings.noticeTitle && (
        <Alert>
          <AlertDescription>
            <strong>{settings.noticeTitle}</strong>
            {settings.noticeContent && (
              <div className="mt-1 whitespace-pre-wrap text-sm">{settings.noticeContent}</div>
            )}
          </AlertDescription>
        </Alert>
      )}
      <div className="grid gap-4 xl:grid-cols-3">
        <StatCard
          title="剩余积分"
          value={formatCredits(data.balance)}
          icon={<CreditCardIcon className="size-4" />}
          hint="可用余额"
          loading={loading}
        />
        <StatCard
          title="累计消耗"
          value={formatCredits(data.totalConsumed)}
          icon={<BotIcon className="size-4" />}
          hint="历史总消耗"
          loading={loading}
        />
        <StatCard
          title="今日消耗"
          value={formatCredits(data.todayConsumed)}
          icon={<SparklesIcon className="size-4" />}
          hint="当天调用消耗"
          loading={loading}
        />
      </div>
      <div className="grid gap-4 xl:grid-cols-[1.3fr_0.7fr]">
        <Card>
          <CardHeader>
            <CardTitle>快速开始</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
            <p>1. 前往「API 密钥」页面创建一个密钥。</p>
            <p>2. 进入「模型列表」查看可用渠道与路由键。</p>
            <p>3. 使用「对话测试」或「图片生成」体验实际调用效果。</p>
            <p>4. 在「账单流水」查看每次调用的积分扣减明细。</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>余额不足？</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-3 text-sm text-muted-foreground">
            <p>前往「兑换中心」使用卡密充值，或通过邀请好友获取返佣积分。</p>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
