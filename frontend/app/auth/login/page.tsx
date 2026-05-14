"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { motion } from "framer-motion";
import { ShieldCheck, Lock, AlertCircle, Loader2 } from "lucide-react";

// /auth/login is the URL Hydra (urls.login) and Kratos (selfservice.flows.login.ui_url)
// redirect to with a login_challenge query parameter. Without that parameter the user
// arrived directly; bounce them back to /login so the OIDC code flow starts cleanly.
type LoginErrorCode =
  | "invalid_credentials"
  | "login_challenge_unknown"
  | "login_flow_expired"
  | "service_unavailable"
  | "unknown";

interface LoginErrorState {
  code: LoginErrorCode;
  message: string;
}

interface AuthLoginResponse {
  status: string;
  data?: { redirect_to?: string };
  error?: string;
  code?: string;
}

function decodeError(status: number, body: AuthLoginResponse): LoginErrorState {
  if (status === 401) {
    return { code: "invalid_credentials", message: "Invalid email or password." };
  }
  if (status === 410) {
    if (body.code === "login_challenge_unknown") {
      return {
        code: "login_challenge_unknown",
        message: "This sign-in link expired. Restart the request.",
      };
    }
    return {
      code: "login_flow_expired",
      message: "Sign-in took too long. Please try again.",
    };
  }
  if (status === 503) {
    return {
      code: "service_unavailable",
      message: "Authentication service is temporarily unavailable.",
    };
  }
  return {
    code: "unknown",
    message: body.error || `Sign-in failed (${status}).`,
  };
}

function AuthLoginInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const loginChallenge = searchParams.get("login_challenge");

  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<LoginErrorState | null>(null);

  useEffect(() => {
    if (!loginChallenge) {
      // User reached /auth/login directly. Restart the Hydra OIDC code flow
      // from /login so we get a fresh challenge.
      router.replace("/login");
    }
  }, [loginChallenge, router]);

  if (!loginChallenge) {
    return <RedirectingFallback />;
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (submitting) return;
    setSubmitting(true);
    setError(null);

    try {
      const response = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          login_challenge: loginChallenge,
          identifier: identifier.trim(),
          password,
        }),
      });

      const body = (await response.json().catch(() => ({}))) as AuthLoginResponse;

      if (!response.ok) {
        const decoded = decodeError(response.status, body);
        setError(decoded);
        // Restart the flow when Hydra/Kratos says the challenge or flow is dead.
        if (decoded.code === "login_challenge_unknown" || decoded.code === "login_flow_expired") {
          setTimeout(() => router.replace("/login"), 1500);
        }
        return;
      }

      const redirectTo = body.data?.redirect_to;
      if (!redirectTo) {
        setError({ code: "unknown", message: "Sign-in succeeded but no redirect target was provided." });
        return;
      }
      window.location.assign(redirectTo);
    } catch (err) {
      console.error("[auth/login] submit failed", err);
      setError({ code: "unknown", message: "Network error; check your connection." });
    } finally {
      setSubmitting(false);
    }
  };

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
            DevHub <span className="text-primary">Sign In</span>
          </h1>
          <p className="text-muted-foreground text-sm font-bold uppercase tracking-widest">
            Authenticate with your DevHub identity
          </p>
        </div>

        <form
          onSubmit={handleSubmit}
          className="glass border-border/60 rounded-[2rem] p-10 shadow-2xl backdrop-blur-2xl space-y-6"
        >
          <div className="space-y-2">
            <label
              htmlFor="identifier"
              className="text-[10px] font-black text-muted-foreground uppercase tracking-widest"
            >
              System ID
            </label>
            <input
              id="identifier"
              name="identifier"
              type="text"
              autoComplete="username"
              required
              autoFocus
              disabled={submitting}
              value={identifier}
              onChange={(e) => setIdentifier(e.target.value)}
              className="w-full bg-card/60 border border-border rounded-2xl px-4 py-3 text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:opacity-50"
              placeholder="e.g. yklee"
            />
          </div>

          <div className="space-y-2">
            <label
              htmlFor="password"
              className="text-[10px] font-black text-muted-foreground uppercase tracking-widest"
            >
              Password
            </label>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              required
              disabled={submitting}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full bg-card/60 border border-border rounded-2xl px-4 py-3 text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 disabled:opacity-50"
              placeholder="••••••••"
            />
          </div>

          {error && (
            <div
              role="alert"
              className="flex items-start gap-2 rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300"
            >
              <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
              <span>{error.message}</span>
            </div>
          )}

          <button
            type="submit"
            disabled={submitting || !identifier || !password}
            className="w-full bg-primary text-primary-foreground font-black py-4 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-2 shadow-lg shadow-primary/20 disabled:opacity-50 disabled:hover:scale-100 uppercase tracking-widest text-xs"
          >
            {submitting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Signing In…
              </>
            ) : (
              <>
                <Lock className="w-4 h-4" />
                Sign In
              </>
            )}
          </button>

          <div className="flex flex-col gap-4 pt-4 text-center">
            <button
              type="button"
              onClick={() => router.push("/auth/signup")}
              className="text-[10px] text-muted-foreground hover:text-primary font-black uppercase tracking-widest transition-colors"
            >
              Don&apos;t have an account? <span className="text-primary underline">Sign Up</span>
            </button>
            
            <button
              type="button"
              onClick={() => alert("Please contact the system administrator (admin@devhub.local) to reset your password.")}
              className="text-[10px] text-muted-foreground/70 hover:text-foreground font-black uppercase tracking-widest transition-colors"
            >
              Forgot Password?
            </button>
          </div>

          <p className="text-[10px] text-muted-foreground text-center font-bold uppercase tracking-widest pt-6 border-t border-border/40">
            Secured by Ory Hydra + Kratos
          </p>
        </form>
      </motion.div>
    </div>
  );
}

function RedirectingFallback() {
  return (
    <div className="min-h-screen bg-background flex items-center justify-center">
      <div className="flex flex-col items-center gap-4">
        <Loader2 className="w-8 h-8 text-primary animate-spin" />
        <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
          Restarting sign-in…
        </p>
      </div>
    </div>
  );
}

export default function AuthLoginPage() {
  // useSearchParams() must be wrapped in a Suspense boundary so the Next.js
  // app router can stream the page during static generation; without this,
  // `next build` errors out with "useSearchParams should be wrapped in a
  // suspense boundary".
  return (
    <Suspense fallback={<RedirectingFallback />}>
      <AuthLoginInner />
    </Suspense>
  );
}
