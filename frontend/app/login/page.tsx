"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { motion } from "framer-motion";
import { AlertCircle, ArrowRight, ShieldCheck } from "lucide-react";
import { authService } from "@/lib/services/auth.service";

// ADR-0020 §4.1.1 sub-carve F (decision B) — `/login` is the canonical entry
// page. `/auth/login` is preserved as a stub redirect for external bookmark
// compatibility. Provider / AuthGuard / callback failures land here with
// `?error=<code>&error_description=<msg>` and the message is rendered above
// the Continue button so a stuck user is not silently re-redirected.

export const ERROR_MESSAGES: Record<string, string> = {
  session_expired: "Your session has expired. Please sign in again.",
  login_failed: "Sign-in failed. Please verify your credentials and try again.",
  unauthorized: "You are not authorized to access that resource. Please sign in again.",
};

// Exported so the helper can be unit-tested without rendering the full page —
// the page itself wires `window.location.assign` for OIDC redirect which jsdom
// cannot resolve. Integration cases live in e2e (TC-AUTH-* family).
export function resolveErrorMessage(
  code: string | null,
  description: string | null,
): string | null {
  if (!code && !description) return null;
  if (description) return description;
  if (code && ERROR_MESSAGES[code]) return ERROR_MESSAGES[code];
  return code;
}

function LoginInner() {
  const searchParams = useSearchParams();
  const errorCode = searchParams.get("error");
  const errorDescription = searchParams.get("error_description");
  const errorMessage = resolveErrorMessage(errorCode, errorDescription);

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
    // When the user arrived with an explicit error code we keep them on this
    // page so the message stays visible. Only auto-bounce on a clean entry.
    if (errorMessage) return;
    handleLogin();
  }, [errorMessage, handleLogin]);

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4 selection:bg-primary/30">
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
            className="inline-flex p-4 rounded-3xl bg-gradient-to-br from-primary/20 to-accent/20 border border-border/60 mb-6 shadow-2xl"
          >
            <ShieldCheck className="w-12 h-12 text-foreground" />
          </motion.div>
          <h1 className="text-4xl font-black text-foreground tracking-tighter uppercase mb-2">
            DevHub <span className="text-primary">Identity</span>
          </h1>
          <p className="text-muted-foreground text-sm font-bold uppercase tracking-widest">
            Unified Authentication Gateway
          </p>
        </div>

        <div className="glass border-border/60 rounded-[2rem] p-10 shadow-2xl backdrop-blur-2xl">
          {errorMessage && (
            <div
              role="alert"
              data-testid="login-error"
              className="mb-6 flex items-start gap-3 rounded-2xl border border-destructive/30 bg-destructive/10 p-4 text-left"
            >
              <AlertCircle className="mt-0.5 h-5 w-5 flex-shrink-0 text-destructive" />
              <div className="space-y-1">
                <p className="text-sm font-bold text-foreground">Sign-in interrupted</p>
                <p className="text-xs text-muted-foreground break-words">{errorMessage}</p>
              </div>
            </div>
          )}

          <p className="text-sm text-muted-foreground text-center mb-8">
            DevHub uses a unified Keycloak OIDC identity flow. Continue to the secure
            sign-in screen to authenticate.
          </p>

          <button
            type="button"
            onClick={handleLogin}
            disabled={isRedirecting}
            className="w-full bg-primary text-primary-foreground font-black py-4 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/20 disabled:opacity-50 disabled:hover:scale-100 group uppercase tracking-widest text-xs"
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
              Need an account?{" "}
              <span className="text-primary ml-1">Contact your system administrator</span>
            </p>
          </div>

          <div className="mt-8 pt-8 border-t border-border/40 text-center">
            <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest">
              Secured by Keycloak OIDC
            </p>
          </div>
        </div>

        <div className="mt-8 flex justify-center gap-6">
          <p className="text-[9px] text-muted-foreground/70 font-bold uppercase">Node: ASIA-01</p>
          <p className="text-[9px] text-muted-foreground/70 font-bold uppercase">v0.5.0-BETA</p>
        </div>
      </motion.div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={null}>
      <LoginInner />
    </Suspense>
  );
}
