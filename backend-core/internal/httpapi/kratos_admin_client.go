package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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

func (c *KratosAdminClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Second}
}

// MockKratosAdmin is a development-only mock that simulates identity creation.
type MockKratosAdmin struct{}

func (m *MockKratosAdmin) CreateIdentity(ctx context.Context, email, name, userID, password string) (string, error) {
	fakeID := fmt.Sprintf("mock-k-id-%s", userID)
	fmt.Printf("[MockKratosAdmin] Identity Created: Email=%s, Name=%s, UserID=%s (Password was received)\n", email, name, userID)
	return fakeID, nil
}
