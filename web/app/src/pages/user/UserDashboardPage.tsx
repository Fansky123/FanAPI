import { CreditCardIcon, SparklesIcon, TrendingUpIcon } from 'lucide-react'

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
      {error ? (
        <Alert variant="destructive">
          <AlertDescription className="flex items-center justify-between">
            <span>{error}</span>
            <Button size="sm" variant="outline" onClick={reload}>重试</Button>
          </AlertDescription>
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
          icon={<CreditCardIcon />}
          loading={loading}
        />
        <StatCard
          title="累计消耗积分"
          value={formatCredits(data.totalConsumed)}
          icon={<TrendingUpIcon />}
          loading={loading}
        />
        <StatCard
          title="今日消耗积分"
          value={formatCredits(data.todayConsumed)}
          icon={<SparklesIcon />}
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
            <p>3. 使用「文本对话」或「图片生成」体验实际调用效果。</p>
            <p>4. 在「积分充值」查看每次调用的积分扣减明细。</p>
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
      {(settings.contactInfo || settings.qqGroupUrl || settings.wechatCsUrl) && (
        <div className="grid gap-4 xl:grid-cols-[1fr_auto]">
          {settings.contactInfo && (
            <Card>
              <CardHeader>
                <CardTitle>联系方式</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-col gap-2 text-sm text-muted-foreground">
                {settings.contactInfo.split('\n').filter(Boolean).map((line, i) => (
                  <p key={i}>{line}</p>
                ))}
              </CardContent>
            </Card>
          )}
          {(settings.qqGroupUrl || settings.wechatCsUrl) && (
            <Card>
              <CardHeader>
                <CardTitle>扫码联系</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-wrap gap-4">
                {settings.qqGroupUrl && (
                  <div className="flex flex-col items-center gap-1">
                    <img src={settings.qqGroupUrl} alt="QQ 交流群" className="h-48 w-48 rounded-lg border object-contain p-1" />
                    <span className="text-xs text-muted-foreground">QQ 交流群</span>
                  </div>
                )}
                {settings.wechatCsUrl && (
                  <div className="flex flex-col items-center gap-1">
                    <img src={settings.wechatCsUrl} alt="微信客服" className="h-48 w-48 rounded-lg border object-contain p-1" />
                    <span className="text-xs text-muted-foreground">微信客服</span>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </>
  )
}
