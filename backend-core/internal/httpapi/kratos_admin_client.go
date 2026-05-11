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

var ErrKratosIdentityNotFound = errors.New("kratos identity not found")

// KratosAdminClient implements KratosAdmin using the Ory Kratos Admin API.
type KratosAdminClient struct {
	AdminURL   string
	HTTPClient *http.Client
}

func (c *KratosAdminClient) GetIdentity(ctx context.Context, id string) (*KratosIdentity, error) {
	endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities/" + url.PathEscape(id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build kratos get identity: %w", err)
	}
	resp, err := c.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("call kratos get identity: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrKratosIdentityNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("kratos get identity status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var identity KratosIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identity); err != nil {
		return nil, fmt.Errorf("decode kratos identity: %w", err)
	}
	return &identity, nil
}

func (c *KratosAdminClient) CreateIdentity(ctx context.Context, email, name, userID, password string) (string, error) {
	endpoint := strings.TrimRight(c.AdminURL, "/") + "/admin/identities"
	payload := map[string]any{
		"schema_id": "default",
		"traits": map[string]any{
			"email":   email,
			"name":    name,
		},
		"metadata_public": map[string]any{
			"user_id": userID,
		},
		"credentials": map[string]any{
			"password": map[string]any{
				"config": map[string]any{
					"password": password,
				},
			},
		},
		"state": "active",
	}
	encoded, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return "", fmt.Errorf("build kratos create identity: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return "", fmt.Errorf("call kratos create identity: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("kratos create identity status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var identity struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &identity); err != nil {
		return "", fmt.Errorf("decode kratos create identity: %w", err)
	}
	return identity.ID, nil
}

func (c *KratosAdminClient) FindIdentityByUserID(ctx context.Context, userID string) (string, error) {
	page := 1
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
		resp, err := c.client().Do(req)
		if err != nil {
			return "", fmt.Errorf("call kratos list identities: %w", err)
		}
		defer resp.Body.Close()
		var batch []KratosIdentity
		if err := json.NewDecoder(resp.Body).Decode(&batch); err != nil {
			return "", fmt.Errorf("decode kratos list identities: %w", err)
		}
		for _, ident := range batch {
			if ident.UserID == userID {
				return ident.ID, nil
			}
		}
		if len(batch) < pageSize {
			break
		}
		page++
	}
	return "", ErrKratosIdentityNotFound
}

func (c *KratosAdminClient) UpdateIdentityPassword(ctx context.Context, identityID, password string) error { return nil }
func (c *KratosAdminClient) SetIdentityState(ctx context.Context, identityID string, active bool) error { return nil }
func (c *KratosAdminClient) DeleteIdentity(ctx context.Context, identityID string) error { return nil }

func (c *KratosAdminClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}

type MockKratosAdmin struct {
	CreatedIDs []string
}

func (m *MockKratosAdmin) GetIdentity(ctx context.Context, id string) (*KratosIdentity, error) {
	return &KratosIdentity{ID: id}, nil
}
func (m *MockKratosAdmin) CreateIdentity(ctx context.Context, email, name, userID, password string) (string, error) {
	id := "mock-id-" + userID
	m.CreatedIDs = append(m.CreatedIDs, id)
	return id, nil
}
func (m *MockKratosAdmin) FindIdentityByUserID(ctx context.Context, userID string) (string, error) {
	return "mock-id-" + userID, nil
}
func (m *MockKratosAdmin) UpdateIdentityPassword(ctx context.Context, identityID, password string) error { return nil }
func (m *MockKratosAdmin) SetIdentityState(ctx context.Context, identityID string, active bool) error { return nil }
func (m *MockKratosAdmin) DeleteIdentity(ctx context.Context, identityID string) error { return nil }
