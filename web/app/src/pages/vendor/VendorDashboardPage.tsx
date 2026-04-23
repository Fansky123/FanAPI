import { PageHeader } from '@/components/shared/PageHeader'
import { TableSkeleton } from '@/components/shared/TableSkeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
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
  const keyCount = profile?.key_count ?? keys.length
  const commissionRatio = profile?.commission_ratio ?? 0
  const payoutRatio = Math.max(0, 1 - commissionRatio)

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
      {profile?.is_active === false ? (
        <Alert>
          <AlertDescription>账号已被禁用，请联系管理员处理后再继续提交或查看号池数据。</AlertDescription>
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
                        : (profile?.commission_ratio != null
                            ? `${((profile.commission_ratio) * 100).toFixed(2)}%`
                            : '—（全局）')}
                </p>
              )}
            </div>
          ))}
        </CardContent>
      </Card>
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>账户余额</CardTitle>
          </CardHeader>
          <CardContent className="flex items-end justify-between gap-3">
            {loading ? <Skeleton className="h-8 w-32" /> : <div className="text-2xl font-semibold text-blue-600">{formatCredits(profile?.balance ?? 0)}</div>}
            <Badge variant="outline">可提现余额</Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>提供 Key 数量</CardTitle>
          </CardHeader>
          <CardContent>
            {loading ? <Skeleton className="h-8 w-20" /> : <div className="text-2xl font-semibold">{keyCount}</div>}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>手续费比例</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            {loading ? <Skeleton className="h-8 w-24" /> : <div className="text-2xl font-semibold">{(commissionRatio * 100).toFixed(2)}%</div>}
            <p className="text-sm text-muted-foreground">平台扣除后，实际到账约 {(payoutRatio * 100).toFixed(2)}%</p>
          </CardContent>
        </Card>
      </div>
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
                <TableHead>渠道</TableHead>
                <TableHead>总成本</TableHead>
                <TableHead>总收益</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead>状态</TableHead>
              </TableRow>
            </TableHeader>
            {loading ? (
              <TableSkeleton cols={7} rows={3} />
            ) : (
              <TableBody>
                {keys.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="py-10 text-center text-muted-foreground">
                      暂无 Key 数据
                    </TableCell>
                  </TableRow>
                ) : (
                  keys.map((key, index) => (
                    <TableRow key={key.id ?? index}>
                      <TableCell>{key.id ?? '-'}</TableCell>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {key.masked_value ?? '-'}
                      </TableCell>
                      <TableCell>{key.channel_name ?? '-'}</TableCell>
                      <TableCell>{formatCredits(key.total_cost ?? 0)}</TableCell>
                      <TableCell>{formatCredits(key.my_earn ?? 0)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {key.created_at ? new Date(key.created_at).toLocaleString('zh-CN') : '-'}
                      </TableCell>
                      <TableCell>
                        <Badge variant={key.is_active === false ? 'secondary' : 'default'}>
                          {key.is_active === false ? '停用' : '启用'}
                        </Badge>
                      </TableCell>
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
