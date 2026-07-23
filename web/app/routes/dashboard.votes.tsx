"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardVotesRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "votes"),
    description: dashboardData.desc(t, "votes"),
    listEndpoint: "/api/admin/vote/list",
    viewPermission: PERMISSIONS.DASHBOARD_VOTE_VIEW,
    detailEndpoint: (id) => `/api/admin/vote/${id}`,
    createEndpoint: "/api/admin/vote/create",
    createPermission: PERMISSIONS.DASHBOARD_VOTE_CREATE,
    updateEndpoint: "/api/admin/vote/update",
    updatePermission: PERMISSIONS.DASHBOARD_VOTE_UPDATE,
    deleteEndpoint: "/api/admin/vote/delete",
    deletePermission: PERMISSIONS.DASHBOARD_VOTE_DELETE,
    deleteMode: "formIds",
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "topicId", label: dashboardData.label(t, "topicId") },
      { name: "userId", label: dashboardData.label(t, "userId") },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "title", label: dashboardData.label(t, "title") },
      { key: "topicId", label: dashboardData.label(t, "topicId") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "type", label: dashboardData.label(t, "voteType") },
      { key: "optionCount", label: dashboardData.label(t, "optionCount") },
      { key: "voteCount", label: dashboardData.label(t, "voteCount") },
      {
        key: "expiredAt",
        label: dashboardData.label(t, "expiredAt"),
        render: (record) => dashboardData.dateCell(record.expiredAt),
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      { name: "id", label: dashboardData.label(t, "id"), type: "number" },
      { name: "title", label: dashboardData.label(t, "title"), required: true, colSpan: 2 },
      { name: "topicId", label: dashboardData.label(t, "topicId"), type: "number", required: true },
      { name: "userId", label: dashboardData.label(t, "userId"), type: "number", required: true },
      {
        name: "type",
        label: dashboardData.label(t, "voteType"),
        type: "select",
        required: true,
        options: [
          { label: t("dashboard.voteTypes.single"), value: 1 },
          { label: t("dashboard.voteTypes.multiple"), value: 2 },
        ],
      },
      { name: "voteNum", label: dashboardData.label(t, "voteNum"), type: "number", min: 1 },
      { name: "optionCount", label: dashboardData.label(t, "optionCount"), type: "number", min: 0 },
      { name: "expiredAt", label: dashboardData.label(t, "expiredAt"), type: "number", min: 0 },
    ],
  }

  return <DashboardDataPage config={config} />
}
