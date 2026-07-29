export type EntityId = string

export interface ApiEnvelope<T> {
  success?: boolean
  data: T
  message?: string
  errorCode?: number
}

export interface PageData<T> {
  cursor?: string
  hasMore: boolean
  results: T[]
}

export interface UserSummary {
  id: EntityId
  username?: string
  nickname?: string
  avatar?: string
  smallAvatar?: string
  description?: string
  backgroundImage?: string
  smallBackgroundImage?: string
  homePage?: string
  createTime?: number
  topicCount?: number
  commentCount?: number
  fansCount?: number
  followCount?: number
  forbidden?: boolean
  followed?: boolean
  roles?: string[]
  permissions?: string[]
  passwordSet?: boolean
}

export interface ImageInfo {
  name?: string
  url?: string
  preview?: string
  size?: number
}

export interface Category {
  id: number
  name: string
  description?: string
  logo?: string
  parentId?: number
  topicCount?: number
  qaCount?: number
  solvedCount?: number
  unsolvedCount?: number
  hasChildren?: boolean
  attachmentConfig?: AttachmentConfig
  children?: Category[]
}

export interface AttachmentConfig {
  enabled?: boolean
  allowedTypes?: string[]
  maxSizeMB?: number
  maxCount?: number
}

export interface Tag {
  id: number
  name: string
  description?: string
}

export interface Attachment {
  id: EntityId
  fileName?: string
  fileUrl?: string
  fileSize?: number
  downloadCount?: number
}

export interface TopicAttachment extends Attachment {
  createTime?: number
}

export interface TopicTocItem {
  id: string
  title: string
  level?: number
}

export interface TopicVoteOption {
  id: number
  content: string
  sortNo?: number
  voteCount?: number
  percent?: number
  voted?: boolean
}

export interface TopicVote {
  id: number
  type?: 1 | 2 | "single" | "multiple" | string
  title?: string
  expiredAt?: number
  voteNum?: number
  optionCount?: number
  voteCount?: number
  expired?: boolean
  voted?: boolean
  optionIds?: number[]
  options?: TopicVoteOption[]
}

export interface Topic {
  id: EntityId
  type?: number
  title?: string
  content?: string
  summary?: string
  createTime?: number
  updateTime?: number
  user: UserSummary
  category?: Category
  tags?: Tag[]
  sticky?: boolean
  recommend?: boolean
  liked?: boolean
  likeCount?: number
  commentCount?: number
  viewCount?: number
  qaStatus?: "solved" | "unsolved" | string
  attachments?: Attachment[]
  toc?: TopicTocItem[]
  favorited?: boolean
  ipLocation?: string
  status?: number
  vote?: TopicVote | null
  acceptedCommentId?: number
}

export interface TopicHideContent {
  content?: string
  exists: boolean
  show: boolean
}

export interface Comment {
  id: number
  user: UserSummary
  entityType?: string
  entityId?: number | string
  contentType?: "text" | "html" | string
  content?: string
  imageList?: ImageInfo[]
  attachments?: Attachment[]
  likeCount?: number
  commentCount?: number
  liked?: boolean
  quoteId?: number
  quote?: Comment | null
  replies?: PageData<Comment> | null
  ipLocation?: string
  status?: number
  createTime?: number
}

export interface SiteNav {
  title: string
  url: string
  openInNewWindow?: boolean
  children?: SiteNav[]
}

export interface SiteConfig {
  language?: string
  siteTitle?: string
  siteDescription?: string
  siteKeywords?: string[]
  baseURL?: string
  siteLogo?: string
  siteNavs?: SiteNav[]
  siteNotification?: string
  recommendTags?: string[]
  defaultCategoryId?: number
  topicListStyle?: "default" | "compact" | string
  topicCaptcha?: boolean
	enableHideContent?: boolean
  attachmentConfig?: AttachmentConfig
  footerLinks?: Array<{
    text?: Record<string, string>
    url?: string
    visible?: boolean
    openInNewWindow?: boolean
  }>
  modules?: {
    topic?: boolean
    qa?: boolean
    [key: string]: boolean | undefined
  }
  loginConfig?: {
    passwordLogin?: { enabled?: boolean }
    githubLogin?: { enabled?: boolean }
    googleLogin?: { enabled?: boolean; clientId?: string }
    weixinLogin?: { enabled?: boolean }
  }
  scriptInjections?: Array<{
    enabled?: boolean
    scriptName?: string
    type?: "external" | "inline" | string
    src?: string
    code?: string
    async?: boolean
    defer?: boolean
    crossorigin?: string
  }>
}

export interface LoginResult {
  user: UserSummary
  token: string
  redirect?: string
}

export interface SearchTopic {
  id: number
  user?: UserSummary
  category?: Category
  tags?: Tag[]
  title?: string
  summary?: string
  createTime?: number
}

export interface SearchUser {
  user?: UserSummary
  nickname?: string
  username?: string
  description?: string
  createTime?: number
}

export interface SearchAllResult {
  topics?: SearchTopic[]
  users?: SearchUser[]
}

export interface Favorite {
  id: number
  entityType?: string
  entityId?: number
  deleted?: boolean
  title?: string
  content?: string
  user?: UserSummary
  url?: string
  createTime?: number
}

export interface UserMessage {
  id: number
  from: UserSummary
  userId?: number
  title?: string
  content?: string
  quoteContent?: string
  type?: number
  detailUrl?: string
  extraData?: string
  status?: number
  createTime?: number
}

export interface BindInfo {
  bind?: boolean
  nickname?: string
}
