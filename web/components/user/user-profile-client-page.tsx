"use client"

import * as React from "react"
import Link from "@/components/common/link"
import { FileText, MessageSquare, UserPlus, Users } from "lucide-react"

import { ArticleList } from "@/components/article/article-list"
import { useCurrentUser } from "@/components/app/app-provider"
import { EmptyState } from "@/components/common/empty-state"
import { LoadMore } from "@/components/common/load-more"
import { PageError, PageLoading } from "@/components/common/page-state"
import { WidgetCard } from "@/components/common/widget-card"
import { TopicListItem } from "@/components/topic/topic-list-item"
import { UserCenterShell } from "@/components/user/user-center-shell"
import { UserFollowList } from "@/components/user/user-follow-list"
import { apiFetch } from "@/lib/api/client"
import type { PageData, Topic, UserSummary } from "@/lib/api/types"
import { useI18n } from "@/lib/i18n/provider"
import { useRouteData, useRouteSegment } from "@/lib/spa-route"
import { useDocumentTitle } from "@/lib/use-document-title"

type UserShellData = {
  user: UserSummary
  fans: UserSummary[]
  followed: UserSummary[]
}

type UserProfileData = UserShellData & { topics: PageData<Topic> }
type UserArticlesData = UserShellData & { articles: PageData<Topic> }
type UserFollowData = UserShellData & { pageData: PageData<UserSummary> }

const emptyPage: PageData<Topic> = { results: [], cursor: "0", hasMore: false }
const emptyArticlePage: PageData<Topic> = {
  results: [],
  cursor: "0",
  hasMore: false,
}
const emptyUserPage: PageData<UserSummary> = {
  results: [],
  cursor: "0",
  hasMore: false,
}

async function loadUserShellData(userId: string): Promise<UserShellData> {
  const [user, fans, followed] = await Promise.all([
    apiFetch<UserSummary>(`/api/user/${userId}`),
    apiFetch<PageData<UserSummary>>("/api/fans/recent/fans", {
      params: { userId },
    })
      .then((data) => data.results || [])
      .catch(() => []),
    apiFetch<PageData<UserSummary>>("/api/fans/recent/follow", {
      params: { userId },
    })
      .then((data) => data.results || [])
      .catch(() => []),
  ])

  return { user, fans, followed }
}

function userDisplayName(user: UserSummary | null | undefined) {
  return (
    user?.nickname || user?.username || (user?.id ? `#${user.id}` : undefined)
  )
}

export function UserProfileClientPage({
  initialUser = null,
}: {
  initialUser?: UserSummary | null
}) {
  const userId = useRouteSegment(1)
  const currentUser = useCurrentUser()
  const { t } = useI18n()
  const load = React.useCallback(async (): Promise<UserProfileData> => {
    const [shell, topics] = await Promise.all([
      loadUserShellData(userId),
      apiFetch<PageData<Topic>>("/api/topic/user_topics", {
        params: { userId },
      }).catch(() => emptyPage),
    ])

    return { ...shell, topics }
  }, [userId])
  const { data, loading, error } = useRouteData(`user:${userId}`, load)
  useDocumentTitle(userDisplayName(data?.user ?? initialUser))

  if (loading) return <PageLoading />
  if (error || !data) return <PageError message={error} />

  const { user, topics, fans, followed } = data
  const loadMoreLabels = {
    loadMore: t("common.loadMore.loadMore"),
    noMore: t("common.loadMore.noMore"),
  }

  return (
    <UserCenterShell
      user={user}
      currentUser={currentUser}
      fans={fans}
      followed={followed}
      t={t}
    >
      <WidgetCard>
        <nav className="mb-2 inline-flex h-9 items-center justify-center rounded-lg bg-muted p-[3px] text-muted-foreground">
          <Link
            href={`/user/${user.id}`}
            className="inline-flex h-full items-center justify-center gap-1.5 rounded-md bg-background px-3 py-1 text-sm font-medium text-foreground shadow-sm"
          >
            <MessageSquare className="h-3.5 w-3.5" aria-hidden="true" />
            <span>{t("pages.user.topics")}</span>
          </Link>
          <Link
            href={`/user/${user.id}/articles`}
            className="inline-flex h-full items-center justify-center gap-1.5 rounded-md px-3 py-1 text-sm font-medium text-foreground/60 hover:text-foreground"
          >
            <FileText className="h-3.5 w-3.5" aria-hidden="true" />
            <span>{t("pages.user.articles")}</span>
          </Link>
        </nav>
        <LoadMore<Topic>
          initialItems={topics.results || []}
          initialCursor={topics.cursor}
          initialHasMore={topics.hasMore}
          initialLoad
          resetKey={`user-topics:${userId}:${topics.cursor}:${topics.hasMore}`}
          labels={loadMoreLabels}
          loadPage={({ cursor }) =>
            apiFetch<PageData<Topic>>("/api/topic/user_topics", {
              params: { userId, cursor },
            })
          }
          renderItems={(items) => (
            <ul className="divide-y divide-border">
              {items.map((topic) => (
                <TopicListItem key={topic.id} topic={topic} t={t} />
              ))}
            </ul>
          )}
          renderEmpty={() => <EmptyState title={t("common.noData")} />}
        />
      </WidgetCard>
    </UserCenterShell>
  )
}

export function UserArticlesClientPage() {
  const userId = useRouteSegment(1)
  const currentUser = useCurrentUser()
  const { t } = useI18n()
  const load = React.useCallback(async (): Promise<UserArticlesData> => {
    const [shell, articles] = await Promise.all([
      loadUserShellData(userId),
      apiFetch<PageData<Topic>>("/api/topic/user_topics", {
        params: { userId, format: "article" },
      }).catch(() => emptyArticlePage),
    ])

    return { ...shell, articles }
  }, [userId])
  const { data, loading, error } = useRouteData(`user-articles:${userId}`, load)

  if (loading) return <PageLoading />
  if (error || !data) return <PageError message={error} />

  const { user, articles, fans, followed } = data
  const loadMoreLabels = {
    loadMore: t("common.loadMore.loadMore"),
    noMore: t("common.loadMore.noMore"),
  }

  return (
    <UserCenterShell
      user={user}
      currentUser={currentUser}
      fans={fans}
      followed={followed}
      t={t}
    >
      <WidgetCard>
        <nav className="mb-2 inline-flex h-9 items-center justify-center rounded-lg bg-muted p-[3px] text-muted-foreground">
          <Link
            href={`/user/${user.id}`}
            className="inline-flex h-full items-center justify-center gap-1.5 rounded-md px-3 py-1 text-sm font-medium text-foreground/60 hover:text-foreground"
          >
            <MessageSquare className="h-3.5 w-3.5" aria-hidden="true" />
            <span>{t("pages.user.topics")}</span>
          </Link>
          <Link
            href={`/user/${user.id}/articles`}
            className="inline-flex h-full items-center justify-center gap-1.5 rounded-md bg-background px-3 py-1 text-sm font-medium text-foreground shadow-sm"
          >
            <FileText className="h-3.5 w-3.5" aria-hidden="true" />
            <span>{t("pages.user.articles")}</span>
          </Link>
        </nav>
        <LoadMore<Topic>
          initialItems={articles.results || []}
          initialCursor={articles.cursor}
          initialHasMore={articles.hasMore}
          initialLoad
          resetKey={`user-articles:${userId}:${articles.cursor}:${articles.hasMore}`}
          labels={loadMoreLabels}
          loadPage={({ cursor }) =>
          apiFetch<PageData<Topic>>("/api/topic/user_topics", {
            params: { userId, cursor, format: "article" },
            })
          }
          renderItems={(items) => <ArticleList articles={items} t={t} />}
          renderEmpty={() => <EmptyState title={t("common.noData")} />}
        />
      </WidgetCard>
    </UserCenterShell>
  )
}

export function UserFansClientPage() {
  return <UserFollowClientPage kind="fans" />
}

export function UserFollowedClientPage() {
  return <UserFollowClientPage kind="followed" />
}

function UserFollowClientPage({ kind }: { kind: "fans" | "followed" }) {
  const userId = useRouteSegment(1)
  const currentUser = useCurrentUser()
  const { t } = useI18n()
  const load = React.useCallback(async (): Promise<UserFollowData> => {
    const [shell, pageData] = await Promise.all([
      loadUserShellData(userId),
      apiFetch<PageData<UserSummary>>(
        kind === "fans" ? "/api/fans/fans" : "/api/fans/followed",
        { params: { userId } }
      ).catch(() => emptyUserPage),
    ])

    return { ...shell, pageData }
  }, [kind, userId])
  const { data, loading, error } = useRouteData(`user-${kind}:${userId}`, load)

  if (loading) return <PageLoading />
  if (error || !data) return <PageError message={error} />

  const { user, pageData, fans, followed } = data
  const labels = {
    loadMore: t("common.loadMore.loadMore"),
    noMore: t("common.loadMore.noMore"),
  }

  return (
    <UserCenterShell
      user={user}
      currentUser={currentUser}
      fans={fans}
      followed={followed}
      t={t}
    >
      <WidgetCard>
        <nav className="mb-2 inline-flex h-9 items-center justify-center rounded-lg bg-muted p-[3px] text-muted-foreground">
          <Link
            href={`/user/${user.id}/fans`}
            className={
              kind === "fans"
                ? "inline-flex h-full items-center justify-center gap-1.5 rounded-md bg-background px-3 py-1 text-sm font-medium text-foreground shadow-sm"
                : "inline-flex h-full items-center justify-center gap-1.5 rounded-md px-3 py-1 text-sm font-medium text-foreground/60 hover:text-foreground"
            }
          >
            <Users className="h-3.5 w-3.5" />
            <span>{t("pages.user.fans")}</span>
          </Link>
          <Link
            href={`/user/${user.id}/followed`}
            className={
              kind === "followed"
                ? "inline-flex h-full items-center justify-center gap-1.5 rounded-md bg-background px-3 py-1 text-sm font-medium text-foreground shadow-sm"
                : "inline-flex h-full items-center justify-center gap-1.5 rounded-md px-3 py-1 text-sm font-medium text-foreground/60 hover:text-foreground"
            }
          >
            <UserPlus className="h-3.5 w-3.5" />
            <span>{t("pages.user.followed")}</span>
          </Link>
        </nav>
        <UserFollowList users={pageData.results || []} />
        <LoadMore<UserSummary>
          initialCursor={pageData.cursor}
          initialHasMore={pageData.hasMore}
          resetKey={`user-${kind}:${userId}:${pageData.cursor}:${pageData.hasMore}`}
          labels={labels}
          loadPage={({ cursor }) =>
            apiFetch<PageData<UserSummary>>(
              kind === "fans" ? "/api/fans/fans" : "/api/fans/followed",
              { params: { userId, cursor } }
            )
          }
          renderItems={(items) => <UserFollowList users={items} />}
          alwaysShowButton
        />
      </WidgetCard>
    </UserCenterShell>
  )
}
