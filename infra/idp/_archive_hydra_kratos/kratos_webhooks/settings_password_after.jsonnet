// Kratos web_hook payload for selfservice.flows.settings.after.password.hooks.
//
// PR-M2-AUDIT (claude/login_usermanagement_finish).
//
// Kept intentionally minimal: identity_id is the primary correlation key
// for audit_logs (FK-like join with users.kratos_identity_id), email is
// the human-readable hint for audit_logs.payload. Event time is taken
// from audit_logs.created_at on the DevHub side — we deliberately do NOT
// pass occurred_at from Kratos to avoid relying on extVars that the
// Kratos jsonnet sandbox does not expose by default.
//
// Any future field additions (e.g. session.identity.traits.full_name)
// extend this jsonnet AND the kratosPasswordChangedPayload struct in
// backend-core/internal/httpapi/kratos_webhook.go together.
function(ctx) {
  identity_id: ctx.identity.id,
  email: if std.objectHas(ctx.identity, 'traits') && std.objectHas(ctx.identity.traits, 'email')
         then ctx.identity.traits.email
         else '',
}
