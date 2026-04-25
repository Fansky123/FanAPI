import type { FormEvent } from 'react'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { ArrowRightIcon, KeyRoundIcon, UserRoundIcon } from 'lucide-react'

import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { getApiErrorMessage } from '@/lib/api/http'
import { authApi } from '@/lib/api/public'
import { setRoleToken, setSiteModePreference } from '@/lib/auth/storage'

export function UserLoginPage() {
  const navigate = useNavigate()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setSubmitting(true)
    setError('')

    try {
      const response = await authApi.login({ username, password })
      setRoleToken('user', response.token)
      setSiteModePreference('user')
      navigate('/dashboard')
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card className="w-full max-w-[440px] border-border/80 bg-card/95 shadow-xl shadow-primary/5 backdrop-blur">
      <CardHeader className="gap-4 px-6 pt-6">
        <Badge variant="secondary" className="w-fit">
          User sign in
        </Badge>
        <div className="flex flex-col gap-2">
          <CardTitle className="text-2xl font-semibold tracking-tight">
            登录用户控制台
          </CardTitle>
          <CardDescription>
            管理 API Key、调用统计、账单余额和生成任务。
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="px-6">
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <div className="flex flex-col gap-2">
            <Label htmlFor="user-login-username">用户名 / 邮箱</Label>
            <div className="relative">
              <UserRoundIcon className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="user-login-username"
                className="h-10 pl-9"
                value={username}
                onChange={(event) => setUsername(event.target.value)}
                placeholder="请输入用户名或邮箱"
                autoComplete="username"
                aria-invalid={Boolean(error)}
                required
              />
            </div>
          </div>
          <div className="flex flex-col gap-2">
            <div className="flex items-center justify-between gap-3">
              <Label htmlFor="user-login-password">密码</Label>
              <Link className="text-xs text-muted-foreground hover:text-foreground" to="/forgot-password">
                忘记密码
              </Link>
            </div>
            <div className="relative">
              <KeyRoundIcon className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="user-login-password"
                className="h-10 pl-9"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="请输入密码"
                autoComplete="current-password"
                aria-invalid={Boolean(error)}
                required
              />
            </div>
          </div>
          {error ? (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          <Button className="h-10 w-full" type="submit" disabled={submitting}>
            {submitting ? '登录中...' : '登录'}
            <ArrowRightIcon data-icon="inline-end" />
          </Button>
        </form>
      </CardContent>
      <CardFooter className="justify-center bg-muted/35 px-6 py-4 text-sm text-muted-foreground">
        <span>还没有账号？</span>
        <Button variant="link" className="h-auto px-1.5" asChild>
          <Link to="/register">立即注册</Link>
        </Button>
      </CardFooter>
    </Card>
  )
}
