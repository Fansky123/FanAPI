import { useEffect, useState } from 'react'
import { KeyRoundIcon, PlusIcon, Trash2Icon } from 'lucide-react'

import { EmptyState } from '@/components/shared/EmptyState'
import { PageHeader } from '@/components/shared/PageHeader'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { Input } from '@/components/ui/input'
import { userApi, type ApiKeyRecord } from '@/lib/api/user'

export function UserKeysPage() {
  const [keys, setKeys] = useState<ApiKeyRecord[]>([])
  const [error, setError] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [createdKey, setCreatedKey] = useState('')
  const [newKeyName, setNewKeyName] = useState('')
  const [newKeyType, setNewKeyType] = useState('low_price')
  const [submitting, setSubmitting] = useState(false)

  async function load() {
    try {
      const response = await userApi.listApiKeys()
      setKeys(Array.isArray(response) ? response : response.api_keys ?? response.keys ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function handleCreate() {
    if (!newKeyName.trim()) {
      setError('请输入密钥名称')
      return
    }
    setSubmitting(true)
    try {
      const response = await userApi.createApiKey(newKeyName.trim(), newKeyType)
      setCreatedKey(String((response as { key?: string }).key ?? ''))
      setCreateOpen(false)
      setNewKeyName('')
      setNewKeyType('low_price')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDelete(id: number | undefined) {
    if (!id) return
    if (!window.confirm('确认永久删除该 API Key 吗？')) return
    try {
      await userApi.deleteApiKey(id)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
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
        description="现在已支持创建和删除密钥，确保用户能在新前端中完成真实调用前的准备工作。"
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <PlusIcon data-icon="inline-start" />
            新建密钥
          </Button>
        }
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      {keys.length === 0 ? (
        <EmptyState
          icon={<KeyRoundIcon className="size-6 text-muted-foreground" />}
          title="还没有 API 密钥"
          description="点击右上角即可创建，生成后的完整密钥只会展示一次。"
        />
      ) : (
        <Card>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>类型</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keys.map((item, index) => (
                <TableRow key={item.id ?? index}>
                  <TableCell className="font-medium">{item.name ?? '未命名'}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {item.key ?? item.masked_key ?? '***'}
                  </TableCell>
                  <TableCell>{item.key_type ?? 'low_price'}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      {item.key ? (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => copyText(item.key as string)}
                        >
                          复制
                        </Button>
                      ) : null}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDelete(item.id)}
                      >
                      <Trash2Icon data-icon="inline-end" />
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
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">名称</label>
              <Input
                value={newKeyName}
                onChange={(event) => setNewKeyName(event.target.value)}
                placeholder="例如：我的项目"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">类型</label>
              <select
                className="flex h-9 w-full rounded-lg border border-input bg-transparent px-3 text-sm outline-none"
                value={newKeyType}
                onChange={(event) => setNewKeyType(event.target.value)}
              >
                <option value="low_price">低价密钥</option>
                <option value="stable">稳定密钥</option>
              </select>
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
    </>
  )
}
