import { useCurrentUser } from "@/components/app/app-provider"
import { RequireUser } from "@/components/auth/require-user"
import { PrivateUserCenterPage } from "@/components/user/private-user-center-page"
import type { Favorite, PageData, UserMessage } from "@/lib/api/types"
import { useI18n } from "@/lib/i18n/provider"
import { useDocumentTitle } from "@/lib/use-document-title"

const emptyFavorites: PageData<Favorite> = {
  results: [],
  cursor: "",
  hasMore: false,
}
const emptyMessages: PageData<UserMessage> = {
  results: [],
  cursor: "",
  hasMore: false,
}
export function PrivateCenter({
  kind,
}: {
  kind: "favorites" | "messages"
}) {
  const { t } = useI18n()
  useDocumentTitle(t(kind === "favorites" ? "user.favorites.title" : "user.messages.title"))
  const data = kind === "favorites" ? emptyFavorites : emptyMessages
  const user = useCurrentUser()
  const redirectPath = kind === "favorites" ? "/user/favorites" : "/user/messages"
  return (
    <RequireUser initialUser={user} redirectPath={redirectPath}>
      <PrivateUserCenterPage
        kind={kind}
        initialData={data as never}
        initialFans={[]}
        initialFollowed={[]}
        serverLoaded={false}
      />
    </RequireUser>
  )
}
