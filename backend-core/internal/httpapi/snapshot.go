package httpapi

import (
	"net/http"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/gin-gonic/gin"
)

type metricResponse struct {
	ID             string  `json:"id"`
	Label          string  `json:"label"`
	Value          string  `json:"value"`
	Trend          string  `json:"trend"`
	TrendDirection string  `json:"trend_direction"`
	NumericValue   float64 `json:"numeric_value,omitempty"`
	Unit           string  `json:"unit,omitempty"`
}

type serviceNodeResponse struct {
	ID              string    `json:"id"`
	Label           string    `json:"label"`
	Kind            string    `json:"kind"`
	Status          string    `json:"status"`
	Region          string    `json:"region,omitempty"`
	CPUPercent      float64   `json:"cpu_percent"`
	MemoryBytes     int64     `json:"memory_bytes"`
	ActiveInstances int       `json:"active_instances"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type serviceEdgeResponse struct {
	ID            string    `json:"id"`
	SourceID      string    `json:"source_id"`
	TargetID      string    `json:"target_id"`
	Label         string    `json:"label"`
	Status        string    `json:"status"`
	LatencyMS     float64   `json:"latency_ms"`
	ThroughputRPS float64   `json:"throughput_rps"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ciRunResponse struct {
	ID              string     `json:"id"`
	RepositoryName  string     `json:"repository_name"`
	Branch          string     `json:"branch"`
	CommitSHA       string     `json:"commit_sha"`
	Status          string     `json:"status"`
	DurationSeconds int        `json:"duration_seconds"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

type ciLogLineResponse struct {
	ID        string    `json:"id"`
	CIRunID   string    `json:"ci_run_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	StepName  string    `json:"step_name"`
}

type riskResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Reason           string    `json:"reason"`
	Impact           string    `json:"impact"`
	Status           string    `json:"status"`
	OwnerLogin       string    `json:"owner_login"`
	SuggestedActions []string  `json:"suggested_actions"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (h Handler) dashboardMetrics(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		role = "developer"
	}

	provider := h.snapshotProvider()
	metrics, ok, err := provider.DashboardMetrics(c.Request.Context(), role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "rejected",
			"error":  "role must be one of developer, manager, system_admin",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   metrics,
		"meta": gin.H{
			"role":   role,
			"count":  len(metrics),
			"source": provider.Source(),
		},
	})
}

func (h Handler) infraNodes(c *gin.Context) {
	provider := h.snapshotProvider()
	nodes, err := provider.InfraNodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   nodes,
		"meta": gin.H{
			"count":  len(nodes),
			"source": provider.Source(),
		},
	})
}

func (h Handler) infraEdges(c *gin.Context) {
	provider := h.snapshotProvider()
	edges, err := provider.InfraEdges(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   edges,
		"meta": gin.H{
			"count":  len(edges),
			"source": provider.Source(),
		},
	})
}

func (h Handler) infraTopology(c *gin.Context) {
	provider := h.snapshotProvider()
	nodes, err := provider.InfraNodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	edges, err := provider.InfraEdges(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data": gin.H{
			"nodes": nodes,
			"edges": edges,
		},
		"meta": gin.H{
			"node_count": len(nodes),
			"edge_count": len(edges),
			"source":     provider.Source(),
		},
	})
}

func (h Handler) ciRuns(c *gin.Context) {
	scope := c.DefaultQuery("scope", "all")
	if h.cfg.DomainStore != nil {
		opts, ok := parseListOptions(c, false)
		if !ok {
			return
		}
		opts.Status = c.Query("status")
		runs, err := h.cfg.DomainStore.ListCIRuns(c.Request.Context(), opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
		if len(runs) > 0 {
			data := make([]ciRunResponse, 0, len(runs))
			for _, run := range runs {
				startedAt := time.Time{}
				if run.StartedAt != nil {
					startedAt = *run.StartedAt
				}
				durationSeconds := 0
				if run.DurationSeconds != nil {
					durationSeconds = *run.DurationSeconds
				}
				data = append(data, ciRunResponse{
					ID:              run.ExternalID,
					RepositoryName:  run.RepositoryName,
					Branch:          run.Branch,
					CommitSHA:       run.CommitSHA,
					Status:          run.Status,
					DurationSeconds: durationSeconds,
					StartedAt:       startedAt,
					FinishedAt:      run.FinishedAt,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"data":   data,
				"meta": gin.H{
					"count":  len(data),
					"scope":  scope,
					"source": "db",
				},
			})
			return
		}
	}

	provider := h.snapshotProvider()
	runs, err := provider.CIRuns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   runs,
		"meta": gin.H{
			"count":  len(runs),
			"scope":  scope,
			"source": provider.Source(),
		},
	})
}

func (h Handler) ciRunLogs(c *gin.Context) {
	ciRunID := c.Param("ci_run_id")
	provider := h.snapshotProvider()
	logs, ok, err := provider.CILogs(c.Request.Context(), ciRunID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "not_found",
			"error":  "ci run logs not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   logs,
		"meta": gin.H{
			"ci_run_id": ciRunID,
			"count":     len(logs),
			"source":    provider.Source(),
		},
	})
}

func (h Handler) criticalRisks(c *gin.Context) {
	if h.cfg.DomainStore != nil {
		risks, err := h.cfg.DomainStore.ListRisks(c.Request.Context(), domain.ListOptions{
			Limit:  50,
			Status: "action_required",
			Impact: "high",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
			return
		}
		if len(risks) > 0 {
			data := make([]riskResponse, 0, len(risks))
			for _, risk := range risks {
				data = append(data, riskFromDomain(risk))
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"data":   data,
				"meta": gin.H{
					"count":  len(data),
					"source": "db",
				},
			})
			return
		}
	}

	provider := h.snapshotProvider()
	risks, err := provider.CriticalRisks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "failed", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   risks,
		"meta": gin.H{
			"count":  len(risks),
			"source": provider.Source(),
		},
	})
}
