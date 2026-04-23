import { useEffect, useRef, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminTask } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

function statusBadge(s: string | undefined) {
  if (s === 'pending') return <Badge className="bg-yellow-500 hover:bg-yellow-500 text-white">排队中</Badge>
  if (s === 'processing') return <Badge className="bg-blue-500 hover:bg-blue-500 text-white">处理中</Badge>
  if (s === 'done') return <Badge className="bg-emerald-600 hover:bg-emerald-600 text-white">成功</Badge>
  if (s === 'failed') return <Badge variant="destructive">失败</Badge>
  return <Badge variant="outline">{s ?? '-'}</Badge>
}

function JsonBlock({ title, value }: { title: string; value: unknown }) {
  if (!value || (typeof value === 'object' && Object.keys(value as object).length === 0)) return null
  return (
    <div className="mb-4">
      <p className="mb-1 text-sm font-semibold">{title}</p>
      <pre className="overflow-auto rounded-lg border bg-muted/40 p-3 font-mono text-xs leading-relaxed whitespace-pre-wrap break-all">
        {JSON.stringify(value, null, 2)}
      </pre>
    </div>
  )
}

export function AdminTasksPage() {
  const [page, setPage] = useState(1)
  const [filters, setFilters] = useState({ task_id: '', user_id: '', type: '', status: '' })
  const [startAt, setStartAt] = useState('')
  const [endAt, setEndAt] = useState('')
  const [queryParams, setQueryParams] = useState<Record<string, unknown>>({ page: 1, size: 20 })

  const { data, loading, error, reload } = useAsync(async () => {
    const res = await adminApi.listTasks(queryParams)
    const tasks = Array.isArray(res) ? res : (res.tasks ?? res.items ?? [])
    const total = Array.isArray(res) ? tasks.length : (res as { total?: number }).total ?? tasks.length
    return { tasks, total }
  }, { tasks: [] as AdminTask[], total: 0 }, [queryParams])

  const pageSize = 20
  const totalPages = Math.ceil(data.total / pageSize)

  // 详情弹窗
  const [detail, setDetail] = useState<AdminTask | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const autoRefreshRef = useRef<ReturnType<typeof setInterval> | null>(null)

  function stopAutoRefresh() {
    if (autoRefreshRef.current) { clearInterval(autoRefreshRef.current); autoRefreshRef.current = null }
  }
  useEffect(() => () => stopAutoRefresh(), [])

  async function openDetail(id: number) {
    setDetailLoading(true)
    stopAutoRefresh()
    try {
      const res = await adminApi.getAdminTask(id)
      const task: AdminTask = (res as { task?: AdminTask }).task ?? (res as AdminTask)
      setDetail(task)
      if (task.status === 'pending' || task.status === 'processing') {
        autoRefreshRef.current = setInterval(async () => {
          const r = await adminApi.getAdminTask(id)
          const t: AdminTask = (r as { task?: AdminTask }).task ?? (r as AdminTask)
          setDetail(t)
          if (t.status !== 'pending' && t.status !== 'processing') stopAutoRefresh()
        }, 3000)
      }
    } finally {
      setDetailLoading(false)
    }
  }

  function closeDetail() {
    stopAutoRefresh()
    setDetail(null)
  }

  function doSearch() {
    const params: Record<string, unknown> = { page: 1, size: pageSize }
    if (filters.task_id) params.task_id = filters.task_id
    if (filters.user_id) params.user_id = filters.user_id
    if (filters.type) params.type = filters.type
    if (filters.status) params.status = filters.status
    if (startAt) params.start_at = startAt.replace('T', ' ') + ':00'
    if (endAt) params.end_at = endAt.replace('T', ' ') + ':00'
    setPage(1)
    setQueryParams(params)
  }

  function resetFilters() {
    setFilters({ task_id: '', user_id: '', type: '', status: '' })
    setStartAt('')
    setEndAt('')
    setPage(1)
    setQueryParams({ page: 1, size: pageSize })
  }

  function changePage(next: number) {
    setPage(next)
    setQueryParams((prev) => ({ ...prev, page: next }))
  }

  const upstreamURL = (req?: Record<string, unknown>) => req?._url as string | undefined
  const upstreamHeaders = (req?: Record<string, unknown>) => {
    if (!req) return undefined
    return req._headers as Record<string, unknown> | undefined
  }
  const upstreamBody = (req?: Record<string, unknown>) => {
    if (!req) return undefined
    const { _url, _headers, ...rest } = req
    return Object.keys(rest).length ? rest : undefined
  }

  return (
    <>
      <PageHeader
        eyebrow="Task Center"
        title="异步任务排障"
        description="查看用户任务状态与完整请求/响应信息，辅助对接排障。"
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
            <label className="text-xs text-muted-foreground">Task ID</label>
            <Input
              className="w-28"
              placeholder="Task ID"
              value={filters.task_id}
              onChange={(e) => setFilters((f) => ({ ...f, task_id: e.target.value }))}
              onKeyDown={(e) => e.key === 'Enter' && doSearch()}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">用户 ID</label>
            <Input
              className="w-28"
              placeholder="用户 ID"
              value={filters.user_id}
              onChange={(e) => setFilters((f) => ({ ...f, user_id: e.target.value }))}
              onKeyDown={(e) => e.key === 'Enter' && doSearch()}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">任务类型</label>
            <Select value={filters.type || '_all'} onValueChange={(v) => setFilters((f) => ({ ...f, type: v === '_all' ? '' : v }))}>
              <SelectTrigger className="w-32"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">全部类型</SelectItem>
                <SelectItem value="image">图片</SelectItem>
                <SelectItem value="video">视频</SelectItem>
                <SelectItem value="audio">音频</SelectItem>
                <SelectItem value="music">音乐</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">状态</label>
            <Select value={filters.status || '_all'} onValueChange={(v) => setFilters((f) => ({ ...f, status: v === '_all' ? '' : v }))}>
              <SelectTrigger className="w-28"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">全部状态</SelectItem>
                <SelectItem value="pending">排队中</SelectItem>
                <SelectItem value="processing">处理中</SelectItem>
                <SelectItem value="done">成功</SelectItem>
                <SelectItem value="failed">失败</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">开始时间</label>
            <Input type="datetime-local" className="w-48" value={startAt} onChange={(e) => setStartAt(e.target.value)} />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">结束时间</label>
            <Input type="datetime-local" className="w-48" value={endAt} onChange={(e) => setEndAt(e.target.value)} />
          </div>
          <Button onClick={doSearch}>查询</Button>
          <Button variant="outline" onClick={resetFilters}>重置</Button>
        </CardContent>
      </Card>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-20">Task ID</TableHead>
              <TableHead className="w-20">用户 ID</TableHead>
              <TableHead className="w-20">类型</TableHead>
              <TableHead className="w-28">状态</TableHead>
              <TableHead className="w-20">渠道</TableHead>
              <TableHead className="w-32 text-right">扣费（cr）</TableHead>
              <TableHead>第三方任务 ID</TableHead>
              <TableHead className="w-40">创建时间</TableHead>
              <TableHead className="w-40">结束时间</TableHead>
              <TableHead className="w-16 text-center">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={10} />
          ) : (
            <TableBody>
              {data.tasks.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={10} className="py-10 text-center text-muted-foreground">
                    暂无任务记录
                  </TableCell>
                </TableRow>
              ) : (
                data.tasks.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell>{row.id ?? '-'}</TableCell>
                    <TableCell className="text-muted-foreground">{row.user_id ?? '-'}</TableCell>
                    <TableCell>{row.type ?? '-'}</TableCell>
                    <TableCell>{statusBadge(row.status)}</TableCell>
                    <TableCell className="text-muted-foreground">{row.channel_id ?? '-'}</TableCell>
                    <TableCell className="text-right font-mono text-sm">
                      {row.credits_charged != null ? row.credits_charged.toLocaleString() : '-'}
                    </TableCell>
                    <TableCell className="max-w-40 truncate font-mono text-xs text-muted-foreground" title={row.upstream_task_id}>
                      {row.upstream_task_id ?? '-'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {(row.status === 'done' || row.status === 'failed') && row.updated_at
                        ? new Date(row.updated_at).toLocaleString('zh-CN')
                        : '-'}
                    </TableCell>
                    <TableCell className="text-center">
                      {row.id != null ? (
                        <Button size="sm" variant="link" className="h-auto p-0" onClick={() => openDetail(row.id!)}>
                          详情
                        </Button>
                      ) : null}
                    </TableCell>
                  </TableRow>
                ))
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

      {/* 详情弹窗 */}
      <Dialog open={Boolean(detail)} onOpenChange={closeDetail}>
        <DialogContent className="max-w-[872px] max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>任务详情 #{detail?.id}</DialogTitle>
          </DialogHeader>
          {detailLoading ? (
            <p className="py-8 text-center text-sm text-muted-foreground">加载中…</p>
          ) : detail ? (
            <div className="space-y-1">
              <div className="grid grid-cols-2 gap-x-4 gap-y-1 rounded-lg border p-4 text-sm">
                <div><span className="text-muted-foreground">Task ID：</span><strong>{detail.id}</strong></div>
                <div><span className="text-muted-foreground">用户 ID：</span><strong>{detail.user_id ?? '-'}</strong></div>
                <div><span className="text-muted-foreground">任务类型：</span>{detail.type ?? '-'}</div>
                <div><span className="text-muted-foreground">状态：</span>{statusBadge(detail.status)}</div>
                <div><span className="text-muted-foreground">渠道 ID：</span>{detail.channel_id ?? '-'}</div>
                <div><span className="text-muted-foreground">扣费：</span>{detail.credits_charged?.toLocaleString() ?? '-'} cr</div>
                {detail.upstream_task_id ? (
                  <div className="col-span-2"><span className="text-muted-foreground">第三方任务 ID：</span><span className="font-mono text-xs">{detail.upstream_task_id}</span></div>
                ) : null}
                {detail.error_msg ? (
                  <div className="col-span-2"><span className="text-muted-foreground">错误信息：</span><span className="text-red-500">{detail.error_msg}</span></div>
                ) : null}
                <div className="col-span-2"><span className="text-muted-foreground">创建时间：</span>{detail.created_at ? new Date(detail.created_at).toLocaleString('zh-CN') : '-'}</div>
                {(detail.status === 'done' || detail.status === 'failed') && detail.updated_at ? (
                  <div className="col-span-2"><span className="text-muted-foreground">完成时间：</span>{new Date(detail.updated_at).toLocaleString('zh-CN')}</div>
                ) : null}
              </div>
              <div className="pt-2">
                <JsonBlock title="用户提交请求体" value={detail.request} />
                {detail.upstream_request && upstreamURL(detail.upstream_request) ? (
                  <div className="mb-4">
                    <p className="mb-1 text-sm font-semibold">上游请求 URL</p>
                    <pre className="overflow-auto rounded-lg border bg-muted/40 p-3 font-mono text-xs text-blue-600 whitespace-pre-wrap break-all">{upstreamURL(detail.upstream_request)}</pre>
                  </div>
                ) : null}
                <JsonBlock title="上游请求头" value={upstreamHeaders(detail.upstream_request)} />
                <JsonBlock title="发送给第三方的请求体" value={upstreamBody(detail.upstream_request)} />
                <JsonBlock title="第三方原始响应体" value={detail.upstream_response} />
                <JsonBlock title="平台标准结果" value={detail.result} />
              </div>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
    </>
  )
}

