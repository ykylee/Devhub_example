"use client";

import { FolderGit2 } from "lucide-react";
import Link from "next/link";
import { MyPendingDevRequestsWidget } from "@/domain/dev-request/view/MyPendingDevRequestsWidget";

export default function DeveloperDashboard() {
  return (
    <div className="space-y-8 animate-fade-in max-w-2xl mx-auto py-10">
      <div className="flex flex-col items-center justify-center text-center space-y-6">
        <div className="w-16 h-16 bg-muted/30 rounded-2xl flex items-center justify-center border border-border shadow-inner">
          <FolderGit2 className="w-8 h-8 text-primary" />
        </div>
        <div className="space-y-2 max-w-md px-4">
          <h2 className="text-xl font-bold text-foreground">Work Status Archived</h2>
          <p className="text-sm text-muted-foreground leading-relaxed">
            이 대시보드 화면은 아카이브되었습니다. 활성화된 운영 기능인 Projects 메뉴를 이용해 주시기 바랍니다.
          </p>
        </div>
        <Link href="/projects" className="bg-primary text-primary-foreground px-6 py-2.5 rounded-xl text-xs font-bold hover:bg-primary/90 transition-all shadow-lg flex items-center gap-2">
          Go to Projects
        </Link>
      </div>

      <div className="border-t border-border/60 pt-8 mt-8">
        <MyPendingDevRequestsWidget />
      </div>
    </div>
  );
}

