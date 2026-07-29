"use client"

import * as React from "react"
import { CheckCircle2, CircleCheckBig, MessageCircle, PenLine, UserCheck } from "lucide-react"
import Link from "@/components/common/link"
import type { MySummary } from "@/lib/api/topics"
import { prettyDate } from "@/lib/format"
import type { TFunction } from "@/lib/i18n"

// ----------------------------------------------------------------
// Activity icon & label per type
// ----------------------------------------------------------------
const activityMeta: Record<
  MySummary["recentActivity"][number]["type"],
  { icon: typeof PenLine; labelKey: string }
> = {
  claimed:       { icon: UserCheck,     labelKey: "component.activitySummary.claimed" },
  answered:      { icon: MessageCircle, labelKey: "component.activitySummary.answered" },
  accepted:      { icon: CheckCircle2,  labelKey: "component.activitySummary.accepted" },
  topic_created: { icon: PenLine,       labelKey: "component.activitySummary.topicCreated" },
}

// ----------------------------------------------------------------
// Main component
// ----------------------------------------------------------------
export function UserActivitySummary({
  data,
  t,
}: {
  data: MySummary
  t: TFunction
}) {
  if (!data.recentActivity?.length) return null

  return (
    <div>
      {/* ---- stat cards ---- */}
      <div className="grid grid-cols-4 gap-3 mb-6">
        <StatCard value={data.topicCount}    label={t("component.activitySummary.topics")} />
        <StatCard value={data.commentCount}  label={t("component.activitySummary.comments")} />
        <StatCard value={data.acceptedCount} label={t("component.activitySummary.acceptedAns")} />
        <StatCard value={data.activeClaims}  label={t("component.activitySummary.claims")} />
      </div>

      {/* ---- activity timeline ---- */}
      <div className="space-y-1">
        {data.recentActivity.map((item, idx) => {
          const meta = activityMeta[item.type]
          const Icon = meta.icon
          return (
            <Link
              key={`${item.type}-${item.topicId}-${idx}`}
              href={`/topic/${item.topicId}`}
              className="flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors hover:bg-accent"
            >
              <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="min-w-0 flex-1 truncate text-muted-foreground">
                {t(meta.labelKey, { title: item.topicTitle })}
              </span>
              <span className="shrink-0 text-xs text-muted-foreground/70">
                {prettyDate(item.time, t)}
              </span>
            </Link>
          )
        })}
      </div>
    </div>
  )
}

function StatCard({ value, label }: { value: number; label: string }) {
  return (
    <div className="rounded-lg border bg-card px-3 py-3 text-center">
      <div className="text-xl font-bold tabular-nums">{value}</div>
      <div className="mt-0.5 text-[11px] text-muted-foreground">{label}</div>
    </div>
  )
}
