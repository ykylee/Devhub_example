package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// KratosSettingsCode classifies the outcomes the /api/v1/account/password
// proxy needs to surface to its caller (L4-C, work_26_05_11-e). The /account
// frontend maps these to its SettingsFlowErrorCode (REAUTH_REQUIRED /
// VALIDATION / FLOW_INIT_FAILED / SUBMIT_FAILED) so the UI can either show
// inline validation or hand the user back through /login.
type KratosSettingsCode string

const (
	// KratosSettingsValidation — Kratos rejected the new password (length,
	// breach list, complexity). Message carries the operator-visible reason.
	KratosSettingsValidation KratosSettingsCode = "validation"
	// KratosSettingsSessionInvalid — the session_token is unknown or expired.
	// Caller returns 401 + REAUTH_REQUIRED so the frontend can re-auth.
	KratosSettingsSessionInvalid KratosSettingsCode = "session_invalid"
	// KratosSettingsPrivilegedRequired — Kratos requires a fresh login flow
	// before privileged actions (privileged_session_max_age=15m). Same UI
	// outcome as session_invalid but a different audit reason.
	KratosSettingsPrivilegedRequired KratosSettingsCode = "privileged_required"
	// KratosSettingsFlowExpired — the flow was created but the lifespan
	// elapsed before submission. Caller can choose to retry the create.
	KratosSettingsFlowExpired KratosSettingsCode = "flow_expired"
)

// KratosSettingsError is returned from the settings flow helpers when Kratos
// reports a specific user-actionable outcome. Other failure modes (network,
// 5xx, decode) surface as plain errors so writeServerError can log them.
type KratosSettingsError struct {
	Code    KratosSettingsCode
	Message string
}

func (e *KratosSettingsError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return string(e.Code) + ": " + e.Message
	}
	return string(e.Code)
}

// IsKratosSettingsError returns the underlying KratosSettingsError when err
// wraps one, nil otherwise. Convenience for handler dispatch.
func IsKratosSettingsError(err error) *KratosSettingsError {
	var se *KratosSettingsError
	if errors.As(err, &se) {
		return se
	}
	return nil
}

// kratosSettingsFlowResponse mirrors the subset of /self-service/settings/api
// that the helpers consume.
type kratosSettingsFlowResponse struct {
	ID    string `json:"id"`
	State string `json:"state"`
	UI    struct {
		Messages []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"messages"`
		Nodes []struct {
			Messages []struct {
				Text string `json:"text"`
				Type string `json:"type"`
			} `json:"messages"`
		} `json:"nodes"`
	} `json:"ui"`
}

type kratosSettingsErrorEnvelope struct {
	Error struct {
		ID      string `json:"id"`
		Code    int    `json:"code"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	} `json:"error"`
	RedirectBrowserTo string `json:"redirect_browser_to"`
}

// CreateSettingsFlow opens an api-mode settings flow on behalf of the user
// identified by sessionToken. Returns the flow id; the caller passes it to
// SubmitSettingsPassword. We do not return ui because the password method
// has fixed fields and no per-flow csrf_token in api-mode.
func (c *KratosClient) CreateSettingsFlow(ctx context.Context, sessionToken string) (string, error) {
	if strings.TrimSpace(c.PublicURL) == "" {
		return "", errors.New("KratosClient.PublicURL is not configured")
	}
	if strings.TrimSpace(sessionToken) == "" {
		return "", &KratosSettingsError{Code: KratosSettingsSessionInvalid, Message: "session_token is empty"}
	}

	endpoint := strings.TrimRight(c.PublicURL, "/") + "/self-service/settings/api"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("build kratos settings flow: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Session-Token", sessionToken)

	resp, err := c.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("call kratos settings flow: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read kratos settings flow: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var flow kratosSettingsFlowResponse
		if err := json.Unmarshal(body, &flow); err != nil {
			return "", fmt.Errorf("decode kratos settings flow: %w", err)
		}
		if flow.ID == "" {
			return "", errors.New("kratos settings flow missing id")
		}
		return flow.ID, nil
	case http.StatusUnauthorized:
		return "", &KratosSettingsError{Code: KratosSettingsSessionInvalid, Message: "session_token rejected by kratos"}
	case http.StatusForbidden:
		env := decodeSettingsErrorEnvelope(body)
		if env.Error.ID == "session_refresh_required" {
			return "", &KratosSettingsError{Code: KratosSettingsPrivilegedRequired, Message: "privileged session required"}
		}
		return "", &KratosSettingsError{Code: KratosSettingsSessionInvalid, Message: env.Error.Reason}
	default:
		return "", fmt.Errorf("kratos settings flow status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

// SubmitSettingsPassword posts a password change against an existing
// settings flow id. Maps Kratos's responses onto KratosSettingsError so the
// /api/v1/account/password handler can pick the right HTTP status without
// re-parsing JSON.
func (c *KratosClient) SubmitSettingsPassword(ctx context.Context, sessionToken, flowID, newPassword string) error {
	if strings.TrimSpace(c.PublicURL) == "" {
		return errors.New("KratosClient.PublicURL is not configured")
	}
	if strings.TrimSpace(sessionToken) == "" {
		return &KratosSettingsError{Code: KratosSettingsSessionInvalid, Message: "session_token is empty"}
	}
	if strings.TrimSpace(flowID) == "" {
		return errors.New("flow_id is required")
	}

	endpoint := strings.TrimRight(c.PublicURL, "/") + "/self-service/settings?flow=" + flowID
	payload := map[string]any{"method": "password", "password": newPassword}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode kratos settings submit: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("build kratos settings submit: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Session-Token", sessionToken)

	resp, err := c.client().Do(req)
	if err != nil {
		return fmt.Errorf("call kratos settings submit: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read kratos settings submit: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var flow kratosSettingsFlowResponse
		if err := json.Unmarshal(body, &flow); err != nil {
			return fmt.Errorf("decode kratos settings submit: %w", err)
		}
		if flow.State == "success" {
			return nil
		}
		// 200 with state="show_form" means validation messages were attached
		// without an HTTP error code; surface them as a validation error.
		if msg := collectSettingsMessages(flow); msg != "" {
			return &KratosSettingsError{Code: KratosSettingsValidation, Message: msg}
		}
		return &KratosSettingsError{Code: KratosSettingsValidation, Message: "settings flow returned 200 without success state"}
	case http.StatusBadRequest:
		// Validation failures arrive as 400 + ui.messages (per-field) or
		// ui.nodes[].messages. Decode using the settings flow envelope so
		// we capture both.
		var flow kratosSettingsFlowResponse
		_ = json.Unmarshal(body, &flow)
		msg := collectSettingsMessages(flow)
		if msg == "" {
			msg = "password did not pass validation"
		}
		return &KratosSettingsError{Code: KratosSettingsValidation, Message: msg}
	case http.StatusUnauthorized:
		return &KratosSettingsError{Code: KratosSettingsSessionInvalid, Message: "session_token rejected by kratos"}
	case http.StatusForbidden:
		env := decodeSettingsErrorEnvelope(body)
		if env.Error.ID == "session_refresh_required" {
			return &KratosSettingsError{Code: KratosSettingsPrivilegedRequired, Message: "privileged session required"}
		}
		return &KratosSettingsError{Code: KratosSettingsSessionInvalid, Message: env.Error.Reason}
	case http.StatusGone:
		return &KratosSettingsError{Code: KratosSettingsFlowExpired, Message: "settings flow expired"}
	default:
		return fmt.Errorf("kratos settings submit status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func decodeSettingsErrorEnvelope(body []byte) kratosSettingsErrorEnvelope {
	var env kratosSettingsErrorEnvelope
	_ = json.Unmarshal(body, &env)
	return env
}

func collectSettingsMessages(flow kratosSettingsFlowResponse) string {
	out := make([]string, 0, 4)
	for _, m := range flow.UI.Messages {
		if m.Text != "" {
			out = append(out, m.Text)
		}
	}
	for _, n := range flow.UI.Nodes {
		for _, m := range n.Messages {
			if m.Text != "" {
				out = append(out, m.Text)
			}
		}
	}
	return strings.Join(out, "; ")
}
