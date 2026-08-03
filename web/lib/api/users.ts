import { serverApiFetch as apiFetch } from "./server"

import type { UserMessage, UserSummary } from "./types"

export function getCurrentUser() {
  return apiFetch<UserSummary | null>("/api/user/current")
}

export function getRecentUserMessages() {
  return apiFetch<{ count?: number; messages?: UserMessage[] }>(
    "/api/user/msg_recent"
  )
}
