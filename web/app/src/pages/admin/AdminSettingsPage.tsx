import { useEffect, useState } from 'react'
import { SaveIcon } from 'lucide-react'

import { PageHeader } from '@/components/shared/PageHeader'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { adminApi } from '@/lib/api/admin'
import { getApiErrorMessage } from '@/lib/api/http'

export function AdminSettingsPage() {
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [rawJson, setRawJson] = useState('{}')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  async function load() {
    try {
      const response = await adminApi.getSettings()
      const maybeSettings = (response as { settings?: unknown }).settings
      const next =
        maybeSettings && typeof maybeSettings === 'object'
          ? (maybeSettings as Record<string, string>)
          : (response as Record<string, string>)
      setSettings(next)
      setRawJson(JSON.stringify(next, null, 2))
    } catch (err) {
      setError(getApiErrorMessage(err))
    }
  }

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load()
  }, [])

  async function saveSettings() {
    setSaving(true)
    try {
      const parsed = JSON.parse(rawJson) as Record<string, string>
      await adminApi.updateSettings(parsed)
      await load()
    } catch (err) {
      setError(getApiErrorMessage(err))
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      <PageHeader
        eyebrow="Configuration"
        title="系统设置"
        description="当前先提供 JSON 级保存能力，确保后台已经具备真实配置修改入口。"
        actions={
          <Button onClick={saveSettings} disabled={saving}>
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
          {Object.entries(settings).map(([key, value]) => (
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
