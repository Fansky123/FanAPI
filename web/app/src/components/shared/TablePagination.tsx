import { ChevronLeft, ChevronRight } from "lucide-react"
import { Button } from "@/components/ui/button"

interface TablePaginationProps {
  current: number
  total: number
  pageSize: number
  onChange: (page: number) => void
}

export function TablePagination({ current, total, pageSize, onChange }: TablePaginationProps) {
  const maxPage = Math.ceil(total / pageSize) || 1
  return (
    <div className="flex items-center justify-end space-x-2 py-4">
      <Button variant="outline" size="sm" onClick={() => onChange(current - 1)} disabled={current <= 1}>
        <ChevronLeft className="h-4 w-4" />
      </Button>
      <div className="text-sm font-medium">
        第 {current} / {maxPage} 页
      </div>
      <Button variant="outline" size="sm" onClick={() => onChange(current + 1)} disabled={current >= maxPage}>
        <ChevronRight className="h-4 w-4" />
      </Button>
    </div>
  )
}
