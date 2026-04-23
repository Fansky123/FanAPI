import { useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
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
import { vendorApi, type VendorKey, type VendorPool } from '@/lib/api/vendor'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function VendorKeysPage() {
  const { data, loading, error: loadError, reload } = useAsync(async () => {
    const [keysRes, poolsRes] = await Promise.all([
      vendorApi.getKeys(),
      vendorApi.getPools(),
    ])
    return {
      keys: Array.isArray(keysRes) ? keysRes : keysRes.items ?? keysRes.keys ?? [] as VendorKey[],
      pools: Array.isArray(poolsRes) ? poolsRes : poolsRes.pools ?? [] as VendorPool[],
    }
  }, { keys: [] as VendorKey[], pools: [] as VendorPool[] })

  const keys = data.keys
  const pools = data.pools

  const [mutError, setMutError] = useState('')
  const [open, setOpen] = useState(false)
  const [poolId, setPoolId] = useState('')
  const [value, setValue] = useState('')
  const [submitResult, setSubmitResult] = useState<{ ok: boolean; message: string } | null>(null)

  const error = loadError || mutError

  async function submit() {
    if (!poolId) {
      setMutError('请选择号池')
      return
    }
    if (!value.trim()) {
      setMutError('请输入要提交的 API Key')
      return
    }
    setMutError('')
    setSubmitResult(null)
    try {
      const selected = pools.find((item) => String(item.id) === poolId)
      const response = await vendorApi.submitKey({
        pool_id: selected?.id,
        channel_id: selected?.channel_id,
        value: value.trim(),
      })
      const resultMessage = typeof response?.message === 'string' ? response.message : 'Key 已成功提交到号池'
      setSubmitResult({ ok: true, message: resultMessage })
      setValue('')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      const message = getApiErrorMessage(err)
      setMutError(message)
      setSubmitResult({ ok: false, message })
    }
  }

  function resetDialog(nextOpen: boolean) {
    setOpen(nextOpen)
    if (!nextOpen) {
      setPoolId('')
      setValue('')
      setSubmitResult(null)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Vendor"
        title="我的 API Key"
        description="上传新 Key，并查看累计消耗与收益。"
        actions={
          <>
            {error ? (
              <Button size="sm" variant="outline" onClick={reload}>
                重试
              </Button>
            ) : null}
            <Button onClick={() => setOpen(true)}>上传新 Key</Button>
          </>
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
              <TableHead>ID</TableHead>
              <TableHead>渠道</TableHead>
              <TableHead>Key</TableHead>
              <TableHead>累计消耗</TableHead>
              <TableHead>我的收益</TableHead>
              <TableHead>添加时间</TableHead>
              <TableHead>状态</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={7} />
          ) : (
            <TableBody>
              {keys.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="py-10 text-center text-muted-foreground">
                    暂无 Key 数据
                  </TableCell>
                </TableRow>
              ) : (
                keys.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell>{row.id ?? '-'}</TableCell>
                    <TableCell>{row.channel_name ?? row.channel_id ?? '-'}</TableCell>
                    <TableCell className="font-mono text-xs">{row.masked_value ?? row.key ?? '-'}</TableCell>
                    <TableCell>{formatCredits(row.total_cost ?? 0)}</TableCell>
                    <TableCell>{formatCredits(row.my_earn ?? row.total_profit ?? 0)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                    </TableCell>
                    <TableCell>
                      <Badge variant={row.is_active === false ? 'secondary' : 'default'}>
                        {row.is_active === false ? '禁用' : '启用'}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>

      <Dialog open={open} onOpenChange={resetDialog}>
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
            {submitResult ? (
              <Alert variant={submitResult.ok ? 'default' : 'destructive'}>
                <AlertDescription>{submitResult.message}</AlertDescription>
              </Alert>
            ) : null}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => resetDialog(false)}>取消</Button>
            <Button onClick={submit} disabled={!poolId || !value.trim()}>验证并提交</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
