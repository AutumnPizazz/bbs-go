"use client"

import * as React from "react"
import { useRouter } from "@/lib/router/navigation"
import { AlertCircle, Plus, Trash2 } from "lucide-react"

import { TagInput } from "@/components/common/tag-input"
import {
  CaptchaChallenge,
  type CaptchaChallengeHandle,
} from "@/components/auth/captcha-field"
import {
  ConfirmDialog,
  type ConfirmDialogState,
} from "@/components/common/confirm-dialog"
import { ContentEditor } from "@/components/editor/content-editor"
import { CategoryQuickSelector } from "@/components/topic/category-selector"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { apiFetch } from "@/lib/api/client"
import {
  getEditorModeOptions,
  getEditorSwitchConfirmMessage,
  type EditorMode,
} from "@/lib/editor-mode"
import type {
  SiteConfig,
  Topic,
  TopicAttachment,
  Category,
  UserSummary,
} from "@/lib/api/types"
import { useI18n } from "@/lib/i18n/provider"
import { formatDate } from "@/lib/format"
import {
  findCategory,
  getFirstCategoryId,
  hasCategory,
  filterCategoryTree,
} from "@/lib/categories"
import { useToastActions } from "@/lib/toast"

type TopicCreateFormState = {
  type: number
  categoryId: number
  title: string
  tags: string[]
  contentType: "html" | "markdown"
  content: string
  hideContent: string
  vote: TopicVoteForm | null
  attachmentIds: string[]
}

type TopicVoteForm = {
  type: 1 | 2
  title: string
  expiredAt: number
  voteNum: number
  options: Array<{ content: string }>
}

const DEFAULT_ATTACHMENT_ACCEPT =
  ".pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.md,.csv,.zip,.rar,.7z,.tar,.gz"

function titleForType(type: number, t: ReturnType<typeof useI18n>["t"]) {
  if (type === 2) return t("pages.topic.create.qa")
  return t("pages.topic.create.post")
}

function publishLabelForType(type: number, t: ReturnType<typeof useI18n>["t"]) {
  if (type === 2) return t("pages.topic.create.qaBtn")
  return t("pages.topic.create.postBtn")
}

function categoryTypeMatches(topicType: number) {
  return (node: Category) =>
    topicType === 2 ? node.type === "qa" : node.type !== "qa"
}

function createInitialForm({
  type,
  categoryId,
  contentType,
}: {
  type: number
  categoryId: number
  contentType: TopicCreateFormState["contentType"]
}): TopicCreateFormState {
  return {
    type,
    categoryId,
    title: "",
    tags: [],
    contentType,
    content: "",
    hideContent: "",
    vote: null,
    attachmentIds: [],
  }
}

function TopicAttachmentField({
  value,
  config,
  categoryId,
  uploading,
  onUploadingChange,
  onChange,
}: {
  value: TopicAttachment[]
  config?: SiteConfig["attachmentConfig"]
  categoryId: number
  uploading: boolean
  onUploadingChange: (value: boolean) => void
  onChange: (value: TopicAttachment[]) => void
}) {
  const { t } = useI18n()
  const { catchError } = useToastActions()
  const inputRef = React.useRef<HTMLInputElement>(null)
  const maxCount = config?.maxCount ?? 5
  const maxSizeMB = config?.maxSizeMB ?? 10
  const hasCountLimit = maxCount > 0
  const hasSizeLimit = maxSizeMB > 0
  const limitHint =
    hasCountLimit && hasSizeLimit
      ? t("pages.topic.create.attachment.limitHint", { maxCount, maxSizeMB })
      : hasCountLimit
        ? t("pages.topic.create.attachment.sizeUnlimitedHint", { maxCount })
        : hasSizeLimit
          ? t("pages.topic.create.attachment.countUnlimitedHint", { maxSizeMB })
          : t("pages.topic.create.attachment.unlimitedHint")
  const configuredAllowedTypes = config?.allowedTypes
  const hasWildcardType =
    Array.isArray(configuredAllowedTypes) &&
    configuredAllowedTypes.some((type) => {
      const normalized = type.trim().toLowerCase()
      return normalized === "*" || normalized === "*/*"
    })
  const accept = hasWildcardType
    ? undefined
    : Array.isArray(configuredAllowedTypes) && configuredAllowedTypes.length
      ? configuredAllowedTypes.join(",")
      : DEFAULT_ATTACHMENT_ACCEPT

  async function upload(file: File) {
    onUploadingChange(true)
    try {
      const body = new FormData()
      body.append("file", file, file.name)
      body.append("categoryId", String(categoryId))
      const attachment = await apiFetch<TopicAttachment>(
        "/api/attachment/upload",
        {
          method: "POST",
          body,
        }
      )
      onChange([...value, attachment])
    } catch (error) {
      catchError(error)
    } finally {
      onUploadingChange(false)
    }
  }

  return (
    <div className="rounded-md border border-dashed bg-muted/20 p-3">
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-sm text-muted-foreground">
          {t("pages.topic.create.attachment.label")}
        </span>
        <span className="text-xs text-muted-foreground">{limitHint}</span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          disabled={uploading || (hasCountLimit && value.length >= maxCount)}
          onClick={() => inputRef.current?.click()}
        >
          {t("pages.topic.create.attachment.add")}
        </Button>
        <input
          ref={inputRef}
          type="file"
          className="hidden"
          accept={accept}
          onChange={(event) => {
            const file = event.currentTarget.files?.[0]
            if (file) void upload(file)
            event.currentTarget.value = ""
          }}
        />
      </div>
      {uploading ? (
        <div className="mt-2 text-xs text-muted-foreground">
          {t("pages.topic.create.attachmentUploading")}
        </div>
      ) : null}
      {value.length ? (
        <ul className="mt-2 space-y-2 text-sm">
          {value.map((attachment, index) => (
            <li
              key={attachment.id || index}
              className="flex flex-col gap-2 rounded border bg-background p-2 sm:flex-row sm:items-center"
            >
              <div className="min-w-0 flex-1">
                <span className="block truncate font-medium">
                  {attachment.fileName}
                </span>
                <span className="text-xs text-muted-foreground">
                  {attachment.fileSize || 0} B
                </span>
              </div>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="text-muted-foreground hover:text-destructive"
                aria-label={t("pages.topic.create.attachment.remove")}
                onClick={() =>
                  onChange(value.filter((_, itemIndex) => itemIndex !== index))
                }
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  )
}

function defaultVote(): TopicVoteForm {
  return {
    type: 1,
    title: "",
    expiredAt: Date.now() + 24 * 60 * 60 * 1000,
    voteNum: 1,
    options: [{ content: "" }, { content: "" }],
  }
}

function cloneVote(vote: TopicVoteForm): TopicVoteForm {
  return {
    type: vote.type === 2 ? 2 : 1,
    title: vote.title || "",
    expiredAt: vote.expiredAt || Date.now() + 24 * 60 * 60 * 1000,
    voteNum: Number(vote.voteNum) || (vote.type === 2 ? 2 : 1),
    options:
      vote.options && vote.options.length >= 2
        ? vote.options.map((option) => ({ content: option.content || "" }))
        : [{ content: "" }, { content: "" }],
  }
}

function VoteEditor({
  vote,
  onChange,
}: {
  vote: TopicVoteForm
  onChange: (vote: TopicVoteForm) => void
}) {
  const { t } = useI18n()
  const dateValue = new Date(vote.expiredAt).toISOString().slice(0, 16)

  return (
    <div className="mt-2 space-y-3 rounded-md border bg-background p-3">
      <Input
        value={vote.title}
        placeholder={t("pages.topic.create.vote.titlePlaceholder")}
        onChange={(event) =>
          onChange({ ...vote, title: event.currentTarget.value })
        }
      />
      <div className="space-y-2">
        {vote.options.map((option, index) => (
          <div key={index} className="flex items-center gap-2">
            <span className="w-5 text-xs text-muted-foreground">
              {index + 1}.
            </span>
            <Input
              value={option.content}
              placeholder={t("pages.topic.create.vote.optionPlaceholder", {
                index: index + 1,
              })}
              onChange={(event) =>
                onChange({
                  ...vote,
                  options: vote.options.map((item, itemIndex) =>
                    itemIndex === index
                      ? { content: event.currentTarget.value }
                      : item
                  ),
                })
              }
            />
            <Button
              type="button"
              variant="outline"
              size="icon-sm"
              disabled={vote.options.length <= 2}
              onClick={() =>
                onChange({
                  ...vote,
                  options: vote.options.filter(
                    (_, itemIndex) => itemIndex !== index
                  ),
                  voteNum: Math.min(vote.voteNum, vote.options.length - 1),
                })
              }
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>
      <Button
        type="button"
        variant="outline"
        size="sm"
        disabled={vote.options.length >= 20}
        onClick={() =>
          onChange({ ...vote, options: [...vote.options, { content: "" }] })
        }
      >
        <Plus className="h-4 w-4" />
        {t("pages.topic.create.vote.addOption")}
      </Button>
      <div className="flex flex-wrap items-center gap-2">
        <Button
          type="button"
          size="sm"
          variant={vote.type === 1 ? "default" : "outline"}
          onClick={() => onChange({ ...vote, type: 1, voteNum: 1 })}
        >
          {t("pages.topic.create.vote.single")}
        </Button>
        <Button
          type="button"
          size="sm"
          variant={vote.type === 2 ? "default" : "outline"}
          onClick={() =>
            onChange({ ...vote, type: 2, voteNum: Math.max(2, vote.voteNum) })
          }
        >
          {t("pages.topic.create.vote.multipleShort")}
        </Button>
        {vote.type === 2 ? (
          <>
            <span className="text-xs text-muted-foreground">
              {t("pages.topic.create.vote.voteNum")}
            </span>
            <Input
              type="number"
              min="1"
              max={vote.options.length}
              className="h-8 w-20"
              value={vote.voteNum}
              onChange={(event) =>
                onChange({
                  ...vote,
                  voteNum: Number(event.currentTarget.value) || 1,
                })
              }
            />
          </>
        ) : null}
      </div>
      <Input
        type="datetime-local"
        value={dateValue}
        onChange={(event) =>
          onChange({
            ...vote,
            expiredAt: new Date(event.currentTarget.value).getTime(),
          })
        }
      />
    </div>
  )
}

function VoteEditorModal({
  open,
  editing,
  vote,
  onOpenChange,
  onChange,
  onConfirm,
}: {
  open: boolean
  editing: boolean
  vote: TopicVoteForm
  onOpenChange: (open: boolean) => void
  onChange: (vote: TopicVoteForm) => void
  onConfirm: () => void
}) {
  const { t } = useI18n()

  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
      <div className="flex max-h-[85vh] w-full max-w-2xl flex-col overflow-hidden rounded-lg border bg-background shadow-lg">
        <div className="border-b px-4 py-3 text-lg font-semibold">
          {editing
            ? t("pages.topic.create.vote.editTitle")
            : t("pages.topic.create.vote.addTitle")}
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          <VoteEditor vote={vote} onChange={onChange} />
        </div>
        <div className="flex justify-end gap-2 border-t px-4 py-3">
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            {t("dialog.cancel")}
          </Button>
          <Button type="button" onClick={onConfirm}>
            {t("dialog.ok")}
          </Button>
        </div>
      </div>
    </div>
  )
}

function validateVote(
  vote: TopicVoteForm | null,
  t: ReturnType<typeof useI18n>["t"],
  msgWarning: (message: string) => void
) {
  if (!vote) return true
  if (!vote.title.trim()) {
    msgWarning(t("pages.topic.create.vote.validateTitle"))
    return false
  }
  if (!vote.expiredAt || vote.expiredAt <= Date.now()) {
    msgWarning(t("pages.topic.create.vote.validateExpiredAt"))
    return false
  }
  const options = vote.options
    .map((option) => option.content.trim())
    .filter(Boolean)
  if (options.length < 2) {
    msgWarning(t("pages.topic.create.vote.validateOptionCount"))
    return false
  }
  if (
    new Set(options.map((item) => item.toLowerCase())).size !== options.length
  ) {
    msgWarning(t("pages.topic.create.vote.validateOptionDuplicate"))
    return false
  }
  if (vote.type === 2 && (vote.voteNum <= 0 || vote.voteNum > options.length)) {
    msgWarning(t("pages.topic.create.vote.validateVoteNum"))
    return false
  }
  return true
}

export function TopicCreateForm({
  contentType,
  currentUser,
  config,
  categoryId,
  categories,
  type,
}: {
  contentType: TopicCreateFormState["contentType"]
  currentUser: UserSummary
  config: SiteConfig | null
  categoryId: number
  categories: Category[]
  type: number
}) {
  const router = useRouter()
  const { t } = useI18n()
  const { catchError, msgWarning } = useToastActions()
  const editorModeOptions = React.useMemo(() => getEditorModeOptions(t), [t])
  const captchaRef = React.useRef<CaptchaChallengeHandle>(null)
  const lastSubmitAtRef = React.useRef(0)
  const [publishing, setPublishing] = React.useState(false)
  const [attachmentList, setAttachmentList] = React.useState<TopicAttachment[]>(
    []
  )
  const [attachmentUploading, setAttachmentUploading] = React.useState(false)
  const [voteModalOpen, setVoteModalOpen] = React.useState(false)
  const [voteEditing, setVoteEditing] = React.useState(false)
  const [voteDraft, setVoteDraft] = React.useState<TopicVoteForm>(defaultVote())
  const [confirmState, setConfirmState] =
    React.useState<ConfirmDialogState>(null)
  const [form, setForm] = React.useState<TopicCreateFormState>(() =>
    createInitialForm({
      type,
      categoryId: categoryId || config?.defaultCategoryId || 0,
      contentType,
    })
  )

  const availableNodes = React.useMemo(
    () => filterCategoryTree(categories, categoryTypeMatches(form.type)),
    [form.type, categories]
  )
  const effectiveCategoryId = hasCategory(availableNodes, form.categoryId)
    ? form.categoryId
    : getFirstCategoryId(availableNodes)
  const effectiveAttachmentConfig =
    findCategory(availableNodes, effectiveCategoryId)?.attachmentConfig ??
    config?.attachmentConfig
  const noQaCategoriesAvailable = form.type === 2 && availableNodes.length === 0
  const featureDisabledMessage = config
    ? form.type === 2 && !config.modules?.qa
        ? t("pages.topic.create.qaFeatureDisabled")
        : form.type === 0 && !config.modules?.topic
          ? t("pages.topic.create.topicFeatureDisabled")
          : null
    : null

  function updateForm(next: Partial<TopicCreateFormState>) {
    setForm((current) => ({ ...current, ...next }))
  }

  function switchEditor(nextContentType: EditorMode) {
    const currentContentType: EditorMode =
      form.contentType === "markdown" ? "markdown" : "html"
    if (nextContentType === currentContentType) {
      return
    }
    if (form.content.trim()) {
      setConfirmState({
        description: getEditorSwitchConfirmMessage(currentContentType, t),
        confirmText: t("common.confirm"),
        onConfirm: () => {
          updateForm({
            content: "",
            contentType: nextContentType,
          })
        },
      })
      return
    }
    updateForm({
      content: "",
      contentType: nextContentType,
    })
  }

  async function publish(
    captcha?: ReturnType<CaptchaChallengeHandle["getCaptcha"]>
  ) {
    if (publishing) return
    const now = Date.now()
    if (now - lastSubmitAtRef.current < 500) {
      return
    }
    lastSubmitAtRef.current = now
    if (form.type === 2 && !hasCategory(availableNodes, effectiveCategoryId)) {
      msgWarning(t("pages.topic.create.noQaCategorySubmit"))
      return
    }
    if (attachmentUploading) {
      msgWarning(t("pages.topic.create.attachmentUploading"))
      return
    }
    if (!validateVote(form.vote, t, msgWarning)) {
      return
    }

    setPublishing(true)
    try {
      const data = await apiFetch<Topic>("/api/topic/create", {
        method: "POST",
        body: {
          ...form,
          categoryId: effectiveCategoryId,
          attachmentIds:
            form.type === 0 ? attachmentList.map((item) => item.id) : [],
          vote: form.vote
            ? {
                ...form.vote,
                voteNum: form.vote.type === 1 ? 1 : form.vote.voteNum,
                options: form.vote.options.map((option) => ({
                  content: option.content.trim(),
                })),
              }
            : null,
          captchaId: captcha?.captchaId || "",
          captchaCode: captcha?.captchaCode || "",
          captchaProtocol: captcha?.captchaProtocol || 2,
        },
      })
      router.push(`/topic/${data.id}`)
    } catch (error) {
      catchError(error)
      setPublishing(false)
      captchaRef.current?.reset()
    }
  }

  function submit() {
    if (config?.topicCaptcha) {
      void captchaRef.current?.open()
      return
    }
    void publish()
  }

  function showVoteEditor(vote?: TopicVoteForm | null) {
    setVoteEditing(Boolean(vote))
    setVoteDraft(vote ? cloneVote(vote) : defaultVote())
    setVoteModalOpen(true)
  }

  function confirmVote() {
    if (!validateVote(voteDraft, t, msgWarning)) {
      return
    }
    updateForm({
      vote: {
        ...voteDraft,
        type: voteDraft.type === 2 ? 2 : 1,
        voteNum: voteDraft.type === 1 ? 1 : voteDraft.voteNum,
        options: voteDraft.options.map((option) => ({
          content: option.content.trim(),
        })),
      },
    })
    setVoteModalOpen(false)
  }

  if (featureDisabledMessage) {
    return (
      <Alert>
        <AlertCircle className="h-4 w-4 shrink-0" />
        <AlertTitle>{featureDisabledMessage}</AlertTitle>
      </Alert>
    )
  }

  if (noQaCategoriesAvailable) {
    return (
      <Alert>
        <AlertCircle className="h-4 w-4 shrink-0" />
        <AlertTitle>{t("pages.topic.create.noQaCategoryTitle")}</AlertTitle>
        <AlertDescription>
          {t("pages.topic.create.noQaCategoryDescription")}
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <>
      <div className="publish-form">
        <div className="form-title">
          <div className="form-title-name">{titleForType(form.type, t)}</div>
          {form.type !== 2 ? (
            <div
              className="editor-mode-switch flex"
              aria-label={t("component.editorMode.switchLabel")}
            >
              <span className="editor-mode-switch-label">
                {t("component.editorMode.label")}
              </span>
              <Tabs
                value={form.contentType === "markdown" ? "markdown" : "html"}
                onValueChange={(value) => switchEditor(value as EditorMode)}
              >
                <TabsList className="h-7 p-0.5 group-data-horizontal/tabs:h-7">
                  {editorModeOptions.map((option) => (
                    <TabsTrigger
                      key={option.value}
                      value={option.value}
                      className="h-6 px-2 py-0 text-xs"
                    >
                      {option.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>
            </div>
          ) : null}
        </div>

        <div className="field">
          <CategoryQuickSelector
            value={effectiveCategoryId}
            categories={availableNodes}
            onChange={(categoryId) => updateForm({ categoryId })}
          />
        </div>

        <div className="field">
          <Input
            value={form.title}
            placeholder={t("pages.topic.create.titlePlaceholder")}
            onChange={(event) => updateForm({ title: event.currentTarget.value })}
          />
        </div>

        <div className="field">
          <ContentEditor
            contentType={form.contentType === "markdown" ? "markdown" : "html"}
            value={form.content}
            placeholder={t("pages.topic.create.contentPlaceholder")}
            height="400px"
            onChange={(content) => updateForm({ content })}
          />
        </div>

        {form.type !== 2 && config?.enableHideContent ? (
          <div className="field">
            <ContentEditor
              contentType="html"
              value={form.hideContent}
              placeholder={t("pages.topic.detail.hideContent")}
              height="200px"
              onChange={(hideContent) => updateForm({ hideContent })}
            />
          </div>
        ) : null}

        <div className="field">
          <TagInput
            value={form.tags}
            recommendTags={config?.recommendTags}
            placeholder={t("component.tagInput.placeholder")}
            onChange={(tags) => updateForm({ tags })}
          />
        </div>

        {form.type === 0 && effectiveAttachmentConfig?.enabled ? (
          <div className="field">
            <TopicAttachmentField
              value={attachmentList}
              config={effectiveAttachmentConfig}
              categoryId={effectiveCategoryId}
              uploading={attachmentUploading}
              onUploadingChange={setAttachmentUploading}
              onChange={setAttachmentList}
            />
          </div>
        ) : null}

        {form.type !== 2 ? (
          <div className="field rounded-md border border-dashed bg-muted/20 p-3">
            <div className="flex flex-wrap items-center gap-2">
              {!form.vote ? (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => showVoteEditor()}
                >
                  {t("pages.topic.create.vote.addEntry")}
                </Button>
              ) : (
                <>
                  <span className="inline-flex items-center rounded-full border border-emerald-300 bg-emerald-50 px-2 py-0.5 text-xs text-emerald-700">
                    {t("pages.topic.create.vote.added")}
                  </span>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => showVoteEditor(form.vote)}
                  >
                    {t("pages.topic.create.vote.edit")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => updateForm({ vote: null })}
                  >
                    {t("pages.topic.create.vote.remove")}
                  </Button>
                </>
              )}
            </div>
            {form.vote ? (
              <div className="mt-2 rounded-md border bg-background px-2 py-2 text-xs text-muted-foreground">
                <div>{form.vote.title}</div>
                <div className="mt-1">
                  {form.vote.type === 1
                    ? t("pages.topic.create.vote.single")
                    : t("pages.topic.create.vote.multiple", {
                        num: form.vote.voteNum,
                      })}
                  {" · "}
                  {form.vote.options.length}{" "}
                  {t("pages.topic.create.vote.optionsCount")}
                </div>
                <div className="mt-1">
                  {t("pages.topic.create.vote.expiredAt")}:{" "}
                  {formatDate(form.vote.expiredAt)}
                </div>
              </div>
            ) : null}
          </div>
        ) : null}

        <div className="form-footer">
          <Button
            type="button"
            disabled={
              publishing || attachmentUploading
            }
            onClick={submit}
          >
            {publishLabelForType(form.type, t)}
          </Button>
        </div>
      </div>
      <VoteEditorModal
        open={voteModalOpen}
        editing={voteEditing}
        vote={voteDraft}
        onOpenChange={setVoteModalOpen}
        onChange={setVoteDraft}
        onConfirm={confirmVote}
      />
      <CaptchaChallenge
        ref={captchaRef}
        onVerified={() => {
          const captcha = captchaRef.current?.getCaptcha()
          void publish(captcha)
        }}
      />
      <ConfirmDialog
        state={confirmState}
        onOpenChange={(open) => {
          if (!open) setConfirmState(null)
        }}
      />
    </>
  )
}
