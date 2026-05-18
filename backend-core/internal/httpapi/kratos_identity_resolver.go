package httpapi

import (
	"context"
	"errors"
	"strings"
)

// resolveIdPSubject returns the Kratos identity_id for a DevHub user.
//
// Fast path (L4-A, work_26_05_11-e): read users.idp_subject from the
// OrganizationStore. When that comes back empty (rows seeded before
// migration 000009, or freshly-created identities the eager path could not
// stamp) we fall back to IdentityAdmin.FindIdentityByUserID's
// /admin/identities page scan and best-effort backfill the column for next
// time.
//
// Returns ErrIdentityNotFound when neither the cache nor Kratos knows
// the user_id. Callers treat that as "user has never been onboarded to
// Kratos".
func (h Handler) resolveIdPSubject(ctx context.Context, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", errors.New("user_id is required")
	}

	// Fast path — DB cache.
	if h.cfg.OrganizationStore != nil {
		if user, err := h.cfg.OrganizationStore.GetUser(ctx, userID); err == nil {
			if id := strings.TrimSpace(user.IdPSubject); id != "" {
				return id, nil
			}
		}
		// GetUser miss or empty cache → fall through to the slow path.
	}

	// Slow path — scan Kratos and lazy-backfill the cache.
	if h.cfg.IdentityAdmin == nil {
		return "", ErrIdentityNotFound
	}
	identityID, err := h.cfg.IdentityAdmin.FindIdentityByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	if h.cfg.OrganizationStore != nil {
		// Best-effort: when the user row is absent (some tests use a bare
		// newMemoryOrganizationStore without CreateUser) SetIdPSubject
		// returns ErrNotFound; that is non-fatal here.
		if setErr := h.cfg.OrganizationStore.SetIdPSubject(ctx, userID, identityID); setErr != nil {
			logRequestCtx(ctx, "[kratos-cache] backfill idp_subject for %s skipped: %v", userID, setErr)
		}
	}
	return identityID, nil
}
