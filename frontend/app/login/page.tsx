"use client";

import { useEffect, useState, useCallback } from "react";
import { motion } from "framer-motion";
import { ArrowRight, ShieldCheck } from "lucide-react";
import Link from "next/link";
import { authService } from "@/lib/services/auth.service";

export default function LoginPage() {
  const [isRedirecting, setIsRedirecting] = useState(false);

  const handleLogin = useCallback(async () => {
    setTimeout(() => setIsRedirecting(true), 0);
    try {
      const url = await authService.getAuthorizeURL();
      window.location.assign(url);
    } catch (error) {
      console.error("[LoginPage] Failed to start OIDC flow:", error);
      setTimeout(() => setIsRedirecting(false), 0);
    }
  }, []);

  useEffect(() => {
    // Automatically initiate OIDC flow on landing to avoid redundant click
    handleLogin();
  }, [handleLogin]);

  return (
    <div className="min-h-screen bg-[#030014] flex items-center justify-center p-4 selection:bg-primary/30">
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 -left-1/4 w-1/2 h-1/2 bg-primary/10 rounded-full blur-[120px]" />
        <div className="absolute bottom-1/4 -right-1/4 w-1/2 h-1/2 bg-accent/10 rounded-full blur-[120px]" />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full max-w-md relative z-10"
      >
        <div className="text-center mb-10">
          <motion.div
            initial={{ scale: 0.5 }}
            animate={{ scale: 1 }}
            className="inline-flex p-4 rounded-3xl bg-gradient-to-br from-primary/20 to-accent/20 border border-white/10 mb-6 shadow-2xl"
          >
            <ShieldCheck className="w-12 h-12 text-white" />
          </motion.div>
          <h1 className="text-4xl font-black text-white tracking-tighter uppercase mb-2">
            DevHub <span className="text-primary">Identity</span>
          </h1>
          <p className="text-muted-foreground text-sm font-bold uppercase tracking-widest">
            Unified Authentication Gateway
          </p>
        </div>

        <div className="glass border-white/10 rounded-[2rem] p-10 shadow-2xl backdrop-blur-2xl">
          <p className="text-sm text-muted-foreground text-center mb-8">
            DevHub uses Ory Hydra and Kratos for identity. Continue to the secure
            sign-in flow to authenticate.
          </p>

          <button
            type="button"
            onClick={handleLogin}
            disabled={isRedirecting}
            className="w-full bg-primary text-white font-black py-4 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/20 disabled:opacity-50 disabled:hover:scale-100 group uppercase tracking-widest text-xs"
          >
            {isRedirecting ? "Redirecting..." : (
              <>
                Continue to Sign In
                <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
              </>
            )}
          </button>

          <div className="text-center pt-8">
            <p className="text-xs text-muted-foreground uppercase tracking-widest font-black">
              New to DevHub?{" "}
              <Link href="/auth/signup" className="text-primary hover:underline ml-1">
                Create Account
              </Link>
            </p>
          </div>

          <div className="mt-8 pt-8 border-t border-white/5 text-center">
            <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest">
              Secured by Ory Hydra + Kratos
            </p>
          </div>
        </div>

        <div className="mt-8 flex justify-center gap-6">
          <p className="text-[9px] text-white/20 font-bold uppercase">Node: ASIA-01</p>
          <p className="text-[9px] text-white/20 font-bold uppercase">v0.5.0-BETA</p>
        </div>
      </motion.div>
    </div>
  );
}
