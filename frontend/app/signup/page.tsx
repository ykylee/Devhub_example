import { redirect } from "next/navigation";

// /signup is a convenience alias for /auth/signup (RM-M3-01 form). The
// canonical route lives at /auth/signup alongside the rest of the auth
// pages (login, callback). sprint claude/work_260513-n: the previous
// PR introduced /signup as a standalone form, but /auth/signup was
// already there with a richer form (Full Name / System ID / Employee
// ID / Password / Confirm Password) and four e2e specs (TC-SIGNUP-01..04
// in tests/e2e/signup.spec.ts). Collapsing /signup into a redirect
// keeps a single source-of-truth.
export default function SignupAlias() {
  redirect("/auth/signup");
}
