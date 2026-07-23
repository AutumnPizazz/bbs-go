"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { useI18n } from "@/lib/i18n/provider"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"

export default function DashboardVoteRecordsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "voteRecords"),
    description: dashboardData.desc(t, "voteRecords"),
    listEndpoint: "/api/admin/vote-record/list",
    viewPermission: PERMISSIONS.DASHBOARD_VOTE_RECORD_VIEW,
    detailEndpoint: (id) => `/api/admin/vote-record/${id}`,
    filters: [
      { name: "id", label: dashboardData.label(t, "id") },
      { name: "userId", label: dashboardData.label(t, "userId") },
      { name: "voteId", label: dashboardData.label(t, "voteId") },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "voteId", label: dashboardData.label(t, "voteId") },
      { key: "optionIds", label: dashboardData.label(t, "optionIds") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "userId", label: dashboardData.label(t, "userId") },
      { key: "voteId", label: dashboardData.label(t, "voteId") },
      { key: "optionIds", label: dashboardData.label(t, "optionIds") },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
  }

  return <DashboardDataPage config={config} />
}
