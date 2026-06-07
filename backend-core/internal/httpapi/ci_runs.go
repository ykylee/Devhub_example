package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// POST /api/v1/ci-runs — sprint mvs/work_260607-h-486-ci-runs-api (N-7 / P0-4).
//
// Issue: #486 — CI Run ingest endpoint (Gitea Actions webhook / direct POST).
// Spec (issue #486 본문 + PR #486 housekeeping comment 의 2차 ID 슬롯 정정):
//   - request body: { repository_id (int64, required), ref (string, required),
//                     status (enum 7종, required),
//                     commit_sha (optional), runner (optional, string),
//                     started_at / finished_at (optional, RFC3339) }
//   - status validation: 위 7종 외 422
//   - repository_id 가드: 존재 검증 (FK violation → 404) + RBAC (best-effort,
//     기존 handler 패턴과 동일하게 route-level auth + handler actor 기반)
//   - idempotency: (repository_id, commit_sha, status, started_at) UNIQUE 충돌
//     시 409 + 기존 row 본문 반환 (Gitea webhook 중복 ingest 안전)
//   - audit emit: ci_run.created (actor, repository_id, status, ref)
//   - metric emit: devhub_ci_runs_total{status} Counter +
//                  devhub_ci_run_ingest_duration_seconds Histogram
//   - 정식 API ID: API-98 (issue #486 housekeeping comment take 2)
//
// 주의: 기존 snapshot.go:280 의 createCIRun (legacy `external_id`+`repository_name`
// 기반) 은 본 sprint 에서 제거. 외부 caller 0 (frontend POST 사용처 0건 확인).
// UpsertCIRun / ciRunLogs GET 은 `external_id` 컬럼 그대로 사용 — 영향 없음.

// validCIStatuses — issue #486 spec 의 7 enum.
var validCIStatuses = map[string]struct{}{
	"queued":    {},
	"running":   {},
	"success":   {},
	"failed":    {},
	"cancelled": {},
	"skipped":   {},
	"unknown":   {},
}

type createCIRunSpecRequest struct {
	RepositoryID int64      `json:"repository_id" binding:"required"`
	Ref          string     `json:"ref" binding:"required"`
	Status       string     `json:"status" binding:"required"`
	CommitSHA    string     `json:"commit_sha,omitempty"`
	Runner       string     `json:"runner,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

func (h Handler) createCIRun(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "failed", "error": "domain store is not configured",
			"code": "create_ci_run.no_store",
		})
		return
	}

	var req createCIRunSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed", "error": err.Error(), "code": "create_ci_run.bind",
		})
		return
	}

	// request body 정합 (strconv / trim / status enum)
	req.Ref = strings.TrimSpace(req.Ref)
	req.CommitSHA = strings.TrimSpace(req.CommitSHA)
	req.Runner = strings.TrimSpace(req.Runner)
	if req.Ref == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "failed", "error": "ref must not be empty", "code": "create_ci_run.ref_empty",
		})
		return
	}
	if _, ok := validCIStatuses[req.Status]; !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "failed",
			"error": fmt.Sprintf("status must be one of: queued, running, success, failed, cancelled, skipped, unknown (got %q)", req.Status),
			"code":  "create_ci_run.status_invalid",
		})
		return
	}
	if req.StartedAt != nil && req.FinishedAt != nil && req.FinishedAt.Before(*req.StartedAt) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"status": "failed", "error": "finished_at must not be before started_at",
			"code": "create_ci_run.time_range",
		})
		return
	}

	// repository_id 가드 — repository 가 실제로 존재하는지 확인. FK constraint
	// 만으로 처리하면 500 으로 새어나가므로, 사전 조회. 조회 실패 / 0 row → 404.
	repo, err := h.lookupRepositoryForCIRun(c.Request.Context(), req.RepositoryID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "not_found", "error": "repository not found",
				"code": "create_ci_run.repo_not_found",
			})
			return
		}
		writeServerError(c, err, "create_ci_run.repo_lookup")
		return
	}

	// RBAC best-effort check — ResourcePipelines:ActionCreate.
	// route-level auth middleware 가 사전 거부하지만, defense-in-depth 로 한 번 더.
	if !h.rbacAllowsCIRunCreate(c) {
		c.JSON(http.StatusForbidden, gin.H{
			"status": "forbidden", "error": "missing ci_run.create permission",
			"code": "create_ci_run.forbidden",
		})
		return
	}

	// external_id 결정. (repo, commit, status, start) 4-tuple 의 SHA256 hash.
	// 동일 4-tuple 재시도 시 동일한 external_id → DB unique (external_id 기존 +
	// 신규 4-col) 둘 다 conflict → ErrConflict → 409 + 기존 row 반환.
	externalID := computeCIRunExternalID(req)

	run := domain.CIRun{
		ExternalID:     externalID,
		RepositoryID:   req.RepositoryID,
		RepositoryName: repo.FullName,
		Branch:         req.Ref,
		CommitSHA:      req.CommitSHA,
		Status:         req.Status,
		StartedAt:      req.StartedAt,
		FinishedAt:     req.FinishedAt,
		Runner:         req.Runner,
	}

	// store.CreateCIRun interface — domain.CIRun 그대로. (signature 변경 없음)
	ciStore, ok := h.cfg.DomainStore.(interface {
		CreateCIRun(context.Context, domain.CIRun) error
		GetCIRunByExternalID(context.Context, string) (domain.CIRun, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "failed", "error": "domain store does not support create/get",
			"code": "create_ci_run.no_method",
		})
		return
	}

	// ingest latency metric — best-effort (panic 무시). handler 전체 duration 측정.
	startTS := time.Now()
	defer func() {
		devhubCIRunIngestDuration.Observe(time.Since(startTS).Seconds())
	}()

	if err := ciStore.CreateCIRun(c.Request.Context(), run); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			// 기존 row 반환. spec: 409 + 기존 row 본문.
			existing, getErr := ciStore.GetCIRunByExternalID(c.Request.Context(), externalID)
			if getErr != nil {
				// existing row 조회 실패 시 409 만 반환 (spec 본문보다 간소화)
				c.JSON(http.StatusConflict, gin.H{
					"status": "conflict", "error": "ci_run already exists",
					"code": "create_ci_run.duplicate",
				})
				return
			}
			c.JSON(http.StatusConflict, gin.H{
				"status": "conflict",
				"error":  "ci_run already exists (idempotency hit)",
				"code":   "create_ci_run.duplicate",
				"data":   ciRunResponseFromDomain(existing),
			})
			return
		case store.IsForeignKeyViolation(err):
			// repository_id 가 사전 조회 후 사라진 race condition. 404.
			c.JSON(http.StatusNotFound, gin.H{
				"status": "not_found", "error": "repository not found (race)",
				"code": "create_ci_run.repo_not_found_race",
			})
			return
		default:
			writeServerError(c, err, "create_ci_run.store")
			return
		}
	}

	// 성공 — 새로 생성된 row 본문 반환
	created, lookupErr := ciStore.GetCIRunByExternalID(c.Request.Context(), externalID)
	if lookupErr != nil {
		// row insert 됐는데 재조회 실패 — 그래도 201 + 핵심 필드 반환
		c.JSON(http.StatusCreated, gin.H{
			"status":         "created",
			"external_id":    externalID,
			"repository_id":  run.RepositoryID,
			"ref":            run.Branch,
			"status_label":   run.Status,
			"commit_sha":     run.CommitSHA,
			"runner":         run.Runner,
			"started_at":     run.StartedAt,
			"finished_at":    run.FinishedAt,
		})
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"status": "created",
			"data":   ciRunResponseFromDomain(created),
		})
	}

	// metric — accepted status
	devhubCIRunsTotal.WithLabelValues(req.Status).Inc()

	// audit — best-effort (recordAuditBestEffort pattern)
	h.recordCIRunAudit(c, req, run, "ci_run.created")
}

// lookupRepositoryForCIRun — repository 존재 검증. domain.DomainStore 가
// ListRepositories 만 있고 단건 조회가 없어서, List 결과에서 ID 매칭.
// 0 row → store.ErrNotFound. 대규모 환경에서는 추후 GetRepositoryByID 추가가
// 자연스러운 다음 sprint. 본 sprint 는 외부 영향 0.
func (h Handler) lookupRepositoryForCIRun(ctx context.Context, repoID int64) (domain.Repository, error) {
	if h.cfg.DomainStore == nil {
		return domain.Repository{}, store.ErrNotFound
	}
	repos, err := h.cfg.DomainStore.ListRepositories(ctx, domain.ListOptions{})
	if err != nil {
		return domain.Repository{}, err
	}
	for _, r := range repos {
		if r.ID == repoID {
			return r, nil
		}
	}
	return domain.Repository{}, store.ErrNotFound
}

// rbacAllowsCIRunCreate — PermissionCache 가 있으면 best-effort 조회.
// cache 가 없거나 nil store 면 true (route-level middleware 가 이미 enforce).
func (h Handler) rbacAllowsCIRunCreate(c *gin.Context) bool {
	if h.cfg.PermissionCache == nil {
		return true
	}
	loginVal, _ := c.Get("devhub_actor_login")
	roleVal, _ := c.Get("devhub_actor_role")
	actorLogin, _ := loginVal.(string)
	actorRole, _ := roleVal.(string)

	// system_admin 은 통과
	if actorRole == string(domain.AppRoleSystemAdmin) {
		return true
	}
	// dev fallback (tests / local) 은 통과
	if v, ok := c.Get("devhub_dev_fallback"); ok {
		if b, _ := v.(bool); b {
			return true
		}
	}

	allowed, err := h.cfg.PermissionCache.Allows(c.Request.Context(), actorLogin, domain.ResourcePipelines, domain.ActionCreate)
	if err != nil {
		// cache lookup 실패 시 deny (안전). 단 audit emit.
		_ = actorLogin
		return false
	}
	return allowed
}

// recordCIRunAudit — best-effort audit emit. AuditStore 가 없거나 nil 이면
// silent skip. recordAuditBestEffort (appview 패턴) 와 동일 의미.
func (h Handler) recordCIRunAudit(c *gin.Context, req createCIRunSpecRequest, run domain.CIRun, action string) {
	_ = req
	if h.cfg.AuditStore == nil {
		return
	}
	loginVal, _ := c.Get("devhub_actor_login")
	actorLogin, _ := loginVal.(string)
	payload := map[string]any{
		"repository_id": run.RepositoryID,
		"ref":           run.Branch,
		"status":        run.Status,
		"commit_sha":    run.CommitSHA,
		"runner":        run.Runner,
	}
	if run.StartedAt != nil {
		payload["started_at"] = run.StartedAt.UTC().Format(time.RFC3339)
	}
	if run.FinishedAt != nil {
		payload["finished_at"] = run.FinishedAt.UTC().Format(time.RFC3339)
	}
	_, _ = h.cfg.AuditStore.CreateAuditLog(c.Request.Context(), domain.AuditLog{
		ActorLogin: actorLogin,
		Action:     action,
		TargetType: "ci_run",
		TargetID:   run.ExternalID,
		Payload:    payload,
		SourceIP:   clientIPFrom(c),
		RequestID:  requestIDFrom(c),
		SourceType: sourceTypeFrom(c),
	})
}

// computeCIRunExternalID — (repo, commit, status, start) 4-tuple hash.
// 같은 tuple 재시도 시 같은 external_id → unique constraint → ErrConflict.
func computeCIRunExternalID(req createCIRunSpecRequest) string {
	startedUnix := int64(0)
	if req.StartedAt != nil {
		startedUnix = req.StartedAt.Unix()
	}
	raw := fmt.Sprintf("%d|%s|%s|%s|%d",
		req.RepositoryID, req.Ref, req.Status, req.CommitSHA, startedUnix,
	)
	sum := sha256.Sum256([]byte(raw))
	// 첫 16 hex (64-bit) 으로 external_id 생성. 사람이 읽기 좋은 prefix
	// "ci-" + 12 hex chars = 16 chars total.
	short := hex.EncodeToString(sum[:8]) // 16 hex chars
	return "ci-" + short[:12]
}

// ciRunResponseFromDomain — domain.CIRun → wire shape. GET /ci-runs response 와
// 같은 shape 으로 통일. duration_seconds, started_at 정규화.
func ciRunResponseFromDomain(run domain.CIRun) gin.H {
	startedAt := time.Time{}
	if run.StartedAt != nil {
		startedAt = *run.StartedAt
	}
	durationSeconds := 0
	if run.DurationSeconds != nil {
		durationSeconds = *run.DurationSeconds
	}
	finishedAt := time.Time{}
	if run.FinishedAt != nil {
		finishedAt = *run.FinishedAt
	}
	return gin.H{
		"id":              run.ID,
		"external_id":     run.ExternalID,
		"repository_id":   run.RepositoryID,
		"repository_name": run.RepositoryName,
		"ref":             run.Branch,
		"commit_sha":      run.CommitSHA,
		"status":          run.Status,
		"runner":          run.Runner,
		"duration_seconds": durationSeconds,
		"started_at":      startedAt.UTC().Format(time.RFC3339Nano),
		"finished_at":     finishedAt.UTC().Format(time.RFC3339Nano),
	}
}