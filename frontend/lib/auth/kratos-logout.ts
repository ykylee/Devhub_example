"use client";

// performKratosBrowserLogout drives the Kratos public /self-service/logout/browser
// flow that destroys the Kratos session cookie. It is the frontend half of
// DEC-1=B: Hydra is owned by the backend (/api/v1/auth/logout), Kratos is owned
// by the browser because Kratos identifies the user via cookie.
//
// On success the function never returns — it navigates to logout_url, which
// finishes by sending the browser to default_browser_return_url (/) per
// kratos.yaml selfservice.flows.logout.after.
//
// On failure (no Kratos session, network error) it falls back to a navigation
// to fallbackReturnTo so the caller still leaves the authenticated state.
export async function performKratosBrowserLogout(fallbackReturnTo: string): Promise<void> {
  const base = (process.env.NEXT_PUBLIC_KRATOS_PUBLIC_URL ?? "http://localhost:4433").replace(/\/$/, "");
  try {
    const res = await fetch(`${base}/self-service/logout/browser`, {
      credentials: "include",
      headers: { Accept: "application/json" },
    });
    if (res.ok) {
      const data = (await res.json()) as { logout_url?: string };
      if (data.logout_url) {
        window.location.assign(data.logout_url);
        return;
      }
    } else {
      console.warn(`[kratos-logout] flow init returned ${res.status}; falling back`);
    }
  } catch (err) {
    console.warn("[kratos-logout] flow init failed (no session?)", err);
  }
  window.location.assign(fallbackReturnTo);
}
