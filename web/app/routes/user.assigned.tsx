"use client"

import * as React from "react"
import { CircleHelpIcon, RefreshCwIcon } from "lucide-react"

import { useCurrentUser } from "@/components/app/app-provider"
import { getMyClaimed, type TopicAssignment, getTopic } from "@/lib/api/topics"
import { useI18n } from "@/lib/i18n/provider"
import Link from "@/components/common/link"
import { Button } from "@/components/ui/button"
import { noindexRouteMeta } from "@/lib/seo"

import { requireUser, requireUserClient } from "../route-helpers/auth"

export async function loader(args: { request: Request }) {
  await requireUser(args)
  return null
}

export async function clientLoader(args: { request: Request }) {
  await requireUserClient(args)
  return null
}

export function meta({
  matches,
}: {
  matches: Array<{ data?: unknown; loaderData?: unknown }>
}) {
  return noindexRouteMeta(matches as any, "My Claims", "我认领的")
}

type AssignmentWithTopic = TopicAssignment & { topicTitle?: string }

export default function UserAssignedRoute() {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const [items, setItems] = React.useState<AssignmentWithTopic[]>([])
  const [loading, setLoading] = React.useState(true)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      const assignments = await getMyClaimed()
      if (!Array.isArray(assignments)) {
        setItems([])
        return
      }
      const enriched = await Promise.all(
        assignments.map(async (a) => {
          try {
            const topic = await getTopic(String(a.topicId))
            return { ...a, topicTitle: topic?.title }
          } catch {
            return { ...a, topicTitle: undefined }
          }
        })
      )
      setItems(enriched)
    } catch {
      setItems([])
    } finally {
      setLoading(false)
    }
  }, [])

  React.useEffect(() => {
    void load()
  }, [load])

  if (!currentUser) {
    return (
      <div className="container mx-auto max-w-3xl px-4 py-12 text-center">
        <p className="text-muted-foreground">{t("common.notLoggedIn")}</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto max-w-3xl px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold">{t("assigned.title")}</h1>
        <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
          <RefreshCwIcon className={loading ? "size-4 animate-spin" : "size-4"} />
          {t("dashboard.actions.refresh")}
        </Button>
      </div>

      {loading ? (
        <p className="text-sm text-muted-foreground">{t("dashboard.loading")}</p>
      ) : items.length === 0 ? (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          <CircleHelpIcon className="mx-auto mb-3 size-8 opacity-40" />
          {t("assigned.empty")}
        </div>
      ) : (
        <div className="grid gap-3">
          {items.map((item) => (
            <Link
              key={item.id}
              href={`/topic/${item.topicId}`}
              className="flex items-center justify-between rounded-lg border p-4 transition-colors hover:bg-accent"
            >
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">
                  {item.topicTitle || `${t("common.topic")} #${item.topicId}`}
                </p>
                <p className="mt-1 text-xs text-muted-foreground">
                  {item.status === "active"
                    ? t("assigned.statusActive")
                    : item.status === "resolved"
                      ? t("assigned.statusResolved")
                      : t("assigned.statusDismissed")}
                </p>
              </div>
              <div className="ml-4 shrink-0">
                <span
                  className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                    item.status === "active"
                      ? "bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-200"
                      : item.status === "resolved"
                        ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-200"
                        : "bg-muted text-muted-foreground"
                  }`}
                >
                  {item.status === "active"
                    ? t("assigned.pending")
                    : item.status === "resolved"
                      ? t("assigned.solved")
                      : t("assigned.dismissed")}
                </span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
