"use client";

import { Suspense, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";

// ADR-0020 §4.1.1 sub-carve F (decision B) — `/login` is the canonical entry
// page. This route is retained as a stub redirect so external bookmarks /
// historical links (issue tracker, deploy docs prior to 2026-05-22) continue
// to land on the canonical page with their query string intact.

function AuthLoginStubInner() {
  const router = useRouter();
  const searchParams = useSearchParams();

  useEffect(() => {
    const query = searchParams.toString();
    router.replace(query ? `/login?${query}` : "/login");
  }, [router, searchParams]);

  return null;
}

export default function AuthLoginStubPage() {
  return (
    <Suspense fallback={null}>
      <AuthLoginStubInner />
    </Suspense>
  );
}
