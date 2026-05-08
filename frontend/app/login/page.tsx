"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import { ArrowRight, ShieldCheck } from "lucide-react";

// OIDC entry point. Defaults assume a local PoC: Hydra public on :4444 issuing tokens for the devhub-frontend client, redirecting back to the SPA root after consent.
const OIDC_LOGIN_URL =
  process.env.NEXT_PUBLIC_OIDC_LOGIN_URL ?? "http://127.0.0.1:4444/oauth2/auth";
const OIDC_CLIENT_ID =
  process.env.NEXT_PUBLIC_OIDC_CLIENT_ID ?? "devhub-frontend";
// /auth/callback is the redirect_uri Hydra has registered for the
// devhub-frontend client (infra/idp/scripts/register-devhub-client.ps1).
// Keep these aligned: changing one without the other breaks the OIDC code
// flow with "redirect_uri_mismatch".
const OIDC_REDIRECT_URI =
  process.env.NEXT_PUBLIC_OIDC_REDIRECT_URI ?? "http://127.0.0.1:3000/auth/callback";
const OIDC_SCOPE =
  process.env.NEXT_PUBLIC_OIDC_SCOPE ?? "openid offline";

function buildAuthorizeURL(): string {
  const url = new URL(OIDC_LOGIN_URL);
  url.searchParams.set("client_id", OIDC_CLIENT_ID);
  url.searchParams.set("response_type", "code");
  url.searchParams.set("redirect_uri", OIDC_REDIRECT_URI);
  url.searchParams.set("scope", OIDC_SCOPE);
  url.searchParams.set("state", crypto.randomUUID());
  return url.toString();
}

export default function LoginPage() {
  const [isRedirecting, setIsRedirecting] = useState(false);

  const handleLogin = () => {
    setIsRedirecting(true);
    window.location.assign(buildAuthorizeURL());
  };

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
