import { serverApiFetch as apiFetch } from "./server"

import type {
  PageData,
  Tag,
  Topic,
  TopicHideContent,
  Category,
  UserSummary,
} from "./types"

export type TopicAssignment = {
  id: number
  topicId: number
  userId: number
  userIdEncode?: string
  assignedBy: number
  status: "active" | "resolved" | "dismissed"
  assignedAt: number
  resolvedAt: number
  createTime: number
  nickname?: string
  username?: string
}

type TopicParams = Record<string, string | number | boolean | undefined>
type SearchTopicParams = {
  keyword: string
  categoryId?: number
  timeRange?: number
  cursor?: string
}

export function getTopics(params: TopicParams = {}) {
  return apiFetch<PageData<Topic>>("/api/topic/topics", { params })
}

export function getCategoryTopics(
  categoryId: string | number,
  params: TopicParams = {}
) {
  return getTopics({ ...params, categoryId })
}

export function getTagTopics(tagId: string | number, cursor?: string) {
  return apiFetch<PageData<Topic>>("/api/topic/tag/topics", {
    params: { tagId, cursor },
  })
}

export function getTopic(id: string) {
  return apiFetch<Topic>(`/api/topic/${id}`)
}

export type TopicEditData = Pick<
  Topic,
  | "id"
  | "type"
  | "title"
  | "content"
  | "attachments"
> & {
  categoryId: number
  contentType?: "html" | "markdown" | string
  hideContent?: string
  tags?: string[] | null
}

export function getTopicEdit(id: string) {
  return apiFetch<TopicEditData>(`/api/topic/edit/${id}`)
}

export function getTopicRecentLikes(id: string) {
  return apiFetch<UserSummary[] | null>(`/api/topic/recentlikes/${id}`)
}

export function getTopicHideContent(id: string) {
  return apiFetch<TopicHideContent>("/api/topic/hide_content", {
    params: { topicId: id },
  })
}

export function getCategory(categoryId: string | number) {
  return apiFetch<Category>("/api/topic/category", { params: { categoryId } })
}

export function getCategories() {
  return apiFetch<Category[]>("/api/topic/categories")
}

export function getCategoryNavs() {
  return apiFetch<Category[]>("/api/topic/category_navs")
}

export function getTag(id: string | number) {
  return apiFetch<Tag>(`/api/tag/${id}`)
}

export function searchTopics(params: SearchTopicParams) {
  return apiFetch<PageData<Topic>>("/api/search/topic", {
    params,
  })
}

// Claim APIs
export function getTopicClaims(topicId: string | number) {
  return apiFetch<TopicAssignment[]>(`/api/topic/${topicId}/claims`)
}

export function claimTopic(topicId: string | number) {
  return apiFetch<TopicAssignment>(`/api/topic/${topicId}/claim`, {
    method: "POST",
  })
}

export function unclaimTopic(topicId: string | number, userId?: number) {
  const params = new URLSearchParams()
  if (userId) params.set("userId", String(userId))
  return apiFetch<null>(`/api/topic/${topicId}/unclaim`, {
    method: "POST",
    body: params,
  })
}

export function dismissClaim(topicId: string | number, userId: number) {
  return apiFetch<null>(`/api/topic/${topicId}/dismiss-claim`, {
    method: "POST",
    body: new URLSearchParams({ userId: String(userId) }),
  })
}

export function getMyClaimed(status?: string, limit?: number) {
  return apiFetch<TopicAssignment[]>("/api/assignment/my", {
    params: { status, limit },
  })
}

export function getMyClaimedCount() {
  return apiFetch<{ count: number }>("/api/assignment/my/count")
}
