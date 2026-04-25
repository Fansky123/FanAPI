import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export type TrendPoint = {
  label: string
  value: number
}

export type DualTrendPoint = {
  label: string
  success: number
  failed: number
}

export function TrendChart({
  title,
  points,
  color,
  formatValue,
}: {
  title: string
  points: TrendPoint[]
  color: string
  formatValue: (value: number) => string
}) {
  const values = points.map((point) => point.value)
  const max = Math.max(...values, 1)
  const width = 100
  const height = 44
  const step = points.length > 1 ? width / (points.length - 1) : width
  const path = points
    .map((point, index) => {
      const x = index * step
      const y = height - (point.value / max) * height
      return `${index === 0 ? 'M' : 'L'} ${x.toFixed(2)} ${y.toFixed(2)}`
    })
    .join(' ')

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="h-48 rounded-xl border border-border/70 bg-muted/20 p-4">
          <svg viewBox={`0 0 ${width} ${height}`} className="h-full w-full overflow-visible" preserveAspectRatio="none">
            {Array.from({ length: 4 }).map((_, index) => {
              const y = (height / 3) * index
              return (
                <line
                  key={index}
                  x1="0"
                  y1={y}
                  x2={width}
                  y2={y}
                  stroke="currentColor"
                  className="text-border/70"
                  strokeDasharray="2 2"
                />
              )
            })}
            <path d={path} fill="none" stroke={color} strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
            {points.map((point, index) => {
              const x = index * step
              const y = height - (point.value / max) * height
              return <circle key={point.label} cx={x} cy={y} r="1.8" fill={color} />
            })}
          </svg>
        </div>
        <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
          {points.map((point) => (
            <div key={point.label} className="rounded-lg border border-border/70 bg-background px-3 py-2">
              <div className="text-xs text-muted-foreground">{point.label}</div>
              <div className="mt-1 text-sm font-medium">{formatValue(point.value)}</div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

export function DualTrendChart({
  title,
  points,
}: {
  title: string
  points: DualTrendPoint[]
}) {
  const max = Math.max(...points.flatMap((point) => [point.success, point.failed]), 1)
  const width = 100
  const height = 44
  const step = points.length > 1 ? width / (points.length - 1) : width

  const buildPath = (selector: (point: DualTrendPoint) => number) =>
    points
      .map((point, index) => {
        const x = index * step
        const value = selector(point)
        const y = height - (value / max) * height
        return `${index === 0 ? 'M' : 'L'} ${x.toFixed(2)} ${y.toFixed(2)}`
      })
      .join(' ')

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <div className="flex items-center gap-2"><span className="size-2 rounded-full bg-blue-500" />成功请求</div>
          <div className="flex items-center gap-2"><span className="size-2 rounded-full bg-orange-500" />失败请求</div>
        </div>
        <div className="h-48 rounded-xl border border-border/70 bg-muted/20 p-4">
          <svg viewBox={`0 0 ${width} ${height}`} className="h-full w-full overflow-visible" preserveAspectRatio="none">
            {Array.from({ length: 4 }).map((_, index) => {
              const y = (height / 3) * index
              return (
                <line
                  key={index}
                  x1="0"
                  y1={y}
                  x2={width}
                  y2={y}
                  stroke="currentColor"
                  className="text-border/70"
                  strokeDasharray="2 2"
                />
              )
            })}
            <path d={buildPath((point) => point.success)} fill="none" stroke="#3b82f6" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
            <path d={buildPath((point) => point.failed)} fill="none" stroke="#f97316" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
        <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
          {points.map((point) => (
            <div key={point.label} className="rounded-lg border border-border/70 bg-background px-3 py-2 text-sm">
              <div className="text-xs text-muted-foreground">{point.label}</div>
              <div className="mt-1 flex items-center gap-3">
                <span className="text-blue-600">成功 {point.success}</span>
                <span className="text-orange-600">失败 {point.failed}</span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
