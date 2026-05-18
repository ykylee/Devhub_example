"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";

export default function LegacyAuthLoginPage() {
  const router = useRouter();

  useEffect(() => {
    // Legacy route kept for backward compatibility with old bookmarks.
    // Keycloak OIDC login always starts from /login.
    router.replace("/login");
  }, [router]);

  return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="flex flex-col items-center gap-6">
        <Loader2 className="w-10 h-10 text-primary animate-spin" />
        <p className="text-xs text-muted-foreground font-bold uppercase tracking-widest">
          Redirecting to sign in...
        </p>
      </div>
    </div>
  );
}
