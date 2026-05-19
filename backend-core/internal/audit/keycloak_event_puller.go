// Package audit — Keycloak event → DevHub audit_logs 통합 (sprint claude/work_260519-u PR-B).
//
// ADR-0019 §5.3 audit event listener carve 의 실 구현. design 문서
// docs/planning/keycloak_event_audit_integration.md.
//
// 책임 분리:
//   - KeycloakAdminClient.ListUserEvents + ListAdminEvents — Keycloak Admin REST API 호출
//     (internal/httpapi/keycloak_admin_client.go, sprint -u PR-B)
//   - EventCursorStore — dedup state 영구화 (internal/store/event_cursors.go, sprint -u PR-B)
//   - KeycloakEventPuller — cron worker (본 파일, sprint -u PR-B skeleton)
//   - main.go wire + Prometheus metric + integration test — 별도 PR-C carve (sprint -v 후보)
//   - audit_logs schema source_type=keycloak_event 정합 — 별도 PR-D carve
//
// HomeLab pull loop + DREQ intake token cron 패턴 정합 — single goroutine + ctx cancellation.
package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/store"
)

// KeycloakEventLister — KeycloakEventPuller 가 의존하는 admin client interface.
// internal/httpapi.KeycloakAdminClient 의 ListUserEvents + ListAdminEvents 부분집합 —
// circular import 회피.
type KeycloakEventLister interface {
	ListUserEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakUserEvent, error)
	ListAdminEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakAdminEvent, error)
}

// KeycloakUserEvent — internal/httpapi.KeycloakUserEvent 의 미러 (circular import 회피).
// main.go wire 시 adapter 로 변환.
type KeycloakUserEvent struct {
	Time     int64
	Type     string
	RealmID  string
	ClientID string
	UserID   string
	IPAddr   string
	Details  map[string]string
	Error    string
}

// KeycloakAdminEvent — internal/httpapi.KeycloakAdminEvent 의 미러.
type KeycloakAdminEvent struct {
	Time          int64
	RealmID       string
	OperationType string
	ResourceType  string
	ResourcePath  string
	AuthUserID    string
	AuthClientID  string
	AuthIPAddr    string
	Error         string
}

// AuditEmitter — Keycloak event 1건마다 audit_logs INSERT 콜백. main.go wire 시 주입.
// design 문서 §3.4 audit_logs INSERT 매핑 정합 (action / target_type / target_id / payload 등).
// sourceEventID 는 puller 의 SHA256 dedup hash — store layer 가 partial UNIQUE INDEX
// (migration 000032, sprint -w PR-D) 로 중복 INSERT 차단 (at-least-once 가드).
type AuditEmitter func(ctx context.Context, action, targetType, targetID, sourceEventID string, payload map[string]any)

// KeycloakEventPullerOptions — cron loop 설정.
type KeycloakEventPullerOptions struct {
	// Interval — tick 주기. default 30s.
	Interval time.Duration
	// MaxEvents — 매 poll 의 page size. default 500.
	MaxEvents int
	// SkipUserEventTypes — 처리 skip 할 event type list (잡음 회피).
	// design §4.3 default = REFRESH_TOKEN / CODE_TO_TOKEN / INTROSPECT_TOKEN.
	SkipUserEventTypes map[string]bool
	// AuditEmitter — nil 이면 audit 생략 (테스트 / dry-run).
	AuditEmitter AuditEmitter
	// Now — 시간 주입 (테스트).
	Now func() time.Time
}

// userEventsCursor / adminEventsCursor — event_cursors table 의 cursor_key.
const (
	userEventsCursor  = "keycloak.events"
	adminEventsCursor = "keycloak.events.admin"
)

// RunKeycloakEventPuller — Keycloak event polling cron loop. ctx cancel 까지 실행.
// design 문서 §3.2 cron worker 패턴 정합 (HomeLab pull / DREQ intake token cron 정합).
func RunKeycloakEventPuller(
	ctx context.Context,
	lister KeycloakEventLister,
	cursors store.EventCursorStore,
	opts KeycloakEventPullerOptions,
) error {
	if opts.Interval <= 0 {
		opts.Interval = 30 * time.Second
	}
	if opts.MaxEvents <= 0 {
		opts.MaxEvents = 500
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	if opts.SkipUserEventTypes == nil {
		opts.SkipUserEventTypes = defaultSkipUserEventTypes()
	}
	InitMetrics()

	log.Printf("[keycloak-event-puller] starting (interval=%s, max_events=%d)", opts.Interval, opts.MaxEvents)

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	// 즉시 1차 tick (운영자가 startup 직후 metric 즉시 확인 — DREQ cron 패턴 정합).
	if err := pullOnce(ctx, lister, cursors, opts, now); err != nil {
		log.Printf("[keycloak-event-puller] initial tick: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Printf("[keycloak-event-puller] stopping: %v", ctx.Err())
			return ctx.Err()
		case <-ticker.C:
			if err := pullOnce(ctx, lister, cursors, opts, now); err != nil {
				log.Printf("[keycloak-event-puller] tick: %v", err)
			}
		}
	}
}

// pullOnce — 1회 poll. user events + admin events 각각 처리.
// 두 branch 가 독립이라 한쪽 error 가 다른 쪽 emit 을 막지 않도록 모두 시도 후 첫 error 보고
// (cursor lag dashboard 가 한 cursor 만 stuck 인 케이스를 식별할 수 있게).
func pullOnce(
	ctx context.Context,
	lister KeycloakEventLister,
	cursors store.EventCursorStore,
	opts KeycloakEventPullerOptions,
	now func() time.Time,
) error {
	var firstErr error
	if err := pullUserEvents(ctx, lister, cursors, opts, now); err != nil {
		ObservePullError("user")
		firstErr = fmt.Errorf("user events: %w", err)
	}
	if err := pullAdminEvents(ctx, lister, cursors, opts, now); err != nil {
		ObservePullError("admin")
		if firstErr == nil {
			firstErr = fmt.Errorf("admin events: %w", err)
		}
	}
	return firstErr
}

// pullUserEvents — Keycloak user events 1회 poll + dedup + audit emit + cursor 갱신.
func pullUserEvents(
	ctx context.Context,
	lister KeycloakEventLister,
	cursors store.EventCursorStore,
	opts KeycloakEventPullerOptions,
	now func() time.Time,
) error {
	cursor, err := loadCursor(ctx, cursors, userEventsCursor, now)
	if err != nil {
		return err
	}
	events, err := lister.ListUserEvents(ctx, cursor.LastEventAt, opts.MaxEvents)
	if err != nil {
		return fmt.Errorf("list user events: %w", err)
	}

	var latestTime time.Time
	var latestHash string
	for _, ev := range events {
		evTime := time.UnixMilli(ev.Time).UTC()
		// 명백히 cursor 이전인 event 는 이미 처리됨. 동일 ms boundary event 는
		// hash dedup 으로 처리 — Keycloak dateFrom 은 inclusive 라 직전 처리 이벤트가
		// 다음 poll 에 반복 등장 (design §3.3 at-least-once dedup).
		if evTime.Before(cursor.LastEventAt) {
			continue
		}
		if opts.SkipUserEventTypes[ev.Type] {
			continue
		}
		evHash := hashUserEvent(ev)
		if evTime.Equal(cursor.LastEventAt) && evHash == cursor.LastEventHash {
			continue
		}
		action, targetType, targetID := mapUserEventToAudit(ev)
		if opts.AuditEmitter != nil {
			opts.AuditEmitter(ctx, action, targetType, targetID, evHash, userEventPayload(ev))
		}
		ObserveEventProcessed("user", action)
		if evTime.After(latestTime) {
			latestTime = evTime
			latestHash = evHash
		}
	}

	if !latestTime.IsZero() {
		if err := cursors.UpsertEventCursor(ctx, store.EventCursor{
			CursorKey:     userEventsCursor,
			LastEventAt:   latestTime,
			LastEventHash: latestHash,
		}); err != nil {
			return fmt.Errorf("upsert cursor %s: %w", userEventsCursor, err)
		}
		ObserveCursorLag(userEventsCursor, now().Sub(latestTime).Seconds())
	} else {
		// no advance — caller 의 dashboard 가 "tick 발생 + 신규 event 없음" 도 가시화하도록 cursor 의 lag emit
		ObserveCursorLag(userEventsCursor, now().Sub(cursor.LastEventAt).Seconds())
	}
	return nil
}

// pullAdminEvents — Keycloak admin events 1회 poll.
func pullAdminEvents(
	ctx context.Context,
	lister KeycloakEventLister,
	cursors store.EventCursorStore,
	opts KeycloakEventPullerOptions,
	now func() time.Time,
) error {
	cursor, err := loadCursor(ctx, cursors, adminEventsCursor, now)
	if err != nil {
		return err
	}
	events, err := lister.ListAdminEvents(ctx, cursor.LastEventAt, opts.MaxEvents)
	if err != nil {
		return fmt.Errorf("list admin events: %w", err)
	}

	var latestTime time.Time
	var latestHash string
	for _, ev := range events {
		evTime := time.UnixMilli(ev.Time).UTC()
		// pullUserEvents 과 동일 boundary 처리 (design §3.3).
		if evTime.Before(cursor.LastEventAt) {
			continue
		}
		evHash := hashAdminEvent(ev)
		if evTime.Equal(cursor.LastEventAt) && evHash == cursor.LastEventHash {
			continue
		}
		action, targetType, targetID := mapAdminEventToAudit(ev)
		if opts.AuditEmitter != nil {
			opts.AuditEmitter(ctx, action, targetType, targetID, evHash, adminEventPayload(ev))
		}
		ObserveEventProcessed("admin", action)
		if evTime.After(latestTime) {
			latestTime = evTime
			latestHash = evHash
		}
	}

	if !latestTime.IsZero() {
		if err := cursors.UpsertEventCursor(ctx, store.EventCursor{
			CursorKey:     adminEventsCursor,
			LastEventAt:   latestTime,
			LastEventHash: latestHash,
		}); err != nil {
			return fmt.Errorf("upsert cursor %s: %w", adminEventsCursor, err)
		}
		ObserveCursorLag(adminEventsCursor, now().Sub(latestTime).Seconds())
	} else {
		ObserveCursorLag(adminEventsCursor, now().Sub(cursor.LastEventAt).Seconds())
	}
	return nil
}

// loadCursor — event_cursors 에서 cursor 조회. row 없으면 now 로 초기화 (첫 run 시
// 과거 event 모두 폭격 회피 — design §3.3).
func loadCursor(ctx context.Context, cursors store.EventCursorStore, key string, now func() time.Time) (store.EventCursor, error) {
	c, err := cursors.GetEventCursor(ctx, key)
	if err == nil {
		return c, nil
	}
	if isNotFound(err) {
		return store.EventCursor{
			CursorKey:   key,
			LastEventAt: now().UTC(),
		}, nil
	}
	return store.EventCursor{}, fmt.Errorf("load cursor %s: %w", key, err)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}

// hashUserEvent / hashAdminEvent — design §3.3 dedup hash (timestamp + type + userId).
// SHA256(time|type|userId|ipAddr) — 동시 다발 event 의 중복 회피.
func hashUserEvent(ev KeycloakUserEvent) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s|%s", ev.Time, ev.Type, ev.UserID, ev.IPAddr)))
	return hex.EncodeToString(h[:])
}

func hashAdminEvent(ev KeycloakAdminEvent) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d|%s:%s|%s|%s", ev.Time, ev.ResourceType, ev.OperationType, ev.ResourcePath, ev.AuthUserID)))
	return hex.EncodeToString(h[:])
}

// defaultSkipUserEventTypes — design §4.3 default skip list. operator 가 env override 가능.
func defaultSkipUserEventTypes() map[string]bool {
	return map[string]bool{
		"REFRESH_TOKEN":    true,
		"CODE_TO_TOKEN":    true,
		"INTROSPECT_TOKEN": true,
	}
}

// mapUserEventToAudit — design §4.1 매핑 표 적용 (15 row). 미매핑 type 은
// `keycloak.event.unknown:<TYPE>` fallback (운영 감시).
func mapUserEventToAudit(ev KeycloakUserEvent) (action, targetType, targetID string) {
	targetID = ev.UserID
	switch ev.Type {
	case "LOGIN":
		return "auth.login.success", "auth", targetID
	case "LOGIN_ERROR":
		return "auth.login.failed", "auth", targetID
	case "LOGOUT":
		return "auth.logout.success", "auth", targetID
	case "LOGOUT_ERROR":
		return "auth.logout.failed", "auth", targetID
	case "REGISTER":
		return "auth.signup.success", "user", targetID
	case "REGISTER_ERROR":
		return "auth.signup.failed", "user", targetID
	case "UPDATE_PASSWORD":
		return "auth.password.changed", "user", targetID
	case "UPDATE_PASSWORD_ERROR":
		return "auth.password.change_failed", "user", targetID
	case "SEND_RESET_PASSWORD":
		return "auth.password.reset_requested", "user", targetID
	case "RESET_PASSWORD":
		return "auth.password.reset_success", "user", targetID
	case "IDENTITY_PROVIDER_LINK_ACCOUNT":
		return "auth.idp.linked", "user", targetID
	case "IDENTITY_PROVIDER_FIRST_LOGIN":
		return "auth.idp.first_login", "user", targetID
	case "VERIFY_EMAIL":
		return "auth.email.verified", "user", targetID
	case "REMOVE_TOTP":
		return "auth.mfa.totp_removed", "user", targetID
	case "UPDATE_TOTP":
		return "auth.mfa.totp_updated", "user", targetID
	default:
		return "keycloak.event.unknown:" + ev.Type, "auth", targetID
	}
}

// mapAdminEventToAudit — design §4.2 매핑 표 (7 row). admin event 의 ResourceType +
// OperationType 조합으로 action 결정.
func mapAdminEventToAudit(ev KeycloakAdminEvent) (action, targetType, targetID string) {
	targetID = ev.ResourcePath
	key := ev.ResourceType + ":" + ev.OperationType
	switch key {
	case "USER:CREATE":
		return "keycloak.user.created", "user", targetID
	case "USER:UPDATE":
		return "keycloak.user.updated", "user", targetID
	case "USER:DELETE":
		return "keycloak.user.deleted", "user", targetID
	case "USER:ACTION":
		return "keycloak.user.action", "user", targetID
	case "REALM_ROLE_MAPPING:CREATE":
		return "keycloak.user.role.granted", "user", targetID
	case "REALM_ROLE_MAPPING:DELETE":
		return "keycloak.user.role.revoked", "user", targetID
	case "CLIENT:UPDATE":
		return "keycloak.client.updated", "client", targetID
	case "REALM:UPDATE":
		return "keycloak.realm.updated", "realm", targetID
	default:
		return "keycloak.admin." + strings.ToLower(key), "realm", targetID
	}
}

func userEventPayload(ev KeycloakUserEvent) map[string]any {
	p := map[string]any{
		"keycloak_event_type": ev.Type,
		"client_id":           ev.ClientID,
		"realm_id":            ev.RealmID,
		"user_id":             ev.UserID,
		"ip_address":          ev.IPAddr,
	}
	if ev.Error != "" {
		p["error"] = ev.Error
	}
	if sessionID, ok := ev.Details["sessionId"]; ok {
		p["session_id"] = sessionID
	}
	return p
}

func adminEventPayload(ev KeycloakAdminEvent) map[string]any {
	p := map[string]any{
		"resource_type":  ev.ResourceType,
		"operation_type": ev.OperationType,
		"resource_path":  ev.ResourcePath,
		"realm_id":       ev.RealmID,
		"auth_user_id":   ev.AuthUserID,
		"auth_client_id": ev.AuthClientID,
		"ip_address":     ev.AuthIPAddr,
	}
	if ev.Error != "" {
		p["error"] = ev.Error
	}
	return p
}
