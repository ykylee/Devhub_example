package httpapi

import "context"

type fakeKratosLogin struct {
	flow              IDPLoginFlow
	flowErr           error
	identity          IDPIdentity
	submitErr         error
	submitCalls       int
	submitIdentifiers []string
	submitPasswords   []string
	settingsFlowID    string
	settingsFlowErr   error
	settingsSubmitErr error
	settingsSubmits   []struct {
		SessionToken string
		FlowID       string
		NewPassword  string
	}
}

func (f *fakeKratosLogin) CreateLoginFlow(_ context.Context) (IDPLoginFlow, error) {
	if f.flowErr != nil {
		return IDPLoginFlow{}, f.flowErr
	}
	return f.flow, nil
}

func (f *fakeKratosLogin) SubmitLogin(_ context.Context, _ IDPLoginFlow, identifier, password string) (IDPIdentity, error) {
	f.submitCalls++
	f.submitIdentifiers = append(f.submitIdentifiers, identifier)
	f.submitPasswords = append(f.submitPasswords, password)
	if f.submitErr != nil {
		return IDPIdentity{}, f.submitErr
	}
	return f.identity, nil
}

func (f *fakeKratosLogin) CreateSettingsFlow(_ context.Context, _ string) (string, error) {
	if f.settingsFlowErr != nil {
		return "", f.settingsFlowErr
	}
	return f.settingsFlowID, nil
}

func (f *fakeKratosLogin) SubmitSettingsPassword(_ context.Context, sessionToken, flowID, newPassword string) error {
	f.settingsSubmits = append(f.settingsSubmits, struct {
		SessionToken string
		FlowID       string
		NewPassword  string
	}{
		SessionToken: sessionToken,
		FlowID:       flowID,
		NewPassword:  newPassword,
	})
	return f.settingsSubmitErr
}
