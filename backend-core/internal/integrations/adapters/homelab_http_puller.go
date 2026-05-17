package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HomeLabHTTPPuller loads HomeLab raw snapshots from an HTTP endpoint.
type HomeLabHTTPPuller struct {
	URL          string
	Token        string
	Client       *http.Client
	RetryMax     int
	RetryBackoff time.Duration
}

func (p HomeLabHTTPPuller) PullSnapshot(ctx context.Context) (HomeLabRawSnapshot, error) {
	endpoint := strings.TrimSpace(p.URL)
	if endpoint == "" {
		return HomeLabRawSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	retryMax := p.RetryMax
	if retryMax < 0 {
		retryMax = 0
	}
	backoff := p.RetryBackoff
	if backoff <= 0 {
		backoff = time.Second
	}

	var lastErr error
	for attempt := 0; attempt <= retryMax; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return HomeLabRawSnapshot{}, fmt.Errorf("build homelab pull request: %w", err)
		}
		if token := strings.TrimSpace(p.Token); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Accept", "application/json")

		raw, retryable, err := doHomeLabHTTPPull(client, req)
		if err == nil {
			return raw, nil
		}
		lastErr = err
		if !retryable || attempt == retryMax {
			break
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return HomeLabRawSnapshot{}, fmt.Errorf("execute homelab pull request: %w", ctx.Err())
		case <-timer.C:
		}
	}
	return HomeLabRawSnapshot{}, lastErr
}

func doHomeLabHTTPPull(client *http.Client, req *http.Request) (HomeLabRawSnapshot, bool, error) {
	resp, err := client.Do(req)
	if err != nil {
		return HomeLabRawSnapshot{}, true, fmt.Errorf("execute homelab pull request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		retryable := resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests
		return HomeLabRawSnapshot{}, retryable, fmt.Errorf("homelab pull request failed: status=%d", resp.StatusCode)
	}

	var raw HomeLabRawSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return HomeLabRawSnapshot{}, true, fmt.Errorf("decode homelab pull response: %w", err)
	}
	if strings.TrimSpace(raw.AgentID) == "" || strings.TrimSpace(raw.SnapshotAt) == "" {
		return HomeLabRawSnapshot{}, false, ErrInvalidHomeLabSnapshot
	}
	if len(raw.Nodes) == 0 && len(raw.Services) == 0 {
		return HomeLabRawSnapshot{}, false, ErrInvalidHomeLabSnapshot
	}
	return raw, false, nil
}
