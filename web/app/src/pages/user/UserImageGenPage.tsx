import { PageHeader } from '@/components/shared/PageHeader'
import { ImageStudioLaunch } from '@/components/shared/ImageStudioLaunch'

export function UserImageGenPage() {
  return <>
    <PageHeader title="图片创作" description="前往图片工作台，开始生成与修图。" />
    <ImageStudioLaunch />
  </>
}
