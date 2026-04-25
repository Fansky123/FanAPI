import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

type Preset = '今天' | '7 天' | '30 天'

function formatLocal(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function presetRange(p: Preset): { start: string; end: string } {
  const end = new Date()
  const start = new Date()
  if (p === '今天') {
    start.setHours(0, 0, 0, 0)
  } else if (p === '7 天') {
    start.setDate(start.getDate() - 7)
  } else if (p === '30 天') {
    start.setDate(start.getDate() - 30)
  }
  return { start: formatLocal(start), end: formatLocal(end) }
}

export function DateRangeFilter({
  startAt,
  endAt,
  onChange,
  label,
}: {
  startAt: string
  endAt: string
  onChange: (next: { startAt: string; endAt: string }) => void
  label?: string
}) {
  function applyPreset(p: Preset) {
    const r = presetRange(p)
    onChange({ startAt: r.start, endAt: r.end })
  }
  return (
    <div className="flex flex-wrap items-center gap-2">
      {label ? <Label className="text-sm text-muted-foreground">{label}</Label> : null}
      <Input
        type="datetime-local"
        value={startAt}
        onChange={(e) => onChange({ startAt: e.target.value, endAt })}
        className="w-[200px]"
      />
      <span className="text-sm text-muted-foreground">至</span>
      <Input
        type="datetime-local"
        value={endAt}
        onChange={(e) => onChange({ startAt, endAt: e.target.value })}
        className="w-[200px]"
      />
      <div className="flex gap-1">
        <Button type="button" size="sm" variant="ghost" className="h-8 px-2 text-xs" onClick={() => applyPreset('今天')}>今天</Button>
        <Button type="button" size="sm" variant="ghost" className="h-8 px-2 text-xs" onClick={() => applyPreset('7 天')}>7 天</Button>
        <Button type="button" size="sm" variant="ghost" className="h-8 px-2 text-xs" onClick={() => applyPreset('30 天')}>30 天</Button>
        {(startAt || endAt) ? (
          <Button
            type="button"
            size="sm"
            variant="ghost"
            className="h-8 px-2 text-xs text-muted-foreground"
            onClick={() => onChange({ startAt: '', endAt: '' })}
          >
            清除
          </Button>
        ) : null}
      </div>
    </div>
  )
}
