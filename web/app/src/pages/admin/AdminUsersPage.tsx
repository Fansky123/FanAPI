import { useState } from 'react'
import { SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { adminApi, type AdminUser } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

type DialogMode = 'recharge' | 'password' | 'group' | 'rebate' | null

function fmtBalance(user: AdminUser) {
  const raw = user.balance ?? (user.balance_credits !== undefined ? user.balance_credits * 1e6 : undefined)
  if (raw === undefined || raw === null) return '-'
  return `¥${(Number(raw) / 1e6).toFixed(4)}`
}

export function AdminUsersPage() {
  const [page, setPage] = useState(1)
  const pageSize = 20

  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const response = await adminApi.listUsers(page, pageSize)
    const users = Array.isArray(response) ? response : response.users ?? response.items ?? []
    const total = Array.isArray(response) ? users.length : (response as { total?: number }).total ?? users.length
    return { users, total }
  }, { users: [] as AdminUser[], total: 0 }, [page])

  const [mutError, setMutError] = useState('')
  const [activeUser, setActiveUser] = useState<AdminUser | null>(null)
  const [dialogMode, setDialogMode] = useState<DialogMode>(null)
  const [value, setValue] = useState('')
  const [confirmPwd, setConfirmPwd] = useState('')
  const [rebatePct, setRebatePct] = useState('')

  const error = loadError || mutError

  function openDialog(user: AdminUser, mode: Exclude<DialogMode, null>) {
    setActiveUser(user)
    setDialogMode(mode)
    setValue(mode === 'group' ? (user.group ?? '') : mode === 'recharge' ? '1000000' : '')
    setConfirmPwd('')
    if (mode === 'rebate') {
      const ratio = user.rebate_ratio
      setRebatePct(ratio != null ? String(parseFloat((ratio * 100).toFixed(2))) : '')
    } else {
      setRebatePct('')
    }
    setMutError('')
  }

  async function submitDialog() {
    if (!activeUser?.id || !dialogMode) return
    if (dialogMode === 'password' && value !== confirmPwd) {
      setMutError('两次密码不一致')
      return
    }
    setMutError('')
    try {
      if (dialogMode === 'recharge') {
        await adminApi.rechargeUser(activeUser.id, Number(value))
      } else if (dialogMode === 'password') {
        await adminApi.resetUserPassword(activeUser.id, value)
      } else if (dialogMode === 'group') {
        await adminApi.setUserGroup(activeUser.id, value)
      } else if (dialogMode === 'rebate') {
        const ratio = rebatePct === '' ? null : parseFloat(rebatePct) / 100
        await adminApi.setUserRebateRatio(activeUser.id, ratio)
      }
      setDialogMode(null)
      setActiveUser(null)
      setValue('')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function toggleAgent(user: AdminUser) {
    if (!user.id) return
    setMutError('')
    const nextRole = user.role === 'agent' ? 'user' : 'agent'
    try {
      await adminApi.setUserRole(user.id, nextRole)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  const totalPages = Math.ceil(data.total / pageSize)

  return (
    <>
      <PageHeader
        eyebrow="Accounts"
        title="用户与余额管理"
        description="查看用户注册状态、余额和手动充值情况，用于日常运营支持。"
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
              <TableHead className="w-14">ID</TableHead>
              <TableHead>用户名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead className="w-16">状态</TableHead>
              <TableHead className="w-20">角色</TableHead>
              <TableHead className="w-28">定价分组</TableHead>
              <TableHead className="w-28">返佣比例</TableHead>
              <TableHead className="w-32">余额（¥）</TableHead>
              <TableHead className="w-40">注册时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={8} />
          ) : (
            <TableBody>
              {data.users.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={10} className="py-10 text-center text-muted-foreground">
                    暂无用户数据
                  </TableCell>
                </TableRow>
              ) : (
                data.users.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell className="text-muted-foreground">{row.id ?? '-'}</TableCell>
                    <TableCell className="font-medium">{row.username ?? '-'}</TableCell>
                    <TableCell className="text-muted-foreground">{row.email ?? '-'}</TableCell>
                    <TableCell>
                      <Badge variant={(row.is_active ?? true) ? 'default' : 'secondary'} className="text-xs">
                        {(row.is_active ?? true) ? '正常' : '禁用'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {row.role ? (
                        <Badge variant={row.role === 'admin' ? 'destructive' : 'secondary'} className="text-xs">
                          {row.role}
                        </Badge>
                      ) : '-'}
                    </TableCell>
                    <TableCell>
                      {row.group ? (
                        <Badge variant="outline" className="text-xs cursor-pointer" onClick={() => openDialog(row, 'group')}>
                          {row.group}
                        </Badge>
                      ) : (
                        <button
                          className="text-xs text-muted-foreground hover:text-foreground underline-offset-2 hover:underline"
                          onClick={() => openDialog(row, 'group')}
                        >
                          默认（点击设置）
                        </button>
                      )}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.rebate_ratio != null
                        ? `${(row.rebate_ratio * 100).toFixed(2)}%`
                        : <span className="text-muted-foreground/50">—（全局）</span>}
                    </TableCell>
                    <TableCell className="font-mono text-sm">{fmtBalance(row)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1.5">
                        <Button size="sm" variant="outline" onClick={() => openDialog(row, 'recharge')}>
                          充值
                        </Button>
                        <Button size="sm" variant="outline" onClick={() => openDialog(row, 'password')}>
                          改密
                        </Button>
                        <Button size="sm" variant="outline" onClick={() => openDialog(row, 'group')}>
                          分组
                        </Button>
                        <Button size="sm" variant="outline" onClick={() => openDialog(row, 'rebate')}>
                          返佣
                        </Button>
                        <Button
                          size="sm"
                          variant={row.role === 'agent' ? 'destructive' : 'secondary'}
                          onClick={() => toggleAgent(row)}
                        >
                          {row.role === 'agent' ? '取消客服' : '设为客服'}
                        </Button>
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
            <span className="text-sm text-muted-foreground">共 {data.total} 位用户</span>
            <div className="flex items-center gap-2">
              <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
                上一页
              </Button>
              <span className="text-sm text-muted-foreground">第 {page} / {totalPages} 页</span>
              <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>
                下一页
              </Button>
            </div>
          </CardContent>
        ) : (
          data.total > 0 ? (
            <CardContent className="border-t py-3">
              <span className="text-sm text-muted-foreground">共 {data.total} 位用户</span>
            </CardContent>
          ) : null
        )}
      </Card>

      <Dialog open={Boolean(dialogMode)} onOpenChange={() => setDialogMode(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {dialogMode === 'recharge'
                ? '手动充值'
                : dialogMode === 'password'
                  ? '重置密码'
                  : dialogMode === 'rebate'
                    ? '设置返佣比例'
                    : '设置定价分组'}
            </DialogTitle>
            <DialogDescription>
              用户：{activeUser?.username ?? activeUser?.email ?? '-'}
            </DialogDescription>
          </DialogHeader>
          {mutError ? (
            <Alert variant="destructive">
              <AlertDescription>{mutError}</AlertDescription>
            </Alert>
          ) : null}
          <div className="flex flex-col gap-3">
            {dialogMode === 'rebate' ? (
              <div className="space-y-1.5">
                <Label>个人返佣比例（%）</Label>
                <Input
                  value={rebatePct}
                  onChange={(event) => setRebatePct(event.target.value)}
                  placeholder="留空=使用全局默认，如：20（代表 20%）"
                />
                <p className="text-xs text-muted-foreground">设置该用户专属的邀请返佣比例，留空则使用系统全局默认值。</p>
              </div>
            ) : (
              <div className="space-y-1.5">
                <Label>
                  {dialogMode === 'recharge'
                    ? '充值积分数（credits）'
                    : dialogMode === 'password'
                      ? '新密码'
                      : '分组名称'}
                </Label>
                <Input
                  value={value}
                  type={dialogMode === 'password' ? 'password' : 'text'}
                  onChange={(event) => setValue(event.target.value)}
                  placeholder={
                    dialogMode === 'recharge'
                      ? '如：1000000（= ¥1）'
                      : dialogMode === 'password'
                        ? '至少 8 位'
                        : '留空=默认定价，如 vip / premium'
                  }
                />
                {dialogMode === 'recharge' && value ? (
                  <p className="text-xs text-muted-foreground">
                    {Number(value).toLocaleString()} credits = ¥{(Number(value) / 1e6).toFixed(6)}
                  </p>
                ) : null}
                {dialogMode === 'group' ? (
                  <p className="text-xs text-muted-foreground">分组名须与渠道 billing_config.pricing_groups 中的键对应</p>
                ) : null}
              </div>
            )}
            {dialogMode === 'password' ? (
              <div className="space-y-1.5">
                <Label>确认密码</Label>
                <Input type="password" value={confirmPwd} onChange={(e) => setConfirmPwd(e.target.value)} placeholder="再次输入" />
              </div>
            ) : null}
            {dialogMode === 'recharge' ? (
              <div className="flex gap-2">
                <Button size="sm" variant="outline" onClick={() => setValue('1000000')}>¥1</Button>
                <Button size="sm" variant="outline" onClick={() => setValue('10000000')}>¥10</Button>
                <Button size="sm" variant="outline" onClick={() => setValue('100000000')}>¥100</Button>
              </div>
            ) : null}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogMode(null)}>
              取消
            </Button>
            <Button onClick={submitDialog}>
              <SaveIcon data-icon="inline-start" />
              确认
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

