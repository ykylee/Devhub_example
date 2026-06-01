"use client";

import { motion } from "framer-motion";
import { User, Mail, Shield, Key, ExternalLink } from "lucide-react";
import { useEffect, useState } from "react";
import { useStore } from "@/lib/store";
import { authService } from "@/domain/auth-session/service/auth.service";
import { ProfileSelfEdit } from "@/components/account/ProfileSelfEdit";

// ADR-0019 / sprint claude/work_260519-ad: self-service password change is
// delegated to the Keycloak Account Console. DevHub no longer proxies the
// flow (the previous Kratos-based proxy was wire-deleted along with the
// rest of the Kratos residual).
//
// Stage 3 (codex P1 + self-review P1-2): the Account Console URL is resolved
// via authService.getAccountConsoleURL() so deployments that surface the
// issuer through /api/runtime-config (server env) keep the link working
// without baking NEXT_PUBLIC_OIDC_ISSUER_URL into the browser bundle.

export default function AccountPage() {
  const { actor } = useStore();
  const [accountConsoleURL, setAccountConsoleURL] = useState("");

  useEffect(() => {
    let cancelled = false;
    authService.getAccountConsoleURL().then((url) => {
      if (!cancelled) setAccountConsoleURL(url);
    }).catch(() => {
      if (!cancelled) setAccountConsoleURL("");
    });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="space-y-8 max-w-4xl mx-auto pb-12">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-black text-foreground dark:text-primary-foreground tracking-tighter uppercase">
          Account <span className="text-primary">Settings</span>
        </h1>
        <p className="text-muted-foreground font-bold text-xs uppercase tracking-widest">
          Manage your DevHub identity and security preferences
        </p>
      </div>

      <ProfileSelfEdit />

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          className="md:col-span-1 space-y-6"
        >
          <div className="glass border-border rounded-3xl p-8 text-center space-y-4">
            <div className="w-24 h-24 mx-auto rounded-2xl bg-gradient-to-br from-primary to-accent flex items-center justify-center border border-border/80 shadow-2xl ring-4 ring-primary/10">
              <User className="w-12 h-12 text-primary-foreground" />
            </div>
            <div className="space-y-1">
              <h2 className="text-xl font-bold text-foreground dark:text-primary-foreground tracking-tight">{actor?.login || "Guest User"}</h2>
              <p className="text-xs font-bold text-primary uppercase tracking-widest">{actor?.role || "Developer"}</p>
            </div>
            <div className="pt-4 border-t border-border/60 space-y-3">
              <div className="flex items-center gap-3 text-xs text-muted-foreground justify-center">
                <Mail className="w-3.5 h-3.5" />
                <span className="font-medium">{actor?.email || `${actor?.login}@example.com`}</span>
              </div>
              <div className="flex items-center gap-3 text-[10px] text-foreground/55 dark:text-primary-foreground/40 justify-center uppercase tracking-widest font-black">
                <Shield className="w-3 h-3" />
                <span>Subject: {actor?.subject?.slice(0, 8) || "N/A"}...</span>
              </div>
            </div>
          </div>

          <div className="glass border-border rounded-3xl p-6 space-y-4">
            <h3 className="text-[10px] font-black text-muted-foreground uppercase tracking-[0.2em] px-2">Session Info</h3>
            <div className="space-y-3">
              <div className="flex items-center justify-between px-2">
                <span className="text-[10px] font-bold text-foreground/60 dark:text-primary-foreground/50 uppercase tracking-widest">Source</span>
                <span className="text-[10px] font-mono text-primary font-bold uppercase">{actor?.source || "Local"}</span>
              </div>
              <div className="flex items-center justify-between px-2">
                <span className="text-[10px] font-bold text-foreground/60 dark:text-primary-foreground/50 uppercase tracking-widest">MFA Status</span>
                <span className="text-[10px] font-bold text-success uppercase">Disabled</span>
              </div>
            </div>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          className="md:col-span-2"
        >
          <div className="glass border-border rounded-3xl overflow-hidden shadow-2xl">
            <div className="bg-muted/30 px-8 py-6 border-b border-border flex items-center gap-4">
              <div className="p-2.5 rounded-xl bg-primary/20 border border-primary/20">
                <Key className="w-5 h-5 text-primary" />
              </div>
              <div>
                <h3 className="text-lg font-bold text-foreground dark:text-primary-foreground tracking-tight">Security Credentials</h3>
                <p className="text-xs text-muted-foreground">Password management lives in the Keycloak Account Console</p>
              </div>
            </div>

            <div className="p-8 space-y-6 text-sm text-foreground/90 dark:text-primary-foreground/90 leading-relaxed">
              <p>
                DevHub uses Keycloak as its single identity provider (ADR-0019). To change your password,
                set up multi-factor authentication, or review your active sessions, open the Keycloak
                Account Console.
              </p>
              <p className="text-xs text-muted-foreground">
                You will sign in to Keycloak directly. After updating your credentials there, return
                here and continue working — your DevHub session keeps its existing access token until
                it expires.
              </p>

              <div className="pt-4 border-t border-border/60 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <p className="text-[10px] text-muted-foreground font-bold uppercase tracking-widest max-w-[260px]">
                  Identity actions are owned by Keycloak. DevHub no longer proxies password changes.
                </p>
                {accountConsoleURL ? (
                  <a
                    href={accountConsoleURL}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="bg-primary text-primary-foreground font-black px-8 py-3.5 rounded-2xl hover:scale-105 active:scale-95 transition-all flex items-center gap-2 shadow-lg shadow-primary/20 uppercase tracking-widest text-xs"
                  >
                    Open Keycloak Console
                    <ExternalLink className="w-4 h-4" />
                  </a>
                ) : (
                  <span
                    data-testid="account-console-unavailable"
                    className="text-[10px] text-warning font-bold uppercase tracking-widest"
                  >
                    OIDC issuer URL is not configured
                  </span>
                )}
              </div>
            </div>
          </div>

          <div className="mt-8 glass border-border rounded-3xl p-8 border-l-4 border-l-warning/50">
            <div className="flex items-start gap-4">
              <div className="p-2.5 rounded-xl bg-warning/10 border border-warning/20 text-warning">
                <Shield className="w-5 h-5" />
              </div>
              <div className="space-y-1">
                <h4 className="text-foreground dark:text-primary-foreground font-bold tracking-tight">Two-Factor Authentication</h4>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  Enable MFA from the Keycloak Account Console under Signing In. DevHub honors any
                  MFA decision Keycloak enforces during sign-in.
                </p>
              </div>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  );
}
