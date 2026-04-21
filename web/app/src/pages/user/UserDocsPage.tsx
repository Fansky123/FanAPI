import { Card, CardContent } from '@/components/ui/card'

export function UserDocsPage() {
  return (
    <Card className="overflow-hidden">
      <CardContent className="p-0">
        <iframe
          className="min-h-[85vh] w-full border-0"
          src="/swagger/index.html"
          title="FanAPI Docs"
        />
      </CardContent>
    </Card>
  )
}
