package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// HomeLabFilePuller loads a HomeLab raw snapshot from a local JSON file.
//
// MaxBytes — payload size limit (sprint claude/work_260518-p, ADR-0015 §6 (1)).
// 0 또는 음수 면 unlimited (legacy behavior). 운영 권장 5 MB (5_242_880).
// 사전 os.Stat 검사 + os.Open + json.NewDecoder streaming 으로 in-memory full
// decode 회피.
type HomeLabFilePuller struct {
	Path     string
	MaxBytes int64
}

func (p HomeLabFilePuller) PullSnapshot(_ context.Context) (HomeLabRawSnapshot, error) {
	path := strings.TrimSpace(p.Path)
	if path == "" {
		return HomeLabRawSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	info, err := os.Stat(path)
	if err != nil {
		return HomeLabRawSnapshot{}, fmt.Errorf("stat homelab pull file: %w", err)
	}
	if p.MaxBytes > 0 && info.Size() > p.MaxBytes {
		// 운영 가드 — oversized payload 가 메모리 압박/DoS 후보. malformed 와
		// 같은 invalid 의미로 분류 (운영자가 size 확장 또는 fixture 정리 결정).
		return HomeLabRawSnapshot{}, fmt.Errorf("homelab pull file exceeds max bytes (%d > %d): %w", info.Size(), p.MaxBytes, ErrInvalidHomeLabSnapshot)
	}
	f, err := os.Open(path)
	if err != nil {
		return HomeLabRawSnapshot{}, fmt.Errorf("open homelab pull file: %w", err)
	}
	defer f.Close()

	var reader io.Reader = f
	if p.MaxBytes > 0 {
		// stat size 가 limit 이하임을 확인했지만 streaming 단계에도 LimitReader 로
		// 추가 가드 (race: file 이 stat 후 grow 한 경우 — concurrent writer).
		reader = io.LimitReader(f, p.MaxBytes+1)
	}
	dec := json.NewDecoder(reader)
	var raw HomeLabRawSnapshot
	if err := dec.Decode(&raw); err != nil {
		// LimitReader 가 limit 초과를 unexpected EOF / io.ErrUnexpectedEOF 또는
		// decoder error 로 surface — 모두 invalid snapshot 으로 분류.
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return HomeLabRawSnapshot{}, fmt.Errorf("homelab pull file truncated or oversized: %w", ErrInvalidHomeLabSnapshot)
		}
		return HomeLabRawSnapshot{}, fmt.Errorf("decode homelab pull file: %w", err)
	}
	// codex hotfix #8 P2 #3 — Decode 만 호출하면 후행 JSON object (예:
	// `{"agent_id":...}{"extra":...}`) 가 silent 무시. 이전 json.Unmarshal 구현은
	// trailing data 를 거부했으므로 회귀 방지를 위해 명시적 EOF 검증.
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return HomeLabRawSnapshot{}, fmt.Errorf("homelab pull file has trailing data after snapshot: %w", ErrInvalidHomeLabSnapshot)
	}
	if strings.TrimSpace(raw.AgentID) == "" || strings.TrimSpace(raw.SnapshotAt) == "" {
		return HomeLabRawSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	if len(raw.Nodes) == 0 && len(raw.Services) == 0 {
		return HomeLabRawSnapshot{}, ErrInvalidHomeLabSnapshot
	}
	return raw, nil
}
