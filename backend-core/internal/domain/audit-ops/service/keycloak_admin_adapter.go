// Package audit 의 internal/httpapi.KeycloakAdminClient → audit.KeycloakEventLister
// adapter (sprint claude/work_260519-v PR-C).
//
// circular import 회피로 audit 패키지가 httpapi 를 import 하지 않게 — adapter 는
// 별도 main.go 측에서 wire 할 수도 있으나, audit 패키지가 httpapi 의 KeycloakAdminClient
// 인터페이스를 의존 없이 받아들이는 방법으로 분리. 본 파일의 KeycloakAdminEventLister
// 는 main.go 가 KeycloakAdminClient 의 ListUserEvents/ListAdminEvents 결과를 audit
// mirror struct 로 변환해 전달하는 adapter wrapper. audit ← httpapi 단방향 의존만.
package service

import (
	"context"
	"time"
)

// HTTPAPIUserEvent — internal/httpapi.KeycloakUserEvent 의 외부 shape (audit 미러로
// 변환할 필드만). main.go adapter 가 httpapi struct → 본 struct 으로 매핑 (필드 이름 동일).
type HTTPAPIUserEvent struct {
	Time     int64
	Type     string
	RealmID  string
	ClientID string
	UserID   string
	IPAddr   string
	Details  map[string]string
	Error    string
}

// HTTPAPIAdminEvent — internal/httpapi.KeycloakAdminEvent 의 외부 shape.
// AuthDetails 는 평탄화된 별도 필드로 전개 (audit mirror 가 평탄 구조이므로).
type HTTPAPIAdminEvent struct {
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

// HTTPAPIEventLister — main.go 가 KeycloakAdminClient 를 wrap 해 본 interface 를
// 충족시킨다 (KeycloakAdminClient.ListUserEvents / ListAdminEvents 가 본 interface 의 시그니처와
// 동일하지만 반환 struct 만 다름 — adapter 가 변환).
type HTTPAPIEventLister interface {
	ListUserEvents(ctx context.Context, dateFrom time.Time, max int) ([]HTTPAPIUserEvent, error)
	ListAdminEvents(ctx context.Context, dateFrom time.Time, max int) ([]HTTPAPIAdminEvent, error)
}

// NewHTTPAPIEventListerAdapter — HTTPAPIEventLister 를 audit.KeycloakEventLister 로
// 변환. main.go wire 시 사용 (sprint -v PR-C).
func NewHTTPAPIEventListerAdapter(httpapiLister HTTPAPIEventLister) KeycloakEventLister {
	return &httpapiEventListerAdapter{src: httpapiLister}
}

type httpapiEventListerAdapter struct {
	src HTTPAPIEventLister
}

func (a *httpapiEventListerAdapter) ListUserEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakUserEvent, error) {
	src, err := a.src.ListUserEvents(ctx, dateFrom, max)
	if err != nil {
		return nil, err
	}
	out := make([]KeycloakUserEvent, len(src))
	for i, ev := range src {
		out[i] = KeycloakUserEvent{
			Time:     ev.Time,
			Type:     ev.Type,
			RealmID:  ev.RealmID,
			ClientID: ev.ClientID,
			UserID:   ev.UserID,
			IPAddr:   ev.IPAddr,
			Details:  ev.Details,
			Error:    ev.Error,
		}
	}
	return out, nil
}

func (a *httpapiEventListerAdapter) ListAdminEvents(ctx context.Context, dateFrom time.Time, max int) ([]KeycloakAdminEvent, error) {
	src, err := a.src.ListAdminEvents(ctx, dateFrom, max)
	if err != nil {
		return nil, err
	}
	out := make([]KeycloakAdminEvent, len(src))
	for i, ev := range src {
		out[i] = KeycloakAdminEvent{
			Time:          ev.Time,
			RealmID:       ev.RealmID,
			OperationType: ev.OperationType,
			ResourceType:  ev.ResourceType,
			ResourcePath:  ev.ResourcePath,
			AuthUserID:    ev.AuthUserID,
			AuthClientID:  ev.AuthClientID,
			AuthIPAddr:    ev.AuthIPAddr,
			Error:         ev.Error,
		}
	}
	return out, nil
}
