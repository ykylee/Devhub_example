package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HomeLabHTTPPuller loads HomeLab raw snapshots from an HTTP endpoint.
//
// MaxBytes — response body size limit (sprint claude/work_260518-p, ADR-0015 §6 (1)).
// 0 또는 음수 면 unlimited (legacy behavior). 운영 권장 5 MB (5_242_880).
// Content-Length header 가 있고 limit 초과 시 사전 reject (network IO 절약),
// 본문 streaming 은 io.LimitReader(body, limit+1) 로 cap.
type HomeLabHTTPPuller struct {
	URL          string
	Token        string
	Client       *http.Client
	RetryMax     int
	RetryBackoff time.Duration
	MaxBytes     int64
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

		raw, retryable, err := doHomeLabHTTPPull(client, req, p.MaxBytes)
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

func doHomeLabHTTPPull(client *http.Client, req *http.Request, maxBytes int64) (HomeLabRawSnapshot, bool, error) {
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

	// Size guard (ADR-0015 §6 (1)) — Content-Length 사전 검사로 network IO 절약,
	// LimitReader 로 streaming cap (Content-Length 미제공 또는 거짓 경우 대비).
	// maxBytes = 0 (default) 면 unlimited — legacy behavior.
	if maxBytes > 0 && resp.ContentLength > 0 && resp.ContentLength > maxBytes {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return HomeLabRawSnapshot{}, false, fmt.Errorf("homelab pull response exceeds max bytes (content-length %d > %d): %w", resp.ContentLength, maxBytes, ErrInvalidHomeLabSnapshot)
	}
	var body io.Reader = resp.Body
	if maxBytes > 0 {
		body = io.LimitReader(resp.Body, maxBytes+1)
	}

	var raw HomeLabRawSnapshot
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		// LimitReader 가 cap 초과 시 unexpected EOF — invalid snapshot 으로 분류
		// (retry 불필요 — 같은 body 재시도해도 같은 결과).
		if maxBytes > 0 && errors.Is(err, io.ErrUnexpectedEOF) {
			return HomeLabRawSnapshot{}, false, fmt.Errorf("homelab pull response oversized or truncated: %w", ErrInvalidHomeLabSnapshot)
		}
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
