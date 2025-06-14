package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GiteaRepo is the response payload from Gitea's create-repo endpoints.
type GiteaRepo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	CloneURL      string `json:"clone_url"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

// GiteaRepoOptions configures a Gitea create-repo call.
type GiteaRepoOptions struct {
	Description string `json:"description,omitempty"`
	Private     bool   `json:"private"`
	AutoInit    bool   `json:"auto_init"`
	Readme      string `json:"readme,omitempty"`
}

// CreateUserRepo creates a repo in the authenticated user's namespace.
// Gitea endpoint: POST /api/v1/user/repos
func (c *GiteaClient) CreateUserRepo(ctx context.Context, name string, options GiteaRepoOptions) (*GiteaRepo, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("gitea: repo name is required")
	}
	body, err := json.Marshal(struct {
		Name string `json:"name"`
		GiteaRepoOptions
	}{Name: name, GiteaRepoOptions: options})
	if err != nil {
		return nil, fmt.Errorf("gitea: encode body: %w", err)
	}
	u := c.BaseURL + "/api/v1/user/repos"
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gitea: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &GiteaAPIError{Class: "network", Message: err.Error(), Cause: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, &GiteaAPIError{
			Class:       giteaErrorClass(resp.StatusCode),
			Message:     strings.TrimSpace(string(errBody)),
			HTTPStatus:  resp.StatusCode,
			Endpoint:    u,
		}
	}
	var repo GiteaRepo
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("gitea: decode response: %w", err)
	}
	return &repo, nil
}

// CreateOrgRepo creates a repo in the given organization namespace.
// Gitea endpoint: POST /api/v1/orgs/{org}/repos
func (c *GiteaClient) CreateOrgRepo(ctx context.Context, org, name string, options GiteaRepoOptions) (*GiteaRepo, error) {
	if strings.TrimSpace(org) == "" {
		return nil, fmt.Errorf("gitea: org is required")
	}
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("gitea: repo name is required")
	}
	body, err := json.Marshal(struct {
		Name string `json:"name"`
		GiteaRepoOptions
	}{Name: name, GiteaRepoOptions: options})
	if err != nil {
		return nil, fmt.Errorf("gitea: encode body: %w", err)
	}
	u := fmt.Sprintf("%s/api/v1/orgs/%s/repos", c.BaseURL, org)
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("gitea: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &GiteaAPIError{Class: "network", Message: err.Error(), Cause: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, &GiteaAPIError{
			Class:       giteaErrorClass(resp.StatusCode),
			Message:     strings.TrimSpace(string(errBody)),
			HTTPStatus:  resp.StatusCode,
			Endpoint:    u,
		}
	}
	var repo GiteaRepo
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("gitea: decode response: %w", err)
	}
	return &repo, nil
}

// GiteaAPIError is a typed error from the Gitea API.
type GiteaAPIError struct {
	Class      string // validation | permission | not_found | rate_limit | server | network
	Message    string
	HTTPStatus int
	Endpoint   string
	Cause      error
}

func (e *GiteaAPIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("gitea %s: %s (status=%d endpoint=%s): %v", e.Class, e.Message, e.HTTPStatus, e.Endpoint, e.Cause)
	}
	return fmt.Sprintf("gitea %s: %s (status=%d endpoint=%s)", e.Class, e.Message, e.HTTPStatus, e.Endpoint)
}

func (e *GiteaAPIError) Unwrap() error { return e.Cause }

func giteaErrorClass(status int) string {
	switch {
	case status == 401 || status == 403:
		return "permission"
	case status == 404:
		return "not_found"
	case status == 422:
		return "validation"
	case status == 429:
		return "rate_limit"
	case status >= 500:
		return "server"
	default:
		return "unknown"
	}
}
