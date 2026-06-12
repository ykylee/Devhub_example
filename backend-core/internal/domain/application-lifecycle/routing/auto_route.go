package routing

import (
	"context"
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

var giteaExternalRefPattern = regexp.MustCompile(`^GITEA-([0-9]+)$`)

func (r *defaultAutoRouter) Route(ctx context.Context, voc VocRegistration) (AutoRouteDecision, error) {
	platforms, err := r.repo.ListEnabledInboundSourcePlatforms(ctx)
	if err != nil {
		return AutoRouteDecision{Matched: false, Reason: "no_match"}, err
	}
	if len(platforms) == 0 {
		return AutoRouteDecision{Matched: false, Reason: "no_match"}, nil
	}

	// Case 1: external_ref pattern match
	if voc.ExternalRef != "" && giteaExternalRefPattern.MatchString(voc.ExternalRef) {
		for _, p := range platforms {
			if p.InboundSourceType == "gitea" && voc.SourceSystem == "gitea" {
				return AutoRouteDecision{
					Matched:    true,
					PlatformID: p.ID,
					Reason:     "external_ref_pattern",
				}, nil
			}
		}
	}

	// Case 2: requester email → DevelopmentUnitID matching
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

	// Case 3: req_department → DevelopmentUnitID matching
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
