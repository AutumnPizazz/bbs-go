"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardDictionariesRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "dictionaryTypes"),
    description: dashboardData.desc(t, "dictionaryTypes"),
    listEndpoint: "/api/admin/dict-type/list",
    listResult: "array",
    viewPermission: PERMISSIONS.DASHBOARD_DICT_TYPE_VIEW,
    detailEndpoint: (id) => `/api/admin/dict-type/${id}`,
    createEndpoint: "/api/admin/dict-type/create",
    createPermission: PERMISSIONS.DASHBOARD_DICT_TYPE_CREATE,
    updateEndpoint: "/api/admin/dict-type/update",
    updatePermission: PERMISSIONS.DASHBOARD_DICT_TYPE_UPDATE,
    deleteEndpoint: "/api/admin/dict-type/delete",
    deletePermission: PERMISSIONS.DASHBOARD_DICT_TYPE_DELETE,
    deleteMode: "formIds",
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "name", label: dashboardData.label(t, "name") },
      { name: "code", label: dashboardData.label(t, "code") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: dashboardData.normalDeletedOptions(t),
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "name", label: dashboardData.label(t, "name") },
      { key: "code", label: dashboardData.label(t, "code") },
      { key: "remark", label: dashboardData.label(t, "remark") },
      {
        key: "status",
        label: dashboardData.label(t, "status"),
        render: (record) => dashboardData.statusCell(t, record.status),
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      { name: "id", label: dashboardData.label(t, "id"), type: "number" },
      { name: "name", label: dashboardData.label(t, "name"), required: true },
      { name: "code", label: dashboardData.label(t, "code"), required: true },
      {
        name: "remark",
        label: dashboardData.label(t, "remark"),
        type: "textarea",
      },
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
