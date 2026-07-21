import { EmptyState } from "@/components/common/empty-state"
import { useI18n } from "@/lib/i18n/provider"

export function PageLoading() {
  return (
    <div className="rounded-lg bg-background p-5 text-sm text-muted-foreground" />
  )
}

export function PageError({ message }: { message?: string | null }) {
  const { t } = useI18n()

  return <EmptyState title={message || t("common.noData")} />
}
