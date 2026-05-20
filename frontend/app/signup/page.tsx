import { redirect } from "next/navigation";

// /signup is a legacy alias that redirects to the canonical /auth/signup route
// (self-signup is currently disabled during Keycloak migration; the canonical
// page renders a "Sign Up Unavailable" notice — see tests/e2e/signup.spec.ts).
export default function SignupAlias() {
  redirect("/auth/signup");
}
