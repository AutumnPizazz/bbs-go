import { HtmlImagePreview } from "@/components/common/image-preview"
import type { Topic } from "@/lib/api/types"

export function TopicContent({ topic }: { topic: Topic }) {
  const content = topic.content || topic.summary || ""

  if (!content) {
    return null
  }

  return (
    <div className="mx-4 mb-4 break-all pt-0 text-[15px] text-foreground">
      <HtmlImagePreview
        html={content}
        className="bbs-content line-numbers wrap-break-word text-base leading-6 antialiased [&_h2]:scroll-mt-20 [&_h3]:scroll-mt-20 [&_h4]:scroll-mt-20 [&_img]:cursor-zoom-in"
      />
    </div>
  )
}
