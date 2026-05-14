package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Application/Repository/Project 관리 API handler stubs (API-43..50, planned 활성화).
//
// 본 sprint (claude/work_260514-a) 의 carve in 은 handler stub + RBAC gate 등록까지.
// 실 store 호출과 응답 body 는 후속 sprint 의 carve out (state.json carve_out 참조).
// 현재는 모두 `501 not_implemented` envelope 응답 — 단, RBAC matrix 가 호출자 role 을
// 먼저 거부하므로 (ADR-0011 §4.1 의 system_admin 일임) developer/manager 는 `403` 을
//받고 system_admin 만 `501` 까지 도달한다.

func (h *Handler) notImplemented(c *gin.Context, codeHint string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"status": "error",
		"error": gin.H{
			"code":    "not_implemented",
			"message": "endpoint scaffolded; PostgreSQL store body and request validation pending (sprint claude/work_260514-a carve out).",
			"hint":    codeHint,
		},
	})
}

// SCM Providers (API-41, API-42)

func (h *Handler) listSCMProviders(c *gin.Context) {
	h.notImplemented(c, "API-41")
}

func (h *Handler) updateSCMProvider(c *gin.Context) {
	h.notImplemented(c, "API-42")
}

// Applications (API-43..47)

func (h *Handler) listApplications(c *gin.Context) {
	h.notImplemented(c, "API-43")
}

func (h *Handler) createApplication(c *gin.Context) {
	h.notImplemented(c, "API-44")
}

func (h *Handler) getApplication(c *gin.Context) {
	h.notImplemented(c, "API-45")
}

func (h *Handler) updateApplication(c *gin.Context) {
	h.notImplemented(c, "API-46")
}

func (h *Handler) archiveApplication(c *gin.Context) {
	h.notImplemented(c, "API-47")
}

// Application-Repository link (API-48..50)

func (h *Handler) listApplicationRepositories(c *gin.Context) {
	h.notImplemented(c, "API-48")
}

func (h *Handler) createApplicationRepository(c *gin.Context) {
	h.notImplemented(c, "API-49")
}

func (h *Handler) deleteApplicationRepository(c *gin.Context) {
	h.notImplemented(c, "API-50")
}
