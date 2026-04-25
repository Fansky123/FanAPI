import { useState } from 'react'
import { TicketIcon } from 'lucide-react'
import { toast } from 'sonner'

import { PageHeader } from '@/components/shared/PageHeader'
import { TableEmpty } from '@/components/shared/TableEmpty'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { userApi, type RedeemRecord } from '@/lib/api/user'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function UserExchangePage() {
  const { data: history, loading, error: loadError, reload } = useAsync(async () => {
    const response = await userApi.getRedeemHistory()
    return Array.isArray(response) ? response : response.records ?? response.list ?? []
  }, [] as RedeemRecord[])

  const [code, setCode] = useState('')
  const [mutError, setMutError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const error = loadError || mutError

  async function redeem() {
    if (!code.trim()) return
    setSubmitting(true)
    setMutError('')
    try {
      const res = await userApi.redeemCard(code.trim()) as { credits?: number; message?: string }
      const credits = typeof res?.credits === 'number' ? res.credits : null
      toast.success(credits ? `兑换成功，获得 ${formatCredits(credits)} 积分` : '兑换成功')
      setCode('')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      const msg = getApiErrorMessage(err)
      setMutError(msg)
      toast.error(msg)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Redeem"
        title="兑换中心"
        description="输入卡密兑换积分，兑换后可查看历史记录。"
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card>
        <CardContent className="flex gap-3 p-6">
          <Input
            value={code}
            onChange={(event) => setCode(event.target.value)}
            placeholder="请输入兑换码"
            onKeyDown={(e) => e.key === 'Enter' && void redeem()}
          />
          <Button onClick={redeem} disabled={submitting}>
            {submitting ? '兑换中...' : '立即兑换'}
          </Button>
        </CardContent>
      </Card>
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>兑换码</TableHead>
              <TableHead>积分数量</TableHead>
              <TableHead>兑换时间</TableHead>
            </TableRow>
          </TableHeader>
          {loading ? (
            <TableSkeleton cols={3} />
          ) : (
            <TableBody>
              {history.length === 0 ? (
                <TableEmpty
                  cols={3}
                  Icon={TicketIcon}
                  title="还没有兑换记录"
                  description="使用上方输入框兑换卡密后，记录会显示在这里。"
                />
              ) : (
                history.map((row, index) => (
                  <TableRow key={row.code ?? index}>
                    <TableCell className="font-mono text-xs">{row.code ?? '-'}</TableCell>
                    <TableCell>{formatCredits(row.credits ?? 0)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">
                      {row.used_at
                        ? new Date(row.used_at).toLocaleString('zh-CN')
                        : '-'}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          )}
        </Table>
      </Card>
    </>
  )
}
