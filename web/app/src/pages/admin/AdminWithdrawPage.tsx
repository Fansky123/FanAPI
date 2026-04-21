import { useEffect, useRef, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminWithdrawal } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

function statusBadge(s: string | undefined) {
  if (s === 'pending') return <Badge className="bg-yellow-500 hover:bg-yellow-500 text-white">待审核</Badge>
  if (s === 'approved') return <Badge className="bg-emerald-600 hover:bg-emerald-600 text-white">已通过</Badge>
  if (s === 'rejected') return <Badge variant="destructive">已拒绝</Badge>
  return <Badge variant="outline">{s ?? '-'}</Badge>
}

function payTypeBadge(t: string | undefined) {
  if (t === 'wechat') return <Badge className="bg-emerald-600 hover:bg-emerald-600 text-white">微信</Badge>
  if (t === 'alipay') return <Badge className="bg-blue-600 hover:bg-blue-600 text-white">支付宝</Badge>
  return <Badge variant="secondary">{t ?? '-'}</Badge>
}

export function AdminWithdrawPage() {
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('pending')
  const [queryParams, setQueryParams] = useState({ page: 1, size: 20, status: 'pending' })

  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [listRes, countRes] = await Promise.all([
      adminApi.listWithdrawals(queryParams),
      adminApi.getPendingWithdrawCount(),
    ])
    const records: AdminWithdrawal[] = (listRes as { records?: AdminWithdrawal[] }).records ?? (Array.isArray(listRes) ? listRes : [])
    const total: number = (listRes as { total?: number }).total ?? records.length
    return { rows: records, total, pendingCount: countRes.count ?? 0 }
  }, { rows: [] as AdminWithdrawal[], total: 0, pendingCount: 0 }, [queryParams])

  // 自动轮询待审核数量（每 30 秒）
  const reloadRef = useRef(reload)
  reloadRef.current = reload
  useEffect(() => {
    const timer = setInterval(() => reloadRef.current(), 30_000)
    return () => clearInterval(timer)
  }, [])

  const rows = data.rows
  const pendingCount = data.pendingCount
  const pageSize = 20
  const totalPages = Math.ceil(data.total / pageSize)

  const [mutError, setMutError] = useState('')
  const [rejecting, setRejecting] = useState<AdminWithdrawal | null>(null)
  const [remark, setRemark] = useState('')
  const [pendingApprove, setPendingApprove] = useState<AdminWithdrawal | null>(null)
  const [viewRow, setViewRow] = useState<AdminWithdrawal | null>(null)

  const error = loadError || mutError

  function doFilter() {
    setPage(1)
    setQueryParams({ page: 1, size: pageSize, status: statusFilter })
  }

  function changePage(next: number) {
    setPage(next)
    setQueryParams((prev) => ({ ...prev, page: next }))
  }

  async function executeApprove(row?: AdminWithdrawal) {
    const target = row ?? pendingApprove
    if (!target?.id) return
    setMutError('')
    try {
      await adminApi.approveWithdrawal(target.id)
      setViewRow(null)
      setPendingApprove(null)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function reject() {
    if (!rejecting?.id) return
    setMutError('')
    try {
      await adminApi.rejectWithdrawal(rejecting.id, remark)
      setRejecting(null)
      setRemark('')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Withdraw"
        title="提现审核"
        description={`当前待处理 ${pendingCount} 条提现申请。`}
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
            <label className="text-xs text-muted-foreground">状态</label>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="pending">待审核</SelectItem>
                <SelectItem value="approved">已通过</SelectItem>
                <SelectItem value="rejected">已拒绝</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button onClick={doFilter}>查询</Button>
        </CardContent>
      </Card>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>用户</TableHead>
              <TableHead>申请时间</TableHead>
              <TableHead>金额</TableHead>
              <TableHead>收款方式</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>备注</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={8} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="py-10 text-center text-muted-foreground">
                    暂无提现申请
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell>{row.id ?? '-'}</TableCell>
                    <TableCell>{row.username ?? '-'}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                    </TableCell>
                    <TableCell>¥{((row.amount ?? 0) / 1_000_000).toFixed(4)}</TableCell>
                    <TableCell>{payTypeBadge(row.payment_type)}</TableCell>
                    <TableCell>{statusBadge(row.status)}</TableCell>
                    <TableCell className="max-w-40 truncate text-xs text-muted-foreground" title={row.admin_remark}>
                      {row.admin_remark ?? '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button size="sm" variant="outline" onClick={() => setViewRow(row)}>
                          查看收款码
                        </Button>
                        {row.status === 'pending' ? (
                          <>
                            <Button size="sm" variant="outline" onClick={() => setPendingApprove(row)}>
                              通过
                            </Button>
                            <Button size="sm" onClick={() => setRejecting(row)}>
                              拒绝
                            </Button>
                          </>
                        ) : null}
                      </div>
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

      {/* 查看收款码弹窗 */}
      <Dialog open={Boolean(viewRow)} onOpenChange={() => setViewRow(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>收款码 — {viewRow?.username}</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col items-center gap-3">
            {viewRow?.payment_qr ? (
              <img
                src={viewRow.payment_qr}
                alt="收款码"
                className="h-60 w-60 rounded-md border object-contain"
              />
            ) : (
              <div className="flex h-60 w-60 items-center justify-center rounded-md border text-muted-foreground text-sm">
                无收款码
              </div>
            )}
            <div className="flex items-center gap-2">
              {payTypeBadge(viewRow?.payment_type)}
              <span className="text-sm">
                提现金额：<strong>¥{((viewRow?.amount ?? 0) / 1_000_000).toFixed(4)}</strong>
              </span>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setViewRow(null)}>关闭</Button>
            {viewRow?.status === 'pending' ? (
              <Button onClick={() => executeApprove(viewRow)}>通过申请</Button>
            ) : null}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(rejecting)} onOpenChange={() => setRejecting(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>拒绝提现申请</DialogTitle>
          </DialogHeader>
          <Textarea
            value={remark}
            onChange={(event) => setRemark(event.target.value)}
            placeholder="填写拒绝原因"
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setRejecting(null)}>
              取消
            </Button>
            <Button onClick={reject}>确认拒绝</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={pendingApprove !== null} onOpenChange={() => setPendingApprove(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认通过</AlertDialogTitle>
            <AlertDialogDescription>
              确认通过 {pendingApprove?.username} 的提现申请吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={() => executeApprove()}>通过</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}


export function AdminWithdrawPage() {
  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [listRes, countRes] = await Promise.all([
      adminApi.listWithdrawals(),
      adminApi.getPendingWithdrawCount(),
    ])
    return { rows: listRes.records ?? [], pendingCount: countRes.count ?? 0 }
  }, { rows: [] as AdminWithdrawal[], pendingCount: 0 })

  const rows = data.rows
  const pendingCount = data.pendingCount

  const [mutError, setMutError] = useState('')
  const [rejecting, setRejecting] = useState<AdminWithdrawal | null>(null)
  const [remark, setRemark] = useState('')
  const [pendingApprove, setPendingApprove] = useState<AdminWithdrawal | null>(null)

  const error = loadError || mutError

  async function executeApprove() {
    if (!pendingApprove?.id) return
    setMutError('')
    try {
      await adminApi.approveWithdrawal(pendingApprove.id)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setPendingApprove(null)
    }
  }

  async function reject() {
    if (!rejecting?.id) return
    setMutError('')
    try {
      await adminApi.rejectWithdrawal(rejecting.id, remark)
      setRejecting(null)
      setRemark('')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Withdraw"
        title="提现审核"
        description={`当前待处理 ${pendingCount} 条提现申请。`}
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
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>用户</TableHead>
              <TableHead>申请时间</TableHead>
              <TableHead>金额</TableHead>
              <TableHead>收款方式</TableHead>
              <TableHead>状态</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={7} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="py-10 text-center text-muted-foreground">
                    暂无提现申请
                  </TableCell>
                </TableRow>
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell>{row.id ?? '-'}</TableCell>
                    <TableCell>{row.username ?? '-'}</TableCell>
                    <TableCell>{row.created_at ?? '-'}</TableCell>
                    <TableCell>{((row.amount ?? 0) / 1_000_000).toFixed(4)} 积分</TableCell>
                    <TableCell>{row.payment_type ?? '-'}</TableCell>
                    <TableCell>{row.status ?? '-'}</TableCell>
                    <TableCell className="text-right">
                      {row.status === 'pending' ? (
                        <div className="flex justify-end gap-2">
                          <Button size="sm" variant="outline" onClick={() => setPendingApprove(row)}>
                            通过
                          </Button>
                          <Button size="sm" onClick={() => setRejecting(row)}>
                            拒绝
                          </Button>
                        </div>
                      ) : null}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>

      <Dialog open={Boolean(rejecting)} onOpenChange={() => setRejecting(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>拒绝提现申请</DialogTitle>
          </DialogHeader>
          <Textarea
            value={remark}
            onChange={(event) => setRemark(event.target.value)}
            placeholder="填写拒绝原因"
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setRejecting(null)}>
              取消
            </Button>
            <Button onClick={reject}>确认拒绝</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={pendingApprove !== null} onOpenChange={() => setPendingApprove(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认通过</AlertDialogTitle>
            <AlertDialogDescription>
              确认通过 {pendingApprove?.username} 的提现申请吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={executeApprove}>通过</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
