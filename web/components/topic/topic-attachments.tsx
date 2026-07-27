import type { Attachment } from "@/lib/api/types"
import type { TFunction } from "@/lib/i18n"
import { formatFileSize } from "@/lib/utils"

export function TopicAttachments({ attachments, t }: { attachments?: Attachment[]; t: TFunction }) {
  if (!attachments?.length) {
    return null
  }

  return (
    <section className="mx-4 mb-4 rounded-md border border-border bg-muted/30 p-3">
      <h2 className="mb-2 text-sm font-medium text-foreground">{t("pages.topic.detail.attachments")}</h2>
      <ul className="space-y-3">
        {attachments.map((attachment) => (
          <li
            key={attachment.id}
            className="flex flex-col gap-2 rounded border border-border bg-background px-3 py-3 text-sm sm:flex-row sm:flex-wrap sm:items-center sm:justify-between"
          >
            <div className="min-w-0 flex-1">
              <div className="truncate font-medium text-foreground">{attachment.fileName || attachment.id}</div>
              <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5 text-muted-foreground">
                {formatFileSize(attachment.fileSize)}
                {typeof attachment.downloadCount === "number" ? (
                  <span>{t("pages.topic.detail.attachmentDownloadCount", { count: attachment.downloadCount })}</span>
                ) : null}
              </div>
            </div>
            <a
              href={`/api/attachment/download/${attachment.id}`}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex h-8 w-full shrink-0 items-center justify-center rounded-md border border-border bg-background px-3 text-sm font-medium hover:bg-muted sm:w-auto"
            >
              {t("pages.topic.detail.download")}
            </a>
          </li>
        ))}
      </ul>
    </section>
  )
}
