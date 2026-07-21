import Link from "@/components/common/link"
import { ChevronRight } from "lucide-react"

import { EmptyState } from "@/components/common/empty-state"
import { WidgetCard } from "@/components/common/widget-card"
import { UserCenterOperations } from "@/components/user/user-center-operations"
import { UserFollowList } from "@/components/user/user-follow-list"
import type { UserSummary } from "@/lib/api/types"
import type { TFunction } from "@/lib/i18n"

export function UserCountsCard({
  user,
  t,
}: {
  user: UserSummary
  t: TFunction
}) {
  return (
    <WidgetCard title={t("component.myCounts.title")}>
      <ul className="extra-info">
        <li>
          <span>{t("component.myCounts.topicCount")}</span>
          <br />
          <b>{user.topicCount ?? 0}</b>
        </li>
        <li>
          <span>{t("component.myCounts.commentCount")}</span>
          <br />
          <b>{user.commentCount ?? 0}</b>
        </li>
      </ul>
    </WidgetCard>
  )
}

export function MyProfileCard({
  user,
  currentUser,
  t,
}: {
  user: UserSummary
  currentUser?: UserSummary | null
  t: TFunction
}) {
  const canEdit = currentUser?.id === user.id
  return (
    <WidgetCard
      title={t("component.myProfile.title")}
      actions={
        canEdit ? (
          <Link href="/user/profile" className="inline-flex items-center gap-1">
            {t("component.myProfile.editProfile")}
            <ChevronRight className="h-4 w-4" />
          </Link>
        ) : null
      }
    >
      <div className="stable">
        <div className="str">
          <div className="slabel">{t("component.myProfile.nickname")}</div>
          <div className="svalue">{user.nickname}</div>
        </div>
        <div className="str">
          <div className="slabel">{t("component.myProfile.description")}</div>
          <div className="svalue">{user.description}</div>
        </div>
        {user.homePage ? (
          <div className="str">
            <div className="slabel">{t("component.myProfile.homePage")}</div>
            <div className="svalue">
              <a href={user.homePage} target="_blank" rel="nofollow noreferrer">
                {user.homePage}
              </a>
            </div>
          </div>
        ) : null}
      </div>
    </WidgetCard>
  )
}

export function FollowWidget({
  title,
  count,
  moreHref,
  users,
  t,
}: {
  title: string
  count?: number
  moreHref: string
  users: UserSummary[]
  t: TFunction
}) {
  return (
    <WidgetCard
      title={
        <>
          <span>{title}</span>
          <span>&nbsp;</span>
          <span>{count ?? 0}</span>
        </>
      }
      actions={
        <Link href={moreHref} className="inline-flex items-center gap-1">
          {t("component.fansWidget.more")}
          <ChevronRight className="h-4 w-4" />
        </Link>
      }
    >
      {users.length ? (
        <UserFollowList users={users} />
      ) : (
        <EmptyState title={t("common.noData")} />
      )}
    </WidgetCard>
  )
}

export function UserCenterSidebar({
  user,
  currentUser,
  fans,
  followed,
  t,
}: {
  user: UserSummary
  currentUser?: UserSummary | null
  fans: UserSummary[]
  followed: UserSummary[]
  t: TFunction
}) {
  return (
    <div className="left-container space-y-4">
      <UserCountsCard user={user} t={t} />
      <MyProfileCard user={user} currentUser={currentUser} t={t} />
      <FollowWidget
        title={t("component.fansWidget.title")}
        count={user.fansCount}
        moreHref={`/user/${user.id}/fans`}
        users={fans}
        t={t}
      />
      <FollowWidget
        title={t("component.followWidget.title")}
        count={user.followCount}
        moreHref={`/user/${user.id}/followed`}
        users={followed}
        t={t}
      />
      <UserCenterOperations user={user} currentUser={currentUser} />
    </div>
  )
}
