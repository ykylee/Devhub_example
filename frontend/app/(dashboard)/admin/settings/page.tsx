"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";

// /admin/settings 직진입은 첫 sub-tab(users) 으로 보낸다. layout 의 가드가
// 동일한 actor.role 검증을 적용하므로 여기서는 redirect 만 수행.
export default function AdminSettingsIndexPage() {
  const router = useRouter();
  useEffect(() => {
    router.replace("/admin/settings/users");
  }, [router]);
  return (
    <div className="flex items-center justify-center py-20">
      <Loader2 className="w-6 h-6 text-orange-400 animate-spin" />
    </div>
  );
}
