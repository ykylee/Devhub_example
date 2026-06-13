"use client";

import Link from "next/link";
import { ShieldAlert, ArrowRight, Settings } from "lucide-react";
import {
  SyncJobQueueWidget,
  SyncJobStatusWidget,
  ProviderHealthWidget,
  DashboardSummaryWidget,
} from "@/components/admin/x1-widgets";

/**
 * /admin landing page — X-1 System Admin 운영 대시보드 (RM-M4-07).
 *
 * v0.1.1 milestone X-1 의 정공법 — admin landing page 의 "Archived" 메시지
 * (v0.1.0 출시 시점, system_admin_catalog_plan_2026-05-27.md §2 옵션 B 의
 * 선행 carve) 를 X-1 widget 4 (sync job queue / status counts / provider
 * health / dashboard summary) + "운영 도구" link 섹션 (Topology v2, Settings)
 * 으로 교체.
 *
 * 권한: system_admin 일임 (Sidebar 의 isSystemAdmin(actor?.role) gate 자동).
 */
export default function AdminDashboard() {
  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center gap-3">
        <div className="w-10 h-10 bg-muted/30 rounded-xl flex items-center justify-center border border-border shadow-inner">
          <ShieldAlert className="w-5 h-5 text-primary" />
        </div>
        <div>
          <h1 className="text-xl font-black text-foreground">System Admin 운영 대시보드</h1>
          <p className="text-xs text-muted-foreground">
            X-1 · Gitea sync job · Provider health · 운영 도구 진입
          </p>
        </div>
      </div>

      {/* X-1 widget 4 grid: 2x2 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <SyncJobQueueWidget />
        <SyncJobStatusWidget />
        <ProviderHealthWidget />
        <DashboardSummaryWidget />
      </div>

      {/* 운영 도구 link 섹션 */}
      <div className="glass border-border rounded-2xl p-5 space-y-3 shadow-sm">
        <h2 className="text-sm font-bold uppercase tracking-wider text-foreground">
          운영 도구
        </h2>
        <div className="flex flex-wrap items-center gap-3">
          <Link
            href="/admin/topology-v2"
            className="glass border-border text-foreground dark:text-primary-foreground px-6 py-2.5 rounded-xl text-xs font-black uppercase tracking-widest hover:bg-muted/40 transition-all flex items-center gap-2"
          >
            <ArrowRight className="w-4 h-4 text-primary" /> Topology v2
          </Link>
          <Link
            href="/admin/settings"
            className="bg-primary text-primary-foreground px-6 py-2.5 rounded-xl text-xs font-bold hover:bg-primary/90 transition-all shadow-lg flex items-center gap-2"
          >
            <Settings className="w-4 h-4 text-primary-foreground" /> Settings
          </Link>
        </div>
      </div>
    </div>
  );
}
