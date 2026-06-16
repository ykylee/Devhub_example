import type { UserRole } from "@/lib/store";

// defaultLandingFor returns the path a user should land on right after login
// (or a forced redirect when they hit a route their role cannot view). The
// mapping is intentionally trivial — DEC-3=C keeps /admin as the system_admin
// landing zone, and developer/manager get their own dashboards.
export function defaultLandingFor(role: UserRole | null | undefined): string {
  switch (role) {
    case "System Admin":
      return "/admin";
    case "Manager":
      return "/manager";
    case "Developer":
    default:
      return "/developer";
  }
}

export function isSystemAdmin(role: UserRole | null | undefined): boolean {
  return role === "System Admin";
}

// pathRequiresSystemAdmin recognises every route that should be visible
// only to system administrators (DEC-3=C: /admin and /admin/settings/*).
// Keep the prefix list narrow — adding more paths here also hides them
// from non-admin sidebars (Sidebar consults this same predicate).
export function pathRequiresSystemAdmin(pathname: string): boolean {
  const adminBase = "/admin";
  const orgBase = "/organization";
  if (pathname === adminBase) return true;
  if (pathname.startsWith(`${adminBase}/`)) return true;
  // /organization stays system_admin-only until PR-S2 retires it in favour
  // of /admin/settings/*. Once that migration ships, drop this clause.
  if (pathname === orgBase) return true;
  if (pathname.startsWith(`${orgBase}/`)) return true;
  return false;
}
