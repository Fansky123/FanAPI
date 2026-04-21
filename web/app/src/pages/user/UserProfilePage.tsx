import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { userApi, type UserProfileResponse } from '@/lib/api/user'
import { useAsync } from '@/hooks/use-async'

function InfoRow({ label, value, loading }: { label: string; value?: string; loading: boolean }) {
  return (
    <div>
      <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
        {label}
      </p>
      {loading ? (
        <Skeleton className="mt-2 h-4 w-28" />
      ) : (
        <p className="mt-2 text-sm">{value ?? '-'}</p>
      )}
    </div>
  )
}

export function UserProfilePage() {
  const { data: profile, loading, error, reload } = useAsync(
    () => userApi.getProfile(),
    null as UserProfileResponse | null,
  )

  return (
    <>
      <PageHeader
        eyebrow="Identity"
        title="个人中心"
        description="账号基本信息，包括用户名、邮箱和所属分组。"
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
          <CardTitle>账号信息</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-3">
          <InfoRow label="用户名" value={profile?.username} loading={loading} />
          <InfoRow label="邮箱" value={profile?.email} loading={loading} />
          <InfoRow label="分组" value={profile?.group} loading={loading} />
        </CardContent>
      </Card>
    </>
  )
}
