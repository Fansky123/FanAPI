import { useRef, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
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
}

export function UserInvitePage() {
  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [inviteRes, qrRes, historyRes] = await Promise.all([
      userApi.getInviteInfo(),
      userApi.getPaymentQR(),
      userApi.listWithdrawHistory(),
    ])
    return {
      info: inviteRes,
      wechatQr: qrRes.wechat_qr ?? '',
      alipayQr: qrRes.alipay_qr ?? '',
      withdrawals: Array.isArray(historyRes)
        ? historyRes
        : historyRes.records ?? historyRes.list ?? [],
    } satisfies InviteData
  }, { info: {}, wechatQr: '', alipayQr: '', withdrawals: [] } as InviteData)

  const [mutError, setMutError] = useState('')
  const [convertOpen, setConvertOpen] = useState(false)
  const [withdrawOpen, setWithdrawOpen] = useState(false)
  const [amount, setAmount] = useState('0')
  const [paymentType, setPaymentType] = useState('wechat')
  const [wechatQrEdit, setWechatQrEdit] = useState('')
  const [alipayQrEdit, setAlipayQrEdit] = useState('')
  const [qrInitialized, setQrInitialized] = useState(false)
  const wechatUploadRef = useRef<HTMLInputElement>(null)
  const alipayUploadRef = useRef<HTMLInputElement>(null)

  // Sync QR fields from loaded data once
  if (!loading && !qrInitialized && (data.wechatQr || data.alipayQr)) {
    setWechatQrEdit(data.wechatQr)
    setAlipayQrEdit(data.alipayQr)
    setQrInitialized(true)
  }

  const error = loadError || mutError

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
      await userApi.convertFrozen(Number(amount))
      setConvertOpen(false)
      setAmount('0')
    })
  }

  async function saveQr() {
    await withMut(async () => { await userApi.savePaymentQR({ wechat_qr: wechatQrEdit, alipay_qr: alipayQrEdit }) })
  }

  async function submitWithdraw() {
    await withMut(async () => {
      await userApi.submitWithdraw(Number(amount), paymentType)
      setWithdrawOpen(false)
      setAmount('0')
    })
  }

  const info = data.info

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
            <Button
              size="sm"
              disabled={loading || !info.invite_code}
              onClick={() => navigator.clipboard.writeText(info.invite_code ?? '')}
            >
              复制邀请码
            </Button>
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
                    <TableCell>{row.payment_type ?? '-'}</TableCell>
                    <TableCell>{row.status ?? '-'}</TableCell>
                    <TableCell>{row.admin_remark ?? '-'}</TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>

      <Dialog open={convertOpen} onOpenChange={setConvertOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>解冻积分</DialogTitle>
          </DialogHeader>
          <Input
            value={amount}
            onChange={(event) => setAmount(event.target.value)}
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
              value={amount}
              onChange={(event) => setAmount(event.target.value)}
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
