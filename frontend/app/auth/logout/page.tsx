"use client";

import { Suspense, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Loader2 } from "lucide-react";
import { authService } from "@/lib/services/auth.service";

function LogoutInner() {
  const searchParams = useSearchParams();

  useEffect(() => {
    const challenge = searchParams.get("logout_challenge");
    if (!challenge) {
      // Direct visit (no Hydra challenge) — nothing to accept; bounce to /login.
      window.location.replace("/login");
      return;
    }
    void authService.completeRPInitiatedLogout(challenge);
  }, [searchParams]);

  return (
    <div className="min-h-screen bg-[#030014] flex items-center justify-center">
      <div className="flex flex-col items-center gap-6">
        <div className="relative">
          <div className="absolute inset-0 bg-primary/20 blur-3xl rounded-full" />
          <Loader2 className="w-12 h-12 text-primary animate-spin relative z-10" />
        </div>
        <div className="text-center space-y-2">
          <h2 className="text-lg font-black text-white uppercase tracking-widest">Signing Out</h2>
          <p className="text-xs text-muted-foreground font-bold uppercase tracking-widest animate-pulse">
            Releasing identity session...
          </p>
        </div>
      </div>
    </div>
  );
}

export default function AuthLogoutPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-[#030014] flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-primary animate-spin" />
      </div>
    }>
      <LogoutInner />
    </Suspense>
  );
}
