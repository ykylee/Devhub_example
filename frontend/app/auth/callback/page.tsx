"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { authService } from "@/lib/services/auth.service";
import { Loader2, AlertCircle } from "lucide-react";
import { motion } from "framer-motion";

function CallbackInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const errorParam = searchParams.get("error");
    const errorDesc = searchParams.get("error_description");

    if (errorParam) {
      setTimeout(() => setError(errorDesc || errorParam), 0);
      return;
    }

    if (!code || !state) {
      setTimeout(() => setError("Missing authorization code or state."), 0);
      return;
    }

    async function processCallback() {
      try {
        // 1. Exchange code for tokens
        await authService.exchangeCode(code!, state!);
        
        // 2. Resolve identity and update store
        await authService.resolveIdentity();
        
        // 3. Success! Redirect to dashboard
        router.replace("/developer");
      } catch (err) {
        console.error("[auth/callback] Error processing callback:", err);
        const msg = err instanceof Error ? err.message : "Authentication failed.";
        setTimeout(() => setError(msg), 0);
      }
    }

    processCallback();
  }, [searchParams, router]);

  if (error) {
    return (
      <div className="min-h-screen bg-[#030014] flex items-center justify-center p-4">
        <motion.div 
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="glass border-white/10 rounded-3xl p-8 max-w-md w-full text-center space-y-6 shadow-2xl"
        >
          <div className="inline-flex p-4 rounded-2xl bg-red-500/10 border border-red-500/20">
            <AlertCircle className="w-8 h-8 text-red-400" />
          </div>
          <h2 className="text-xl font-bold text-white">Authentication Failed</h2>
          <p className="text-muted-foreground text-sm leading-relaxed">{error}</p>
          <button
            onClick={() => router.replace("/login")}
            className="w-full py-3 bg-white/5 hover:bg-white/10 border border-white/10 rounded-xl text-white text-sm font-bold transition-all"
          >
            Back to Sign In
          </button>
        </motion.div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#030014] flex items-center justify-center">
      <div className="flex flex-col items-center gap-6">
        <div className="relative">
          <div className="absolute inset-0 bg-primary/20 blur-3xl rounded-full" />
          <Loader2 className="w-12 h-12 text-primary animate-spin relative z-10" />
        </div>
        <div className="text-center space-y-2">
          <h2 className="text-lg font-black text-white uppercase tracking-widest">Verifying Identity</h2>
          <p className="text-xs text-muted-foreground font-bold uppercase tracking-widest animate-pulse">
            Establishing secure connection to cluster...
          </p>
        </div>
      </div>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-[#030014] flex items-center justify-center">
        <Loader2 className="w-8 h-8 text-primary animate-spin" />
      </div>
    }>
      <CallbackInner />
    </Suspense>
  );
}
