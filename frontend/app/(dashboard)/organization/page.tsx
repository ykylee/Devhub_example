"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";

// /organization 은 PR-S2 에서 /admin/settings 하위로 이전됐다. 외부 북마크
// 와 docs/wiki 링크가 끊기지 않도록 짧은 deprecation 기간 동안 redirect 만
// 유지한다. AuthGuard 가 같은 가드를 적용하므로 비-admin 은 자기 default
// landing 으로 먼저 보내진다.
export default function OrganizationDeprecatedPage() {
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
