import { WidgetCard } from "@/components/common/widget-card"
import { useI18n } from "@/lib/i18n/provider"
import { noindexRouteMeta } from "@/lib/seo"
import { useDocumentTitle } from "@/lib/use-document-title"

export function meta({
  matches,
}: {
  matches: Array<{ data?: unknown; loaderData?: unknown }>
}) {
  return noindexRouteMeta(matches, "Sign up", "注册")
}

export default function SignupRoute() {
  const { t } = useI18n()
  useDocumentTitle(t("user.signup.title"))
  return (
    <section className="main">
      <div className="container mx-auto">
        <WidgetCard
          className="mx-auto max-w-[600px]"
          title={t("user.signup.title")}
          bodyClassName="pt-0"
        >
          <p className="p-6 text-center text-muted-foreground">
            {t("user.signup.adminOnly")}
          </p>
        </WidgetCard>
      </div>
    </section>
  )
}
