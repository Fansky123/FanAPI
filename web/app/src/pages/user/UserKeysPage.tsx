import { useState } from 'react'
import { KeyRoundIcon, PlusIcon, Trash2Icon } from 'lucide-react'

import { EmptyState } from '@/components/shared/EmptyState'
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
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { NativeSelect } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { userApi, type ApiKeyRecord } from '@/lib/api/user'
import { useAsync } from '@/hooks/use-async'

export function UserKeysPage() {
  const { data: keys, loading, error: loadError, reload } = useAsync(async () => {
    const response = await userApi.listApiKeys()
    return Array.isArray(response) ? response : response.api_keys ?? response.keys ?? []
  }, [] as ApiKeyRecord[])

  const [mutError, setMutError] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [createdKey, setCreatedKey] = useState('')
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyType, setNewKeyType] = useState('low_price')
  const [submitting, setSubmitting] = useState(false)
  const [pendingDeleteId, setPendingDeleteId] = useState<number | undefined>()

  const error = loadError || mutError

  async function handleCreate() {
    if (!newKeyName.trim()) {
      setMutError('请输入密钥名称')
      return
    }
    setSubmitting(true)
    setMutError('')
    try {
      const response = await userApi.createApiKey(newKeyName.trim(), newKeyType)
      setCreatedKey(String((response as { key?: string }).key ?? ''))
      setCreateOpen(false)
      setNewKeyName('')
      setNewKeyType('low_price')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function executeDelete() {
    if (!pendingDeleteId) return
    setMutError('')
    try {
      await userApi.deleteApiKey(pendingDeleteId)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setPendingDeleteId(undefined)
    }
  }

  function copyText(text: string) {
    navigator.clipboard.writeText(text).catch(() => undefined)
  }

  return (
    <>
      <PageHeader
        eyebrow="Security"
        title="API 密钥"
        description="管理用于 API 调用鉴权的密钥，创建后的完整密钥只会展示一次。"
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <PlusIcon data-icon="inline-start" />
            新建密钥
          </Button>
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {loading ? (
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableSkeleton cols={5} rows={3} />
          </Table>
        </Card>
      ) : keys.length === 0 ? (
        <EmptyState
          icon={<KeyRoundIcon className="size-6 text-muted-foreground" />}
          title="还没有 API 密钥"
          description="点击右上角「新建密钥」即可创建，生成后的完整密钥只会展示一次。"
        />
      ) : (
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keys.map((item, index) => (
                <TableRow key={item.id ?? index}>
                  <TableCell className="font-medium">{item.name ?? '未命名'}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {item.viewable
                      ? (item.raw_key ?? item.key ?? '***')
                      : item.key_prefix
                        ? `${item.key_prefix}...`
                        : (item.masked_key ?? '***')}
                  </TableCell>
                  <TableCell>{item.key_type ?? 'low_price'}</TableCell>
                  <TableCell>
                    <Badge variant={item.is_active === false ? 'secondary' : 'default'}>
                      {item.is_active === false ? '停用' : '启用'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      {item.viewable && (item.raw_key || item.key) ? (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => copyText((item.raw_key ?? item.key) as string)}
                        >
                          复制
                        </Button>
                      ) : null}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setPendingDeleteId(item.id)}
                      >
                        <Trash2Icon data-icon="inline-start" />
                        删除
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      )}

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建 API 密钥</DialogTitle>
            <DialogDescription>
              创建后会返回一次性明文，关闭后只能看到遮罩形式。
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <Label>名称</Label>
              <Input
                value={newKeyName}
                onChange={(event) => setNewKeyName(event.target.value)}
                placeholder="例如：我的项目"
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label>类型</Label>
              <NativeSelect
                value={newKeyType}
                onChange={(event) => setNewKeyType(event.target.value)}
              >
                <option value="low_price">低价密钥</option>
                <option value="stable">稳定密钥</option>
              </NativeSelect>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              取消
            </Button>
            <Button onClick={handleCreate} disabled={submitting}>
              {submitting ? '创建中...' : '创建'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={Boolean(createdKey)} onOpenChange={() => setCreatedKey('')}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>密钥已创建</DialogTitle>
            <DialogDescription>请立即复制保存，这个明文值后续不会再次展示。</DialogDescription>
          </DialogHeader>
          <div className="rounded-xl border border-border/70 bg-muted/25 p-4 font-mono text-xs break-all">
            {createdKey}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreatedKey('')}>
              关闭
            </Button>
            <Button onClick={() => copyText(createdKey)}>复制密钥</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={pendingDeleteId !== undefined}
        onOpenChange={() => setPendingDeleteId(undefined)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确认永久删除该 API Key 吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={executeDelete}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}


