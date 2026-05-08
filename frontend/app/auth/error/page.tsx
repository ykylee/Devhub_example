"use client";

import { Suspense } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { AlertTriangle } from "lucide-react";

// Hydra (urls.error) and Kratos (selfservice.flows.error.ui_url) redirect
// here when the OIDC or self-service flow itself fails — bad client_id,
// expired challenge, schema validation, etc. We display the human-readable
// hint and offer a single "restart" link back to /login.

interface ErrorPayload {
  id?: string;
  reason?: string;
  description?: string;
}

function decode(searchParams: URLSearchParams): ErrorPayload {
  // Both Hydra (?error=...&error_description=...) and Kratos (?id=<error_id>)
  // funnel through here. Kratos requires a follow-up GET /self-service/errors
  // call to fetch the structured payload; for now we display whatever is in
  // the query string and link the operator at the id for follow-up.
  return {
    id: searchParams.get("id") ?? searchParams.get("error") ?? undefined,
    reason: searchParams.get("error_hint") ?? searchParams.get("reason") ?? undefined,
    description: searchParams.get("error_description") ?? searchParams.get("message") ?? undefined,
  };
}

function AuthErrorInner() {
  const searchParams = useSearchParams();
  const payload = decode(new URLSearchParams(searchParams.toString()));

  return (
    <div className="min-h-screen bg-[#030014] flex items-center justify-center p-4">
      <div className="w-full max-w-md glass border-white/10 rounded-[2rem] p-10 shadow-2xl backdrop-blur-2xl text-center space-y-6">
        <div className="inline-flex p-4 rounded-3xl bg-red-500/15 border border-red-500/30">
          <AlertTriangle className="w-10 h-10 text-red-300" />
        </div>
        <div className="space-y-2">
          <h1 className="text-2xl font-black text-white tracking-tighter">Sign-in failed</h1>
          <p className="text-sm text-muted-foreground">
            {payload.description || payload.reason || "The authentication flow returned an error."}
          </p>
          {payload.id && (
            <p className="text-[10px] text-white/30 font-mono break-all">id: {payload.id}</p>
          )}
        </div>
        <Link
          href="/login"
          className="inline-flex items-center justify-center w-full bg-primary text-white font-black py-3 rounded-2xl hover:scale-[1.02] active:scale-[0.98] transition-all uppercase tracking-widest text-xs"
        >
          Restart sign-in
        </Link>
      </div>
    </div>
  );
}

export default function AuthErrorPage() {
  return (
    <Suspense fallback={null}>
      <AuthErrorInner />
    </Suspense>
  );
}
