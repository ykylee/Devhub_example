package httpapi

import (
	"net/http"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

type rbacRoleResponse struct {
	Role        string `json:"role"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type rbacResourceResponse struct {
	Resource    string `json:"resource"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type rbacPermissionResponse struct {
	Permission  string `json:"permission"`
	Label       string `json:"label"`
	Rank        int    `json:"rank"`
	Description string `json:"description"`
}

type rbacPolicyResponse struct {
	Roles       []rbacRoleResponse           `json:"roles"`
	Resources   []rbacResourceResponse       `json:"resources"`
	Permissions []rbacPermissionResponse     `json:"permissions"`
	Matrix      map[string]map[string]string `json:"matrix"`
}

func (h Handler) getRBACPolicy(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   defaultRBACPolicy(),
		"meta": gin.H{
			"policy_version": "2026-05-07.default",
			"source":         "static_default_policy",
			"editable":       false,
		},
	})
}

func defaultRBACPolicy() rbacPolicyResponse {
	return rbacPolicyResponse{
		Roles: []rbacRoleResponse{
			{
				Role:        string(domain.AppRoleDeveloper),
				Label:       "Developer",
				Description: "개발자 대시보드, 본인 관련 repository/CI/risk 조회 권한",
			},
			{
				Role:        string(domain.AppRoleManager),
				Label:       "Manager",
				Description: "팀 운영, risk triage, 승인 전 command 생성 권한",
			},
			{
				Role:        string(domain.AppRoleSystemAdmin),
				Label:       "System Admin",
				Description: "시스템 설정, 조직/사용자 관리, 운영 command 관리 권한",
			},
		},
		Resources: []rbacResourceResponse{
			{Resource: "repositories", Label: "Repositories", Description: "repository, issue, pull request metadata"},
			{Resource: "ci_runs", Label: "CI Runs", Description: "CI run, build log, deployment state"},
			{Resource: "risks", Label: "Risks", Description: "risk list, mitigation command, decision audit"},
			{Resource: "commands", Label: "Commands", Description: "service action and mitigation command lifecycle"},
			{Resource: "organization", Label: "Organization", Description: "users, org units, memberships"},
			{Resource: "system_config", Label: "System Config", Description: "runtime config, adapters, infrastructure controls"},
		},
		Permissions: []rbacPermissionResponse{
			{Permission: "none", Label: "None", Rank: 0, Description: "접근 불가"},
			{Permission: "read", Label: "Read", Rank: 10, Description: "조회 가능"},
			{Permission: "write", Label: "Write", Rank: 20, Description: "생성/수정 또는 command 생성 가능"},
			{Permission: "admin", Label: "Admin", Rank: 30, Description: "관리 작업과 위험 작업 승인 가능"},
		},
		Matrix: map[string]map[string]string{
			string(domain.AppRoleDeveloper): {
				"repositories":  "read",
				"ci_runs":       "read",
				"risks":         "read",
				"commands":      "none",
				"organization":  "none",
				"system_config": "none",
			},
			string(domain.AppRoleManager): {
				"repositories":  "write",
				"ci_runs":       "read",
				"risks":         "write",
				"commands":      "write",
				"organization":  "read",
				"system_config": "none",
			},
			string(domain.AppRoleSystemAdmin): {
				"repositories":  "admin",
				"ci_runs":       "admin",
				"risks":         "admin",
				"commands":      "admin",
				"organization":  "admin",
				"system_config": "admin",
			},
		},
	}
}
