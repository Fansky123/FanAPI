import { useState } from 'react'
import { FileClockIcon, Search, Loader2 } from 'lucide-react'

import { DateRangeFilter } from '@/components/shared/DateRangeFilter'
import { PageHeader } from '@/components/shared/PageHeader'
import { TableEmpty } from '@/components/shared/TableEmpty'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminLog } from '@/lib/api/admin'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

function JsonBlock({
  title,
  value,
  maxHeight = 'max-h-60',
}: {
  title: string
  value: unknown
  maxHeight?: string
}) {
  if (!value || (typeof value === 'object' && Object.keys(value as object).length === 0)) {
    return null
  }

  return (
    <div>
      <div className="mb-2 font-semibold">{title}</div>
      <pre className={`overflow-auto rounded-md bg-muted p-3 text-xs whitespace-pre-wrap break-all ${maxHeight}`}>
        {JSON.stringify(value, null, 2)}
      </pre>
    </div>
  )
}

function InfoItem({
  label,
  value,
  className,
}: {
  label: string
  value: React.ReactNode
  className?: string
}) {
  return (
    <div className={className}>
      <div className="mb-1 text-muted-foreground">{label}</div>
      <div>{value}</div>
    </div>
  )
}

export function AdminLogsPage() {
  const [page, setPage] = useState(1)
  const pageSize = 20
  const [filters, setFilters] = useState({
    model: '', user_id: '', channel_id: '', status: '', corr_id: '', startAt: '', endAt: '',
  })

  const { data, loading, error, reload } = useAsync(async () => {
    const params: Record<string, unknown> = { page, page_size: pageSize }
    if (filters.model) params.model = filters.model
    if (filters.user_id) params.user_id = filters.user_id
    if (filters.channel_id) params.channel_id = filters.channel_id
    if (filters.status) params.status = filters.status
    if (filters.corr_id) params.corr_id = filters.corr_id
    if (filters.startAt) params.start_at = new Date(filters.startAt).getTime()
    if (filters.endAt) params.end_at = new Date(filters.endAt).getTime()

    const res = await adminApi.listLogs(params)
    return {
      logs: (Array.isArray(res) ? res : res.logs ?? res.items ?? []) as AdminLog[],
      total: (!Array.isArray(res) ? (res.total ?? 0) : 0) as number,
    }
  }, { logs: [] as AdminLog[], total: 0 })

  const rows = data.logs
  const total = data.total
  const totalPages = Math.ceil(total / pageSize)

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [currentLog, setCurrentLog] = useState<AdminLog | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)

  async function openDetail(basicLog: AdminLog) {
    setDrawerOpen(true)
    setDetailLoading(true)
    setCurrentLog({ ...basicLog })
    try {
      const res = await adminApi.getLog(basicLog.id!)
      setCurrentLog({ ...res, credits_charged: basicLog.credits_charged, cost_charged: basicLog.cost_charged })
    } catch (e) {
      console.error(e)
    } finally {
      setDetailLoading(false)
    }
  }

  function handleSearch() { setPage(1); setTimeout(reload, 0) }
  function handleReset() {
    setFilters({ model: '', user_id: '', channel_id: '', status: '', corr_id: '', startAt: '', endAt: '' })
    setPage(1)
    setTimeout(reload, 0)
  }

  function renderStatus(status?: string) {
    if (status === 'ok') return <Badge variant="secondary" className="bg-green-100 text-green-800">成功</Badge>
    if (status === 'error') return <Badge variant="destructive">失败</Badge>
    if (status === 'refunded') return <Badge variant="outline" className="border-orange-200 text-orange-600">已退款</Badge>
    if (status === 'pending') return <Badge variant="secondary">进行中</Badge>
    return <Badge variant="outline">{status ?? '-'}</Badge>
  }

  function renderUsageSummary(row: AdminLog) {
    if (!row.usage) {
      return <span className="text-muted-foreground/50">—</span>
    }

    return (
      <div className="flex flex-col gap-1 text-xs">
        <span>
          ↑{row.usage.prompt_tokens?.toLocaleString() ?? '-'} / ↓{row.usage.completion_tokens?.toLocaleString() ?? '-'}
        </span>
        {row.usage.cache_creation_tokens || row.usage.cache_read_tokens ? (
          <span className="text-muted-foreground">
            写缓存 {row.usage.cache_creation_tokens?.toLocaleString() ?? 0}
            {' / '}
            命缓存 {row.usage.cache_read_tokens?.toLocaleString() ?? 0}
          </span>
        ) : null}
      </div>
    )
  }

  return (
    <>
      <PageHeader
        eyebrow="Observability"
        title="调用日志"
        description="查看平台所有 API 调用日志记录。"
        actions={error ? <Button size="sm" variant="outline" onClick={reload}>重试</Button> : null}
      />
      {error ? <Alert variant="destructive" className="mb-4"><AlertDescription>{error}</AlertDescription></Alert> : null}

      <Card className="mb-4">
        <CardContent className="pt-6">
          <div className="flex flex-wrap items-center gap-3">
            <Input placeholder="模型名称" value={filters.model}
              onChange={e => setFilters({ ...filters, model: e.target.value })} className="w-[160px]"
              onKeyDown={e => e.key === 'Enter' && handleSearch()} />
            <Input placeholder="用户 ID" value={filters.user_id}
              onChange={e => setFilters({ ...filters, user_id: e.target.value })} className="w-[100px]"
              onKeyDown={e => e.key === 'Enter' && handleSearch()} />
            <Input placeholder="渠道 ID" value={filters.channel_id}
              onChange={e => setFilters({ ...filters, channel_id: e.target.value })} className="w-[100px]"
              onKeyDown={e => e.key === 'Enter' && handleSearch()} />
            <Input placeholder="Corr ID" value={filters.corr_id}
              onChange={e => setFilters({ ...filters, corr_id: e.target.value })} className="w-[220px] font-mono text-xs"
              onKeyDown={e => e.key === 'Enter' && handleSearch()} />
            <NativeSelect value={filters.status} onChange={e => setFilters({ ...filters, status: e.target.value })} className="w-[140px]">
              <option value="">全部状态</option>
              <option value="ok">成功 (ok)</option>
              <option value="error">失败 (error)</option>
              <option value="refunded">已退款 (refunded)</option>
              <option value="pending">进行中 (pending)</option>
            </NativeSelect>
            <DateRangeFilter
              startAt={filters.startAt}
              endAt={filters.endAt}
              onChange={({ startAt, endAt }) => setFilters({ ...filters, startAt, endAt })}
            />
            <Button onClick={handleSearch}><Search className="mr-2 h-4 w-4" />查询</Button>
            <Button variant="outline" onClick={handleReset}>重置</Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[60px]">ID</TableHead>
              <TableHead>模型</TableHead>
              <TableHead>用户 ID</TableHead>
              <TableHead>用户名</TableHead>
              <TableHead>渠道 ID</TableHead>
              <TableHead>上游 API Key</TableHead>
              <TableHead>相关 ID</TableHead>
              <TableHead>Token 用量</TableHead>
              <TableHead className="text-right">消耗积分</TableHead>
              <TableHead className="text-right">上游成本</TableHead>
              <TableHead className="text-center">上游状态</TableHead>
              <TableHead className="text-center">状态</TableHead>
              <TableHead>时间</TableHead>
              <TableHead className="text-center">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? <TableSkeleton cols={13} rows={10} /> : (
            <TableBody>
              {rows.length === 0 ? (
                <TableEmpty cols={13} Icon={FileClockIcon} title="还没有调用日志" description="所有 LLM 调用记录会在此处汇总。" />
              ) : rows.map((row, idx) => (
                <TableRow key={row.id ?? idx}>
                  <TableCell className="text-muted-foreground">{row.id ?? '-'}</TableCell>
                  <TableCell className="font-medium max-w-[180px] truncate" title={row.model}>{row.model ?? '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{row.user_id ?? '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{row.username ?? '-'}</TableCell>
                  <TableCell className="text-muted-foreground">{row.channel_id ?? '-'}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground max-w-[180px] truncate" title={row.upstream_api_key}>
                    {row.upstream_api_key ?? '-'}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground max-w-[220px] truncate" title={row.corr_id}>{row.corr_id ?? '-'}</TableCell>
                  <TableCell>
                    {renderUsageSummary(row)}
                  </TableCell>
                  <TableCell className="text-right">
                    {row.credits_charged ? (
                      <span className="font-semibold text-red-500">-{formatCredits(row.credits_charged)}</span>
                    ) : <span className="text-muted-foreground/50">—</span>}
                  </TableCell>
                  <TableCell className="text-right">
                    {row.cost_charged ? (
                      <span className="text-amber-600 dark:text-amber-400">-{formatCredits(row.cost_charged)}</span>
                    ) : <span className="text-muted-foreground/50">—</span>}
                  </TableCell>
                  <TableCell className="text-center">
                    {row.upstream_status ?? <span className="text-muted-foreground/50">—</span>}
                  </TableCell>
                  <TableCell className="text-center">{renderStatus(row.status)}</TableCell>
                  <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                    {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                  </TableCell>
                  <TableCell className="text-center">
                    <Button variant="ghost" size="sm" onClick={() => openDetail(row)}>详情</Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          )}
        </Table>
        {totalPages > 0 && (
          <div className="flex items-center justify-between px-4 py-4 border-t">
            <div className="text-sm text-muted-foreground">共 {total} 条数据</div>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" disabled={page <= 1}
                onClick={() => { setPage(p => p - 1); setTimeout(reload, 0) }}>上一页</Button>
              <div className="text-sm">第 {page} / {totalPages || 1} 页</div>
              <Button variant="outline" size="sm" disabled={page >= totalPages}
                onClick={() => { setPage(p => p + 1); setTimeout(reload, 0) }}>下一页</Button>
            </div>
          </div>
        )}
      </Card>

      <Sheet open={drawerOpen} onOpenChange={setDrawerOpen}>
        <SheetContent className="w-[min(96vw,1160px)] sm:max-w-[1160px] overflow-y-auto">
          <SheetHeader className="mb-6"><SheetTitle>日志详情</SheetTitle></SheetHeader>
          {detailLoading ? (
            <div className="flex justify-center py-10"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
          ) : currentLog ? (
            <div className="space-y-6 text-sm">
              <div className="grid grid-cols-2 gap-4">
                <InfoItem label="ID" value={<div className="font-mono">{currentLog.id ?? '—'}</div>} />
                <InfoItem label="状态" value={renderStatus(currentLog.status)} />
                <InfoItem label="模型" value={<div className="font-medium">{currentLog.model ?? '—'}</div>} className="col-span-2" />
                <InfoItem label="用户 ID" value={currentLog.user_id ?? '—'} />
                <InfoItem label="渠道 ID" value={currentLog.channel_id ?? '—'} />
                <InfoItem label="API Key ID" value={currentLog.api_key_id ?? '—'} />
                <InfoItem label="流式" value={currentLog.is_stream ? '是' : '否'} />
                <InfoItem label="Corr ID" value={<div className="font-mono text-xs break-all">{currentLog.corr_id ?? '—'}</div>} className="col-span-2" />
                <InfoItem label="上游 HTTP 状态" value={currentLog.upstream_status ?? '—'} />
                <InfoItem label="上游请求方法" value={currentLog.upstream_method ?? 'POST'} />
                <InfoItem label="上游 URL" value={<div className="break-all">{currentLog.upstream_url ?? '—'}</div>} className="col-span-2" />
                <InfoItem label="输入 Tokens" value={currentLog.usage?.prompt_tokens ?? '—'} />
                <InfoItem
                  label="输出 Tokens"
                  value={
                    <div>
                      {currentLog.usage?.completion_tokens ?? '—'}
                      {currentLog.usage?.estimated ? <Badge variant="outline" className="ml-2">估算</Badge> : null}
                    </div>
                  }
                />
                <InfoItem label="缓存写入" value={currentLog.usage?.cache_creation_tokens?.toLocaleString() ?? '—'} />
                <InfoItem label="缓存命中" value={currentLog.usage?.cache_read_tokens?.toLocaleString() ?? '—'} />
                <InfoItem
                  label="消耗积分"
                  value={
                    <div className={currentLog.credits_charged ? 'font-medium text-red-500' : ''}>
                      {currentLog.credits_charged ? `-${formatCredits(currentLog.credits_charged)}` : '—'}
                    </div>
                  }
                />
                <InfoItem
                  label="上游成本"
                  value={
                    <div className={currentLog.cost_charged ? 'font-medium text-amber-600 dark:text-amber-400' : ''}>
                      {currentLog.cost_charged ? `-${formatCredits(currentLog.cost_charged)}` : '—'}
                    </div>
                  }
                />
                <InfoItem
                  label="毛利"
                  value={
                    (() => {
                      const c = currentLog.credits_charged ?? 0
                      const u = currentLog.cost_charged ?? 0
                      const profit = c - u
                      if (!c && !u) return <span className="text-muted-foreground">—</span>
                      const sign = profit >= 0 ? '+' : '−'
                      const cls = profit >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-500'
                      return <div className={`font-medium ${cls}`}>{sign}{formatCredits(Math.abs(profit))}</div>
                    })()
                  }
                />
                <InfoItem label="请求时间" value={currentLog.created_at ? new Date(currentLog.created_at).toLocaleString('zh-CN') : '—'} />
                <InfoItem label="完成时间" value={currentLog.updated_at ? new Date(currentLog.updated_at).toLocaleString('zh-CN') : '—'} className="col-span-2" />
              </div>
              {currentLog.error_msg && (
                <div>
                  <div className="font-semibold mb-2 text-red-600">错误信息</div>
                  <div className="rounded-md bg-red-50 p-3 text-sm whitespace-pre-wrap text-red-900">{currentLog.error_msg}</div>
                </div>
              )}
              <JsonBlock title="Token 用量详情" value={currentLog.usage} maxHeight="max-h-40" />
              <JsonBlock title="客户端请求" value={currentLog.client_request} />
              <JsonBlock title="上游请求头" value={currentLog.upstream_headers} maxHeight="max-h-40" />
              <JsonBlock title="上游请求体" value={currentLog.upstream_request} />
              <JsonBlock title="上游响应" value={currentLog.upstream_response} />
              <JsonBlock title="客户端响应" value={currentLog.client_response} />
            </div>
          ) : null}
        </SheetContent>
      </Sheet>
    </>
  )
}
