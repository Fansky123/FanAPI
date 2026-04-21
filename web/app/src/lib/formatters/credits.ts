export function formatCredits(value: number | undefined | null) {
  if (!value) return '0.00'
  return (value / 1e6).toFixed(2)
}
