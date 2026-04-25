import { useEffect, useRef, useState } from 'react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { authApi } from '@/lib/api/public'
import { userApi, type UserProfileResponse } from '@/lib/api/user'
import { useAsync } from '@/hooks/use-async'

export function UserProfilePage() {
  const { data: profile, loading, error: loadError, reload } = useAsync(
    () => userApi.getProfile(),
    null as UserProfileResponse | null,
  )

  // 修改密码
  const [pwdForm, setPwdForm] = useState({ old_password: '', new_password: '', confirm: '' })
  const [pwdError, setPwdError] = useState('')
  const [pwdSuccess, setPwdSuccess] = useState('')
  const [pwdLoading, setPwdLoading] = useState(false)

  async function changePassword() {
    setPwdError('')
    setPwdSuccess('')
    if (pwdForm.new_password.length < 8) { setPwdError('新密码不少于 8 位'); return }
    if (pwdForm.new_password !== pwdForm.confirm) { setPwdError('两次密码不一致'); return }
    setPwdLoading(true)
    try {
      await userApi.changePassword({ old_password: pwdForm.old_password, new_password: pwdForm.new_password })
      setPwdSuccess('密码已修改成功')
      setPwdForm({ old_password: '', new_password: '', confirm: '' })
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setPwdError(getApiErrorMessage(err))
    } finally {
      setPwdLoading(false)
    }
  }

  // 邮箱绑定
  const [emailInput, setEmailInput] = useState('')
  const [codeInput, setCodeInput] = useState('')
  const [emailError, setEmailError] = useState('')
  const [emailSuccess, setEmailSuccess] = useState('')
  const [emailLoading, setEmailLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => () => { if (timerRef.current) clearInterval(timerRef.current) }, [])

  async function sendCode() {
    if (!emailInput) { setEmailError('请填写邮箱'); return }
    setEmailError('')
    try {
      await authApi.sendCode(emailInput)
      setCountdown(60)
      timerRef.current = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) { clearInterval(timerRef.current!); timerRef.current = null; return 0 }
          return c - 1
        })
      }, 1000)
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setEmailError(getApiErrorMessage(err))
    }
  }

  async function bindEmail() {
    if (!emailInput || !codeInput) { setEmailError('请填写邮箱和验证码'); return }
    setEmailError('')
    setEmailLoading(true)
    try {
      await userApi.bindEmail({ email: emailInput, code: codeInput })
      setEmailSuccess('邮箱绑定成功')
      setEmailInput('')
      setCodeInput('')
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setEmailError(getApiErrorMessage(err))
    } finally {
      setEmailLoading(false)
    }
  }

  const initial = profile?.username?.[0] ?? profile?.email?.[0] ?? '?'
  const balanceYuan = profile?.balance != null ? (profile.balance / 1e6).toFixed(4) : '--'

  return (
    <>
      <PageHeader
        eyebrow="Identity"
        title="个人中心"
        description="账号基本信息与安全设置。"
        actions={
          loadError ? (
            <Button size="sm" variant="outline" onClick={reload}>
              重试
            </Button>
          ) : null
        }
      />
      {loadError ? (
        <Alert variant="destructive">
          <AlertDescription>{loadError}</AlertDescription>
        </Alert>
      ) : null}

      {/* 个人信息卡 */}
      <Card>
        <CardContent className="flex flex-col items-center gap-4 py-8 sm:flex-row sm:items-start sm:gap-8">
          {/* 头像 */}
          <div className="flex size-20 shrink-0 items-center justify-center rounded-full bg-primary text-3xl font-bold text-primary-foreground">
            {loading ? '?' : initial.toUpperCase()}
          </div>
          <div className="flex flex-col gap-2">
            {loading ? (
              <>
                <Skeleton className="h-6 w-32" />
                <Skeleton className="h-4 w-48" />
              </>
            ) : (
              <>
                <p className="text-xl font-semibold">{profile?.username ?? '-'}</p>
                <p className="text-xl font-semibold">用户ID {profile?.id ?? '-'}</p>
                <p className="text-sm text-muted-foreground">{profile?.email ?? '未绑定邮箱'}</p>
                <div className="flex items-center gap-2">
                  {profile?.group ? <Badge variant="secondary">{profile.group}</Badge> : null}
                  <span className="text-sm text-muted-foreground">余额：<strong>¥{balanceYuan}</strong></span>
                </div>
              </>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 修改密码 */}
      <Card>
        <CardHeader>
          <CardTitle>修改密码</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 max-w-md">
          {pwdError ? <Alert variant="destructive"><AlertDescription>{pwdError}</AlertDescription></Alert> : null}
          {pwdSuccess ? <Alert><AlertDescription className="text-emerald-600">{pwdSuccess}</AlertDescription></Alert> : null}
          <div className="space-y-1">
            <label className="text-sm font-medium">当前密码</label>
            <Input type="password" value={pwdForm.old_password} onChange={(e) => setPwdForm((f) => ({ ...f, old_password: e.target.value }))} />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">新密码</label>
            <Input type="password" value={pwdForm.new_password} onChange={(e) => setPwdForm((f) => ({ ...f, new_password: e.target.value }))} placeholder="至少 8 位" />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">确认新密码</label>
            <Input type="password" value={pwdForm.confirm} onChange={(e) => setPwdForm((f) => ({ ...f, confirm: e.target.value }))} />
          </div>
          <Button onClick={changePassword} disabled={pwdLoading}>
            {pwdLoading ? '保存中…' : '保存密码'}
          </Button>
        </CardContent>
      </Card>

      {/* 邮箱绑定 */}
      <Card>
        <CardHeader>
          <CardTitle>邮箱绑定</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 max-w-md">
          {emailError ? <Alert variant="destructive"><AlertDescription>{emailError}</AlertDescription></Alert> : null}
          {emailSuccess ? <Alert><AlertDescription className="text-emerald-600">{emailSuccess}</AlertDescription></Alert> : null}
          {profile?.email ? (
            <div className="flex items-center gap-2 text-sm">
              <span className="text-emerald-600">✓</span>
              <span>{profile.email}</span>
              <span className="text-muted-foreground">已绑定，可用于找回密码</span>
            </div>
          ) : (
            <>
              <div className="space-y-1">
                <label className="text-sm font-medium">邮箱地址</label>
                <Input type="email" value={emailInput} onChange={(e) => setEmailInput(e.target.value)} placeholder="example@email.com" />
              </div>
              <div className="flex gap-2">
                <div className="flex-1 space-y-1">
                  <label className="text-sm font-medium">验证码</label>
                  <Input value={codeInput} onChange={(e) => setCodeInput(e.target.value)} placeholder="6 位验证码" />
                </div>
                <div className="flex items-end">
                  <Button variant="outline" disabled={countdown > 0} onClick={sendCode}>
                    {countdown > 0 ? `${countdown}s` : '发送验证码'}
                  </Button>
                </div>
              </div>
              <Button onClick={bindEmail} disabled={emailLoading}>
                {emailLoading ? '绑定中…' : '绑定邮箱'}
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </>
  )
}

