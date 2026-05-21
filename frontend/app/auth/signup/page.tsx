"use client";

import Link from "next/link";
import { ShieldAlert } from "lucide-react";

export default function SignUpPage() {
  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-6">
      <div className="max-w-xl w-full glass border-border/60 rounded-[2.5rem] p-10 shadow-2xl text-center space-y-6">
        <div className="w-16 h-16 bg-warning/15 border border-warning/30 rounded-2xl mx-auto flex items-center justify-center">
          <ShieldAlert className="w-9 h-9 text-amber-300" />
        </div>
        <div className="space-y-2">
          <h1 className="text-3xl font-black text-foreground tracking-tighter uppercase">
            Sign Up Unavailable
          </h1>
          <p className="text-sm text-muted-foreground leading-relaxed">
            Self-signup is temporarily disabled during the Keycloak migration.
            Please request account provisioning from your system administrator.
          </p>
        </div>

        <div className="flex flex-col gap-3">
          <Link
            href="/auth/login"
            className="w-full bg-primary text-primary-foreground font-black py-3 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all uppercase tracking-widest text-xs"
          >
            Back to Sign In
          </Link>
          <p className="text-[11px] text-muted-foreground">
            Contact: <span className="text-foreground font-semibold">admin@devhub.local</span>
          </p>
        </div>
      </div>
    </div>
  );
}
