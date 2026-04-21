import { useMemo, useState } from 'react'
import { CopyIcon, SaveIcon } from 'lucide-react'

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
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { agentApi, type AgentUser } from '@/lib/api/agent'
import { useAsync } from '@/hooks/use-async'

function formatCredits(n: number | undefined) {
  if (!n) return '¥0.0000'
  return `¥${(n / 1e6).toFixed(4)}`
}

export function AgentDashboardPage() {
  const [page, setPage] = useState(1)
  const pageSize = 50

  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [usersRes, inviteRes] = await Promise.all([
      agentApi.listUsers(page, pageSize),
      agentApi.getInvite(),
    ])
    const rawUsers = Array.isArray(usersRes) ? usersRes : usersRes.users ?? usersRes.items ?? []
    const total = Array.isArray(usersRes) ? rawUsers.length : (usersRes as { total?: number }).total ?? rawUsers.length
    return {
      users: rawUsers,
      total,
      inviteCode: inviteRes.invite_code ?? '',
      wechatQR: inviteRes.wechat_qr ?? '',
    }
  }, { users: [] as AgentUser[], total: 0, inviteCode: '', wechatQR: '' }, [page])

  const [mutError, setMutError] = useState('')
  const error = loadError || mutError

  // 充值弹窗
  const [rechargeTarget, setRechargeTarget] = useState<AgentUser | undefined>()
  const [rechargeAmount, setRechargeAmount] = useState('1000000')
  const [recharging, setRecharging] = useState(false)

  // 二维码弹窗
  const [qrOpen, setQrOpen] = useState(false)
  const [qrInput, setQrInput] = useState('')
  const [savingQR, setSavingQR] = useState(false)

  const inviteLink = useMemo(() => {
    if (!data.inviteCode) return ''
    return `${location.origin}/register?ref=${data.inviteCode}`
  }, [data.inviteCode])

  function copyLink() {
    if (!inviteLink) return
    navigator.clipboard.writeText(inviteLink).catch(() => {
      const el = document.createElement('input')
      el.value = inviteLink
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
    })
  }

  function openRecharge(user: AgentUser) {
    setRechargeTarget(user)
    setRechargeAmount('1000000')
    setMutError('')
  }

  async function doRecharge() {
    if (!rechargeTarget?.id) return
    setRecharging(true)
    setMutError('')
    try {
      await agentApi.rechargeUser(rechargeTarget.id, Number(rechargeAmount || '0'))
      setRechargeTarget(undefined)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setRecharging(false)
    }
  }

  async function saveQR() {
    if (!qrInput.trim()) return
    setSavingQR(true)
    setMutError('')
    try {
      await agentApi.updateWechatQR(qrInput.trim())
      setQrOpen(false)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setSavingQR(false)
    }
  }

  const totalPages = Math.ceil(data.total / pageSize)

  return (
    <>
      <PageHeader
        eyebrow="Agent"
        title="Agent 工作台"
        description="管理您名下的用户，查看充值消费明细。"
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

      {/* 邀请链接卡片 */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="text-xs font-semibold uppercase tracking-wide text-emerald-600 mb-1">Invite Link</p>
              <CardTitle className="text-base">我的邀请链接</CardTitle>
              <p className="mt-1 text-sm text-muted-foreground">
                将链接发给客户，通过链接注册的用户将自动绑定到您名下，您可对其进行充值管理。
              </p>
            </div>
            <Badge variant="secondary" className="shrink-0 whitespace-nowrap">
              已邀请 {data.total} 人
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2">
            <Input
              readOnly
              value={inviteLink || '加载中...'}
              className="font-mono text-xs"
            />
            <Button variant="outline" size="sm" onClick={copyLink} disabled={!inviteLink}>
              <CopyIcon className="mr-1 h-3.5 w-3.5" />
              复制
            </Button>
            <Button variant="outline" size="sm" onClick={() => { setQrInput(data.wechatQR); setQrOpen(true) }}>
              微信二维码
            </Button>
          </div>
          {data.wechatQR ? (
            <div className="mt-3 flex items-center gap-2">
              <img src={data.wechatQR} alt="微信二维码" className="h-10 w-10 rounded border object-contain p-0.5" />
              <span className="text-xs text-muted-foreground">已设置微信二维码</span>
            </div>
          ) : null}
        </CardContent>
      </Card>

      {/* 用户列表 */}
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-baseline gap-3">
            <CardTitle className="text-base">我邀请的用户</CardTitle>
            <span className="text-xs text-muted-foreground">余额不足 ¥1 的用户以红色标出</span>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-16">ID</TableHead>
                <TableHead>用户名</TableHead>
                <TableHead>邮箱</TableHead>
                <TableHead className="text-right">当前余额</TableHead>
                <TableHead className="text-right">累计充值</TableHead>
                <TableHead className="text-right">累计消费</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            {loading ? (
              <TableSkeleton cols={7} />
            ) : (
              <TableBody>
                {data.users.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="py-10 text-center text-muted-foreground">
                      暂无邀请用户
                    </TableCell>
                  </TableRow>
                ) : (
                  data.users.map((row, index) => {
                    const balance = row.balance ?? (row.balance_credits ? row.balance_credits * 1e6 : 0)
                    const isLow = balance < 1000000
                    return (
                      <TableRow key={row.id ?? index} className={isLow ? 'bg-red-50/50' : undefined}>
                        <TableCell className="text-muted-foreground">{row.id ?? '-'}</TableCell>
                        <TableCell className="font-medium">{row.username ?? '-'}</TableCell>
                        <TableCell className="text-muted-foreground">{row.email ?? '—'}</TableCell>
                        <TableCell className={`text-right font-mono text-sm ${isLow ? 'font-bold text-red-500' : ''}`}>
                          {formatCredits(balance)}
                        </TableCell>
                        <TableCell className="text-right font-mono text-sm text-emerald-600">
                          {formatCredits(row.total_recharge)}
                        </TableCell>
                        <TableCell className="text-right font-mono text-sm text-amber-600">
                          {formatCredits(row.total_spend)}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button size="sm" variant="outline" onClick={() => openRecharge(row)}>
                            充值
                          </Button>
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            )}
          </Table>
        </CardContent>
        {totalPages > 1 ? (
          <div className="flex items-center justify-end gap-2 border-t px-4 py-3">
            <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              上一页
            </Button>
            <span className="text-sm text-muted-foreground">
              第 {page} / {totalPages} 页
            </span>
            <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
              下一页
            </Button>
          </div>
        ) : null}
      </Card>

      {/* 充值弹窗 */}
      <AlertDialog open={rechargeTarget !== undefined} onOpenChange={() => setRechargeTarget(undefined)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>为用户充值</AlertDialogTitle>
            <AlertDialogDescription>
              为用户 <strong>{rechargeTarget?.username}</strong> 手动充值积分（credits）。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="space-y-3 py-2">
            <div className="space-y-2">
              <label className="text-sm font-medium">充值金额（credits）</label>
              <Input
                type="number"
                value={rechargeAmount}
                onChange={(e) => setRechargeAmount(e.target.value)}
                placeholder="如 1000000"
              />
              {rechargeAmount ? (
                <p className="text-xs text-muted-foreground">
                  {Number(rechargeAmount).toLocaleString()} credits ≈ ¥{(Number(rechargeAmount) / 1e6).toFixed(4)}
                </p>
              ) : null}
            </div>
            <div className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => setRechargeAmount('1000000')}>¥1</Button>
              <Button size="sm" variant="outline" onClick={() => setRechargeAmount('10000000')}>¥10</Button>
              <Button size="sm" variant="outline" onClick={() => setRechargeAmount('100000000')}>¥100</Button>
            </div>
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={doRecharge} disabled={recharging || !rechargeAmount || Number(rechargeAmount) <= 0}>
              {recharging ? '充值中...' : '确认充值'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* 微信二维码弹窗 */}
      <Dialog open={qrOpen} onOpenChange={setQrOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>设置微信二维码</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            {data.wechatQR ? (
              <div className="flex justify-center">
                <img
                  src={data.wechatQR}
                  alt="当前微信二维码"
                  className="h-40 w-40 rounded-lg border object-contain p-2"
                />
              </div>
            ) : (
              <div className="flex h-32 items-center justify-center rounded-lg border-2 border-dashed text-sm text-muted-foreground">
                暂未设置二维码
              </div>
            )}
            <div className="space-y-2">
              <label className="text-sm font-medium">图片 URL</label>
              <Input
                value={qrInput}
                onChange={(e) => setQrInput(e.target.value)}
                placeholder="粘贴微信二维码图片 URL"
              />
              <p className="text-xs text-muted-foreground">
                设置后，通过您邀请链接注册并登录的用户将看到此二维码。
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setQrOpen(false)}>取消</Button>
            <Button onClick={saveQR} disabled={savingQR || !qrInput.trim()}>
              <SaveIcon className="mr-1.5 h-4 w-4" />
              {savingQR ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

