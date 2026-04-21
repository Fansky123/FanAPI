import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { vendorApi, type VendorKey, type VendorProfile } from '@/lib/api/vendor'
import { formatCredits } from '@/lib/formatters/credits'
import { useAsync } from '@/hooks/use-async'

export function VendorDashboardPage() {
  const { data, loading, error, reload } = useAsync(async () => {
    const [profileResponse, keysResponse] = await Promise.all([
      vendorApi.getProfile(),
      vendorApi.getKeys(),
    ])
    return {
      profile: profileResponse as VendorProfile,
      keys: Array.isArray(keysResponse) ? keysResponse : keysResponse.items ?? keysResponse.keys ?? [] as VendorKey[],
    }
  }, { profile: null as VendorProfile | null, keys: [] as VendorKey[] })

  const profile = data.profile
  const keys = data.keys

  return (
    <>
      <PageHeader
        eyebrow="Vendor"
        title="Vendor 工作台"
        description="供应侧资料与名下 Key 概览。"
        actions={
          error ? (
            <Button size="sm" variant="outline" onClick={reload}>
              重试
            </Button>
          ) : null
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card>
        <CardHeader>
          <CardTitle>供应侧资料</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-4">
          {(['用户名', '邮箱', '余额', '佣金比例'] as const).map((label) => (
            <div key={label}>
              <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
                {label}
              </p>
              {loading ? (
                <Skeleton className="mt-2 h-4 w-24" />
              ) : (
                <p className="mt-2 text-sm">
                  {label === '用户名'
                    ? (profile?.username ?? profile?.name ?? '-')
                    : label === '邮箱'
                      ? (profile?.email ?? '-')
                      : label === '余额'
                        ? formatCredits(profile?.balance ?? 0)
                        : (profile?.commission_ratio ?? '-')}
                </p>
              )}
            </div>
          ))}
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
            {loading ? (
              <TableSkeleton cols={5} rows={3} />
            ) : (
              <TableBody>
                {keys.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="py-10 text-center text-muted-foreground">
                      暂无 Key 数据
                    </TableCell>
                  </TableRow>
                ) : (
                  keys.map((key, index) => (
                    <TableRow key={key.id ?? index}>
                      <TableCell>{key.id ?? '-'}</TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {key.key ?? '-'}
                      </TableCell>
                      <TableCell>{key.key_type ?? '-'}</TableCell>
                      <TableCell>{formatCredits(key.total_cost ?? 0)}</TableCell>
                      <TableCell>{formatCredits(key.total_profit ?? 0)}</TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            )}
          </Table>
        </CardContent>
      </Card>
    </>
  )
}
