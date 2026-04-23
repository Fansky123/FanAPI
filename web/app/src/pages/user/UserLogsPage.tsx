import { useState } from 'react'
import { Search, Loader2 } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { userApi, type UserLog } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function UserLogsPage() {
  const [page, setPage] = useState(1)
  const pageSize = 20
  const [filters, setFilters] = useState({ model: '', status: '', startAt: '', endAt: '' })
  
  const { data, loading, error, reload } = useAsync(async () => {
    const params: Record<string, any> = { page, page_size: pageSize }
    if (filters.model) params.model = filters.model
    if (filters.status) params.status = filters.status
    if (filters.startAt) params.start_at = filters.startAt.replace('T', ' ') + ':00'
    if (filters.endAt) params.end_at = filters.endAt.replace('T', ' ') + ':00'
    
    const res = await userApi.listLogs(params)
    return {
      logs: (Array.isArray(res) ? res : res.items ?? res.logs ?? []) as UserLog[],
      total: (res && !Array.isArray(res) ? res.total : 0) as number
    }
  }, { logs: [] as UserLog[], total: 0 })

  const rows = data.logs
  const total = data.total
  const totalPages = Math.ceil(total / pageSize)

  // Drawer state
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [currentLog, setCurrentLog] = useState<UserLog | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)

  async function openDetail(basicLog: UserLog) {
    setDrawerOpen(true)
    setDetailLoading(true)
    setCurrentLog({ ...basicLog })
    try {
      const res = await userApi.getLog(basicLog.id!)
      setCurrentLog({ ...res, credits_charged: basicLog.credits_charged ?? basicLog.cost_credits }) // Merge list data
    } catch (e) {
      console.error(e)
    } finally {
      setDetailLoading(false)
    }
  }

  function handleSearch() {
    setPage(1)
    setTimeout(reload, 0)
  }

  function handleReset() {
    setFilters({ model: '', status: '', startAt: '', endAt: '' })
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

  return (
    <>
      <PageHeader
        eyebrow="Observability"
        title="调用日志"
        description="查看带有过滤、分页、和详情查看的完整 API 调用记录。"
        actions={
          error ? (
            <Button size="sm" variant="outline" onClick={reload}>重试</Button>
          ) : null
        }
      />
      {error ? (
        <Alert variant="destructive" className="mb-4"><AlertDescription>{error}</AlertDescription></Alert>
      ) : null}
      
      <Card className="mb-4">
        <CardContent className="pt-6">
          <div className="flex flex-wrap items-center gap-3">
            <Input 
              placeholder="模型名称" 
              value={filters.model} 
              onChange={e => setFilters({...filters, model: e.target.value})}
              className="w-[180px]"
              onKeyDown={e => e.key === 'Enter' && handleSearch()}
            />
            <NativeSelect 
              value={filters.status} 
              onChange={e => setFilters({...filters, status: e.target.value})}
              className="w-[140px]"
            >
              <option value="">全部状态</option>
              <option value="ok">成功 (ok)</option>
              <option value="error">失败 (error)</option>
              <option value="refunded">已退款 (refunded)</option>
              <option value="pending">进行中 (pending)</option>
            </NativeSelect>
            <Input 
              type="datetime-local" 
              value={filters.startAt} 
              onChange={e => setFilters({...filters, startAt: e.target.value})}
              className="w-[200px]"
            />
            <span className="text-muted-foreground text-sm">至</span>
            <Input 
              type="datetime-local" 
              value={filters.endAt} 
              onChange={e => setFilters({...filters, endAt: e.target.value})}
              className="w-[200px]"
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
              <TableHead>模型</TableHead>
              <TableHead>请求时间</TableHead>
              <TableHead className="text-right">输入 Tokens</TableHead>
              <TableHead className="text-right">输出 Tokens</TableHead>
              <TableHead className="text-right">消耗积分</TableHead>
              <TableHead className="text-center">状态</TableHead>
              <TableHead className="text-center">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={7} rows={8} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="py-10 text-center text-muted-foreground">暂无调用日志</TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell className="font-medium max-w-[200px] truncate" title={row.model}>{row.model ?? '-'}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}</TableCell>
                    <TableCell className="text-right">
                      {row.usage?.prompt_tokens != null ? (
                        <div className="text-sm">
                          {row.usage.prompt_tokens.toLocaleString()}
                          {row.usage.cache_read_tokens ? <div className="text-[10px] text-muted-foreground leading-tight mt-1">命中 {row.usage.cache_read_tokens.toLocaleString()}</div> : null}
                          {row.usage.cache_creation_tokens ? <div className="text-[10px] text-muted-foreground leading-tight">写入 {row.usage.cache_creation_tokens.toLocaleString()}</div> : null}
                        </div>
                      ) : <span className="text-muted-foreground/50">—</span>}
                    </TableCell>
                    <TableCell className="text-right">
                      {row.usage?.completion_tokens != null ? (
                        <div className="flex flex-col items-end gap-1">
                          <span className="text-sm">{row.usage.completion_tokens.toLocaleString()}</span>
                          {row.usage.estimated ? <Badge variant="outline" className="text-[10px] h-4 px-1 py-0 border-orange-200 text-orange-600">估算</Badge> : null}
                        </div>
                      ) : <span className="text-muted-foreground/50">—</span>}
                    </TableCell>
                    <TableCell className="text-right">
                      {(row.credits_charged ?? row.cost_credits) ? (
                        <span className="font-semibold text-red-500">-{formatCredits(row.credits_charged ?? row.cost_credits)}</span>
                      ) : <span className="text-muted-foreground/50">—</span>}
                    </TableCell>
                    <TableCell className="text-center">{renderStatus(row.status)}</TableCell>
                    <TableCell className="text-center">
                      <Button variant="ghost" size="sm" onClick={() => openDetail(row)}>详情</Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
        {totalPages > 0 && (
          <div className="flex items-center justify-between px-4 py-4 border-t">
            <div className="text-sm text-muted-foreground">共 {total} 条数据</div>
            <div className="flex items-center space-x-2">
              <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => { setPage(p => p - 1); setTimeout(reload, 0) }}>上一页</Button>
              <div className="text-sm">第 {page} / {totalPages} 页</div>
              <Button variant="outline" size="sm" disabled={page >= totalPages} onClick={() => { setPage(p => p + 1); setTimeout(reload, 0) }}>下一页</Button>
            </div>
          </div>
        )}
      </Card>

      <Sheet open={drawerOpen} onOpenChange={setDrawerOpen}>
        <SheetContent className="w-[min(96vw,960px)] sm:max-w-[960px] overflow-y-auto">
          <SheetHeader className="mb-6"><SheetTitle>日志详情</SheetTitle></SheetHeader>
          {detailLoading ? (
            <div className="flex justify-center py-10"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
          ) : currentLog ? (
            <div className="space-y-6">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div><div className="text-muted-foreground mb-1">ID</div><div className="font-mono text-xs">{currentLog.id}</div></div>
                <div><div className="text-muted-foreground mb-1">状态</div><div>{renderStatus(currentLog.status)}</div></div>
                <div className="col-span-2"><div className="text-muted-foreground mb-1">模型</div><div className="font-medium">{currentLog.model}</div></div>
                <div className="col-span-2"><div className="text-muted-foreground mb-1">Corr ID</div><div className="font-mono text-xs break-all">{currentLog.corr_id}</div></div>
                
                <div><div className="text-muted-foreground mb-1">输入 Tokens</div><div>{currentLog.usage?.prompt_tokens ?? '—'}</div></div>
                <div>
                  <div className="text-muted-foreground mb-1">输出 Tokens {currentLog.usage?.estimated && <Badge variant="outline" className="ml-1 text-[10px] h-4 py-0">估算</Badge>}</div>
                  <div>{currentLog.usage?.completion_tokens ?? '—'}</div>
                </div>
                <div><div className="text-muted-foreground mb-1">消耗积分</div><div className={(currentLog.credits_charged ?? currentLog.cost_credits) ? "text-red-500 font-medium" : ""}>{(currentLog.credits_charged ?? currentLog.cost_credits) ? `-${formatCredits(currentLog.credits_charged ?? currentLog.cost_credits)}` : '—'}</div></div>
                <div><div className="text-muted-foreground mb-1">流式</div><div>{currentLog.is_stream ? '是' : '否'}</div></div>
                
                <div><div className="text-muted-foreground mb-1">请求时间</div><div>{currentLog.created_at ? new Date(currentLog.created_at).toLocaleString('zh-CN') : '-'}</div></div>
                <div><div className="text-muted-foreground mb-1">完成时间</div><div>{currentLog.status !== 'pending' && currentLog.updated_at ? new Date(currentLog.updated_at).toLocaleString('zh-CN') : '—'}</div></div>
              </div>

              {currentLog.error_msg && (
                <div>
                  <div className="text-sm font-semibold mb-2 text-red-600">错误信息</div>
                  <div className="bg-red-50 text-red-900 p-3 rounded-md text-sm whitespace-pre-wrap">{currentLog.error_msg}</div>
                </div>
              )}

              {currentLog.client_request && (
                <div>
                  <div className="text-sm font-semibold mb-2">您发送的请求</div>
                  <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-md text-xs font-mono overflow-x-auto break-all whitespace-pre-wrap max-h-[300px]">
                    {JSON.stringify(currentLog.client_request, null, 2)}
                  </pre>
                </div>
              )}
              {currentLog.upstream_headers && (
                <div>
                  <div className="text-sm font-semibold mb-2">上游请求头</div>
                  <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-md text-xs font-mono overflow-x-auto break-all whitespace-pre-wrap max-h-[200px]">
                    {JSON.stringify(currentLog.upstream_headers, null, 2)}
                  </pre>
                </div>
              )}
              {currentLog.upstream_request && (
                <div>
                  <div className="text-sm font-semibold mb-2">上游请求体</div>
                  <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-md text-xs font-mono overflow-x-auto break-all whitespace-pre-wrap max-h-[300px]">
                    {JSON.stringify(currentLog.upstream_request, null, 2)}
                  </pre>
                </div>
              )}
              {currentLog.upstream_response && (
                <div>
                  <div className="text-sm font-semibold mb-2">上游响应</div>
                  <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-md text-xs font-mono overflow-x-auto break-all whitespace-pre-wrap max-h-[300px]">
                    {JSON.stringify(currentLog.upstream_response, null, 2)}
                  </pre>
                </div>
              )}
              {currentLog.client_response && (
                <div>
                  <div className="text-sm font-semibold mb-2">返回给您的响应</div>
                  <pre className="bg-zinc-950 text-zinc-50 p-4 rounded-md text-xs font-mono overflow-x-auto break-all whitespace-pre-wrap max-h-[300px]">
                    {JSON.stringify(currentLog.client_response, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          ) : null}
        </SheetContent>
      </Sheet>
    </>
  )
}
