import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
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
import { adminApi, type AdminKeyPool, type AdminPoolKey } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

export function AdminKeyPoolsPage() {
  const [pools, setPools] = useState<AdminKeyPool[]>([])
  const [keys, setKeys] = useState<AdminPoolKey[]>([])
  const [activePool, setActivePool] = useState<AdminKeyPool | null>(null)
  const [error, setError] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [keyOpen, setKeyOpen] = useState(false)
  const [name, setName] = useState('')
  const [channelId, setChannelId] = useState('1')
  const [keyValue, setKeyValue] = useState('')
  const [priority, setPriority] = useState('0')

  async function loadPools() {
    try {
      const response = await adminApi.listKeyPools()
      setPools(Array.isArray(response) ? response : response.pools ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadPools()
  }, [])

  async function openKeys(pool: AdminKeyPool) {
    setActivePool(pool)
    setKeyOpen(true)
    try {
      const response = await adminApi.listPoolKeys(pool.id as number)
      setKeys(Array.isArray(response) ? response : response.keys ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function createPool() {
    try {
      await adminApi.createKeyPool({ channel_id: Number(channelId), name })
      setCreateOpen(false)
      setName('')
      setChannelId('1')
      await loadPools()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function togglePool(pool: AdminKeyPool) {
    if (!pool.id) return
    try {
      await adminApi.toggleKeyPool(pool.id)
      await loadPools()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function toggleVendor(pool: AdminKeyPool) {
    if (!pool.id) return
    try {
      await adminApi.toggleVendorSubmittable(pool.id)
      await loadPools()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function deletePool(pool: AdminKeyPool) {
    if (!pool.id) return
    if (!window.confirm(`确认删除号池 ${pool.name ?? pool.id} 吗？`)) return
    try {
      await adminApi.deleteKeyPool(pool.id)
      await loadPools()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function addKey() {
    if (!activePool?.id) return
    try {
      await adminApi.addPoolKey(activePool.id, { value: keyValue, priority: Number(priority) })
      setKeyValue('')
      setPriority('0')
      await openKeys(activePool)
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function updateKey(row: AdminPoolKey) {
    if (!row.id) return
    try {
      await adminApi.updatePoolKey(row.id, {
        priority: row.priority ?? 0,
        is_active: row.is_active ?? true,
      })
      await openKeys(activePool as AdminKeyPool)
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function removeKey(row: AdminPoolKey) {
    if (!row.id) return
    try {
      await adminApi.removePoolKey(row.id)
      await openKeys(activePool as AdminKeyPool)
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Key Pools"
        title="号池管理"
        description="支持创建号池、启停、切换号商上传，以及管理池内 Key。"
        actions={<Button onClick={() => setCreateOpen(true)}>新建号池</Button>}
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>ID</TableHead>
              <TableHead>名称</TableHead>
              <TableHead>渠道 ID</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>号商上传</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pools.map((pool, index) => (
              <TableRow key={pool.id ?? index}>
                <TableCell>{pool.id ?? '-'}</TableCell>
                <TableCell>{pool.name ?? '-'}</TableCell>
                <TableCell>{pool.channel_id ?? '-'}</TableCell>
                <TableCell>
                  <Badge variant={pool.is_active ? 'default' : 'secondary'}>
                    {pool.is_active ? '启用' : '停用'}
                  </Badge>
                </TableCell>
                <TableCell>
                  <Badge variant={pool.vendor_submittable ? 'default' : 'secondary'}>
                    {pool.vendor_submittable ? '开放' : '关闭'}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button size="sm" variant="outline" onClick={() => openKeys(pool)}>
                      管理 Keys
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => togglePool(pool)}>
                      {pool.is_active ? '停用' : '启用'}
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => toggleVendor(pool)}>
                      切上传
                    </Button>
                    <Button size="sm" onClick={() => deletePool(pool)}>
                      删除
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>新建号池</DialogTitle></DialogHeader>
          <div className="space-y-4">
            <Input value={channelId} onChange={(event) => setChannelId(event.target.value)} placeholder="渠道 ID" />
            <Input value={name} onChange={(event) => setName(event.target.value)} placeholder="号池名称" />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>取消</Button>
            <Button onClick={createPool}>创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={keyOpen} onOpenChange={setKeyOpen}>
        <DialogContent className="max-w-3xl">
          <DialogHeader><DialogTitle>{activePool?.name ?? ''} - Key 管理</DialogTitle></DialogHeader>
          <div className="flex gap-3">
            <Input value={keyValue} onChange={(event) => setKeyValue(event.target.value)} placeholder="Key 值" />
            <Input value={priority} onChange={(event) => setPriority(event.target.value)} placeholder="优先级" />
            <Button onClick={addKey}>添加 Key</Button>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>优先级</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keys.map((row, index) => (
                <TableRow key={row.id ?? index}>
                  <TableCell>{row.id ?? '-'}</TableCell>
                  <TableCell className="font-mono text-xs">{row.value ?? '-'}</TableCell>
                  <TableCell>{row.priority ?? 0}</TableCell>
                  <TableCell>{row.is_active === false ? '停用' : '启用'}</TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-2">
                      <Button size="sm" variant="outline" onClick={() => updateKey(row)}>保存</Button>
                      <Button size="sm" onClick={() => removeKey(row)}>删除</Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </DialogContent>
      </Dialog>
    </>
  )
}
