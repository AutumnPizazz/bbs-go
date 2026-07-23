"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardVoteOptionsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "voteOptions"),
    description: dashboardData.desc(t, "voteOptions"),
    listEndpoint: "/api/admin/vote-option/list",
    viewPermission: PERMISSIONS.DASHBOARD_VOTE_OPTION_VIEW,
    detailEndpoint: (id) => `/api/admin/vote-option/${id}`,
    createEndpoint: "/api/admin/vote-option/create",
    createPermission: PERMISSIONS.DASHBOARD_VOTE_OPTION_CREATE,
    updateEndpoint: "/api/admin/vote-option/update",
    updatePermission: PERMISSIONS.DASHBOARD_VOTE_OPTION_UPDATE,
    deleteEndpoint: "/api/admin/vote-option/delete",
    deletePermission: PERMISSIONS.DASHBOARD_VOTE_OPTION_DELETE,
    deleteMode: "formIds",
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "voteId", label: dashboardData.label(t, "voteId") },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "voteId", label: dashboardData.label(t, "voteId") },
      { key: "content", label: dashboardData.label(t, "content") },
      { key: "sortNo", label: dashboardData.label(t, "sortNo") },
      { key: "voteCount", label: dashboardData.label(t, "voteCount") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      { name: "id", label: dashboardData.label(t, "id"), type: "number" },
      { name: "voteId", label: dashboardData.label(t, "voteId"), type: "number", required: true },
      { name: "content", label: dashboardData.label(t, "content"), required: true, colSpan: 2 },
      { name: "sortNo", label: dashboardData.label(t, "sortNo"), type: "number", min: 0 },
    ],
  }

  return <DashboardDataPage config={config} />
}
