package view

import (
	"github.com/devhub/backend-core/internal/shared/httphelp"
	"context"
	"errors"
	"strings"
)

// resolveIdPSubject returns the IdP identity_id for a DevHub user.
//
// Fast path: read users.idp_subject from the OrganizationStore. When that
// comes back empty (rows seeded before migration 000009/000030, or freshly-
// created identities the eager path could not stamp) we fall back to
// IdentityAdmin.FindIdentityByUserID and best-effort backfill the column
// for next time.
//
// Returns httphelp.httphelp.ErrIdentityNotFound when neither the cache nor the IdP knows
// the user_id.
func (h *AuthHandler) resolveIdPSubject(ctx context.Context, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", errors.New("user_id is required")
	}

	if h.cfg.OrganizationStore != nil {
		if user, err := h.cfg.OrganizationStore.GetUser(ctx, userID); err == nil {
			if id := strings.TrimSpace(user.IdPSubject); id != "" {
				return id, nil
			}
		}
	}

	if h.cfg.IdentityAdmin == nil {
		return "", httphelp.ErrIdentityNotFound
	}
	identityID, err := h.cfg.IdentityAdmin.FindIdentityByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	if h.cfg.OrganizationStore != nil {
		if setErr := h.cfg.OrganizationStore.SetIdPSubject(ctx, userID, identityID); setErr != nil {
			httphelp.LogRequestCtx(ctx, "[idp-cache] backfill idp_subject for %s skipped: %v", userID, setErr)
		}
	}
	return identityID, nil
}
