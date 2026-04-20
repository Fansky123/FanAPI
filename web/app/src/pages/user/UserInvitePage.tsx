import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
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
import { Textarea } from '@/components/ui/textarea'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type InviteInfo, type WithdrawRecord } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'

export function UserInvitePage() {
  const [info, setInfo] = useState<InviteInfo>({})
  const [withdrawals, setWithdrawals] = useState<WithdrawRecord[]>([])
  const [error, setError] = useState('')
  const [convertOpen, setConvertOpen] = useState(false)
  const [withdrawOpen, setWithdrawOpen] = useState(false)
  const [amount, setAmount] = useState('0')
  const [paymentType, setPaymentType] = useState('wechat')
  const [wechatQr, setWechatQr] = useState('')
  const [alipayQr, setAlipayQr] = useState('')

  async function load() {
    try {
      const [inviteRes, qrRes, historyRes] = await Promise.all([
        userApi.getInviteInfo(),
        userApi.getPaymentQR(),
        userApi.listWithdrawHistory(),
      ])
      setInfo(inviteRes)
      setWechatQr(qrRes.wechat_qr ?? '')
      setAlipayQr(qrRes.alipay_qr ?? '')
      setWithdrawals(
        Array.isArray(historyRes) ? historyRes : historyRes.records ?? historyRes.list ?? []
      )
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function convert() {
    try {
      await userApi.convertFrozen(Number(amount))
      setConvertOpen(false)
      setAmount('0')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function saveQr() {
    try {
      await userApi.savePaymentQR({ wechat_qr: wechatQr, alipay_qr: alipayQr })
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function submitWithdraw() {
    try {
      await userApi.submitWithdraw(Number(amount), paymentType)
      setWithdrawOpen(false)
      setAmount('0')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Invite"
        title="邀请中心"
        description="支持查看邀请码、冻结返佣、解冻积分和提现申请。"
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader><CardTitle>邀请码</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="font-mono">{info.invite_code ?? '-'}</div>
            <Button size="sm" onClick={() => navigator.clipboard.writeText(info.invite_code ?? '')}>
              复制邀请码
            </Button>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>已邀请人数</CardTitle></CardHeader>
          <CardContent className="text-2xl font-semibold">{info.invite_count ?? 0}</CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle>冻结返佣</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <div className="text-2xl font-semibold">{formatCredits(info.frozen_balance ?? 0)}</div>
            <div className="flex gap-2">
              <Button size="sm" onClick={() => setConvertOpen(true)}>解冻</Button>
              <Button size="sm" variant="outline" onClick={() => setWithdrawOpen(true)}>提现</Button>
            </div>
          </CardContent>
        </Card>
      </div>
      <Card>
        <CardHeader><CardTitle>收款码</CardTitle></CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          <Textarea value={wechatQr} onChange={(event) => setWechatQr(event.target.value)} placeholder="微信收款码（URL 或 base64）" />
          <Textarea value={alipayQr} onChange={(event) => setAlipayQr(event.target.value)} placeholder="支付宝收款码（URL 或 base64）" />
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
          <TableBody>
            {withdrawals.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.created_at ?? '-'}</TableCell>
                <TableCell>{formatCredits(row.amount ?? 0)}</TableCell>
                <TableCell>{row.payment_type ?? '-'}</TableCell>
                <TableCell>{row.status ?? '-'}</TableCell>
                <TableCell>{row.admin_remark ?? '-'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      <Dialog open={convertOpen} onOpenChange={setConvertOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>解冻积分</DialogTitle></DialogHeader>
          <Input value={amount} onChange={(event) => setAmount(event.target.value)} placeholder="0 表示全部" />
          <DialogFooter>
            <Button variant="outline" onClick={() => setConvertOpen(false)}>取消</Button>
            <Button onClick={convert}>确认解冻</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={withdrawOpen} onOpenChange={setWithdrawOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>申请提现</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <Input value={amount} onChange={(event) => setAmount(event.target.value)} placeholder="提现积分数量" />
            <select className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none" value={paymentType} onChange={(event) => setPaymentType(event.target.value)}>
              <option value="wechat">微信</option>
              <option value="alipay">支付宝</option>
            </select>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setWithdrawOpen(false)}>取消</Button>
            <Button onClick={submitWithdraw}>提交提现</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
