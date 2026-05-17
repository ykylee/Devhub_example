package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// HomeLabFilePuller loads a HomeLab raw snapshot from a local JSON file.
type HomeLabFilePuller struct {
	Path string
}

func (p HomeLabFilePuller) PullSnapshot(_ context.Context) (HomeLabRawSnapshot, error) {
	path := strings.TrimSpace(p.Path)
	if path == "" {
		return HomeLabRawSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return HomeLabRawSnapshot{}, fmt.Errorf("read homelab pull file: %w", err)
	}
	var raw HomeLabRawSnapshot
	if err := json.Unmarshal(payload, &raw); err != nil {
		return HomeLabRawSnapshot{}, fmt.Errorf("decode homelab pull file: %w", err)
	}
	if strings.TrimSpace(raw.AgentID) == "" || strings.TrimSpace(raw.SnapshotAt) == "" {
		return HomeLabRawSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	if len(raw.Nodes) == 0 && len(raw.Services) == 0 {
		return HomeLabRawSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	return raw, nil
}
