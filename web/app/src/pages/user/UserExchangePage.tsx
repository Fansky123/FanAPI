import { useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
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
      await userApi.redeemCard(code.trim())
      setCode('')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
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
                <TableRow>
                  <TableCell colSpan={3} className="py-10 text-center text-muted-foreground">
                    暂无兑换记录
                  </TableCell>
                </TableRow>
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
