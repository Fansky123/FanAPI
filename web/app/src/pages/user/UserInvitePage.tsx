import { useEffect, useRef, useState } from 'react'
import { toast } from 'sonner'

import { PageHeader } from '@/components/shared/PageHeader'
import { TablePagination } from '@/components/shared/TablePagination'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { NativeSelect } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { userApi, type InviteInfo, type WithdrawRecord } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

type InviteData = {
  info: InviteInfo
  wechatQr: string
  alipayQr: string
  withdrawals: WithdrawRecord[]
  withdrawalsTotal: number
}

const withdrawPageSize = 20

function withdrawStatusBadge(status: string | undefined) {
  if (status === 'pending') return <Badge className="bg-yellow-500 hover:bg-yellow-500 text-white">待审核</Badge>
  if (status === 'approved') return <Badge className="bg-emerald-600 hover:bg-emerald-600 text-white">已通过</Badge>
  if (status === 'rejected') return <Badge variant="destructive">已拒绝</Badge>
  return <Badge variant="outline">{status ?? '-'}</Badge>
}

function paymentTypeLabel(type: string | undefined) {
  if (type === 'wechat') return '微信'
  if (type === 'alipay') return '支付宝'
  return type ?? '-'
}

export function UserInvitePage() {
  const [historyPage, setHistoryPage] = useState(1)

  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [inviteRes, qrRes, historyRes] = await Promise.all([
      userApi.getInviteInfo(),
      userApi.getPaymentQR(),
      userApi.listWithdrawHistory(historyPage, withdrawPageSize),
    ])
    const withdrawals = Array.isArray(historyRes)
      ? historyRes
      : historyRes.records ?? historyRes.list ?? []
    return {
      info: inviteRes,
      wechatQr: qrRes.wechat_qr ?? '',
      alipayQr: qrRes.alipay_qr ?? '',
      withdrawals,
      withdrawalsTotal: Array.isArray(historyRes)
        ? withdrawals.length
        : historyRes.total ?? withdrawals.length,
    } satisfies InviteData
  }, { info: {}, wechatQr: '', alipayQr: '', withdrawals: [], withdrawalsTotal: 0 } as InviteData, [historyPage])

  const [mutError, setMutError] = useState('')
  const [convertOpen, setConvertOpen] = useState(false)
  const [withdrawOpen, setWithdrawOpen] = useState(false)
  const [convertAmount, setConvertAmount] = useState('0')
  const [withdrawAmount, setWithdrawAmount] = useState('0')
  const [paymentType, setPaymentType] = useState('wechat')
  const [wechatQrEdit, setWechatQrEdit] = useState('')
  const [alipayQrEdit, setAlipayQrEdit] = useState('')
  const [qrInitialized, setQrInitialized] = useState(false)
  const wechatUploadRef = useRef<HTMLInputElement>(null)
  const alipayUploadRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (loading || qrInitialized) {
      return
    }
    setWechatQrEdit(data.wechatQr)
    setAlipayQrEdit(data.alipayQr)
    setQrInitialized(true)
  }, [data.alipayQr, data.wechatQr, loading, qrInitialized])

  const error = loadError || mutError
  const info = data.info
  const inviteLink = info.invite_code ? `${window.location.origin}/register?ref=${info.invite_code}` : ''

  async function withMut(fn: () => Promise<void>) {
    setMutError('')
    try {
      await fn()
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  function readLocalFile(file: File, onDone: (base64: string) => void) {
    const reader = new FileReader()
    reader.onload = (event) => {
      const result = event.target?.result
      if (typeof result === 'string') {
        onDone(result)
      }
    }
    reader.readAsDataURL(file)
  }

  function pickQr(kind: 'wechat' | 'alipay', file: File | undefined) {
    if (!file) {
      return
    }

    readLocalFile(file, (base64) => {
      if (kind === 'wechat') {
        setWechatQrEdit(base64)
        return
      }
      setAlipayQrEdit(base64)
    })
  }

  async function convert() {
    await withMut(async () => {
      await userApi.convertFrozen(Number(convertAmount))
      setConvertOpen(false)
      setConvertAmount('0')
    })
  }

  async function saveQr() {
    await withMut(async () => {
      await userApi.savePaymentQR({ wechat_qr: wechatQrEdit, alipay_qr: alipayQrEdit })
      toast.success('收款码已保存')
    })
  }

  async function submitWithdraw() {
    const amountValue = Number(withdrawAmount)
    const currentQr = paymentType === 'wechat' ? wechatQrEdit.trim() : alipayQrEdit.trim()
    if (!amountValue || amountValue < 0) {
      setMutError('请输入有效的提现积分数量')
      return
    }
    if (!currentQr) {
      setMutError(`请先保存${paymentType === 'wechat' ? '微信' : '支付宝'}收款码`)
      return
    }
    await withMut(async () => {
      await userApi.submitWithdraw(amountValue, paymentType)
      setWithdrawOpen(false)
      setWithdrawAmount('0')
      setHistoryPage(1)
    })
  }

  return (
    <>
      <PageHeader
        eyebrow="Invite"
        title="邀请中心"
        description="查看邀请码、冻结返佣、解冻积分和提现申请。"
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
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>邀请码</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2 text-sm">
            <div className="font-mono">{loading ? '加载中...' : (info.invite_code ?? '-')}</div>
            <div className="flex flex-wrap gap-2">
              <Button
                size="sm"
                disabled={loading || !info.invite_code}
                onClick={() => navigator.clipboard.writeText(info.invite_code ?? '')}
              >
                复制邀请码
              </Button>
              <Button
                size="sm"
                variant="outline"
                disabled={loading || !inviteLink}
                onClick={() => navigator.clipboard.writeText(inviteLink)}
              >
                复制邀请链接
              </Button>
            </div>
            <div className="rounded-lg border border-dashed border-border/70 bg-muted/20 px-3 py-2 text-xs text-muted-foreground">
              {loading ? '邀请链接加载中...' : (inviteLink || '暂无邀请链接')}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>已邀请人数</CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold">
            {loading ? '-' : (info.invite_count ?? 0)}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>冻结返佣</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-3">
            <div className="text-2xl font-semibold">
              {loading ? '-' : formatCredits(info.frozen_balance ?? 0)}
            </div>
            <div className="flex gap-2">
              <Button size="sm" disabled={loading} onClick={() => setConvertOpen(true)}>
                解冻
              </Button>
              <Button
                size="sm"
                variant="outline"
                disabled={loading}
                onClick={() => setWithdrawOpen(true)}
              >
                提现
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>收款码</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm font-medium">微信收款码</span>
              <div className="flex items-center gap-2">
                <input
                  ref={wechatUploadRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(event) => {
                    pickQr('wechat', event.target.files?.[0])
                    event.target.value = ''
                  }}
                />
                <Button size="sm" variant="outline" onClick={() => wechatUploadRef.current?.click()}>
                  本地上传
                </Button>
                {wechatQrEdit ? (
                  <Button size="sm" variant="ghost" onClick={() => setWechatQrEdit('')}>
                    清空
                  </Button>
                ) : null}
              </div>
            </div>
            <Textarea
              value={wechatQrEdit}
              onChange={(event) => setWechatQrEdit(event.target.value)}
              placeholder="微信收款码（URL 或 base64）"
            />
            <div className="rounded-xl border border-dashed border-border/80 bg-muted/20 p-3">
              {wechatQrEdit ? (
                <img
                  src={wechatQrEdit}
                  alt="微信收款码预览"
                  className="max-h-56 rounded-md border bg-background object-contain"
                />
              ) : (
                <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
                  暂无微信收款码
                </div>
              )}
            </div>
          </div>
          <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm font-medium">支付宝收款码</span>
              <div className="flex items-center gap-2">
                <input
                  ref={alipayUploadRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(event) => {
                    pickQr('alipay', event.target.files?.[0])
                    event.target.value = ''
                  }}
                />
                <Button size="sm" variant="outline" onClick={() => alipayUploadRef.current?.click()}>
                  本地上传
                </Button>
                {alipayQrEdit ? (
                  <Button size="sm" variant="ghost" onClick={() => setAlipayQrEdit('')}>
                    清空
                  </Button>
                ) : null}
              </div>
            </div>
            <Textarea
              value={alipayQrEdit}
              onChange={(event) => setAlipayQrEdit(event.target.value)}
              placeholder="支付宝收款码（URL 或 base64）"
            />
            <div className="rounded-xl border border-dashed border-border/80 bg-muted/20 p-3">
              {alipayQrEdit ? (
                <img
                  src={alipayQrEdit}
                  alt="支付宝收款码预览"
                  className="max-h-56 rounded-md border bg-background object-contain"
                />
              ) : (
                <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">
                  暂无支付宝收款码
                </div>
              )}
            </div>
          </div>
          <div className="md:col-span-2">
            <Button onClick={saveQr}>保存收款码</Button>
          </div>
        </CardContent>
      </Card>
      <Alert>
        <AlertDescription>
          您邀请的用户每次消费后，返佣积分会先冻结到邀请账户；您可以解冻为可用积分，也可以在保存收款码后发起提现申请。
        </AlertDescription>
      </Alert>
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>时间</TableHead>
              <TableHead>积分数量</TableHead>
              <TableHead>收款方式</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>备注</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={5} rows={3} />
          ) : (
            <TableBody>
              {data.withdrawals.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="py-10 text-center text-muted-foreground">
                    暂无提现记录
                  </TableCell>
                </TableRow>
              ) : (
                data.withdrawals.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                    </TableCell>
                    <TableCell>{formatCredits(row.amount ?? 0)}</TableCell>
                    <TableCell>{paymentTypeLabel(row.payment_type)}</TableCell>
                    <TableCell>{withdrawStatusBadge(row.status)}</TableCell>
                    <TableCell>{row.admin_remark ?? '-'}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
        {!loading && data.withdrawalsTotal > withdrawPageSize ? (
          <CardContent className="border-t">
            <TablePagination
              current={historyPage}
              total={data.withdrawalsTotal}
              pageSize={withdrawPageSize}
              onChange={setHistoryPage}
            />
          </CardContent>
        ) : null}
      </Card>

      <Dialog open={convertOpen} onOpenChange={setConvertOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>解冻积分</DialogTitle>
          </DialogHeader>
          <Input
            value={convertAmount}
            onChange={(event) => setConvertAmount(event.target.value)}
            placeholder="0 表示全部"
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setConvertOpen(false)}>
              取消
            </Button>
            <Button onClick={convert}>确认解冻</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={withdrawOpen} onOpenChange={setWithdrawOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>申请提现</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4">
            <Input
              value={withdrawAmount}
              onChange={(event) => setWithdrawAmount(event.target.value)}
              placeholder="提现积分数量"
            />
            <NativeSelect
              value={paymentType}
              onChange={(event) => setPaymentType(event.target.value)}
            >
              <option value="wechat">微信</option>
              <option value="alipay">支付宝</option>
            </NativeSelect>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setWithdrawOpen(false)}>
              取消
            </Button>
            <Button onClick={submitWithdraw}>提交提现</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
