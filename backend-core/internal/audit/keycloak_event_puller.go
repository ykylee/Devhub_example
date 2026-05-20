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
	// UserSync — ADR-0020 sub-carve C (sprint -k, issue #212). nil 이면 sync
	// 생략 (audit 만 emit). USER:UPDATE / USER:DELETE / GROUP_MEMBERSHIP
	// CREATE/DELETE 처리 시 호출. main.go 가 user_sync.go 의 SyncUserProfile /
	// SyncUserMembership / MarkUserDeactivated 로 wire.
	UserSync UserSyncCallback
	// Now — 시간 주입 (테스트).
	Now func() time.Time
}

// UserSyncCallback — admin event 처리 시 DevHub `users` 컬럼 sync 위한
// callback. action 은 SyncUserAction (`profile` / `membership` / `status`),
// identityID 는 Keycloak user UUID (ResourcePath 에서 파싱), userIDHint 는
// 가능한 경우 username (USER:DELETE 시 caller 가 cache 한 값, 없으면 빈
// 문자열). 본 callback 안에서 error 처리 (metric / log) 는 caller 책임.
type UserSyncCallback func(ctx context.Context, action SyncUserAction, identityID, userIDHint string)

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
	var latestEmittable bool // sprint -aa codex hotfix #11 — same-ms boundary 에서 emit-able 우선
	for _, ev := range events {
		evTime := time.UnixMilli(ev.Time).UTC()
		// 명백히 cursor 이전인 event 는 이미 처리됨. 동일 ms boundary event 는
		// hash dedup 으로 처리 — Keycloak dateFrom 은 inclusive 라 직전 처리 이벤트가
		// 다음 poll 에 반복 등장 (design §3.3 at-least-once dedup).
		if evTime.Before(cursor.LastEventAt) {
			continue
		}
		evHash := hashUserEvent(ev)
		if evTime.Equal(cursor.LastEventAt) && evHash == cursor.LastEventHash {
			continue
		}
		isSkip := opts.SkipUserEventTypes[ev.Type]
		// cursor 는 skip type 까지 포함해 advance — skip-only page (REFRESH_TOKEN 등으로
		// 가득 찬 페이지) 에서도 cursor 가 멈추지 않게 (sprint -y codex hotfix #10 P1-A).
		// latestHash 는 emit-able event 의 hash 를 우선 — same-ms 에 skip+emit 가 함께
		// 들어오면 emit-able event 의 hash 가 cursor.LastEventHash 에 저장되어, 다음 tick
		// 의 boundary dedup 이 정확히 emit-able event 를 차단 (sprint -aa codex hotfix #11
		// P2 정정 — skip event 의 hash 가 latestHash 로 set 되어 emit-able event 가 매 tick
		// re-emit + ObserveEventProcessed counter 누적되던 side effect 해소).
		if evTime.After(latestTime) {
			latestTime = evTime
			latestHash = evHash
			latestEmittable = !isSkip
		} else if evTime.Equal(latestTime) && !isSkip && !latestEmittable {
			// same-ms 에서 emit-able 가 처음 등장 → skip event 의 hash 를 overwrite.
			latestHash = evHash
			latestEmittable = true
		}
		if isSkip {
			continue
		}
		action, targetType, targetID := mapUserEventToAudit(ev)
		if opts.AuditEmitter != nil {
			opts.AuditEmitter(ctx, action, targetType, targetID, evHash, userEventPayload(ev))
		}
		ObserveEventProcessed("user", action)
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
		// ADR-0020 sub-carve C (sprint -k, issue #212) — DevHub `users` 컬럼
		// sync. UserSync callback nil 이면 audit 만 emit (이전 sprint -u~-y 동작
		// 동등 — backward compatible). callback 안에서 error / metric 처리.
		if opts.UserSync != nil {
			if syncAction, identityID, userIDHint := classifyAdminEventForSync(ev); syncAction != "" && identityID != "" {
				opts.UserSync(ctx, syncAction, identityID, userIDHint)
			}
		}
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
// 과거 event 모두 폭격 회피 — design §3.3) **+ 즉시 UPSERT 로 영구화**.
//
// 영구화 이유 (sprint -y codex hotfix #10 P1-B 정정): in-memory 초기화만 하면 첫
// poll 이 빈 결과인 경우 latestTime zero → upsert path 진입 안 함 → 다음 tick 이 또
// 새 now() 로 reinit → 두 tick 사이 발생한 Keycloak event 는 dateFrom 이 이미
// 지나간 시각이라 응답에서 제외되어 영구 누락. 따라서 row 가 없으면 즉시 seed
// upsert 후 반환.
func loadCursor(ctx context.Context, cursors store.EventCursorStore, key string, now func() time.Time) (store.EventCursor, error) {
	c, err := cursors.GetEventCursor(ctx, key)
	if err == nil {
		return c, nil
	}
	if isNotFound(err) {
		seed := store.EventCursor{
			CursorKey:   key,
			LastEventAt: now().UTC(),
		}
		if upsertErr := cursors.UpsertEventCursor(ctx, seed); upsertErr != nil {
			return store.EventCursor{}, fmt.Errorf("seed cursor %s on first-run init: %w", key, upsertErr)
		}
		return seed, nil
	}
	return store.EventCursor{}, fmt.Errorf("load cursor %s: %w", key, err)
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}

// hashUserEvent / hashAdminEvent — design §3.3 dedup hash. store-level partial UNIQUE
// INDEX (migration 000032) 의 source_event_id 로도 사용되므로 distinguishing 필드 모두
// 포함 (sprint -y codex hotfix #10 P2-D — burst 동시 ms event 의 client/realm 차이를
// hash 가 반영해 audit 손실 회피).
func hashUserEvent(ev KeycloakUserEvent) string {
	// time + type + userId + ipAddr + clientId + realmId + sessionId (details)
	// — burst LOGIN 시 (같은 ms, 같은 user, 다른 client) 의 distinct event 식별.
	sessionID := ev.Details["sessionId"]
	h := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s|%s|%s|%s|%s",
		ev.Time, ev.Type, ev.UserID, ev.IPAddr, ev.ClientID, ev.RealmID, sessionID)))
	return hex.EncodeToString(h[:])
}

func hashAdminEvent(ev KeycloakAdminEvent) string {
	// time + resourceType:operationType + resourcePath + authUserID + authClientID
	// + authIPAddr + realmID — burst admin action 의 distinct event 식별.
	h := sha256.Sum256([]byte(fmt.Sprintf("%d|%s:%s|%s|%s|%s|%s|%s",
		ev.Time, ev.ResourceType, ev.OperationType, ev.ResourcePath,
		ev.AuthUserID, ev.AuthClientID, ev.AuthIPAddr, ev.RealmID)))
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

// classifyAdminEventForSync — ADR-0020 sub-carve C (sprint -k, issue #212).
// admin event 를 검사해 (1) DevHub `users` 컬럼 sync 가 필요한지, (2) 어떤
// SyncUserAction 인지, (3) identity_id 와 (가능하면) username hint 를 반환한다.
//
// 분류:
//   USER:UPDATE → SyncActionProfile (email/display_name/status sync)
//   USER:DELETE → SyncActionStatus (soft delete, users.status=deactivated)
//   GROUP_MEMBERSHIP:CREATE/DELETE → SyncActionMembership (users.role 재계산)
//   그 외 → ("", "", "")  noop
//
// identity_id 는 ResourcePath 에서 파싱. username hint 는 현재 admin event
// payload 에 없으므로 빈 문자열 (USER:DELETE 시 admin client 가 user lookup
// 불가능 — caller 가 별도 cache 또는 best-effort).
func classifyAdminEventForSync(ev KeycloakAdminEvent) (SyncUserAction, string, string) {
	key := ev.ResourceType + ":" + ev.OperationType
	identityID := ParseIdentityIDFromResourcePath(ev.ResourcePath)
	switch key {
	case "USER:UPDATE":
		return SyncActionProfile, identityID, ""
	case "USER:DELETE":
		return SyncActionStatus, identityID, ""
	case "GROUP_MEMBERSHIP:CREATE", "GROUP_MEMBERSHIP:DELETE":
		return SyncActionMembership, identityID, ""
	}
	return "", "", ""
}

// mapAdminEventToAudit — design §4.2 매핑 표. admin event 의 ResourceType +
// OperationType 조합으로 action 결정. ADR-0020 §5.3.1 (sprint -k) 가 GROUP_MEMBERSHIP
// CREATE/DELETE row 신규 추가 — group composite role 매핑 (sub-carve C 의 핵심
// 시그널).
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
	// ADR-0020 sub-carve C (sprint -k, issue #212) — group composite role 매핑
	// (keycloak_groups_rbac_mapping.md sprint -f 권장 B). GROUP_MEMBERSHIP event 는
	// group ↔ user 관계 변경. devhub-{role}s group 가입/탈퇴가 곧 role 변경 신호.
	case "GROUP_MEMBERSHIP:CREATE":
		return "keycloak.user.group.joined", "user", targetID
	case "GROUP_MEMBERSHIP:DELETE":
		return "keycloak.user.group.left", "user", targetID
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
