import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { PageHeader } from '@/components/shared/PageHeader'
import { TablePagination } from '@/components/shared/TablePagination'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { userApi } from '@/lib/api/user'
import { payApi } from '@/lib/api/pay'
import { useAsync } from '@/hooks/use-async'
import { useSiteSettings } from '@/hooks/use-site-settings'
import { useAuth } from '@/hooks/use-auth'
import { Badge } from '@/components/ui/badge'
import { formatCredits } from '@/lib/formatters/credits'
import { cn } from '@/lib/utils'
import { Check, Info, Loader2, RefreshCcw, Wallet } from 'lucide-react'
import { toast } from 'sonner'
import type { PaymentOrder } from '@/lib/api/user'

function cx(...classes: (string | undefined | null | false)[]) {
  return classes.filter(Boolean).join(' ')
}

export function UserBillingPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const defaultTab = searchParams.get('tab') || 'recharge'
  
  const { settings } = useSiteSettings()
  const { user } = useAuth()
  const { data: balanceCredits, reload: reloadBalance } = useAsync(async () => {
    const res = await userApi.getBalance()
    return res.balance_credits ?? 0
  }, 0)

  // Recharge State
  const [selectedPlan, setSelectedPlan] = useState<number>(-1)
  const [selectedAmount, setSelectedAmount] = useState<number | ''>('')
  const [payMethod, setPayMethod] = useState<'wechat' | 'alipay'>('wechat')
  const [isPaying, setIsPaying] = useState(false)
  const [payUrl, setPayUrl] = useState<string>('')
  const [currentOutTradeNo, setCurrentOutTradeNo] = useState<string>('')
  const [showPayFrame, setShowPayFrame] = useState(false)

  // Transaction State
  const [txPage, setTxPage] = useState(1)
  const [txTaskIdFilter, setTxTaskIdFilter] = useState('')
  const [txCorrIdFilter, setTxCorrIdFilter] = useState('')
  const { data: txData, reload: txReload } = useAsync(async () => {
    const res = await userApi.getTransactions(
      txPage,
      20,
      txTaskIdFilter || undefined,
      txCorrIdFilter || undefined,
    )
    return {
      items: Array.isArray(res) ? res : res.items ?? res.transactions ?? [],
      total: !Array.isArray(res) ? res.total ?? 0 : 0
    }
  }, { items: [], total: 0 } as { items: unknown[]; total: number }, [txPage, txTaskIdFilter, txCorrIdFilter])

  // Orders State
  const [orderPage, setOrderPage] = useState(1)
  const { data: orderData, reload: orderReload } = useAsync(async () => {
    const res = await userApi.getPaymentOrders(orderPage, 20)
    return {
      items: res.orders || [],
      total: res.total || 0
    }
  }, { items: [], total: 0 } as any, [orderPage])

  // Update URL on tab change
  const handleTabChange = (val: string) => {
    setSearchParams({ tab: val })
  }

  const handlePaymentSuccess = () => {
    toast.success('充值成功')
    setShowPayFrame(false)
    txReload()
    orderReload()
    reloadBalance()
  }

  // Payment logic
  const handlePay = async () => {
    if (!selectedAmount || Number(selectedAmount) < 0.01) {
      toast.error('请输入有效的充值金额')
      return
    }
    
    const amount = Number(selectedAmount)
    setIsPaying(true)
    try {
      if (settings.payApplyEnabled) {
        const payFlat = payMethod === 'wechat' ? 1 : 2
        const res = await payApi.createPayApplyOrder({ amount, pay_flat: payFlat, pay_from: 'pc' })
        if (res.pay_url) {
          setPayUrl(res.pay_url)
          setCurrentOutTradeNo(res.out_trade_no || "")
          setShowPayFrame(true)
          startPolling(res.out_trade_no || "")
        }
      } else if (settings.epayEnabled) {
        const type = payMethod === 'wechat' ? 'wxpay' : 'alipay'
        const res = await payApi.createEpayOrder(amount, type)
        if (res.pay_url) {
          window.open(res.pay_url, '_blank')
          setCurrentOutTradeNo(res.out_trade_no || "")
          setShowPayFrame(true) // Also show to ask user if paid
        }
      }
    } catch (e: any) {
      toast.error(e.message || '支付发起失败')
    } finally {
      setIsPaying(false)
    }
  }

  const startPolling = (outTradeNo: string) => {
    const timer = setInterval(async () => {
      try {
        const res = await payApi.getOrderStatus(outTradeNo)
        if (res.status === 'paid') {
          clearInterval(timer)
          handlePaymentSuccess()
        }
      } catch (e) {
        // ignore polling errors
      }
    }, 3000)

    // Cleanup timer after some time or when dialog unmounts - handled roughly here via a timeout
    setTimeout(() => clearInterval(timer), 300000) // 5 minutes max
  }

  const txTypeLabel = (type: string) => {
    const map: Record<string, string> = {
      recharge: '充值',
      consume: '消费',
      refund: '退款',
      invite_rebate: '邀请返利',
    }
    return map[type] || type
  }

  const txSign = (type: string) => (['consume'].includes(type) ? '-' : '+')
  const txAmtColor = (type: string) => (['consume'].includes(type) ? 'text-red-500' : 'text-green-600')

  const orderStatusLabel = (status: string) => {
    switch (status) {
      case 'pending': return '待支付'
      case 'paid': return '已完成'
      case 'failed': return '失败'
      default: return status ?? '未知'
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title="我的订单"
        description="管理您的积分余额、充值、以及查看流水账单。"
      />

      <Tabs value={defaultTab} onValueChange={handleTabChange} className="w-full">
        <TabsList className="mb-6">
          <TabsTrigger value="recharge">积分充值</TabsTrigger>
          <TabsTrigger value="transactions">余额流水</TabsTrigger>
          <TabsTrigger value="orders">订单记录</TabsTrigger>
        </TabsList>

        <TabsContent value="recharge" className="space-y-6">
          <div className="grid gap-6 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>当前积分</CardTitle>
                <CardDescription className="flex items-center gap-2">
                  <Info className="h-4 w-4" /> 积分永不过期，随时可用
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-baseline gap-2">
                  <span className="text-4xl font-bold tracking-tight">{formatCredits(balanceCredits)}</span>
                  <span className="text-muted-foreground">积分</span>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>账户信息</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">用户名</span>
                  <span className="font-medium">{user?.username || '—'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">分组</span>
                  <span className="font-medium">{user?.group || '默认'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">邮箱</span>
                  <span className="font-medium text-green-600">{user?.email || '未绑定'}</span>
                </div>
              </CardContent>
            </Card>
          </div>

          {(settings.epayEnabled || settings.payApplyEnabled) && (
            <Card>
              <CardHeader>
                <CardTitle>选择充值套餐</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-3">
                  {settings.plans?.map((plan, i) => (
                    <div
                      key={i}
                      onClick={() => {
                        setSelectedPlan(i)
                        setSelectedAmount(plan.amount)
                      }}
                      className={cn(
                        "relative cursor-pointer rounded-xl border-2 p-4 transition-all hover:border-primary/50",
                        selectedPlan === i ? "border-primary bg-primary/5 shadow-md" : "border-border"
                      )}
                    >
                      {selectedPlan === i && (
                        <div className="absolute top-2 right-2 rounded-full bg-primary p-1 text-primary-foreground">
                          <Check className="h-3 w-3" />
                        </div>
                      )}
                      <div className="mb-2 font-semibold">
                        {plan.credits} 积分
                        {plan.bonus ? <span className="text-xs text-orange-500"> (+{plan.bonus})</span> : ''}
                      </div>
                      <div className="flex items-baseline gap-1">
                        <span className="text-sm">￥</span>
                        <span className="text-2xl font-bold">{plan.amount}</span>
                        {plan.origin_amount && (
                          <span className="text-xs text-muted-foreground line-through ml-2">￥{plan.origin_amount}</span>
                        )}
                      </div>
                      {plan.desc && <div className="mt-2 text-xs text-muted-foreground">{plan.desc}</div>}
                    </div>
                  ))}
                  {(!settings.plans?.length || settings.allowCustom) && (
                    <div className="flex flex-col justify-center rounded-xl border-2 border-border p-4">
                      <div className="mb-2 font-semibold">自定义金额</div>
                      <div className="flex items-center gap-2">
                        <span>￥</span>
                        <Input
                          type="number"
                          value={selectedAmount}
                          onChange={(e) => {
                            setSelectedPlan(-1)
                            setSelectedAmount(e.target.value ? Number(e.target.value) : '')
                          }}
                          min={1}
                          max={10000}
                          placeholder="请输入金额"
                        />
                      </div>
                    </div>
                  )}
                </div>

                <div className="mt-8 flex flex-col items-center gap-6">
                  <div className="flex gap-4">
                    <Button
                      variant={payMethod === 'wechat' ? 'default' : 'outline'}
                      className={cx("h-12 w-32 border-2", payMethod === 'wechat' ? 'border-green-600 bg-green-50 text-green-700 hover:bg-green-100 hover:text-green-800' : '')}
                      onClick={() => setPayMethod('wechat')}
                    >
                      微信支付
                    </Button>
                    <Button
                      variant={payMethod === 'alipay' ? 'default' : 'outline'}
                      className={cx("h-12 w-32 border-2", payMethod === 'alipay' ? 'border-blue-600 bg-blue-50 text-blue-700 hover:bg-blue-100 hover:text-blue-800' : '')}
                      onClick={() => setPayMethod('alipay')}
                    >
                      支付宝
                    </Button>
                  </div>

                  <Button size="lg" className="w-64 rounded-full text-lg" onClick={handlePay} disabled={isPaying || !selectedAmount}>
                    {isPaying && <Loader2 className="mr-2 h-5 w-5 animate-spin" />}
                    立即支付 ￥{(Number(selectedAmount) || 0).toFixed(2)}
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="transactions">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-4">
              <CardTitle>流水明细</CardTitle>
              <div className="flex w-full max-w-2xl items-center gap-2">
                <Input
                  placeholder="按任务 ID 查询"
                  value={txTaskIdFilter}
                  onChange={(e) => setTxTaskIdFilter(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && setTxPage(1)}
                />
                <Input
                  placeholder="按 Corr ID 查询"
                  value={txCorrIdFilter}
                  onChange={(e) => setTxCorrIdFilter(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && setTxPage(1)}
                />
                <Button variant="secondary" onClick={() => setTxPage(1)}>查询</Button>
                <Button variant="outline" size="icon" onClick={() => txReload()}><RefreshCcw className="h-4 w-4" /></Button>
              </div>
            </CardHeader>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>类型</TableHead>
                  <TableHead>积分变动</TableHead>
                  <TableHead>操作后余额</TableHead>
                  <TableHead>关联任务</TableHead>
                  <TableHead>时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {!txData?.items?.length ? (
                  <TableRow><TableCell colSpan={5} className="text-center py-6 text-muted-foreground">暂无流水记录</TableCell></TableRow>
                ) : (
                  txData.items.map((row: any) => (
                    <TableRow key={row.id}>
                      <TableCell><Badge variant="outline">{txTypeLabel(row.type)}</Badge></TableCell>
                      <TableCell className={cn("font-medium", txAmtColor(row.type))}>
                        {txSign(row.type)} {formatCredits(Math.abs(row.credits || row.amount || 0))} 积分
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {row.balance_after != null ? `${formatCredits(row.balance_after)} 积分` : '—'}
                      </TableCell>
                      <TableCell className="font-mono text-xs text-blue-500">
                        {row.metrics?.task_id ? `#${row.metrics.task_id}` : '—'}
                      </TableCell>
                      <TableCell>{row.created_at}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
            {txData?.total > 0 && (
              <div className="p-4 border-t">
                <TablePagination current={txPage} total={txData.total} pageSize={20} onChange={setTxPage} />
              </div>
            )}
          </Card>
        </TabsContent>

        <TabsContent value="orders">
          <Card>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>订单号</TableHead>
                  <TableHead>充值金额</TableHead>
                  <TableHead>到账积分</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>支付时间</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {!orderData?.items?.length ? (
                  <TableRow><TableCell colSpan={5} className="text-center py-6 text-muted-foreground">暂无订单记录</TableCell></TableRow>
                ) : (
                  orderData.items.map((row: PaymentOrder) => (
                    <TableRow key={row.id}>
                      <TableCell className="font-mono text-xs">{row.out_trade_no}</TableCell>
                      <TableCell className="font-semibold text-blue-600">￥{row.amount.toFixed(2)}</TableCell>
                      <TableCell className="font-semibold text-green-600">
                        {row.credits ? `+${formatCredits(row.credits)} 积分` : '—'}
                      </TableCell>
                      <TableCell>
                        <Badge variant={row.status === 'paid' ? 'default' : 'secondary'}>{orderStatusLabel(row.status)}</Badge>
                      </TableCell>
                      <TableCell>{row.paid_at || row.created_at}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
            {orderData?.total > 0 && (
              <div className="p-4 border-t">
                <TablePagination current={orderPage} total={orderData.total} pageSize={20} onChange={setOrderPage} />
              </div>
            )}
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={showPayFrame} onOpenChange={setShowPayFrame}>
        <DialogContent className="max-w-[400px]">
          <DialogHeader>
            <DialogTitle>扫描二维码支付</DialogTitle>
            <DialogDescription>
              请使用扫码完成支付，支付成功后系统将自动到账。
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col items-center justify-center p-4">
            {settings.payApplyEnabled && payUrl ? (
               <iframe src={payUrl} title="pay url" className="w-[280px] h-[280px] border-0" />
            ) : (
              <div className="py-8 text-center text-muted-foreground">
                <Wallet className="h-12 w-12 mx-auto mb-4 opacity-50" />
                即将跳转到支付网关...
              </div>
            )}
            <div className="mt-6 text-sm text-center">
              长按保存或截图扫码
              <br />
              <span className="text-muted-foreground text-xs mt-2 inline-block">单号: {currentOutTradeNo}</span>
            </div>
            
            <div className="mt-6 flex w-full gap-2">
              <Button variant="outline" className="w-full" onClick={() => setShowPayFrame(false)}>取消支付</Button>
              <Button className="w-full" onClick={() => {
                setShowPayFrame(false);
                txReload();
                orderReload();
                reloadBalance();
              }}>我已支付</Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
