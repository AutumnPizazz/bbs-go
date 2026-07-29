"use client"

import * as React from "react"
import { CheckCircle2, CircleHelp, Clock, RefreshCwIcon } from "lucide-react"

import { useCurrentUser } from "@/components/app/app-provider"
import { getMyClaimed, type TopicAssignment, getTopic } from "@/lib/api/topics"
import { useI18n } from "@/lib/i18n/provider"
import Link from "@/components/common/link"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { noindexRouteMeta } from "@/lib/seo"
import { prettyDate } from "@/lib/format"
import { cn } from "@/lib/utils"

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

type FilterTab = "all" | "active" | "resolved" | "dismissed"

type AssignmentEnriched = TopicAssignment & {
  topicTitle?: string
  topicCategory?: string
  topicSolved?: boolean
}

export default function UserAssignedRoute() {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const [items, setItems] = React.useState<AssignmentEnriched[]>([])
  const [loading, setLoading] = React.useState(true)
  const [tab, setTab] = React.useState<FilterTab>("all")

  const load = React.useCallback(async (status?: string) => {
    setLoading(true)
    try {
      const assignments = await getMyClaimed(status || undefined)
      if (!Array.isArray(assignments)) {
        setItems([])
        return
      }
      const enriched = await Promise.all(
        assignments.map(async (a) => {
          try {
            const topic = await getTopic(String(a.topicId))
            return {
              ...a,
              topicTitle: topic?.title,
              topicCategory: topic?.category?.name,
              topicSolved: topic?.qaStatus === "solved",
            }
          } catch {
            return { ...a }
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
    void load(tab === "all" ? undefined : tab)
  }, [load, tab])

  const activeCount = items.filter((i) => i.status === "active").length
  const resolvedCount = items.filter((i) => i.status === "resolved").length
  const total = activeCount + resolvedCount
  const progressPct = total > 0 ? Math.round((resolvedCount / total) * 100) : 0

  if (!currentUser) {
    return (
      <div className="container mx-auto max-w-3xl px-4 py-12 text-center">
        <p className="text-muted-foreground">{t("common.notLoggedIn")}</p>
      </div>
    )
  }

  const tabs: { key: FilterTab; label: string }[] = [
    { key: "all", label: t("assigned.filterAll") },
    { key: "active", label: t("assigned.pending") },
    { key: "resolved", label: t("assigned.solved") },
    { key: "dismissed", label: t("assigned.dismissed") },
  ]

  return (
    <div className="container mx-auto max-w-3xl px-4 py-8">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold">{t("assigned.title")}</h1>
        <Button variant="outline" size="sm" onClick={() => load(tab === "all" ? undefined : tab)} disabled={loading}>
          <RefreshCwIcon className={loading ? "size-4 animate-spin" : "size-4"} />
          {t("dashboard.actions.refresh")}
        </Button>
      </div>

      {/* ---- progress bar ---- */}
      {total > 0 ? (
        <div className="mb-6 rounded-lg border px-4 py-3">
          <div className="flex items-center justify-between text-sm mb-2">
            <span className="text-muted-foreground">
              {t("assigned.progress", { resolved: String(resolvedCount), total: String(total) })}
            </span>
            <span className="font-medium tabular-nums">{progressPct}%</span>
          </div>
          <div className="h-2 w-full rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-emerald-500 transition-all"
              style={{ width: `${progressPct}%` }}
            />
          </div>
          <div className="mt-2 flex gap-4 text-xs text-muted-foreground">
            <span className="inline-flex items-center gap-1">
              <span className="h-2 w-2 rounded-full bg-amber-400" />
              {t("assigned.pending")}: {activeCount}
            </span>
            <span className="inline-flex items-center gap-1">
              <span className="h-2 w-2 rounded-full bg-emerald-500" />
              {t("assigned.solved")}: {resolvedCount}
            </span>
          </div>
        </div>
      ) : null}

      {/* ---- filter tabs ---- */}
      <div className="mb-4 flex gap-1 rounded-lg bg-muted p-1">
        {tabs.map((tItem) => (
          <button
            key={tItem.key}
            type="button"
            className={cn(
              "flex-1 rounded-md px-1.5 py-1.5 text-xs font-medium transition-colors sm:px-3 sm:text-sm",
              tab === tItem.key
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            )}
            onClick={() => setTab(tItem.key)}
          >
            {tItem.label}
          </button>
        ))}
      </div>

      {/* ---- list ---- */}
      {loading ? (
        <p className="text-sm text-muted-foreground">{t("dashboard.loading")}</p>
      ) : items.length === 0 ? (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          <CircleHelp className="mx-auto mb-3 size-8 opacity-40" />
          {t("assigned.empty")}
        </div>
      ) : (
        <div className="grid min-w-0 gap-3">
          {items.map((item) => (
            <Link
              key={item.id}
              href={`/topic/${item.topicId}`}
              className="flex flex-col gap-2 rounded-lg border p-4 transition-colors hover:bg-accent sm:flex-row sm:items-start sm:justify-between min-w-0"
            >
              <div className="min-w-0 flex-1">
                <div className="flex min-w-0 items-center gap-2">
                  <p className="truncate text-sm font-medium">
                    {item.topicTitle || `${t("common.topic")} #${item.topicId}`}
                  </p>
                  {item.topicSolved ? (
                    <CheckCircle2 className="h-3.5 w-3.5 shrink-0 text-emerald-500" />
                  ) : null}
                </div>
                <div className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5 text-xs text-muted-foreground">
                  {item.topicCategory ? (
                    <span className="inline-flex items-center rounded-sm bg-accent px-1.5 py-0.5 text-[11px]">
                      {item.topicCategory}
                    </span>
                  ) : null}
                  <span className="inline-flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {prettyDate(item.assignedAt, t)}
                  </span>
                </div>
              </div>
              <Badge
                variant="secondary"
                className={cn(
                  "w-fit shrink-0 sm:ml-3",
                  item.status === "active" && "border-amber-200 bg-amber-50 text-amber-700",
                  item.status === "resolved" && "border-emerald-200 bg-emerald-50 text-emerald-700"
                )}
              >
                {item.status === "active"
                  ? t("assigned.pending")
                  : item.status === "resolved"
                    ? t("assigned.solved")
                    : t("assigned.dismissed")}
              </Badge>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
