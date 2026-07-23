"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardDictionaryItemsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "dictionaryItems"),
    description: dashboardData.desc(t, "dictionaryItems"),
    listEndpoint: "/api/admin/dict/list",
    listResult: "array",
    viewPermission: PERMISSIONS.DASHBOARD_DICT_VIEW,
    detailEndpoint: (id) => `/api/admin/dict/${id}`,
    createEndpoint: "/api/admin/dict/create",
    createPermission: PERMISSIONS.DASHBOARD_DICT_CREATE,
    updateEndpoint: "/api/admin/dict/update",
    updatePermission: PERMISSIONS.DASHBOARD_DICT_UPDATE,
    deleteEndpoint: "/api/admin/dict/delete",
    deletePermission: PERMISSIONS.DASHBOARD_DICT_DELETE,
    deleteMode: "formIds",
    sortEndpoint: "/api/admin/dict/update_sort",
    sortPermission: PERMISSIONS.DASHBOARD_DICT_SORT,
    tree: true,
    treeDefaultCollapsed: true,
    treeIndentKey: "name",
    filters: [
      {
        name: "typeId",
        label: dashboardData.label(t, "dictType"),
        type: "select",
        optionsEndpoint: "/api/admin/dict-type/list",
        optionLabel: (record) => String(record.name || record.code || record.id),
        optionValue: (record) => record.id as number,
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "name", label: dashboardData.label(t, "name") },
      { key: "label", label: dashboardData.label(t, "label") },
      { key: "value", label: dashboardData.label(t, "value") },
      { key: "sortNo", label: dashboardData.label(t, "sortNo") },
      {
        key: "status",
        label: dashboardData.label(t, "status"),
        render: (record) => dashboardData.statusCell(t, record.status),
      },
    ],
    formFields: [
      { name: "id", label: dashboardData.label(t, "id"), type: "number" },
      {
        name: "typeId",
        label: dashboardData.label(t, "dictType"),
        type: "select",
        required: true,
        optionsEndpoint: "/api/admin/dict-type/list",
        optionLabel: (record) => String(record.name || record.code || record.id),
        optionValue: (record) => record.id as number,
      },
      { name: "parentId", label: dashboardData.label(t, "parentId"), type: "number" },
      { name: "name", label: dashboardData.label(t, "name"), required: true },
      { name: "label", label: dashboardData.label(t, "label"), required: true },
      { name: "value", label: dashboardData.label(t, "value"), required: true },
      { name: "sortNo", label: dashboardData.label(t, "sortNo"), type: "number" },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        required: true,
        options: dashboardData.normalDeletedOptions(t),
      },
    ],
  }

  return <DashboardDataPage config={config} />
}
