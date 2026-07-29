"use client"

import Link from "@/components/common/link"
import * as React from "react"
import {
  ChevronDown,
  ChevronRight,
  LayoutGridIcon,
  MoreHorizontalIcon,
} from "lucide-react"

import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer"
import { ScrollArea } from "@/components/ui/scroll-area"
import { apiFetch } from "@/lib/api/client"
import type { Category } from "@/lib/api/types"
import { useI18n } from "@/lib/i18n/provider"
import { cn } from "@/lib/utils"

// ----------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------

function nodeHref(node: Category) {
  return `/topics/category/${node.id}`
}

/** Collect ancestor ids from root → parent of target (excludes target). */
function findAncestorIds(
  tree: Category[],
  targetId: number,
  ancestors: number[] = []
): number[] | null {
  for (const node of tree) {
    if (node.id === targetId) return ancestors
    if (node.children?.length) {
      const r = findAncestorIds(node.children, targetId, [...ancestors, node.id])
      if (r) return r
    }
  }
  return null
}

// ----------------------------------------------------------------
// Chevron toggle (rendered INSIDE the Link – no z-index conflict)
// ----------------------------------------------------------------

function TreeChevron({
  canExpand,
  expanded,
  onToggle,
}: {
  canExpand: boolean
  expanded: boolean
  onToggle: () => void
}) {
  if (!canExpand) {
    // Spacer so leaf nodes align with expandable siblings.
    return <span className="inline-flex h-7 w-7 shrink-0" />
  }

  return (
    <span
      role="button"
      tabIndex={0}
      className="inline-flex h-7 w-7 shrink-0 cursor-pointer items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
      onClick={(e) => {
        e.preventDefault()
        e.stopPropagation()
        onToggle()
      }}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault()
          e.stopPropagation()
          onToggle()
        }
      }}
      aria-label={expanded ? "Collapse" : "Expand"}
    >
      {expanded ? (
        <ChevronDown className="h-4 w-4" />
      ) : (
        <ChevronRight className="h-4 w-4" />
      )}
    </span>
  )
}

// ----------------------------------------------------------------
// Recursive tree node
// ----------------------------------------------------------------

function CategoryTreeNode({
  node,
  depth,
  expandedIds,
  onToggle,
  currentCategoryId,
  currentRootCategoryId,
  isMobile,
}: {
  node: Category
  depth: number
  expandedIds: Set<number>
  onToggle: (id: number) => void
  currentCategoryId?: number
  currentRootCategoryId?: number
  isMobile?: boolean
}) {
  const { t } = useI18n()
  const expanded = expandedIds.has(node.id)
  const canExpand = !!(
    node.hasChildren ||
    (node.children && node.children.length > 0)
  )
  const isRootActive = currentRootCategoryId === node.id
  const isLeafActive = currentCategoryId === node.id
  const active = isRootActive || (isLeafActive && node.id === currentRootCategoryId)

  const badgeEl = node.unsolvedCount ? (
    <span className="ml-auto shrink-0 rounded-full bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700 dark:bg-amber-900/50 dark:text-amber-300">
      {node.unsolvedCount}
    </span>
  ) : null

  const logoEl =
    node.logo ? (
      <i
        className="node-logo shrink-0"
        style={{ backgroundImage: `url(${node.logo})` }}
      />
    ) : (
      <i className="node-logo shrink-0" />
    )

  const indent = depth * 16

  // --- mobile: inside Drawer ---
  if (isMobile) {
    return (
      <>
        <div className={cn("flex w-full items-center", active && "active")}>
          <DrawerClose asChild>
            <Link
              href={nodeHref(node)}
              className={cn(
                "topics-mobile-category-item flex-1",
                active && "active"
              )}
              style={{ paddingLeft: `${indent + 8}px` }}
            >
              <TreeChevron
                canExpand={canExpand}
                expanded={expanded}
                onToggle={() => onToggle(node.id)}
              />
              {logoEl}
              <span>{node.name}</span>
              {badgeEl}
            </Link>
          </DrawerClose>
        </div>
        {expanded && node.children?.length ? (
          <div className="ml-3">
            {node.children.map((child) => (
              <CategoryTreeNode
                key={child.id}
                node={child}
                depth={depth + 1}
                expandedIds={expandedIds}
                onToggle={onToggle}
                currentCategoryId={currentCategoryId}
                currentRootCategoryId={currentRootCategoryId}
                isMobile
              />
            ))}
          </div>
        ) : null}
      </>
    )
  }

  // --- desktop ---
  return (
    <li
      className={cn("select-none", active && "active")}
      data-node-id={node.id}
    >
      <Link
        href={nodeHref(node)}
        className="flex items-center gap-2 py-1"
        style={{ paddingLeft: `${indent + 8}px` }}
      >
        <TreeChevron
          canExpand={canExpand}
          expanded={expanded}
          onToggle={() => onToggle(node.id)}
        />
        {logoEl}
        <div className="node-name truncate">{node.name}</div>
        {badgeEl}
      </Link>
      {expanded && node.children?.length ? (
        <ul className="dock-nav-list">
          {node.children.map((child) => (
            <CategoryTreeNode
              key={child.id}
              node={child}
              depth={depth + 1}
              expandedIds={expandedIds}
              onToggle={onToggle}
              currentCategoryId={currentCategoryId}
              currentRootCategoryId={currentRootCategoryId}
            />
          ))}
        </ul>
      ) : null}
    </li>
  )
}

// ----------------------------------------------------------------
// Main exported component
// ----------------------------------------------------------------

export function TopicsNavContent({
  initialCategories,
  currentCategoryId,
  currentRootCategoryId,
}: {
  initialCategories: Category[]
  currentCategoryId?: number
  currentRootCategoryId?: number
}) {
  const { t } = useI18n()
  const [categories, setCategories] = React.useState(initialCategories)
  const [mobileDrawerOpen, setMobileDrawerOpen] = React.useState(false)
  const mobileScrollRef = React.useRef<HTMLDivElement>(null)

  // --- fetch full tree client-side ---
  React.useEffect(() => {
    let mounted = true
    void apiFetch<Category[]>("/api/topic/category_navs")
      .then((data) => {
        if (mounted) setCategories(data)
      })
      .catch(() => undefined)
    return () => {
      mounted = false
    }
  }, [])

  // --- expand / collapse state ---
  const targetId =
    currentCategoryId !== undefined && currentCategoryId > 0
      ? currentCategoryId
      : currentRootCategoryId || 0

  function buildExpandedSet(tree: Category[], tgt: number): Set<number> {
    const s = new Set<number>()
    // Root nodes start expanded by default.
    for (const cat of tree) {
      if (cat.id > 0) s.add(cat.id)
    }
    if (tgt > 0) {
      const ancestors = findAncestorIds(tree, tgt) ?? []
      for (const id of ancestors) s.add(id)
    }
    return s
  }

  const [expandedIds, setExpandedIds] = React.useState<Set<number>>(() =>
    buildExpandedSet(categories, targetId)
  )

  // Fully rebuild expanded set when tree or target changes.
  const prevKeyRef = React.useRef("")
  React.useEffect(() => {
    const key = `${targetId}|${categories.map((c) => c.id).join(",")}`
    if (key !== prevKeyRef.current) {
      prevKeyRef.current = key
      setExpandedIds(buildExpandedSet(categories, targetId))
    }
  }, [categories, targetId])

  const handleToggle = React.useCallback((id: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }, [])

  // --- derived ---
  const visibleCategories = categories.filter((node) => node.id > 0)
  const allCategoryLabel = t("pages.topics.allCategories")
  const moreCategoriesLabel = t("pages.topics.moreCategories")
  const activeNodeId =
    currentCategoryId !== undefined && currentCategoryId <= 0
      ? "all"
      : String(currentRootCategoryId || currentCategoryId || "")

  // Scroll mobile row so active node is visible.
  React.useEffect(() => {
    if (!activeNodeId) return
    window.requestAnimationFrame(() => {
      const el = mobileScrollRef.current?.querySelector<HTMLElement>(
        `[data-node-id="${activeNodeId}"]`
      )
      el?.scrollIntoView({ behavior: "smooth", block: "nearest", inline: "center" })
    })
  }, [activeNodeId, visibleCategories.length])

  // --- render ---

  function renderDesktopTree() {
    return (
      <ul className="dock-nav-list">
        <li
          className={cn(
            currentCategoryId !== undefined && currentCategoryId <= 0 && "active"
          )}
          data-node-id="all"
        >
          <Link href="/topics" className="flex items-center gap-2 py-1" style={{ paddingLeft: "8px" }}>
            {/* Spacer matching TreeChevron width so icon aligns with category logos */}
            <span className="inline-flex h-7 w-7 shrink-0" />
            <LayoutGridIcon className="node-logo node-logo-icon" aria-hidden="true" />
            <div className="node-name">{allCategoryLabel}</div>
          </Link>
        </li>
        {visibleCategories.map((node) => (
          <CategoryTreeNode
            key={node.id}
            node={node}
            depth={0}
            expandedIds={expandedIds}
            onToggle={handleToggle}
            currentCategoryId={currentCategoryId}
            currentRootCategoryId={currentRootCategoryId}
          />
        ))}
      </ul>
    )
  }

  function renderMobileTree() {
    return (
      <div className="topics-mobile-category-list">
        <DrawerClose asChild>
          <Link
            href="/topics"
            className={cn(
              "topics-mobile-category-item",
              currentCategoryId !== undefined && currentCategoryId <= 0 && "active"
            )}
          >
            <LayoutGridIcon aria-hidden="true" />
            <span>{allCategoryLabel}</span>
          </Link>
        </DrawerClose>
        {visibleCategories.map((node) => (
          <CategoryTreeNode
            key={`mobile-${node.id}`}
            node={node}
            depth={0}
            expandedIds={expandedIds}
            onToggle={handleToggle}
            currentCategoryId={currentCategoryId}
            currentRootCategoryId={currentRootCategoryId}
            isMobile
          />
        ))}
      </div>
    )
  }

  return (
    <div className="topics-nav">
      <nav className="dock-nav">
        {/* Desktop */}
        <ScrollArea className="topics-scroll-area dock-nav-desktop-scroll">
          {renderDesktopTree()}
        </ScrollArea>

        {/* Mobile */}
        <div className="dock-nav-mobile-row">
          <div
            ref={mobileScrollRef}
            className="dock-nav-scroll"
            aria-label={allCategoryLabel}
          >
            {renderDesktopTree()}
          </div>
          <Drawer open={mobileDrawerOpen} onOpenChange={setMobileDrawerOpen}>
            <DrawerTrigger asChild>
              <button
                type="button"
                className="topics-mobile-category-more"
                aria-label={moreCategoriesLabel}
                title={moreCategoriesLabel}
              >
                <MoreHorizontalIcon aria-hidden="true" />
                <span>{t("pages.topic.categorySelector.more")}</span>
              </button>
            </DrawerTrigger>
            <DrawerContent className="topics-mobile-category-drawer">
              <DrawerHeader>
                <DrawerTitle>{moreCategoriesLabel}</DrawerTitle>
              </DrawerHeader>
              {renderMobileTree()}
            </DrawerContent>
          </Drawer>
        </div>
      </nav>
    </div>
  )
}
