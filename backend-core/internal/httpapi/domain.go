package httpapi

import (
	"net/http"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

type repositoryResponse struct {
	ID            int64     `json:"id"`
	GiteaID       int64     `json:"gitea_repository_id,omitempty"`
	FullName      string    `json:"full_name"`
	OwnerLogin    string    `json:"owner_login,omitempty"`
	Name          string    `json:"name"`
	CloneURL      string    `json:"clone_url,omitempty"`
	HTMLURL       string    `json:"html_url,omitempty"`
	DefaultBranch string    `json:"default_branch,omitempty"`
	Private       bool      `json:"private"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type issueResponse struct {
	ID             int64      `json:"id"`
	GiteaID        int64      `json:"gitea_issue_id,omitempty"`
	RepositoryName string     `json:"repository_name"`
	Number         int64      `json:"number"`
	Title          string     `json:"title"`
	State          string     `json:"state"`
	AuthorLogin    string     `json:"author_login,omitempty"`
	AssigneeLogin  string     `json:"assignee_login,omitempty"`
	HTMLURL        string     `json:"html_url,omitempty"`
	OpenedAt       *time.Time `json:"opened_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type pullRequestResponse struct {
	ID             int64      `json:"id"`
	GiteaID        int64      `json:"gitea_pull_request_id,omitempty"`
	RepositoryName string     `json:"repository_name"`
	Number         int64      `json:"number"`
	Title          string     `json:"title"`
	State          string     `json:"state"`
	AuthorLogin    string     `json:"author_login,omitempty"`
	HeadBranch     string     `json:"head_branch,omitempty"`
	BaseBranch     string     `json:"base_branch,omitempty"`
	HeadSHA        string     `json:"head_sha,omitempty"`
	HTMLURL        string     `json:"html_url,omitempty"`
	MergedAt       *time.Time `json:"merged_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func riskFromDomain(risk domain.Risk) riskResponse {
	return riskResponse{
		ID:               risk.RiskKey,
		Title:            risk.Title,
		Reason:           risk.Reason,
		Impact:           risk.Impact,
		Status:           risk.Status,
		OwnerLogin:       risk.OwnerLogin,
		SuggestedActions: risk.SuggestedActions,
		CreatedAt:        risk.CreatedAt,
		UpdatedAt:        risk.UpdatedAt,
	}
}

func (h Handler) repositories(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, false)
	if !ok {
		return
	}
	repositories, err := h.cfg.DomainStore.ListRepositories(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_repositories")
		return
	}

	data := make([]repositoryResponse, 0, len(repositories))
	for _, repository := range repositories {
		data = append(data, repositoryResponse{
			ID:            repository.ID,
			GiteaID:       repository.GiteaID,
			FullName:      repository.FullName,
			OwnerLogin:    repository.OwnerLogin,
			Name:          repository.Name,
			CloneURL:      repository.CloneURL,
			HTMLURL:       repository.HTMLURL,
			DefaultBranch: repository.DefaultBranch,
			Private:       repository.Private,
			UpdatedAt:     repository.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func (h Handler) issues(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, true)
	if !ok {
		return
	}
	issues, err := h.cfg.DomainStore.ListIssues(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_issues")
		return
	}

	data := make([]issueResponse, 0, len(issues))
	for _, issue := range issues {
		data = append(data, issueResponse{
			ID:             issue.ID,
			GiteaID:        issue.GiteaID,
			RepositoryName: issue.RepositoryName,
			Number:         issue.Number,
			Title:          issue.Title,
			State:          issue.State,
			AuthorLogin:    issue.AuthorLogin,
			AssigneeLogin:  issue.AssigneeLogin,
			HTMLURL:        issue.HTMLURL,
			OpenedAt:       issue.OpenedAt,
			ClosedAt:       issue.ClosedAt,
			UpdatedAt:      issue.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func (h Handler) pullRequests(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, true)
	if !ok {
		return
	}
	pullRequests, err := h.cfg.DomainStore.ListPullRequests(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_pull_requests")
		return
	}

	data := make([]pullRequestResponse, 0, len(pullRequests))
	for _, pullRequest := range pullRequests {
		data = append(data, pullRequestResponse{
			ID:             pullRequest.ID,
			GiteaID:        pullRequest.GiteaID,
			RepositoryName: pullRequest.RepositoryName,
			Number:         pullRequest.Number,
			Title:          pullRequest.Title,
			State:          pullRequest.State,
			AuthorLogin:    pullRequest.AuthorLogin,
			HeadBranch:     pullRequest.HeadBranch,
			BaseBranch:     pullRequest.BaseBranch,
			HeadSHA:        pullRequest.HeadSHA,
			HTMLURL:        pullRequest.HTMLURL,
			MergedAt:       pullRequest.MergedAt,
			ClosedAt:       pullRequest.ClosedAt,
			UpdatedAt:      pullRequest.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func (h Handler) risks(c *gin.Context) {
	if h.cfg.DomainStore == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "domain store is not configured",
		})
		return
	}

	opts, ok := parseListOptions(c, false)
	if !ok {
		return
	}
	opts.Impact = c.Query("impact")
	risks, err := h.cfg.DomainStore.ListRisks(c.Request.Context(), opts)
	if err != nil {
		writeServerError(c, err, "domain.list_risks")
		return
	}

	data := make([]riskResponse, 0, len(risks))
	for _, risk := range risks {
		data = append(data, riskFromDomain(risk))
	}
	c.JSON(http.StatusOK, listEnvelope(data, opts))
}

func parseListOptions(c *gin.Context, includeState bool) (domain.ListOptions, bool) {
	limit, err := parseBoundedInt(c.DefaultQuery("limit", "50"), 1, 100)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "limit must be an integer between 1 and 100"})
		return domain.ListOptions{}, false
	}
	offset, err := parseBoundedInt(c.DefaultQuery("offset", "0"), 0, 100000)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "rejected", "error": "offset must be a non-negative integer"})
		return domain.ListOptions{}, false
	}
	opts := domain.ListOptions{
		Limit:          limit,
		Offset:         offset,
		RepositoryName: c.Query("repository_name"),
		Status:         c.Query("status"),
	}
	if includeState {
		opts.State = c.Query("state")
	}
	return opts, true
}

func listEnvelope[T any](data []T, opts domain.ListOptions) gin.H {
	meta := gin.H{
		"limit":  opts.Limit,
		"offset": opts.Offset,
		"count":  len(data),
	}
	if opts.RepositoryName != "" {
		meta["repository_name"] = opts.RepositoryName
	}
	if opts.State != "" {
		meta["state"] = opts.State
	}
	if opts.Status != "" {
		meta["status"] = opts.Status
	}
	if opts.Impact != "" {
		meta["impact"] = opts.Impact
	}
	return gin.H{
		"status": "ok",
		"data":   data,
		"meta":   meta,
	}
}
