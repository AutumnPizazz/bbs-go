"use client"

import * as React from "react"
import { XIcon, HandshakeIcon, Loader2Icon, CheckCircle2Icon } from "lucide-react"

import { useCurrentUser } from "@/components/app/app-provider"
import { Button } from "@/components/ui/button"
import { useI18n } from "@/lib/i18n/provider"
import { msgError, msgSuccess } from "@/lib/toast"
import type { Topic } from "@/lib/api/types"
import {
  getTopicClaims,
  claimTopic,
  unclaimTopic,
  dismissClaim,
  type TopicAssignment,
} from "@/lib/api/topics"

function userLabel(a: TopicAssignment) {
  return a.nickname || a.username || `${"用户"}#${a.userId}`
}

export function TopicClaimPanel({ topic }: { topic: Topic }) {
  const { t } = useI18n()
  const currentUser = useCurrentUser()
  const [claims, setClaims] = React.useState<TopicAssignment[]>([])
  const [acting, setActing] = React.useState(false)

  if (topic.type !== 2) return null
  if (!currentUser) return null

  const isAuthor = topic.user?.id === currentUser.id
  const isAdmin = (currentUser.roles || "").includes("owner") || (currentUser.roles || "").includes("admin")
  const myActiveClaim = claims.find((c) => c.userIdEncode === currentUser.id && c.status === "active")
  const hasActiveClaim = !!myActiveClaim

  const load = React.useCallback(async () => {
    try {
      const list = await getTopicClaims(topic.id)
      setClaims(Array.isArray(list) ? list : [])
    } catch { /* silent */ }
  }, [topic.id])

  React.useEffect(() => { void load() }, [load])

  async function handleClaim() {
    setActing(true)
    try {
      await claimTopic(topic.id)
      msgSuccess(t("component.claim.claimed"))
      await load()
    } catch (err) {
      msgError(err instanceof Error ? err.message : t("dashboard.errors.actionFailed"))
    } finally { setActing(false) }
  }

  async function handleUnclaim() {
    setActing(true)
    try {
      await unclaimTopic(topic.id)
      await load()
    } catch (err) {
      msgError(err instanceof Error ? err.message : t("dashboard.errors.actionFailed"))
    } finally { setActing(false) }
  }

  async function handleDismiss(userId: number) {
    try {
      await dismissClaim(topic.id, userId)
      await load()
    } catch (err) {
      msgError(err instanceof Error ? err.message : t("dashboard.errors.actionFailed"))
    }
  }

  const activeClaims = claims.filter((c) => c.status === "active")
  const resolvedClaims = claims.filter((c) => c.status === "resolved")

  return (
    <div className="rounded-lg border bg-card p-4">
      <div className="mb-3 flex items-center gap-2">
        <HandshakeIcon className="size-4 text-muted-foreground" />
        <h3 className="text-sm font-medium">{t("component.claim.title")}</h3>
      </div>

      {/* Claim / Unclaim button for all users who are not the author */}
      {!isAuthor && (
        hasActiveClaim ? (
          <div className="mb-2 flex items-center justify-between rounded-md bg-muted/50 px-3 py-2">
            <span className="flex items-center gap-2 text-sm text-muted-foreground">
              <CheckCircle2Icon className="size-4 text-emerald-500" />
              {t("component.claim.youClaimed")}
            </span>
            <button
              type="button"
              className="text-xs text-muted-foreground underline hover:text-foreground"
              onClick={() => void handleUnclaim()}
              disabled={acting}
            >
              {t("component.claim.cancel")}
            </button>
          </div>
        ) : (
          <Button
            variant="default"
            size="sm"
            className="mb-2 w-full"
            onClick={() => void handleClaim()}
            disabled={acting}
          >
            {acting ? <Loader2Icon className="mr-2 size-4 animate-spin" /> : <HandshakeIcon className="mr-2 size-4" />}
            {t("component.claim.claimIt")}
          </Button>
        )
      )}

      {/* Active claimers */}
      {activeClaims.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {activeClaims.map((c) => (
            <span
              key={c.id}
              className="inline-flex items-center gap-1 rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900 dark:text-amber-300"
            >
              {userLabel(c)}
              {c.userIdEncode === currentUser.id ? ` (${t("component.claim.you")})` : ""}
              {(isAuthor || isAdmin || c.userIdEncode === currentUser.id) ? (
                <button
                  type="button"
                  className="ml-0.5 rounded-full p-0.5 hover:bg-black/10 dark:hover:bg-white/10"
                  onClick={() => {
                    if (c.userIdEncode === currentUser.id && !isAuthor && !isAdmin) {
                      void handleUnclaim()
                    } else {
                      void handleDismiss(Number(c.userId))
                    }
                  }}
                  aria-label={t("component.claim.remove")}
                >
                  <XIcon className="size-3" />
                </button>
              ) : null}
            </span>
          ))}
        </div>
      ) : (
        <p className="text-xs text-muted-foreground">{t("component.claim.noClaims")}</p>
      )}

      {/* Resolved claims */}
      {resolvedClaims.length > 0 ? (
        <div className="mt-2 flex flex-wrap gap-2">
          {resolvedClaims.map((c) => (
            <span
              key={c.id}
              className="inline-flex items-center gap-1 rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300"
            >
              {userLabel(c)}
            </span>
          ))}
        </div>
      ) : null}
    </div>
  )
}
