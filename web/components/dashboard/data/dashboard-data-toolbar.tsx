"use client"

import {
  BookmarkIcon,
  ListChecksIcon,
  PlusIcon,
  RefreshCwIcon,
  SearchIcon,
  XIcon,
} from "lucide-react"

import type { AdminFormValue } from "@/lib/api/admin"
import { Button } from "@/components/ui/button"
import { NativeSelect, NativeSelectOption } from "@/components/ui/native-select"

import { DashboardDataFilterControl } from "./dashboard-data-filter-control"
import type {
  DashboardDataFilter,
  DashboardDataOption,
} from "./dashboard-data-types"

export function DashboardDataToolbar({
  filters,
  values,
  asyncOptions,
  loading,
  canCreate,
  error,
  searchLabel,
  refreshLabel,
  createLabel,
  saveLabel,
  onFilterChange,
  onRefresh,
  onCreate,
  onSaveFilters,
  savedViews,
  onLoadView,
  loadViewLabel,
  bulkActions,
  selectedCount,
  selectedLabel,
  clearSelectionLabel,
  onClearSelection,
}: {
  filters?: DashboardDataFilter[]
  values: Record<string, AdminFormValue>
  asyncOptions: Record<string, DashboardDataOption[]>
  loading: boolean
  canCreate: boolean
  error: string | null
  searchLabel: string
  refreshLabel: string
  createLabel: string
  saveLabel: string
  onFilterChange: (name: string, value: AdminFormValue) => void
  onRefresh: () => void
  onCreate: () => void
  onSaveFilters: () => void
  savedViews: string[]
  onLoadView: (name: string) => void
  loadViewLabel: string
  bulkActions?: Array<{ label: string; onClick: () => void }>
  selectedCount: number
  selectedLabel: (count: number) => string
  clearSelectionLabel: string
  onClearSelection: () => void
}) {
  return (
    <div className="rounded-lg border bg-[var(--dashboard-panel)] p-3 text-card-foreground shadow-xs">
      <div className="flex flex-wrap items-end gap-2">
        {filters?.map((filter) => (
          <DashboardDataFilterControl
            key={filter.name}
            filter={filter}
            value={values[filter.name]}
            options={[
              ...(filter.options ?? []),
              ...(asyncOptions[filter.name] ?? []),
            ]}
            onChange={(value) => onFilterChange(filter.name, value)}
          />
        ))}
        <Button onClick={onRefresh} disabled={loading}>
          <SearchIcon />
          {searchLabel}
        </Button>
        <Button
          variant="outline"
          size="icon"
          onClick={onRefresh}
          disabled={loading}
        >
          <RefreshCwIcon />
          <span className="sr-only">{refreshLabel}</span>
        </Button>
        <Button variant="outline" onClick={onSaveFilters} disabled={loading}>
          <BookmarkIcon />
          {saveLabel}
        </Button>
        {savedViews.length ? (
          <NativeSelect
            value=""
            onChange={(event) => onLoadView(event.target.value)}
            aria-label={loadViewLabel}
          >
            <NativeSelectOption value="">{loadViewLabel}</NativeSelectOption>
            {savedViews.map((view) => (
              <NativeSelectOption key={view} value={view}>
                {view}
              </NativeSelectOption>
            ))}
          </NativeSelect>
        ) : null}
        {canCreate ? (
          <Button className="ml-auto" onClick={onCreate}>
            <PlusIcon />
            {createLabel}
          </Button>
        ) : null}
      </div>

      {selectedCount > 0 && bulkActions?.length ? (
        <div className="mt-3 flex flex-wrap items-center gap-2 border-t pt-3">
          <span className="mr-1 inline-flex items-center gap-1.5 text-sm font-medium">
            <ListChecksIcon className="size-4" />
            {selectedLabel(selectedCount)}
          </span>
          <Button type="button" variant="ghost" onClick={onClearSelection}>
            <XIcon />
            {clearSelectionLabel}
          </Button>
          {bulkActions.map((action) => (
            <Button
              key={action.label}
              type="button"
              variant="outline"
              onClick={action.onClick}
            >
              <ListChecksIcon />
              {action.label}
            </Button>
          ))}
        </div>
      ) : null}

      {error ? (
        <div className="mt-3 rounded-md border border-destructive/25 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      ) : null}
    </div>
  )
}
