"use client"

import * as React from "react"
import {
  CheckIcon,
  ChevronRightIcon,
  ChevronsUpDownIcon,
  SearchIcon,
  XIcon,
} from "lucide-react"
import { Popover as PopoverPrimitive } from "radix-ui"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useI18n } from "@/lib/i18n/provider"
import { cn } from "@/lib/utils"

export type DashboardSelectOption = {
  label: string
  value: string | number | boolean
  depth?: number
  searchText?: string
}

type DashboardSelectProps = {
  value?: unknown
  options: DashboardSelectOption[]
  placeholder?: string
  emptyLabel?: string
  emptyValue?: string
  searchPlaceholder?: string
  noResultsLabel?: string
  disabled?: boolean
  allowClear?: boolean
  className?: string
  triggerClassName?: string
  contentClassName?: string
  portalContainer?: HTMLElement | null
  onValueChange: (value: string | undefined) => void
}

const DEFAULT_EMPTY_VALUE = "__empty__"

type DashboardHierarchicalOption = DashboardSelectOption & {
  depth: number
  parentValue?: string | number | boolean
  ancestorValues: string[]
}

function useDashboardHierarchy(
  options: DashboardSelectOption[],
  search: string,
  open: boolean,
  selectedValues: string[]
) {
  const hierarchicalOptions = React.useMemo(() => {
    const stack: DashboardSelectOption[] = []
    return options.map((option) => {
      const depth = Math.max(0, Number(option.depth || 0))
      stack.length = depth
      const parentValue = depth > 0 ? stack[depth - 1]?.value : undefined
      stack[depth] = option
      return { ...option, depth, parentValue }
    })
  }, [options])
  const optionsWithAncestors = React.useMemo<DashboardHierarchicalOption[]>(
    () => {
      const stack: DashboardSelectOption[] = []
      return hierarchicalOptions.map((option) => {
        const ancestorValues = Array.from({ length: option.depth || 0 })
          .map((_, index) => stack[index])
          .filter(
            (ancestor): ancestor is DashboardSelectOption => Boolean(ancestor)
          )
          .map((ancestor) => String(ancestor.value))
        stack[option.depth || 0] = option
        return { ...option, ancestorValues }
      })
    },
    [hierarchicalOptions]
  )
  const parentValues = React.useMemo(
    () =>
      new Set(
        hierarchicalOptions
          .map((option) => option.parentValue)
          .filter(
            (value): value is string | number | boolean => value !== undefined
          )
          .map(String)
      ),
    [hierarchicalOptions]
  )
  const [expandedValues, setExpandedValues] = React.useState<Set<string>>(
    () => new Set()
  )
  const normalizedSearch = search.trim().toLowerCase()
  const filteredOptions = React.useMemo(() => {
    if (!normalizedSearch) {
      return optionsWithAncestors.filter(
        (option) =>
          option.depth === 0 ||
          option.ancestorValues.every((value) => expandedValues.has(value))
      )
    }
    const matchingValues = new Set<string>()
    optionsWithAncestors.forEach((option) => {
      const text = `${option.label} ${option.searchText || ""} ${option.value}`.toLowerCase()
      if (text.includes(normalizedSearch)) {
        matchingValues.add(String(option.value))
        option.ancestorValues.forEach((value) => matchingValues.add(value))
      }
    })
    return optionsWithAncestors.filter((option) =>
      matchingValues.has(String(option.value))
    )
  }, [expandedValues, normalizedSearch, optionsWithAncestors])

  function toggleExpanded(valueToToggle: string) {
    setExpandedValues((current) => {
      const next = new Set(current)
      if (next.has(valueToToggle)) next.delete(valueToToggle)
      else next.add(valueToToggle)
      return next
    })
  }

  React.useEffect(() => {
    if (!open || !selectedValues.length) return
    const selectedValueSet = new Set(selectedValues)
    const selectedAncestorValues = optionsWithAncestors.flatMap((option) =>
      selectedValueSet.has(String(option.value)) ? option.ancestorValues : []
    )
    if (!selectedAncestorValues.length) return
    setExpandedValues((current) => {
      const next = new Set(current)
      selectedAncestorValues.forEach((value) => next.add(value))
      return next
    })
  }, [open, optionsWithAncestors, selectedValues])

  return {
    expandedValues,
    filteredOptions,
    normalizedSearch,
    parentValues,
    toggleExpanded,
  }
}

function HierarchyExpandButton({
  hasChildren,
  searching,
  expanded,
  onToggle,
}: {
  hasChildren: boolean
  searching: boolean
  expanded: boolean
  onToggle: () => void
}) {
  const { t } = useI18n()
  if (!hasChildren || searching) {
    return <span className="size-6 shrink-0" aria-hidden="true" />
  }

  const label = t(
    expanded
      ? "common.accessibility.collapse"
      : "common.accessibility.expand"
  )
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-xs"
      aria-label={label}
      aria-expanded={expanded}
      title={label}
      onClick={onToggle}
    >
      <ChevronRightIcon
        className={cn(
          "size-3 transition-transform",
          expanded && "rotate-90"
        )}
      />
    </Button>
  )
}

export function DashboardSelect({
  value,
  options,
  placeholder,
  emptyLabel,
  emptyValue = DEFAULT_EMPTY_VALUE,
  searchPlaceholder,
  noResultsLabel = "-",
  disabled,
  allowClear = true,
  className,
  triggerClassName,
  contentClassName,
  portalContainer,
  onValueChange,
}: DashboardSelectProps) {
  const { t } = useI18n()
  const [open, setOpen] = React.useState(false)
  const [search, setSearch] = React.useState("")
  const selectedValue =
    value === undefined || value === null || value === ""
      ? emptyValue
      : String(value)
  const selectedOption = options.find(
    (option) => String(option.value) === selectedValue
  )
  const selectedLabel =
    selectedValue === emptyValue
      ? placeholder
      : selectedOption?.label || placeholder
  const canClear = allowClear && selectedValue !== emptyValue
  const hierarchySelectedValues = React.useMemo(
    () => [selectedValue],
    [selectedValue]
  )
  const {
    expandedValues,
    filteredOptions,
    normalizedSearch,
    parentValues,
    toggleExpanded,
  } = useDashboardHierarchy(options, search, open, hierarchySelectedValues)

  function selectValue(nextValue: string) {
    onValueChange(nextValue === emptyValue ? undefined : nextValue)
    setOpen(false)
    setSearch("")
  }

  function clearValue(event: React.MouseEvent<HTMLElement>) {
    event.preventDefault()
    event.stopPropagation()
    selectValue(emptyValue)
  }

  return (
    <PopoverPrimitive.Root open={open} onOpenChange={setOpen}>
      <PopoverPrimitive.Trigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            "w-full justify-between px-3 font-normal",
            (!selectedOption || selectedValue === emptyValue) &&
              "text-muted-foreground",
            triggerClassName,
            className
          )}
        >
          <span className="line-clamp-1 text-left">
            {selectedLabel || placeholder}
          </span>
          {canClear ? (
            <span
              aria-label={t("common.accessibility.clearSelection")}
              className="flex size-4 shrink-0 items-center justify-center rounded-sm opacity-50 transition-opacity hover:opacity-100"
              onClick={clearValue}
              onPointerDown={(event) => {
                event.preventDefault()
                event.stopPropagation()
              }}
            >
              <XIcon className="size-4" />
            </span>
          ) : (
            <ChevronsUpDownIcon className="size-4 shrink-0 opacity-50" />
          )}
        </Button>
      </PopoverPrimitive.Trigger>
      <PopoverPrimitive.Portal container={portalContainer ?? undefined}>
        <PopoverPrimitive.Content
          align="start"
          sideOffset={4}
          className={cn(
            "z-50 w-[var(--radix-popover-trigger-width)] rounded-md border bg-popover p-1 text-popover-foreground shadow-md outline-none data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95",
            contentClassName
          )}
        >
          <div className="flex items-center gap-2 border-b px-2 py-1.5">
            <SearchIcon className="size-4 shrink-0 text-muted-foreground" />
            <Input
              autoFocus
              className="h-8 border-0 px-0 shadow-none focus-visible:ring-0"
              value={search}
              placeholder={searchPlaceholder || placeholder}
              onChange={(event) => setSearch(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Escape") setOpen(false)
              }}
            />
          </div>
          <div className="max-h-64 overflow-y-auto py-1">
            {allowClear && emptyLabel ? (
              <ComboboxOptionButton
                label={emptyLabel}
                selected={selectedValue === emptyValue}
                onSelect={() => selectValue(emptyValue)}
              />
            ) : null}
            {filteredOptions.length ? (
              filteredOptions.map((option) => {
                const optionValue = String(option.value)
                const hasChildren = parentValues.has(optionValue)
                const expanded =
                  Boolean(normalizedSearch) || expandedValues.has(optionValue)
                return (
                  <div key={optionValue} className="flex items-center gap-1">
                    <HierarchyExpandButton
                      hasChildren={hasChildren}
                      searching={Boolean(normalizedSearch)}
                      expanded={expanded}
                      onToggle={() => toggleExpanded(optionValue)}
                    />
                    <ComboboxOptionButton
                      label={option.label}
                      selected={selectedValue === optionValue}
                      onSelect={() => selectValue(optionValue)}
                      depth={option.depth}
                    />
                  </div>
                )
              })
            ) : (
              <div className="px-2 py-3 text-sm text-muted-foreground">
                {noResultsLabel}
              </div>
            )}
          </div>
        </PopoverPrimitive.Content>
      </PopoverPrimitive.Portal>
    </PopoverPrimitive.Root>
  )
}

export function DashboardMultiSelect({
  value,
  options,
  placeholder,
  searchPlaceholder,
  noResultsLabel = "-",
  disabled,
  className,
  triggerClassName,
  contentClassName,
  portalContainer,
  onValueChange,
}: {
  value?: unknown
  options: DashboardSelectOption[]
  placeholder?: string
  searchPlaceholder?: string
  noResultsLabel?: string
  disabled?: boolean
  className?: string
  triggerClassName?: string
  contentClassName?: string
  portalContainer?: HTMLElement | null
  onValueChange: (value: string[]) => void
}) {
  const { t } = useI18n()
  const [open, setOpen] = React.useState(false)
  const [search, setSearch] = React.useState("")
  const selectedValues = React.useMemo(
    () => (Array.isArray(value) ? value.map((item) => String(item)) : []),
    [value]
  )
  const selectedOptions = selectedValues
    .map((selectedValue) =>
      options.find((option) => String(option.value) === selectedValue)
    )
    .filter((option): option is DashboardSelectOption => Boolean(option))
  const selectedLabel = selectedOptions.length
    ? selectedOptions.map((option) => option.label).join(", ")
    : placeholder
  const {
    expandedValues,
    filteredOptions,
    normalizedSearch,
    parentValues,
    toggleExpanded,
  } = useDashboardHierarchy(options, search, open, selectedValues)

  function toggleValue(nextValue: string) {
    const nextValues = selectedValues.includes(nextValue)
      ? selectedValues.filter((item) => item !== nextValue)
      : [...selectedValues, nextValue]
    onValueChange(nextValues)
  }

  function clearValue(event: React.MouseEvent<HTMLElement>) {
    event.preventDefault()
    event.stopPropagation()
    onValueChange([])
  }

  return (
    <PopoverPrimitive.Root open={open} onOpenChange={setOpen}>
      <PopoverPrimitive.Trigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            "w-full justify-between px-3 font-normal",
            !selectedOptions.length && "text-muted-foreground",
            triggerClassName,
            className
          )}
        >
          <span className="line-clamp-1 text-left">
            {selectedLabel || placeholder}
          </span>
          {selectedOptions.length ? (
            <span
              aria-label={t("common.accessibility.clearSelection")}
              className="flex size-4 shrink-0 items-center justify-center rounded-sm opacity-50 transition-opacity hover:opacity-100"
              onClick={clearValue}
              onPointerDown={(event) => {
                event.preventDefault()
                event.stopPropagation()
              }}
            >
              <XIcon className="size-4" />
            </span>
          ) : (
            <ChevronsUpDownIcon className="size-4 shrink-0 opacity-50" />
          )}
        </Button>
      </PopoverPrimitive.Trigger>
      <PopoverPrimitive.Portal container={portalContainer ?? undefined}>
        <PopoverPrimitive.Content
          align="start"
          sideOffset={4}
          className={cn(
            "z-50 w-[var(--radix-popover-trigger-width)] rounded-md border bg-popover p-1 text-popover-foreground shadow-md outline-none data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95",
            contentClassName
          )}
        >
          <div className="flex items-center gap-2 border-b px-2 py-1.5">
            <SearchIcon className="size-4 shrink-0 text-muted-foreground" />
            <Input
              autoFocus
              className="h-8 border-0 px-0 shadow-none focus-visible:ring-0"
              value={search}
              placeholder={searchPlaceholder || placeholder}
              onChange={(event) => setSearch(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Escape") setOpen(false)
              }}
            />
          </div>
          <div className="max-h-64 overflow-y-auto py-1">
            {filteredOptions.length ? (
              filteredOptions.map((option) => {
                const optionValue = String(option.value)
                const hasChildren = parentValues.has(optionValue)
                const expanded =
                  Boolean(normalizedSearch) || expandedValues.has(optionValue)
                return (
                  <div key={optionValue} className="flex items-center gap-1">
                    <HierarchyExpandButton
                      hasChildren={hasChildren}
                      searching={Boolean(normalizedSearch)}
                      expanded={expanded}
                      onToggle={() => toggleExpanded(optionValue)}
                    />
                    <ComboboxOptionButton
                      label={option.label}
                      selected={selectedValues.includes(optionValue)}
                      onSelect={() => toggleValue(optionValue)}
                      depth={option.depth}
                    />
                  </div>
                )
              })
            ) : (
              <div className="px-2 py-3 text-sm text-muted-foreground">
                {noResultsLabel}
              </div>
            )}
          </div>
        </PopoverPrimitive.Content>
      </PopoverPrimitive.Portal>
    </PopoverPrimitive.Root>
  )
}

function ComboboxOptionButton({
  label,
  selected,
  onSelect,
  depth = 0,
}: {
  label: string
  selected: boolean
  onSelect: () => void
  depth?: number
}) {
  return (
    <button
      type="button"
      className="flex min-w-0 flex-1 items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm outline-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground"
      style={{ paddingLeft: `${0.5 + depth * 1.25}rem` }}
      onClick={onSelect}
    >
      <CheckIcon
        className={cn("size-4", selected ? "opacity-100" : "opacity-0")}
      />
      <span className="line-clamp-1">{label}</span>
    </button>
  )
}
