import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
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
import { NativeSelect } from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { vendorApi, type VendorKey, type VendorPool } from '@/lib/api/vendor'
import { formatCredits } from '@/lib/formatters/credits'

export function VendorKeysPage() {
  const [keys, setKeys] = useState<VendorKey[]>([])
  const [pools, setPools] = useState<VendorPool[]>([])
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [open, setOpen] = useState(false)
  const [poolId, setPoolId] = useState('')
  const [value, setValue] = useState('')

  async function load() {
    try {
      const [keysRes, poolsRes] = await Promise.all([
        vendorApi.getKeys(),
        vendorApi.getPools(),
      ])
      setKeys(Array.isArray(keysRes) ? keysRes : keysRes.items ?? keysRes.keys ?? [])
      setPools(Array.isArray(poolsRes) ? poolsRes : poolsRes.pools ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function submit() {
    if (!poolId) {
      setError('请选择号池')
      return
    }
    if (!value.trim()) {
      setError('请输入要提交的 API Key')
      return
    }
    try {
      setError('')
      const selected = pools.find((item) => String(item.id) === poolId)
      await vendorApi.submitKey({
        pool_id: selected?.id,
        channel_id: selected?.channel_id,
        value: value.trim(),
      })
      setSuccess('Key 已通过校验并提交到号池')
      setOpen(false)
      setPoolId('')
      setValue('')
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Vendor"
        title="我的 API Key"
        description="支持上传新 Key，并查看累计消耗与收益。"
        actions={<Button onClick={() => setOpen(true)}>上传新 Key</Button>}
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      {success ? (
        <Card className="border-emerald-500/20 bg-emerald-500/5">
          <CardContent className="py-4 text-sm text-emerald-700">{success}</CardContent>
        </Card>
      ) : null}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>渠道</TableHead>
              <TableHead>Key</TableHead>
              <TableHead>累计消耗</TableHead>
              <TableHead>我的收益</TableHead>
              <TableHead>状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {keys.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell>{row.channel_name ?? row.channel_id ?? '-'}</TableCell>
                <TableCell className="font-mono text-xs">{row.masked_value ?? row.key ?? '-'}</TableCell>
                <TableCell>{formatCredits(row.total_cost ?? 0)}</TableCell>
                <TableCell>{formatCredits(row.my_earn ?? row.total_profit ?? 0)}</TableCell>
                <TableCell>{row.is_active === false ? '停用' : '启用'}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>上传新 Key</DialogTitle>
          </DialogHeader>
          <div className="flex flex-col gap-4">
            <NativeSelect value={poolId} onChange={(event) => setPoolId(event.target.value)}>
              <option value="">选择号池</option>
              {pools.map((pool) => (
                <option key={pool.id} value={String(pool.id)}>
                  {pool.channel_name}（{pool.name}）
                </option>
              ))}
            </NativeSelect>
            <Input value={value} onChange={(event) => setValue(event.target.value)} placeholder="请输入 API Key" />
            {pools.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                当前没有开放给号商上传的号池，请先让管理员在后台开启。
              </p>
            ) : null}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>取消</Button>
            <Button onClick={submit} disabled={!poolId || !value.trim()}>验证并提交</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
