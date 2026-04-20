import { useEffect, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getApiErrorMessage } from '@/lib/api/http'
import { userApi, type UserProfileResponse } from '@/lib/api/user'

export function UserProfilePage() {
  const [profile, setProfile] = useState<UserProfileResponse | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    async function load() {
      try {
        const response = await userApi.getProfile()
        setProfile(response)
      } catch (err) {
        setError(getApiErrorMessage(err))
      }
    }

    void load()
  }, [])

  return (
    <>
      <PageHeader
        eyebrow="Identity"
        title="个人中心"
        description="这里优先聚合账号身份信息，后续会继续补邮箱绑定、密码修改和返佣相关能力。"
      />
      {error ? (
        <Card className="border-destructive/25 bg-destructive/5">
          <CardContent className="py-4 text-sm text-destructive">{error}</CardContent>
        </Card>
      ) : null}
      <Card>
        <CardHeader>
          <CardTitle>账号信息</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-3">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              用户名
            </p>
            <p className="mt-2 text-sm">{profile?.username ?? '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              邮箱
            </p>
            <p className="mt-2 text-sm">{profile?.email ?? '-'}</p>
          </div>
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
              分组
            </p>
            <p className="mt-2 text-sm">{profile?.group ?? '-'}</p>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
