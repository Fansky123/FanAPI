import { useEffect, useState } from 'react'
import { SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
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
import { getApiErrorMessage } from '@/lib/api/http'

type DialogMode = 'recharge' | 'password' | 'group' | null

export function AdminUsersPage() {
  const [rows, setRows] = useState<AdminUser[]>([])
  const [error, setError] = useState('')
  const [activeUser, setActiveUser] = useState<AdminUser | null>(null)
  const [dialogMode, setDialogMode] = useState<DialogMode>(null)
  const [value, setValue] = useState('')

  async function load() {
    try {
      const response = await adminApi.listUsers()
      setRows(Array.isArray(response) ? response : response.items ?? response.users ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  function openDialog(user: AdminUser, mode: Exclude<DialogMode, null>) {
    setActiveUser(user)
    setDialogMode(mode)
    setValue(mode === 'group' ? user.group ?? '' : '')
  }

  async function submitDialog() {
    if (!activeUser?.id || !dialogMode) return
    try {
      if (dialogMode === 'recharge') {
        await adminApi.rechargeUser(activeUser.id, Number(value))
      } else if (dialogMode === 'password') {
        await adminApi.resetUserPassword(activeUser.id, value)
      } else if (dialogMode === 'group') {
        await adminApi.setUserGroup(activeUser.id, value)
      }
      setDialogMode(null)
      setActiveUser(null)
      setValue('')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function toggleAgent(user: AdminUser & { role?: string }) {
    if (!user.id) return
    const nextRole = user.role === 'agent' ? 'user' : 'agent'
    try {
      await adminApi.setUserRole(user.id, nextRole)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Accounts"
        title="用户管理"
        description="已补上最小运营动作，后台现在可以直接充值、改密、调分组和切换客服角色。"
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
              <TableHead>用户名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>分组</TableHead>
              <TableHead>余额</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.id ?? '-'}</TableCell>
                <TableCell className="font-medium">{row.username ?? '-'}</TableCell>
                <TableCell>{row.email ?? '-'}</TableCell>
                <TableCell>{row.group ?? '-'}</TableCell>
                <TableCell>{row.balance_credits ?? row.balance ?? '-'}</TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button size="sm" variant="outline" onClick={() => openDialog(row, 'recharge')}>
                      充值
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => openDialog(row, 'password')}>
                      改密
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => openDialog(row, 'group')}>
                      分组
                    </Button>
                    <Button size="sm" onClick={() => toggleAgent(row as AdminUser & { role?: string })}>
                      切角色
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      <Dialog open={Boolean(dialogMode)} onOpenChange={() => setDialogMode(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {dialogMode === 'recharge'
                ? '手动充值'
                : dialogMode === 'password'
                  ? '重置密码'
                  : '设置分组'}
            </DialogTitle>
            <DialogDescription>
              当前用户：{activeUser?.username ?? activeUser?.email ?? '-'}
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <Label>
              {dialogMode === 'recharge'
                ? '充值 credits'
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
                  ? '例如：1000000'
                  : dialogMode === 'password'
                    ? '至少 8 位'
                    : '例如：vip'
              }
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogMode(null)}>
              取消
            </Button>
            <Button onClick={submitDialog}>
              <SaveIcon data-icon="inline-start" />
              保存
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
