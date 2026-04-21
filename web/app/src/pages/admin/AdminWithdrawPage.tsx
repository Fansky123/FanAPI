import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
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
import { getApiErrorMessage } from '@/lib/api/http'

export function AdminWithdrawPage() {
  const [rows, setRows] = useState<AdminWithdrawal[]>([])
  const [pendingCount, setPendingCount] = useState(0)
  const [error, setError] = useState('')
  const [rejecting, setRejecting] = useState<AdminWithdrawal | null>(null)
  const [remark, setRemark] = useState('')
  const [pendingApprove, setPendingApprove] = useState<AdminWithdrawal | null>(null)

  async function load() {
    try {
      const [listRes, countRes] = await Promise.all([
        adminApi.listWithdrawals(),
        adminApi.getPendingWithdrawCount(),
      ])
      setRows(listRes.records ?? [])
      setPendingCount(countRes.count ?? 0)
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function approve(row: AdminWithdrawal) {
    if (!row.id) return
    setPendingApprove(row)
  }

  async function executeApprove() {
    if (!pendingApprove?.id) return
    try {
      await adminApi.approveWithdrawal(pendingApprove.id)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setPendingApprove(null)
    }
  }

  async function reject() {
    if (!rejecting?.id) return
    try {
      await adminApi.rejectWithdrawal(rejecting.id, remark)
      setRejecting(null)
      setRemark('')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Withdraw"
        title="提现审核"
        description={`当前待处理 ${pendingCount} 条，已支持通过与拒绝操作。`}
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
          <TableBody>
            {rows.map((row, index) => (
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
                      <Button size="sm" variant="outline" onClick={() => approve(row)}>
                        通过
                      </Button>
                      <Button size="sm" onClick={() => setRejecting(row)}>
                        拒绝
                      </Button>
                    </div>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
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
