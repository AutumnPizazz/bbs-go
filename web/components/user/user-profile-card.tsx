"use client"

import * as React from "react"
import Link from "@/components/common/link"

import { UserAvatar } from "@/components/common/avatar"
import { FollowButton } from "@/components/user/follow-button"
import { BackgroundUploadButton } from "@/components/user/image-upload"
import type { UserSummary } from "@/lib/api/types"

type UserProfileSummary = UserSummary & {
  smallBackgroundImage?: string
}

function displayName(user: UserProfileSummary) {
  return user.nickname || user.username || `#${user.id}`
}

export function UserProfileCard({
  user,
  currentUser,
}: {
  user: UserProfileSummary
  currentUser?: UserSummary | null
}) {
  const [backgroundImage, setBackgroundImage] = React.useState(
    user.smallBackgroundImage || user.backgroundImage || ""
  )
  const isOwner = currentUser?.id === user.id

  return (
    <section
      className="profile"
      style={
        backgroundImage
          ? { backgroundImage: `url(${backgroundImage})` }
          : undefined
      }
    >
      {isOwner ? (
        <BackgroundUploadButton onUploaded={setBackgroundImage} />
      ) : null}
      <div className="profile-avatar">
        <UserAvatar user={user} size={100} />
      </div>
      <div className="profile-info">
        <div className="metas">
          <div className="nickname-row">
            <span className="nickname">
              <Link
                href={`/user/${user.id}`}
                className="text-foreground hover:underline"
              >
                {displayName(user)}
              </Link>
            </span>
          </div>
          {user.description ? (
            <div className="description">
              <p>{user.description}</p>
            </div>
          ) : null}
        </div>
        <div className="action-btns">
          {isOwner ? null : (
            <FollowButton userId={user.id} initialFollowed={user.followed} />
          )}
        </div>
      </div>
    </section>
  )
}
