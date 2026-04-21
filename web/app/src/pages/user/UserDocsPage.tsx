import { PageHeader } from '@/components/shared/PageHeader'
import { Card, CardContent } from '@/components/ui/card'

export function UserDocsPage() {
  return (
    <>
      <PageHeader
        eyebrow="Reference"
        title="接口文档"
        description="直接复用后端 Swagger / OpenAPI 文档入口，保证文档与线上接口一致。"
      />
      <Card className="overflow-hidden">
        <CardContent className="p-0">
          <iframe
            className="min-h-[75vh] w-full border-0"
            src="/docs"
            title="FanAPI Docs"
          />
        </CardContent>
      </Card>
    </>
  )
}
