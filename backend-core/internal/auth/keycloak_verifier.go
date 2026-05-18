package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/devhub/backend-core/internal/httpapi"
)

// KeycloakJWKSVerifier is a PR-A skeleton for Keycloak-based bearer token
// verification. Real JWT/JWKS verification is implemented in PR-B.
type KeycloakJWKSVerifier struct {
	IssuerURL string
	JWKSURL   string
	ClientID  string
}

func (v *KeycloakJWKSVerifier) VerifyBearerToken(_ context.Context, _ string) (httpapi.AuthenticatedActor, error) {
	if strings.TrimSpace(v.IssuerURL) == "" && strings.TrimSpace(v.JWKSURL) == "" {
		return httpapi.AuthenticatedActor{}, errors.New("KeycloakJWKSVerifier requires DEVHUB_OIDC_ISSUER_URL or DEVHUB_OIDC_JWKS_URL")
	}
	return httpapi.AuthenticatedActor{}, errors.New("KeycloakJWKSVerifier is not implemented yet")
}
