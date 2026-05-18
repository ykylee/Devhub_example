package httpapi

import "context"

// PasswordAuthClient is the backend contract for current-password verification
// and privileged settings flow based password updates.
type PasswordAuthClient interface {
	CreateLoginFlow(ctx context.Context) (IDPLoginFlow, error)
	SubmitLogin(ctx context.Context, flow IDPLoginFlow, identifier, password string) (IDPIdentity, error)
	CreateSettingsFlow(ctx context.Context, sessionToken string) (string, error)
	SubmitSettingsPassword(ctx context.Context, sessionToken, flowID, newPassword string) error
}
