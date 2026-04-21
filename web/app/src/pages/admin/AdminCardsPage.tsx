import { useEffect, useState } from 'react'
import { RefreshCwIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
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
import { adminApi, type AdminCard } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

export function AdminCardsPage() {
  const [rows, setRows] = useState<AdminCard[]>([])
  const [error, setError] = useState('')
  const [generateOpen, setGenerateOpen] = useState(false)
  const [resultOpen, setResultOpen] = useState(false)
  const [generatedCards, setGeneratedCards] = useState<AdminCard[]>([])
  const [count, setCount] = useState('10')
  const [amount, setAmount] = useState('10')
  const [note, setNote] = useState('')
  const [pendingDeleteCard, setPendingDeleteCard] = useState<AdminCard | undefined>()

  async function load() {
    try {
      const response = await adminApi.listCards()
      setRows(response.cards ?? [])
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function generateCards() {
    try {
      const response = await adminApi.generateCards({
        count: Number(count),
        credits: Math.round(Number(amount) * 1_000_000),
        note,
      })
      setGeneratedCards(response.cards ?? [])
      setGenerateOpen(false)
      setResultOpen(true)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  async function deleteCard(row: AdminCard) {
    if (!row.id) return
    setPendingDeleteCard(row)
  }

  async function executeDeleteCard() {
    if (!pendingDeleteCard?.id) return
    try {
      await adminApi.deleteCard(pendingDeleteCard.id)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setPendingDeleteCard(undefined)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Cards"
        title="卡密管理"
        description="支持批量生成和删除未使用卡密，满足后台最小运营需求。"
        actions={
          <Button onClick={() => setGenerateOpen(true)}>
            生成卡密
          </Button>
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
              <TableHead>兑换码</TableHead>
              <TableHead>面值</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>备注</TableHead>
              <TableHead>生成时间</TableHead>
              <TableHead className="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row, index) => (
              <TableRow key={row.id ?? index}>
                <TableCell className="font-mono text-xs">{row.code ?? '-'}</TableCell>
                <TableCell>{((row.credits ?? 0) / 1_000_000).toFixed(4)}</TableCell>
                <TableCell>{row.status ?? '-'}</TableCell>
                <TableCell>{row.note ?? '-'}</TableCell>
                <TableCell>{row.created_at ?? '-'}</TableCell>
                <TableCell className="text-right">
                  {row.status === 'unused' ? (
                    <Button size="sm" variant="outline" onClick={() => deleteCard(row)}>
                      删除
                    </Button>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
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
            <Button variant="outline" onClick={() => setGenerateOpen(false)}>
              取消
            </Button>
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
            <Button variant="outline" onClick={() => setResultOpen(false)}>
              关闭
            </Button>
            <Button
              onClick={() =>
                navigator.clipboard.writeText(
                  generatedCards
                    .map((card) => `${card.code} ${(card.credits ?? 0) / 1_000_000}元`)
                    .join('\n')
                )
              }
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
