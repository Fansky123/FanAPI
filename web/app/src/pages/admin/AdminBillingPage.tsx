import { useState } from 'react'
import { WalletCardsIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableEmpty } from '@/components/shared/TableEmpty'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminTransaction, type AdminTransactionSummary } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

function toMoney(cny: number | undefined) {
  return (cny ?? 0).toFixed(4)
}

function txTypeLabel(t: string | undefined) {
  return ({ charge: '扣费', refund: '退款', recharge: '充值', hold: '预扣', settle: '结算' } as Record<string, string>)[t ?? ''] ?? (t ?? '-')
}

function txTypeVariant(t: string | undefined): 'destructive' | 'secondary' | 'outline' {
  if (t === 'charge' || t === 'hold' || t === 'settle') return 'destructive'
  if (t === 'refund' || t === 'recharge') return 'secondary'
  return 'outline'
}

function profitOf(row: AdminTransaction) {
  if (row.profit != null) return row.profit
  const amount = row.amount ?? 0
  const cost = row.cost ?? 0
  if (row.type === 'refund') return -(amount - cost)
  if (row.type === 'charge' || row.type === 'settle' || row.type === 'hold') return amount - cost
  return 0
}

export function AdminBillingPage() {
  const [page, setPage] = useState(1)
  const [startAt, setStartAt] = useState('')
  const [endAt, setEndAt] = useState('')
  const [searchParams, setSearchParams] = useState<Record<string, unknown>>({ page: 1 })

  const { data, loading, error, reload } = useAsync(async () => {
    const res = await adminApi.listTransactions(searchParams)
    const transactions = Array.isArray(res) ? res : res.transactions ?? res.items ?? []
    const total = Array.isArray(res) ? transactions.length : (res as { total?: number }).total ?? transactions.length
    const summary: AdminTransactionSummary = Array.isArray(res) ? {} : (res as { summary?: AdminTransactionSummary }).summary ?? {}
    return { transactions, total, summary }
  }, { transactions: [] as AdminTransaction[], total: 0, summary: {} as AdminTransactionSummary }, [searchParams])

  const pageSize = 20
  const totalPages = Math.ceil(data.total / pageSize)

  function doSearch() {
    const params: Record<string, unknown> = { page: 1, size: pageSize }
    if (startAt) params.start_at = startAt.replace('T', ' ') + ':00'
    if (endAt) params.end_at = endAt.replace('T', ' ') + ':00'
    setPage(1)
    setSearchParams(params)
  }

  function changePage(next: number) {
    setPage(next)
    setSearchParams((prev) => ({ ...prev, page: next }))
  }

  return (
    <>
      <PageHeader
        eyebrow="Finance"
        title="账单流水与利润统计"
        description="按时间范围查看平台收入、成本和利润，支持运营复盘与对账。"
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

      {/* 过滤栏 */}
      <Card>
        <CardContent className="flex flex-wrap items-end gap-3 py-4">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">开始时间</label>
            <Input type="datetime-local" value={startAt} onChange={(e) => setStartAt(e.target.value)} className="w-52" />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">结束时间</label>
            <Input type="datetime-local" value={endAt} onChange={(e) => setEndAt(e.target.value)} className="w-52" />
          </div>
          <Button onClick={doSearch}>查询</Button>
        </CardContent>
      </Card>

      {/* 汇总卡片 */}
      {(data.summary.revenue != null || data.summary.cost != null) ? (
        <div className="grid gap-4 sm:grid-cols-3">
          <Card>
            <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">收入</CardTitle></CardHeader>
            <CardContent><p className="text-2xl font-bold">¥{toMoney(data.summary.revenue)}</p></CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">成本</CardTitle></CardHeader>
            <CardContent><p className="text-2xl font-bold">¥{toMoney(data.summary.cost)}</p></CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2"><CardTitle className="text-sm font-medium text-muted-foreground">利润</CardTitle></CardHeader>
            <CardContent><p className="text-2xl font-bold text-blue-600">¥{toMoney(data.summary.profit)}</p></CardContent>
          </Card>
        </div>
      ) : null}

      {/* 流水表格 */}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-14">ID</TableHead>
              <TableHead className="w-16">用户 ID</TableHead>
              <TableHead className="w-20">类型</TableHead>
              <TableHead className="w-36 text-right">金额（CNY）</TableHead>
              <TableHead className="w-32 text-right">成本（CNY）</TableHead>
              <TableHead className="w-32 text-right">利润（CNY）</TableHead>
              <TableHead className="w-16">渠道</TableHead>
              <TableHead>关联 ID</TableHead>
              <TableHead className="w-40">时间</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={9} />
          ) : (
            <TableBody>
              {data.transactions.length === 0 ? (
                <TableEmpty
                  cols={9}
                  Icon={WalletCardsIcon}
                  title="还没有账单记录"
                  description="平台累计的所有积分流水会汇总在此处。"
                />
              ) : (
                data.transactions.map((row, index) => {
                  const amount = row.amount ?? 0
                  const isDebit = ['charge', 'hold', 'settle'].includes(row.type ?? '')
                  const profit = profitOf(row)
                  return (
                    <TableRow key={row.id ?? index}>
                      <TableCell className="text-muted-foreground">{row.id ?? '-'}</TableCell>
                      <TableCell className="text-muted-foreground">{row.user_id ?? '-'}</TableCell>
                      <TableCell>
                        <Badge variant={txTypeVariant(row.type)} className="text-xs">
                          {txTypeLabel(row.type)}
                        </Badge>
                      </TableCell>
                      <TableCell className={`text-right font-mono text-sm ${isDebit ? 'text-red-500' : 'text-emerald-600'}`}>
                        {isDebit ? '-' : '+'}{Math.abs(amount).toLocaleString('zh-CN', { minimumFractionDigits: 4, maximumFractionDigits: 4 })}
                      </TableCell>
                      <TableCell className="text-right font-mono text-sm text-muted-foreground">
                        {(row.cost ?? 0).toLocaleString('zh-CN', { minimumFractionDigits: 4, maximumFractionDigits: 4 })}
                      </TableCell>
                      <TableCell className={`text-right font-mono text-sm ${profit >= 0 ? 'text-blue-600' : 'text-red-500'}`}>
                        {profit.toLocaleString('zh-CN', { minimumFractionDigits: 4, maximumFractionDigits: 4 })}
                      </TableCell>
                      <TableCell className="text-muted-foreground">{row.channel_id ?? '-'}</TableCell>
                      <TableCell className="max-w-32 truncate text-xs text-muted-foreground" title={row.corr_id}>
                        {row.corr_id ?? '-'}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          )}
        </Table>
        {totalPages > 1 ? (
          <CardContent className="flex items-center justify-between border-t py-3">
            <span className="text-sm text-muted-foreground">共 {data.total} 条记录</span>
            <div className="flex items-center gap-2">
              <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => changePage(page - 1)}>上一页</Button>
              <span className="text-sm text-muted-foreground">第 {page} / {totalPages} 页</span>
              <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => changePage(page + 1)}>下一页</Button>
            </div>
          </CardContent>
        ) : null}
      </Card>
    </>
  )
}

