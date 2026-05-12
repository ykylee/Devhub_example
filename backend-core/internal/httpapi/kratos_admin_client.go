package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrKratosIdentityNotFound is returned when no Kratos identity matches the
// supplied DevHub user_id (or the underlying GET returns 404).
var ErrKratosIdentityNotFound = errors.New("kratos identity not found")

// KratosAdminClient drives identity management on the Ory Kratos Admin API.
type KratosAdminClient struct {
	AdminURL   string
	HTTPClient *http.Client
}

// KratosCreateIdentityRequest is the payload for POST /admin/identities.
type KratosCreateIdentityRequest struct {
	SchemaID string `json:"schema_id"`
	State    string `json:"state"`
	Traits   struct {
		SystemID    string `json:"system_id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
	} `json:"traits"`
	MetadataPublic struct {
		UserID string `json:"user_id"`
	} `json:"metadata_public"`
	Credentials struct {
		Password struct {
			Config struct {
				Password string `json:"password"`
			} `json:"config"`
		} `json:"password"`
	} `json:"credentials"`
}

func (c *KratosAdminClient) CreateIdentity(ctx context.Context, email, name, userID, password string) (string, error) {
	if strings.TrimSpace(c.AdminURL) == "" {
		return "", fmt.Errorf("KratosAdminClient.AdminURL is not configured")
	}

	reqBody := KratosCreateIdentityRequest{
		SchemaID: "devhub_user",
		State:    "active",
	}
	reqBody.Traits.SystemID = userID
	reqBody.Traits.Email = email
	reqBody.Traits.DisplayName = name
	reqBody.MetadataPublic.UserID = userID
	reqBody.Credentials.Password.Config.Password = password

	encoded, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("encode kratos create identity: %w", err)
	}

	endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return "", fmt.Errorf("build kratos create identity request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("call kratos create identity: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read kratos create identity response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("kratos create identity status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var raw struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", fmt.Errorf("decode kratos create identity response: %w", err)
	}

	return raw.ID, nil
}

// FindIdentityByUserID locates a Kratos identity whose metadata_public.user_id
// equals the supplied DevHub user_id.
//
// Fast path: Kratos /admin/identities supports a `credentials_identifier`
// query parameter that filters on the password method's identifier. By
// convention (seedLocalAdmin + globalSetup), DevHub identities populate
// `traits.system_id == metadata_public.user_id`, so a single query with
// credentials_identifier=<user_id> returns the right identity in O(1). We
// still validate metadata_public.user_id on the result so a divergent
// operator setup (system_id != user_id) falls back to the scan instead of
// silently returning a wrong identity.
//
// Slow path: page through /admin/identities and match metadata_public.user_id
// in-process. Used when the fast path returns nothing or the metadata does
// not align — e.g., identities created before this code path was wired.
// Capped at 40 pages × 250 = 10k identities to avoid hammering Kratos when
// metadata_public.user_id was never populated.
func (c *KratosAdminClient) FindIdentityByUserID(ctx context.Context, userID string) (string, error) {
	if strings.TrimSpace(c.AdminURL) == "" {
		return "", fmt.Errorf("KratosAdminClient.AdminURL is not configured")
	}
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("user_id is required")
	}

	if id, ok, err := c.findByCredentialsIdentifier(ctx, userID); err != nil {
		return "", err
	} else if ok {
		return id, nil
	}

	// Kratos /admin/identities uses 0-based pagination (verified against
	// v26.2.0 — page=0 returns the first batch, page=1 returns the second,
	// …). The earlier 1-based start silently returned an empty first page
	// and short-circuited to ErrKratosIdentityNotFound even when the user
	// existed; that masked seedLocalAdmin recovery in e2e setups.
	page := 0
	const pageSize = 250
	for {
		params := url.Values{}
		params.Set("page", fmt.Sprintf("%d", page))
		params.Set("per_page", fmt.Sprintf("%d", pageSize))
		endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities?" + params.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return "", fmt.Errorf("build kratos list identities: %w", err)
		}
		req.Header.Set("Accept", "application/json")

		resp, err := c.client().Do(req)
		if err != nil {
			return "", fmt.Errorf("call kratos list identities: %w", err)
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("read kratos list identities: %w", readErr)
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("kratos list identities status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var batch []struct {
			ID             string `json:"id"`
			MetadataPublic struct {
				UserID string `json:"user_id"`
			} `json:"metadata_public"`
		}
		if err := json.Unmarshal(body, &batch); err != nil {
			return "", fmt.Errorf("decode kratos list identities: %w", err)
		}
		for _, ident := range batch {
			if ident.MetadataPublic.UserID == userID {
				return ident.ID, nil
			}
		}
		if len(batch) < pageSize {
			return "", ErrKratosIdentityNotFound
		}
		page++
		if page > 39 { // 10k cap for the PoC scan (pages 0..39 = 40 batches × 250)
			return "", ErrKratosIdentityNotFound
		}
	}
}

// findByCredentialsIdentifier attempts the O(1) Kratos query and validates
// that the returned identity still carries metadata_public.user_id matching
// the DevHub user_id. Returns (id, true, nil) on a confirmed match,
// (_, false, nil) for any miss (empty result / metadata mismatch / unexpected
// shape) so the caller falls through to the slow scan, and (_, false, err)
// only for network / HTTP-level failures the caller must surface.
func (c *KratosAdminClient) findByCredentialsIdentifier(ctx context.Context, userID string) (string, bool, error) {
	params := url.Values{}
	params.Set("credentials_identifier", userID)
	endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", false, fmt.Errorf("build kratos credentials_identifier query: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return "", false, fmt.Errorf("call kratos credentials_identifier query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false, fmt.Errorf("read kratos credentials_identifier query: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// Older Kratos releases reject the parameter; treat any non-200 as a
		// silent miss so the slow scan still runs.
		return "", false, nil
	}

	var batch []struct {
		ID             string `json:"id"`
		MetadataPublic struct {
			UserID string `json:"user_id"`
		} `json:"metadata_public"`
	}
	if err := json.Unmarshal(body, &batch); err != nil {
		return "", false, nil
	}
	for _, ident := range batch {
		if ident.MetadataPublic.UserID == userID {
			return ident.ID, true, nil
		}
	}
	return "", false, nil
}

// getIdentityRaw fetches the identity as a generic map so we can mutate one
// field and round-trip via PUT without depending on the full Kratos schema.
func (c *KratosAdminClient) getIdentityRaw(ctx context.Context, identityID string) (map[string]any, error) {
	endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities/" + url.PathEscape(identityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build kratos get identity: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("call kratos get identity: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read kratos get identity: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrKratosIdentityNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kratos get identity status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode kratos get identity: %w", err)
	}
	return raw, nil
}

// UpdateIdentityPassword overwrites the password credential for the given
// identity. Kratos has no first-class admin "set password" endpoint; the
// supported pattern is GET -> mutate credentials.password.config -> PUT.
func (c *KratosAdminClient) UpdateIdentityPassword(ctx context.Context, identityID, password string) error {
	if strings.TrimSpace(c.AdminURL) == "" {
		return fmt.Errorf("KratosAdminClient.AdminURL is not configured")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}
	current, err := c.getIdentityRaw(ctx, identityID)
	if err != nil {
		return err
	}
	creds, _ := current["credentials"].(map[string]any)
	if creds == nil {
		creds = map[string]any{}
	}
	creds["password"] = map[string]any{
		"config": map[string]any{"password": password},
	}
	current["credentials"] = creds
	return c.putIdentity(ctx, identityID, current)
}

// SetIdentityState toggles the identity between "active" and "inactive". An
// inactive identity cannot complete the login flow, which is the property
// account-disable relies on.
func (c *KratosAdminClient) SetIdentityState(ctx context.Context, identityID string, active bool) error {
	if strings.TrimSpace(c.AdminURL) == "" {
		return fmt.Errorf("KratosAdminClient.AdminURL is not configured")
	}
	current, err := c.getIdentityRaw(ctx, identityID)
	if err != nil {
		return err
	}
	if active {
		current["state"] = "active"
	} else {
		current["state"] = "inactive"
	}
	return c.putIdentity(ctx, identityID, current)
}

// DeleteIdentity removes the identity from Kratos. Use after DevHub
// users.status -> "disabled" or alongside DevHub user delete.
func (c *KratosAdminClient) DeleteIdentity(ctx context.Context, identityID string) error {
	if strings.TrimSpace(c.AdminURL) == "" {
		return fmt.Errorf("KratosAdminClient.AdminURL is not configured")
	}
	endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities/" + url.PathEscape(identityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build kratos delete identity: %w", err)
	}
	resp, err := c.client().Do(req)
	if err != nil {
		return fmt.Errorf("call kratos delete identity: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrKratosIdentityNotFound
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("kratos delete identity status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *KratosAdminClient) putIdentity(ctx context.Context, identityID string, body map[string]any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode kratos put identity: %w", err)
	}
	endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities/" + url.PathEscape(identityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("build kratos put identity: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return fmt.Errorf("call kratos put identity: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrKratosIdentityNotFound
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("kratos put identity status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func (c *KratosAdminClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}

// MockKratosAdmin is a development-only mock that simulates identity
// management. Tracks the calls so handler tests can assert on them.
type MockKratosAdmin struct {
	CreatedIDs       []string
	PasswordResets   []string
	StateChanges     map[string]bool
	DeletedIDs       []string
	FindIDOverride   map[string]string
	// FindCalls counts how many times FindIdentityByUserID was invoked. The
	// L4-A cache hit test asserts this stays at zero when the DevHub users
	// row already carries a kratos_identity_id.
	FindCalls       int
	FindError       error
	UpdatePassError error
	SetStateError   error
	DeleteError     error
}

func (m *MockKratosAdmin) CreateIdentity(_ context.Context, email, name, userID, password string) (string, error) {
	fakeID := fmt.Sprintf("mock-k-id-%s", userID)
	m.CreatedIDs = append(m.CreatedIDs, fakeID)
	fmt.Printf("[MockKratosAdmin] Identity Created: Email=%s, Name=%s, UserID=%s (Password was received)\n", email, name, userID)
	_ = password
	return fakeID, nil
}

func (m *MockKratosAdmin) FindIdentityByUserID(_ context.Context, userID string) (string, error) {
	m.FindCalls++
	if m.FindError != nil {
		return "", m.FindError
	}
	if id, ok := m.FindIDOverride[userID]; ok {
		return id, nil
	}
	return fmt.Sprintf("mock-k-id-%s", userID), nil
}

func (m *MockKratosAdmin) UpdateIdentityPassword(_ context.Context, identityID, password string) error {
	if m.UpdatePassError != nil {
		return m.UpdatePassError
	}
	m.PasswordResets = append(m.PasswordResets, identityID)
	_ = password
	return nil
}

func (m *MockKratosAdmin) SetIdentityState(_ context.Context, identityID string, active bool) error {
	if m.SetStateError != nil {
		return m.SetStateError
	}
	if m.StateChanges == nil {
		m.StateChanges = map[string]bool{}
	}
	m.StateChanges[identityID] = active
	return nil
}

func (m *MockKratosAdmin) DeleteIdentity(_ context.Context, identityID string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	m.DeletedIDs = append(m.DeletedIDs, identityID)
	return nil
}
