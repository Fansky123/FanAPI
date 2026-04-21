export function formatCredits(value: number | undefined | null) {
  if (value == null || value === 0) return '0.00'
  const credits = value / 1e6
  const abs = Math.abs(credits)
  if (abs >= 0.01) return credits.toFixed(2)
  if (abs >= 0.0001) return credits.toFixed(6)
  return credits.toFixed(8)
}
