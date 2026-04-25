import { useState } from 'react'
import { CreditCardIcon, RefreshCwIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableEmpty } from '@/components/shared/TableEmpty'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { copyToClipboard } from '@/lib/clipboard'
import { adminApi, type AdminCard } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

export function AdminCardsPage() {
  const [statusFilter, setStatusFilter] = useState('')
  const [queryParams, setQueryParams] = useState<Record<string, unknown>>({})

  const { data: rows, loading, error: loadError, reload } = useAsync(async () => {
    const response = await adminApi.listCards(queryParams)
    return response.cards ?? []
  }, [] as AdminCard[], [queryParams])

  const [mutError, setMutError] = useState('')
  const [generateOpen, setGenerateOpen] = useState(false)
  const [resultOpen, setResultOpen] = useState(false)
  const [generatedCards, setGeneratedCards] = useState<AdminCard[]>([])
  const [count, setCount] = useState('10')
  const [amount, setAmount] = useState('10')
  const [note, setNote] = useState('')
  const [pendingDeleteCard, setPendingDeleteCard] = useState<AdminCard | undefined>()

  const error = loadError || mutError

  async function generateCards() {
    setMutError('')
    try {
      const response = await adminApi.generateCards({
        count: Number(count),
        credits: Math.round(Number(amount) * 1_000_000),
        note,
      })
      setGeneratedCards(response.cards ?? [])
      setGenerateOpen(false)
      setResultOpen(true)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    }
  }

  async function executeDeleteCard() {
    if (!pendingDeleteCard?.id) return
    setMutError('')
    try {
      await adminApi.deleteCard(pendingDeleteCard.id)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setPendingDeleteCard(undefined)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Cards"
        title="卡密管理"
        description="批量生成和删除未使用卡密。"
        actions={
          <>
            {error ? (
              <Button size="sm" variant="outline" onClick={reload}>
                重试
              </Button>
            ) : null}
            <Button onClick={() => setGenerateOpen(true)}>生成卡密</Button>
          </>
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card>
        <CardContent className="flex items-end gap-3 py-4">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">状态过滤</label>
            <Select value={statusFilter || '_all'} onValueChange={(v) => setStatusFilter(v === '_all' ? '' : v)}>
              <SelectTrigger className="w-32"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="_all">全部</SelectItem>
                <SelectItem value="unused">未使用</SelectItem>
                <SelectItem value="used">已使用</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button onClick={() => setQueryParams(statusFilter ? { status: statusFilter } : {})}>查询</Button>
          <Button variant="outline" onClick={() => { setStatusFilter(''); setQueryParams({}) }}>重置</Button>
        </CardContent>
      </Card>

      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>兑换码</TableHead>
              <TableHead>面值</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>备注</TableHead>
              <TableHead>生成时间</TableHead>
              <TableHead>使用时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={7} />
          ) : (
            <TableBody>
              {rows.length === 0 ? (
                <TableEmpty
                  cols={7}
                  Icon={CreditCardIcon}
                  title="还没有卡密"
                  description="使用上方「批量生成」创建卡密后会显示在这里。"
                />
              ) : (
                rows.map((row, index) => (
                  <TableRow key={row.id ?? index}>
                    <TableCell
                      className="font-mono text-xs cursor-pointer hover:text-primary"
                      onClick={() => {
                        void copyToClipboard(row.code ?? '', { successMessage: '兑换码已复制' })
                      }}
                      title="点击复制"
                    >{row.code ?? '-'}</TableCell>
                    <TableCell>¥{((row.credits ?? 0) / 1_000_000).toFixed(4)}</TableCell>
                    <TableCell>
                      <Badge variant={row.status === 'unused' ? 'default' : 'secondary'}>
                        {row.status === 'unused' ? '未使用' : '已使用'}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{row.note ?? '-'}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.created_at ? new Date(row.created_at).toLocaleString('zh-CN') : '-'}
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.used_at ? new Date(row.used_at).toLocaleString('zh-CN') : '—'}
                    </TableCell>
                    <TableCell className="text-right">
                      {row.status === 'unused' ? (
                        <Button size="sm" variant="outline" onClick={() => setPendingDeleteCard(row)}>
                          删除
                        </Button>
                      ) : null}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>

      <Dialog open={generateOpen} onOpenChange={setGenerateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>批量生成卡密</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4">
            <Input value={count} onChange={(event) => setCount(event.target.value)} placeholder="数量" />
            <Input value={amount} onChange={(event) => setAmount(event.target.value)} placeholder="面值（元）" />
            <Input value={note} onChange={(event) => setNote(event.target.value)} placeholder="备注" />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setGenerateOpen(false)}>取消</Button>
            <Button onClick={generateCards}>生成</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={resultOpen} onOpenChange={setResultOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>生成结果</DialogTitle>
          </DialogHeader>
          <div className="max-h-96 overflow-auto rounded-xl border border-border/70 bg-muted/25 p-4 font-mono text-xs">
            {generatedCards.map((card) => `${card.code} ${(card.credits ?? 0) / 1_000_000}元`).join('\n')}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setResultOpen(false)}>关闭</Button>
            <Button
              onClick={() => {
                void copyToClipboard(
                  generatedCards
                    .map((card) => `${card.code} ${(card.credits ?? 0) / 1_000_000}元`)
                    .join('\n'),
                  { successMessage: '卡密列表已复制' }
                )
              }}
            >
              <RefreshCwIcon data-icon="inline-start" />
              复制全部
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog open={pendingDeleteCard !== undefined} onOpenChange={() => setPendingDeleteCard(undefined)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确认删除卡密 {pendingDeleteCard?.code ?? pendingDeleteCard?.id} 吗？此操作不可撤销。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={executeDeleteCard}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
