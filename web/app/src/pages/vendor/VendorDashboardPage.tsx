import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getApiErrorMessage } from '@/lib/api/http'
import { vendorApi, type VendorKey, type VendorProfile } from '@/lib/api/vendor'
import { formatCredits } from '@/lib/formatters/credits'

export function VendorDashboardPage() {
  const [profile, setProfile] = useState<VendorProfile | null>(null)
  const [keys, setKeys] = useState<VendorKey[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const [profileResponse, keysResponse] = await Promise.all([
          vendorApi.getProfile(),
          vendorApi.getKeys(),
        ])
        setProfile(profileResponse)
        setKeys(Array.isArray(keysResponse) ? keysResponse : keysResponse.items ?? keysResponse.keys ?? [])
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader eyebrow="Vendor" title="Vendor 工作台" description="Vendor 端已接入资料与 key 数据，后续继续补提交流程和筛选器。" />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <Card>
        <CardHeader>
          <CardTitle>供应侧资料</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-4">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              用户名
            </p>
            <p className="mt-2 text-sm">{profile?.username ?? profile?.name ?? '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              邮箱
            </p>
            <p className="mt-2 text-sm">{profile?.email ?? '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              余额
            </p>
            <p className="mt-2 text-sm">{formatCredits(profile?.balance ?? 0)}</p>
          </div>
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              佣金比例
            </p>
            <p className="mt-2 text-sm">{profile?.commission_ratio ?? '-'}</p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>名下 Key</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Key</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>总成本</TableHead>
                <TableHead>总收益</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {keys.map((key, index) => (
                <TableRow key={key.id ?? index}>
                  <TableCell>{key.id ?? '-'}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {key.key ?? '-'}
                  </TableCell>
                  <TableCell>{key.key_type ?? '-'}</TableCell>
                  <TableCell>{formatCredits(key.total_cost ?? 0)}</TableCell>
                  <TableCell>{formatCredits(key.total_profit ?? 0)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
