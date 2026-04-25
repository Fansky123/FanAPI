import { useState } from 'react'
import { Link } from 'react-router-dom'
import {
  ArrowRightIcon,
  BlocksIcon,
  CheckIcon,
  CopyIcon,
  CreditCardIcon,
  KeySquareIcon,
  MessageSquareIcon,
  ImageIcon,
  ShoppingCartIcon,
  SparklesIcon,
  TrendingUpIcon,
} from 'lucide-react'
import { toast } from 'sonner'

import { StatCard } from '@/components/shared/StatCard'
import { TrendChart, DualTrendChart } from '@/components/shared/TrendChart'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { userApi, type UserStatsResponse } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'
import { useSiteSettings } from '@/hooks/use-site-settings'

type GuideStep = {
  num: number
  text: string
  to: string
  Icon: React.ComponentType<{ className?: string }>
}

const guideSteps: GuideStep[] = [
  { num: 1, text: '点击左侧【API 密钥】创建密钥', to: '/keys', Icon: KeySquareIcon },
  { num: 2, text: '点击左侧【模型列表】查看模型 ID 和接口调用地址', to: '/models', Icon: BlocksIcon },
  { num: 3, text: '点击左侧【文本对话】在线体验所有 AI 聊天模型', to: '/playground', Icon: MessageSquareIcon },
  { num: 4, text: '点击左侧【图片生成】在线体验所有图片生成模型', to: '/image-gen', Icon: ImageIcon },
  { num: 5, text: '点击左侧【积分充值】充值积分', to: '/billing', Icon: ShoppingCartIcon },
]

export function UserDashboardPage() {
  const { data, loading, error, reload } = useAsync(async () => {
    const [balance, stats] = await Promise.all([userApi.getBalance(), userApi.getStats()])
    return {
      balance: balance.balance_credits ?? 0,
      totalConsumed: stats.total_consumed ?? 0,
      todayConsumed: stats.today_consumed ?? 0,
      stats,
    }
  }, { balance: 0, totalConsumed: 0, todayConsumed: 0, stats: {} as UserStatsResponse })

  const { settings } = useSiteSettings()
  const [copied, setCopied] = useState(false)
  const apiBase = typeof window !== 'undefined' ? window.location.origin : ''

  function copyApiBase() {
    if (!apiBase) return
    navigator.clipboard.writeText(apiBase).then(
      () => {
        setCopied(true)
        toast.success('已复制 API 网关地址')
        window.setTimeout(() => setCopied(false), 2000)
      },
      () => toast.error('复制失败，请手动复制'),
    )
  }

  const trends = buildTrends(data.stats)

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
          variant="info"
        />
        <StatCard
          title="累计消耗积分"
          value={formatCredits(data.totalConsumed)}
          icon={<TrendingUpIcon />}
          loading={loading}
          variant="success"
        />
        <StatCard
          title="今日消耗积分"
          value={formatCredits(data.todayConsumed)}
          icon={<SparklesIcon />}
          loading={loading}
          variant="warning"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>快速开始</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          {guideSteps.map((step) => (
            <Link
              key={step.num}
              to={step.to}
              className="group flex items-center gap-3 rounded-lg border border-border/60 px-3 py-2.5 text-sm transition hover:border-primary/40 hover:bg-primary/5"
            >
              <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
                {step.num}
              </span>
              <step.Icon className="size-4 text-muted-foreground" />
              <span className="flex-1 text-foreground">{step.text}</span>
              <span className="hidden items-center gap-1 text-xs text-primary opacity-0 transition group-hover:opacity-100 sm:flex">
                立即前往 <ArrowRightIcon className="size-3" />
              </span>
            </Link>
          ))}
        </CardContent>
      </Card>

      <Alert>
        <AlertDescription className="flex flex-wrap items-center gap-2">
          <span className="text-sm">本站大模型接口网关：</span>
          <code className="rounded bg-muted px-2 py-1 font-mono text-xs">{apiBase || '—'}</code>
          <Button size="sm" variant="ghost" onClick={copyApiBase} disabled={!apiBase}>
            {copied ? <CheckIcon className="size-3.5" /> : <CopyIcon className="size-3.5" />}
            <span className="ml-1">{copied ? '已复制' : '复制'}</span>
          </Button>
          <span className="text-xs text-muted-foreground">
            将模型基址替换为该网关，完全兼容 OpenAI 协议。
          </span>
        </AlertDescription>
      </Alert>

      {loading ? (
        <div className="grid gap-4 xl:grid-cols-2">
          <Card><CardContent className="p-6"><Skeleton className="h-64 w-full" /></CardContent></Card>
          <Card><CardContent className="p-6"><Skeleton className="h-64 w-full" /></CardContent></Card>
        </div>
      ) : trends.creditsTrend.length > 0 ? (
        <div className="grid gap-4 xl:grid-cols-2">
          <TrendChart
            title="积分消耗趋势（最近 7 天）"
            points={trends.creditsTrend}
            color="#2563eb"
            formatValue={(value) => `${value.toFixed(2)} 积分`}
          />
          <DualTrendChart title="请求次数统计（最近 7 天）" points={trends.requestTrend} />
        </div>
      ) : null}

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

function buildTrends(stats: UserStatsResponse) {
  const days: string[] = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date()
    d.setDate(d.getDate() - i)
    days.push(`${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`)
  }
  const dailyCredits = stats.daily_credits ?? []
  const dailyRequests = stats.daily_requests ?? []
  const creditsTrend = days.map((label) => {
    const entry = dailyCredits.find((row) => row.day === label)
    return { label, value: (entry?.credits ?? 0) / 1e6 }
  })
  const requestTrend = days.map((label) => {
    const entry = dailyRequests.find((row) => row.day === label)
    return { label, success: entry?.success ?? 0, failed: entry?.failed ?? 0 }
  })
  const hasAny = creditsTrend.some((p) => p.value > 0) || requestTrend.some((p) => p.success + p.failed > 0)
  return hasAny ? { creditsTrend, requestTrend } : { creditsTrend: [], requestTrend: [] }
}
