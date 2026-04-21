import { useEffect, useState } from 'react'
import { BotIcon, CreditCardIcon, SparklesIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'

type DashboardState = {
  balance: number
  totalConsumed: number
  todayConsumed: number
}

export function UserDashboardPage() {
  const [state, setState] = useState<DashboardState>({
    balance: 0,
    totalConsumed: 0,
    todayConsumed: 0,
  })
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [balance, stats] = await Promise.all([
          userApi.getBalance(),
          userApi.getStats(),
        ])
        setState({
          balance: balance.balance_credits ?? 0,
          totalConsumed: stats.total_consumed ?? 0,
          todayConsumed: stats.today_consumed ?? 0,
        })
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="User overview"
        title="用户数据看板"
        description="这里优先呈现最关键的账户与消费信息，避免旧版首页的视觉噪音和信息堆叠。"
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <div className="grid gap-4 xl:grid-cols-3">
        <StatCard
          title="剩余积分"
          value={formatCredits(state.balance)}
          icon={<CreditCardIcon className="size-4" />}
          hint="单位已按 credits 转换为展示值"
        />
        <StatCard
          title="累计消耗"
          value={formatCredits(state.totalConsumed)}
          icon={<BotIcon className="size-4" />}
          hint="来自现有 /user/stats 接口"
        />
        <StatCard
          title="今日消耗"
          value={formatCredits(state.todayConsumed)}
          icon={<SparklesIcon className="size-4" />}
          hint="用于快速判断当天调用变化"
        />
      </div>
      <div className="grid gap-4 xl:grid-cols-[1.3fr_0.7fr]">
        <Card>
          <CardHeader>
            <CardTitle>重构后的首页原则</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm text-muted-foreground">
            <p>1. 先展示关键数据，不先展示装饰内容。</p>
            <p>2. 信息层级固定，后续新增模块不能打破首页节奏。</p>
            <p>3. 样式统一继承设计系统，不再页面单独做卡片视觉。</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>下一步建议</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm text-muted-foreground">
            <p>创建 API 密钥后，再查看模型列表与账单页。</p>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
