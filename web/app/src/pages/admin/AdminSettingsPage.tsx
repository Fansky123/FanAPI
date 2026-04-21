import { useState } from 'react'
import { SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { adminApi } from '@/lib/api/admin'
import { useAsync } from '@/hooks/use-async'

export function AdminSettingsPage() {
  const { data: settings, loading, error: loadError, reload } = useAsync(async () => {
    const response = await adminApi.getSettings()
    const maybeSettings = (response as { settings?: unknown }).settings
    return maybeSettings && typeof maybeSettings === 'object'
      ? (maybeSettings as Record<string, string>)
      : (response as Record<string, string>)
  }, {} as Record<string, string>)

  const [rawJson, setRawJson] = useState('')
  const [jsonInitialized, setJsonInitialized] = useState(false)
  const [mutError, setMutError] = useState('')
  const [saving, setSaving] = useState(false)

  // Sync JSON editor from loaded data once
  if (!loading && !jsonInitialized) {
    setRawJson(JSON.stringify(settings, null, 2))
    setJsonInitialized(true)
  }

  const error = loadError || mutError

  async function saveSettings() {
    setSaving(true)
    setMutError('')
    try {
      const parsed = JSON.parse(rawJson) as Record<string, string>
      await adminApi.updateSettings(parsed)
      reload()
    } catch (err) {
      const { getApiErrorMessage } = await import('@/lib/api/http')
      setMutError(getApiErrorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Configuration"
        title="系统设置"
        description="查看和修改平台全局配置项。"
        actions={
          <Button onClick={saveSettings} disabled={saving || loading}>
            <SaveIcon data-icon="inline-start" />
            {saving ? '保存中...' : '保存设置'}
          </Button>
        }
      />
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}
      <Card>
        <CardHeader>
          <CardTitle>当前配置快照</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-2">
          {loading
            ? Array.from({ length: 4 }).map((_, i) => (
                <div key={i} className="rounded-xl border border-border/70 bg-muted/20 p-4">
                  <Skeleton className="h-3 w-20" />
                  <Skeleton className="mt-2 h-4 w-32" />
                </div>
              ))
            : Object.entries(settings).map(([key, value]) => (
                <div key={key} className="rounded-xl border border-border/70 bg-muted/20 p-4">
                  <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">
                    {key}
                  </p>
                  <p className="mt-2 break-all text-sm">{value || '-'}</p>
                </div>
              ))}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>JSON 编辑器</CardTitle>
        </CardHeader>
        <CardContent>
          <Textarea
            className="min-h-96 font-mono text-xs"
            value={rawJson}
            onChange={(event) => setRawJson(event.target.value)}
          />
        </CardContent>
      </Card>
    </>
  )
}
