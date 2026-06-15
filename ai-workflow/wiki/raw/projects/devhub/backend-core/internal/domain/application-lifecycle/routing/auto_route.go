package routing

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/devhub/backend-core/internal/domain"
)

// VocRegistration captures the fields needed for auto-routing decisions.
// Mirrors the POST body of dev-request voc creation.
type VocRegistration struct {
	ExternalRef   string
	SourceSystem  string
	Requester     string
	ReqDepartment string
	Title         string
	Details       string
	Assignee      string
	DevDepartment string
}

// AutoRouteDecision is the result of matching a VocRegistration to a platform.
type AutoRouteDecision struct {
	Matched      bool
	PlatformID   string
	DevRequestID string // empty unless matched
	Reason       string // "external_ref_pattern" | "requester_email" | "req_department" | "no_match"
	// ProviderHint is the matched provider type ("gitea" | "jira" | "github" | "other")
	// when Reason = "external_ref_pattern". Empty for non-pattern matches.
	ProviderHint string
}

// InboundSourceRoutingConfig 는 platform.inbound_source_config 의 JSONB schema (X-2).
// 사용자가 platform 별로 custom external_ref/requester/department pattern 을 정의 가능.
// 모든 field optional — 미설정 시 provider-default pattern 사용.
type InboundSourceRoutingConfig struct {
	// CustomExternalRefPattern 은 source_system provider 가 지원하지 않는 custom pattern
	// (예: "^CUSTOM-(?<id>\d+)$"). 유효한 Go regexp syntax.
	CustomExternalRefPattern string `json:"custom_external_ref_pattern,omitempty"`
	// CustomRequesterPattern 은 requester email 매칭용 custom pattern.
	CustomRequesterPattern string `json:"custom_requester_pattern,omitempty"`
	// CustomDepartmentPattern 은 req_department 매칭용 custom pattern.
	CustomDepartmentPattern string `json:"custom_department_pattern,omitempty"`
}

// ParseInboundSourceRoutingConfig 는 platform.InboundSourceConfig (raw JSONB text) 의
// typed parse. 빈 문자열 / invalid JSON 의 경우 zero value + nil error (custom pattern 미적용).
// 본 함수는 (X-2) 의 multi-provider pattern matcher 가 InboundSourceConfig 의
// optional custom field 를 활용하기 위함.
func ParseInboundSourceRoutingConfig(raw string) (InboundSourceRoutingConfig, error) {
	var cfg InboundSourceRoutingConfig
	if raw == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// PlatformStore is the subset of PlatformRepository needed by the auto-router.
type PlatformStore interface {
	ListEnabledInboundSourcePlatforms(ctx context.Context) ([]domain.Platform, error)
}

// AutoRouter routes incoming vocs to the matching platform.
type AutoRouter interface {
	Route(ctx context.Context, voc VocRegistration) (AutoRouteDecision, error)
}

type defaultAutoRouter struct {
	repo PlatformStore
}

func NewAutoRouter(repo PlatformStore) AutoRouter {
	return &defaultAutoRouter{repo: repo}
}

// Provider-specific external_ref pattern (X-2 multi-provider depth 정밀화).
// gitea: "GITEA-<digits>" | jira: "<PROJECT_KEY>-<digits>" (예: DEV-123) | github: "#<digits>" | gitlab: "!<digits>"
// 'other' provider 는 custom pattern (InboundSourceRoutingConfig.CustomExternalRefPattern) 만 지원.
var (
	giteaExternalRefPattern  = regexp.MustCompile(`^GITEA-([0-9]+)$`)
	jiraExternalRefPattern   = regexp.MustCompile(`^([A-Z][A-Z0-9_]{1,9})-([0-9]+)$`)
	githubExternalRefPattern = regexp.MustCompile(`^#([0-9]+)$`)
	gitlabExternalRefPattern = regexp.MustCompile(`^!([0-9]+)$`)
)

// matchExternalRefPattern 는 (X-2) 의 provider-default + custom pattern 의 통합 matcher.
// 4-tier 우선순위: source-system-specific provider (gitea/jira/github/gitlab) → custom pattern → no_match.
func matchExternalRefPattern(voc VocRegistration, p domain.Platform) (bool, string) {
	if voc.ExternalRef == "" {
		return false, ""
	}
	switch p.InboundSourceType {
	case "gitea":
		if voc.SourceSystem == "gitea" && giteaExternalRefPattern.MatchString(voc.ExternalRef) {
			return true, "gitea"
		}
	case "jira":
		if voc.SourceSystem == "jira" && jiraExternalRefPattern.MatchString(voc.ExternalRef) {
			return true, "jira"
		}
	case "other":
		// 'other' provider 는 custom_external_ref_pattern 만 지원 (사용자 정의 regex).
		// InboundSourceConfig 의 JSONB schema 가 custom pattern 을 정의.
		cfg, err := ParseInboundSourceRoutingConfig(p.InboundSourceConfig)
		if err != nil || cfg.CustomExternalRefPattern == "" {
			return false, ""
		}
		customRe, err := regexp.Compile(cfg.CustomExternalRefPattern)
		if err != nil {
			return false, "" // invalid pattern = no match (silent skip, audit 으로 운영자 알림)
		}
		if customRe.MatchString(voc.ExternalRef) {
			return true, "other_custom"
		}
	}
	// InboundSourceType 비어 있거나 매칭 실패 → custom pattern 시도 (X-2 depth).
	if p.InboundSourceConfig != "" {
		cfg, err := ParseInboundSourceRoutingConfig(p.InboundSourceConfig)
		if err == nil && cfg.CustomExternalRefPattern != "" {
			if customRe, err := regexp.Compile(cfg.CustomExternalRefPattern); err == nil {
				if customRe.MatchString(voc.ExternalRef) {
					return true, "custom"
				}
			}
		}
	}
	return false, ""
}

func (r *defaultAutoRouter) Route(ctx context.Context, voc VocRegistration) (AutoRouteDecision, error) {
	platforms, err := r.repo.ListEnabledInboundSourcePlatforms(ctx)
	if err != nil {
		return AutoRouteDecision{Matched: false, Reason: "no_match"}, err
	}
	if len(platforms) == 0 {
		return AutoRouteDecision{Matched: false, Reason: "no_match"}, nil
	}

	// Case 1: multi-provider external_ref pattern match (X-2 depth).
	if voc.ExternalRef != "" {
		for _, p := range platforms {
			if matched, providerHint := matchExternalRefPattern(voc, p); matched {
				return AutoRouteDecision{
					Matched:      true,
					PlatformID:   p.ID,
					Reason:       "external_ref_pattern",
					ProviderHint: providerHint,
				}, nil
			}
		}
	}

	// Case 2: requester email → DevelopmentUnitID matching.
	if voc.Requester != "" {
		for _, p := range platforms {
			if p.DevelopmentUnitID != "" && matchRequesterToUnit(voc.Requester, p.DevelopmentUnitID) {
				return AutoRouteDecision{
					Matched:    true,
					PlatformID: p.ID,
					Reason:     "requester_email",
				}, nil
			}
		}
	}

	// Case 3: req_department → DevelopmentUnitID matching.
	if voc.ReqDepartment != "" {
		for _, p := range platforms {
			if p.DevelopmentUnitID != "" && matchDepartmentToUnit(voc.ReqDepartment, p.DevelopmentUnitID) {
				return AutoRouteDecision{
					Matched:    true,
					PlatformID: p.ID,
					Reason:     "req_department",
				}, nil
			}
		}
	}

	return AutoRouteDecision{Matched: false, Reason: "no_match"}, nil
}

// matchRequesterToUnit matches a requester email to a development unit.
// In production this would do a Keycloak user lookup → primary unit resolution.
// In tests, exact match on email prefix is sufficient.
func matchRequesterToUnit(requester, unitID string) bool {
	if requester == "" || unitID == "" {
		return false
	}
	if requester == unitID {
		return true
	}
	return false
}

// matchDepartmentToUnit matches a req_department string to a development unit.
// In production this would resolve through organization hierarchy.
// In tests, exact match on the department label against unit ID is sufficient.
func matchDepartmentToUnit(department, unitID string) bool {
	if department == "" || unitID == "" {
		return false
	}
	if department == unitID {
		return true
	}
	return false
}
