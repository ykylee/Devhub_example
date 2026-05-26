"use client";

import { Users } from "lucide-react";
import Link from "next/link";

export default function ManagerDashboard() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[500px] text-center space-y-6 animate-fade-in">
      <div className="w-16 h-16 bg-muted/30 rounded-2xl flex items-center justify-center border border-border shadow-inner">
        <Users className="w-8 h-8 text-primary" />
      </div>
      <div className="space-y-2 max-w-md px-4">
        <h2 className="text-xl font-bold text-foreground">Quality Status Archived</h2>
        <p className="text-sm text-muted-foreground leading-relaxed">
          이 대시보드 화면은 아카이브되었습니다. 활성화된 운영 기능인 Projects 메뉴를 이용해 주시기 바랍니다.
        </p>
      </div>
      <Link href="/projects" className="bg-primary text-primary-foreground px-6 py-2.5 rounded-xl text-xs font-bold hover:bg-primary/90 transition-all shadow-lg flex items-center gap-2">
        Go to Projects
      </Link>
    </div>
  );
}
