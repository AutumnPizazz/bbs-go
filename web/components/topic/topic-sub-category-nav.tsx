"use client"

import Link from "@/components/common/link"
import * as React from "react"
import { ChevronDown, ChevronRight } from "lucide-react"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { ScrollArea, ScrollBar } from "@/components/ui/scroll-area"
import type { Category } from "@/lib/api/types"
import { useI18n } from "@/lib/i18n/provider"
import { cn } from "@/lib/utils"

// ----------------------------------------------------------------
// tiny helpers – no memo, no callback, plain functions
// ----------------------------------------------------------------

function findAncestors(tree: Category[], targetId: number): number[] {
  function walk(nodes: Category[], path: number[]): number[] | null {
    for (const n of nodes) {
      if (n.id === targetId) return path
      if (n.children?.length) {
        const r = walk(n.children, [...path, n.id])
        if (r) return r
      }
    }
    return null
  }
  return walk(tree, []) ?? []
}

function isOnActivePath(
  tree: Category[],
  nodeId: number,
  currentCategoryId: number
): boolean {
  if (nodeId === currentCategoryId) return true
  const ancestors = findAncestors(tree, currentCategoryId)
  return ancestors.includes(nodeId)
}

// ----------------------------------------------------------------
// Main component
// ----------------------------------------------------------------

export function TopicSubCategoryNav({
  categoryTree,
  currentCategoryId,
}: {
  categoryTree: Category[]
  currentCategoryId: number
}) {
  const { t } = useI18n()

  // ---------- local state: which node ids have their children visible ----------
  const [expandedIds, setExpandedIds] = React.useState<number[]>(() => {
    if (currentCategoryId <= 0) return []
    const ancestors = findAncestors(categoryTree, currentCategoryId)
    // root + ancestors + self
    const rootId = ancestors.length > 0 ? ancestors[0] : currentCategoryId
    return [rootId, ...ancestors, currentCategoryId]
  })

  // reset when category changes
  const prevCatRef = React.useRef(currentCategoryId)
  if (prevCatRef.current !== currentCategoryId) {
    prevCatRef.current = currentCategoryId
    // mutate during render is safe here because we immediately setState below
  }
  React.useEffect(() => {
    if (currentCategoryId <= 0) {
      setExpandedIds([])
      return
    }
    const ancestors = findAncestors(categoryTree, currentCategoryId)
    const rootId = ancestors.length > 0 ? ancestors[0] : currentCategoryId
    setExpandedIds([rootId, ...ancestors, currentCategoryId])
  }, [categoryTree, currentCategoryId])

  function toggle(id: number) {
    setExpandedIds((prev) => {
      if (prev.includes(id)) {
        return prev.filter((x) => x !== id)
      }
      return [...prev, id]
    })
  }

  // ---------- build rows (plain computation during render) ----------
  const rows: { parentId: number; parentName: string; depth: number; children: Category[] }[] = []

  function walk(nodes: Category[], depth: number) {
    for (const node of nodes) {
      const kids = node.children?.length ? node.children : []
      if (expandedIds.includes(node.id) && kids.length > 0) {
        rows.push({
          parentId: node.id,
          parentName: node.name,
          depth,
          children: kids,
        })
      }
      if (node.children?.length) {
        walk(node.children, depth + 1)
      }
    }
  }
  walk(categoryTree, 0)

  if (!rows.length) return null

  // ---------- render ----------
  return (
    <div className="border-b border-border px-4 pt-3 pb-2">
      {rows.map((row, rowIdx) => {
        const isLast = rowIdx === rows.length - 1

        return (
          <div
            key={`${row.parentId}-${row.depth}`}
            className={cn("min-w-0", !isLast && "mb-2")}
          >
            {/* header for nested rows */}
            {row.depth > 0 ? (
              <div
                className="mb-1 flex items-center gap-1 pl-1 text-[11px] text-muted-foreground"
                style={{ paddingLeft: `${(row.depth - 1) * 16}px` }}
              >
                <ChevronDown className="h-3 w-3 shrink-0" />
                <Link
                  href={`/topics/category/${row.parentId}`}
                  className="truncate font-medium text-foreground/80 hover:text-foreground hover:underline"
                >
                  {row.parentName}
                </Link>
              </div>
            ) : null}

            {/* pills row */}
            <div
              className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-2 text-xs sm:text-[13px]"
              style={{
                paddingLeft: row.depth > 0 ? `${(row.depth - 1) * 16 + 20}px` : "0px",
              }}
            >
              <div className="min-w-0">
                <ScrollArea className="topic-sub-node-scroll min-w-0 whitespace-nowrap [&_[data-slot=scroll-area-viewport]]:overflow-y-hidden">
                  <div className="flex w-max flex-nowrap items-center gap-1 pr-1 pb-1">
                    {/* "All" pill for root row */}
                    {row.depth === 0 ? (
                      <span className="inline-flex shrink-0 items-center rounded-md text-sm font-medium whitespace-nowrap">
                        <Link
                          href={`/topics/category/${row.parentId}`}
                          className={cn(
                            "inline-flex items-center rounded-md px-3 py-1 transition-colors",
                            currentCategoryId === row.parentId
                              ? "bg-primary text-primary-foreground shadow-sm"
                              : "bg-muted text-muted-foreground hover:bg-muted/80 hover:text-foreground"
                          )}
                        >
                          {t("pages.topics.allCategories")}
                        </Link>
                      </span>
                    ) : null}

                    {row.children.map((child) => {
                      const hasKids =
                        child.hasChildren ||
                        (child.children && child.children.length > 0)
                      const expanded = expandedIds.includes(child.id)
                      const onPath = isOnActivePath(
                        categoryTree,
                        child.id,
                        currentCategoryId
                      )

                      return (
                        <span
                          key={child.id}
                          className="inline-flex shrink-0 items-center rounded-md text-sm font-medium whitespace-nowrap"
                        >
                          <Link
                            href={`/topics/category/${child.id}`}
                            className={cn(
                              "inline-flex items-center px-3 py-1 transition-colors",
                              hasKids
                                ? "rounded-l-md rounded-r-none"
                                : "rounded-md",
                              onPath
                                ? "bg-primary text-primary-foreground shadow-sm"
                                : "bg-muted text-muted-foreground hover:bg-muted/80 hover:text-foreground"
                            )}
                          >
                            {child.name}
                          </Link>
                          {hasKids ? (
                            <span
                              className={cn(
                                "inline-flex shrink-0 cursor-pointer items-center self-stretch rounded-r-md border-l px-1.5 transition-colors",
                                onPath
                                  ? "border-primary-foreground/30 bg-primary text-primary-foreground hover:bg-primary/90"
                                  : "border-muted-foreground/20 bg-muted text-muted-foreground hover:bg-muted/80 hover:text-foreground"
                              )}
                              onClick={(e) => {
                                // Use onMouseDown + preventDefault to kill any
                                // navigation before it starts, then toggle on click.
                                e.preventDefault()
                                e.stopPropagation()
                                toggle(child.id)
                              }}
                              role="button"
                              tabIndex={0}
                              onKeyDown={(e) => {
                                if (e.key === "Enter" || e.key === " ") {
                                  e.preventDefault()
                                  e.stopPropagation()
                                  toggle(child.id)
                                }
                              }}
                              aria-label={expanded ? "Collapse" : "Expand"}
                            >
                              {expanded ? (
                                <ChevronDown className="h-3.5 w-3.5" />
                              ) : (
                                <ChevronRight className="h-3.5 w-3.5" />
                              )}
                            </span>
                          ) : null}
                        </span>
                      )
                    })}
                  </div>
                  <ScrollBar orientation="horizontal" />
                </ScrollArea>
              </div>

              {/* overflow dropdown */}
              {row.children.length > 3 ? (
                <div className="pb-0.5">
                  <DropdownMenu modal={false}>
                    <DropdownMenuTrigger asChild>
                      <button
                        type="button"
                        className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
                        aria-label={t("pages.topics.moreSubCategories")}
                      >
                        <ChevronDown className="h-4 w-4" />
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                      align="end"
                      className="max-h-[min(60vh,420px)] w-[280px] sm:w-[320px] md:w-[360px]"
                    >
                      {row.depth === 0 ? (
                        <>
                          <DropdownMenuItem
                            className={cn(
                              currentCategoryId === row.parentId &&
                                "bg-accent text-accent-foreground"
                            )}
                            asChild
                          >
                            <Link href={`/topics/category/${row.parentId}`}>
                              {t("pages.topics.allCategories")}
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                        </>
                      ) : null}
                      {row.children.map((child) => {
                        const hasKids =
                          child.hasChildren ||
                          (child.children && child.children.length > 0)
                        const expanded = expandedIds.includes(child.id)
                        const onPath = isOnActivePath(
                          categoryTree,
                          child.id,
                          currentCategoryId
                        )
                        return (
                          <DropdownMenuItem
                            key={`menu-${child.id}`}
                            className={cn(
                              onPath && "bg-accent text-accent-foreground"
                            )}
                            asChild
                          >
                            <div className="flex w-full items-center gap-1">
                              <Link
                                href={`/topics/category/${child.id}`}
                                className="min-w-0 flex-1 truncate"
                              >
                                {child.name}
                              </Link>
                              {hasKids ? (
                                <button
                                  type="button"
                                  className="flex h-6 w-6 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                                  onClick={(e) => {
                                    e.preventDefault()
                                    e.stopPropagation()
                                    toggle(child.id)
                                  }}
                                  aria-label={expanded ? "Collapse" : "Expand"}
                                >
                                  {expanded ? (
                                    <ChevronDown className="h-3.5 w-3.5" />
                                  ) : (
                                    <ChevronRight className="h-3.5 w-3.5" />
                                  )}
                                </button>
                              ) : null}
                            </div>
                          </DropdownMenuItem>
                        )
                      })}
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              ) : null}
            </div>
          </div>
        )
      })}
    </div>
  )
}
