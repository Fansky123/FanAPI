import { useState } from 'react'

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
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
