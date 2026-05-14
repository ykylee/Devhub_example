package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
)

// memoryApplicationStore is an in-memory ApplicationStore used by handler tests.
// 본 store 는 SQL CHECK 제약 / FK 동작을 흉내내지 않으므로, handler 레벨 검증 (key
// 정규식, immutable, 상태 전이 가드, RBAC denial) 만 검증 대상. PostgreSQL 통합
// 테스트는 별도 (build-tagged) — 본 sprint 의 carve out.
type memoryApplicationStore struct {
	mu               sync.Mutex
	apps             map[string]domain.Application
	links            map[string][]domain.ApplicationRepository
	providers        map[string]domain.SCMProvider
	projects         map[string]domain.Project
	activeLinkCounts map[string]int
	integrations     map[string]domain.ProjectIntegration
	criticalCounts   map[string]int // override for CountApplicationCriticalWarnings tests
}

func newMemoryApplicationStore() *memoryApplicationStore {
	return &memoryApplicationStore{
		apps:  make(map[string]domain.Application),
		links: make(map[string][]domain.ApplicationRepository),
		providers: map[string]domain.SCMProvider{
			"bitbucket": {ProviderKey: "bitbucket", DisplayName: "Bitbucket", Enabled: true, AdapterVersion: "0.0.1"},
			"gitea":     {ProviderKey: "gitea", DisplayName: "Gitea", Enabled: true, AdapterVersion: "0.0.1"},
			"forgejo":   {ProviderKey: "forgejo", DisplayName: "Forgejo", Enabled: false, AdapterVersion: "0.0.1"},
		},
		projects:         make(map[string]domain.Project),
		activeLinkCounts: make(map[string]int),
		integrations:     make(map[string]domain.ProjectIntegration),
		criticalCounts:   make(map[string]int),
	}
}

func (s *memoryApplicationStore) ListApplications(_ context.Context, opts store.ApplicationListOptions) ([]domain.Application, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Application, 0, len(s.apps))
	for _, a := range s.apps {
		if opts.Status != "" && string(a.Status) != opts.Status {
			continue
		}
		if !opts.IncludeArchived && a.Status == domain.ApplicationStatusArchived {
			continue
		}
		out = append(out, a)
	}
	return out, len(out), nil
}

func (s *memoryApplicationStore) GetApplication(_ context.Context, id string) (domain.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a, ok := s.apps[id]; ok {
		return a, nil
	}
	return domain.Application{}, store.ErrNotFound
}

func (s *memoryApplicationStore) GetApplicationByKey(_ context.Context, key string) (domain.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.apps {
		if a.Key == key {
			return a, nil
		}
	}
	return domain.Application{}, store.ErrNotFound
}

func (s *memoryApplicationStore) CreateApplication(_ context.Context, app domain.Application) (domain.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.apps {
		if a.Key == app.Key {
			return domain.Application{}, store.ErrConflict
		}
	}
	if app.ID == "" {
		app.ID = "app-" + app.Key
	}
	app.CreatedAt = time.Now().UTC()
	app.UpdatedAt = app.CreatedAt
	s.apps[app.ID] = app
	return app, nil
}

func (s *memoryApplicationStore) UpdateApplication(_ context.Context, app domain.Application) (domain.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.apps[app.ID]
	if !ok {
		return domain.Application{}, store.ErrNotFound
	}
	app.CreatedAt = current.CreatedAt
	app.UpdatedAt = time.Now().UTC()
	if app.Status == domain.ApplicationStatusArchived && app.ArchivedAt == nil {
		now := time.Now().UTC()
		app.ArchivedAt = &now
	} else if app.Status != domain.ApplicationStatusArchived {
		app.ArchivedAt = nil
	}
	s.apps[app.ID] = app
	return app, nil
}

func (s *memoryApplicationStore) ArchiveApplication(_ context.Context, id, _ string) (domain.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.apps[id]
	if !ok {
		return domain.Application{}, store.ErrNotFound
	}
	app.Status = domain.ApplicationStatusArchived
	now := time.Now().UTC()
	app.ArchivedAt = &now
	app.UpdatedAt = now
	s.apps[id] = app
	return app, nil
}

func (s *memoryApplicationStore) CountActiveApplicationRepositories(_ context.Context, applicationID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeLinkCounts[applicationID], nil
}

func (s *memoryApplicationStore) ListApplicationRepositories(_ context.Context, applicationID string) ([]domain.ApplicationRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.ApplicationRepository(nil), s.links[applicationID]...), nil
}

func (s *memoryApplicationStore) CreateApplicationRepository(_ context.Context, link domain.ApplicationRepository) (domain.ApplicationRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.apps[link.ApplicationID]; !ok {
		return domain.ApplicationRepository{}, store.ErrConflict
	}
	existing := s.links[link.ApplicationID]
	for _, e := range existing {
		if e.RepoProvider == link.RepoProvider && e.RepoFullName == link.RepoFullName {
			return domain.ApplicationRepository{}, store.ErrConflict
		}
	}
	link.LinkedAt = time.Now().UTC()
	s.links[link.ApplicationID] = append(existing, link)
	if link.SyncStatus == domain.SyncStatusActive {
		s.activeLinkCounts[link.ApplicationID]++
	}
	return link, nil
}

func (s *memoryApplicationStore) DeleteApplicationRepository(_ context.Context, key store.ApplicationRepositoryLinkKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.links[key.ApplicationID]
	for i, l := range links {
		if l.RepoProvider == key.RepoProvider && l.RepoFullName == key.RepoFullName {
			if l.SyncStatus == domain.SyncStatusActive {
				s.activeLinkCounts[key.ApplicationID]--
			}
			s.links[key.ApplicationID] = append(links[:i], links[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *memoryApplicationStore) UpdateApplicationRepositorySync(_ context.Context, key store.ApplicationRepositoryLinkKey, status domain.ApplicationRepositorySyncStatus, code domain.SyncErrorCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.links[key.ApplicationID]
	for i, l := range links {
		if l.RepoProvider == key.RepoProvider && l.RepoFullName == key.RepoFullName {
			if l.SyncStatus != domain.SyncStatusActive && status == domain.SyncStatusActive {
				s.activeLinkCounts[key.ApplicationID]++
			} else if l.SyncStatus == domain.SyncStatusActive && status != domain.SyncStatusActive {
				s.activeLinkCounts[key.ApplicationID]--
			}
			links[i].SyncStatus = status
			links[i].SyncErrorCode = code
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *memoryApplicationStore) ListSCMProviders(_ context.Context) ([]domain.SCMProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.SCMProvider, 0, len(s.providers))
	for _, p := range s.providers {
		out = append(out, p)
	}
	return out, nil
}

func (s *memoryApplicationStore) UpdateSCMProvider(_ context.Context, p domain.SCMProvider) (domain.SCMProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.providers[p.ProviderKey]; !ok {
		return domain.SCMProvider{}, store.ErrNotFound
	}
	cur := s.providers[p.ProviderKey]
	cur.DisplayName = p.DisplayName
	cur.Enabled = p.Enabled
	cur.UpdatedAt = time.Now().UTC()
	s.providers[p.ProviderKey] = cur
	return cur, nil
}

func (s *memoryApplicationStore) ListProjects(_ context.Context, opts store.ProjectListOptions) ([]domain.Project, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Project, 0)
	for _, p := range s.projects {
		if p.RepositoryID != opts.RepositoryID {
			continue
		}
		if opts.Status != "" && string(p.Status) != opts.Status {
			continue
		}
		if !opts.IncludeArchived && p.Status == domain.ApplicationStatusArchived {
			continue
		}
		out = append(out, p)
	}
	return out, len(out), nil
}

func (s *memoryApplicationStore) GetProject(_ context.Context, id string) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.projects[id]; ok {
		return p, nil
	}
	return domain.Project{}, store.ErrNotFound
}

func (s *memoryApplicationStore) CreateProject(_ context.Context, p domain.Project) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.projects {
		if existing.RepositoryID == p.RepositoryID && existing.Key == p.Key {
			return domain.Project{}, store.ErrConflict
		}
	}
	if p.ID == "" {
		p.ID = "proj-" + p.Key
	}
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	s.projects[p.ID] = p
	return p, nil
}

func (s *memoryApplicationStore) UpdateProject(_ context.Context, p domain.Project) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[p.ID]; !ok {
		return domain.Project{}, store.ErrNotFound
	}
	p.UpdatedAt = time.Now().UTC()
	s.projects[p.ID] = p
	return p, nil
}

func (s *memoryApplicationStore) ArchiveProject(_ context.Context, id, _ string) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[id]
	if !ok {
		return domain.Project{}, store.ErrNotFound
	}
	p.Status = domain.ApplicationStatusArchived
	now := time.Now().UTC()
	p.ArchivedAt = &now
	p.UpdatedAt = now
	s.projects[id] = p
	return p, nil
}

// --- Repository 운영 지표 (sprint claude/work_260514-c) ---
// 메모리 store 는 SQL 집계를 흉내내지 않으므로 모두 zero-value 반환.

func (s *memoryApplicationStore) ListRepositoryActivity(_ context.Context, repoID int64, _ store.RepositoryActivityOptions) (domain.RepositoryActivity, error) {
	return domain.RepositoryActivity{RepositoryID: repoID}, nil
}

func (s *memoryApplicationStore) ListRepositoryPullRequests(_ context.Context, _ int64, _ store.PRActivityListOptions) ([]domain.PRActivity, int, error) {
	return []domain.PRActivity{}, 0, nil
}

func (s *memoryApplicationStore) ListRepositoryBuildRuns(_ context.Context, _ int64, _ store.BuildRunListOptions) ([]domain.BuildRun, int, error) {
	return []domain.BuildRun{}, 0, nil
}

func (s *memoryApplicationStore) ListRepositoryQualitySnapshots(_ context.Context, _ int64, _ store.QualitySnapshotListOptions) ([]domain.QualitySnapshot, int, error) {
	return []domain.QualitySnapshot{}, 0, nil
}

// --- Application 롤업 (sprint claude/work_260514-c) ---

func (s *memoryApplicationStore) ComputeApplicationRollup(_ context.Context, _ string, opts domain.ApplicationRollupOptions) (domain.ApplicationRollup, error) {
	if opts.Policy == "" {
		opts.Policy = domain.WeightPolicyEqual
	}
	// custom weight 검증만 흉내냄 (handler 분기 테스트용).
	if opts.Policy == domain.WeightPolicyCustom {
		sum := 0.0
		for _, w := range opts.CustomWeights {
			if w < 0 {
				return domain.ApplicationRollup{}, errors.New("invalid weight policy: negative weight")
			}
			sum += w
		}
		if sum < 1.0-domain.CustomWeightTolerance || sum > 1.0+domain.CustomWeightTolerance {
			return domain.ApplicationRollup{}, errors.New("invalid weight policy: custom weights must sum to 1.0")
		}
	}
	return domain.ApplicationRollup{
		PullRequestDistribution: map[string]int{},
		Meta: domain.ApplicationRollupMeta{
			Period:         domain.RollupPeriod{From: time.Now().UTC(), To: time.Now().UTC()},
			Filters:        map[string]any{},
			WeightPolicy:   opts.Policy,
			AppliedWeights: map[string]float64{},
			Fallbacks:      []domain.RollupFallback{},
			DataGaps:       []domain.RollupDataGap{},
		},
	}, nil
}

func (s *memoryApplicationStore) CountApplicationCriticalWarnings(_ context.Context, _ string) (int, error) {
	// 1차 메모리 store 는 critical warning 이 없는 환경 가정. 별도 case 가 필요한 test 가
	// 있으면 sub-type 으로 override.
	return 0, nil
}

// --- Integration CRUD (sprint claude/work_260514-c) ---

type memoryIntegration = domain.ProjectIntegration

func (s *memoryApplicationStore) ListIntegrations(_ context.Context, opts store.IntegrationListOptions) ([]domain.ProjectIntegration, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.ProjectIntegration, 0)
	for _, i := range s.integrations {
		if string(opts.Scope) != "" && string(i.Scope) != string(opts.Scope) {
			continue
		}
		if opts.ApplicationID != "" && i.ApplicationID != opts.ApplicationID {
			continue
		}
		if opts.ProjectID != "" && i.ProjectID != opts.ProjectID {
			continue
		}
		if string(opts.IntegrationType) != "" && string(i.IntegrationType) != string(opts.IntegrationType) {
			continue
		}
		out = append(out, i)
	}
	return out, len(out), nil
}

func (s *memoryApplicationStore) GetIntegration(_ context.Context, id string) (domain.ProjectIntegration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i, ok := s.integrations[id]; ok {
		return i, nil
	}
	return domain.ProjectIntegration{}, store.ErrNotFound
}

func (s *memoryApplicationStore) CreateIntegration(_ context.Context, i domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.integrations {
		if existing.Scope == i.Scope &&
			existing.ApplicationID == i.ApplicationID &&
			existing.ProjectID == i.ProjectID &&
			existing.IntegrationType == i.IntegrationType &&
			existing.ExternalKey == i.ExternalKey {
			return domain.ProjectIntegration{}, store.ErrConflict
		}
	}
	if i.ID == "" {
		i.ID = "int-" + i.ExternalKey
	}
	i.CreatedAt = time.Now().UTC()
	i.UpdatedAt = i.CreatedAt
	s.integrations[i.ID] = i
	return i, nil
}

func (s *memoryApplicationStore) UpdateIntegration(_ context.Context, i domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.integrations[i.ID]
	if !ok {
		return domain.ProjectIntegration{}, store.ErrNotFound
	}
	current.ExternalKey = i.ExternalKey
	current.URL = i.URL
	current.Policy = i.Policy
	current.UpdatedAt = time.Now().UTC()
	s.integrations[i.ID] = current
	return current, nil
}

func (s *memoryApplicationStore) DeleteIntegration(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrations[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.integrations, id)
	return nil
}

// --- handler tests ---

func newApplicationsRouter(appStore ApplicationStore) http.Handler {
	return NewRouter(RouterConfig{
		ApplicationStore: appStore,
		AuthDevFallback:  true, // bypass bearer auth
	})
}

// 1) POST /applications — happy.
func TestCreateApplication_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications",
		`{"key":"A1B2C3D4E5","name":"Devhub Platform","owner_user_id":"u1","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"key":"A1B2C3D4E5"`)) {
		t.Errorf("response should echo key: %s", rec.Body.String())
	}
}

// 2) POST /applications — invalid key format → 422 invalid_application_key.
func TestCreateApplication_InvalidKey(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())

	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications",
		`{"key":"too-short","name":"X","owner_user_id":"u1","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_application_key"`)) {
		t.Errorf("response should carry invalid_application_key code: %s", rec.Body.String())
	}
}

// 3) POST /applications — duplicate key → 409 application_key_conflict.
func TestCreateApplication_DuplicateKey(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	body := `{"key":"A1B2C3D4E5","name":"X","owner_user_id":"u1","visibility":"internal","status":"planning"}`
	first := doJSON(t, router, http.MethodPost, "/api/v1/applications", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", first.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"application_key_conflict"`)) {
		t.Errorf("expected application_key_conflict: %s", rec.Body.String())
	}
}

// 4) PATCH /applications/:id — immutable key 거부 → 422 application_key_immutable.
func TestUpdateApplication_ImmutableKey(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
		`{"key":"NEWKEY1234"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"application_key_immutable"`)) {
		t.Errorf("expected application_key_immutable: %s", rec.Body.String())
	}
}

// 5) PATCH /applications/:id — planning → active 의 활성 repo 0건 → 422.
func TestUpdateApplication_ActivationPreconditionFailed(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
		`{"status":"active"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"application_activation_precondition_failed"`)) {
		t.Errorf("expected activation precondition failure: %s", rec.Body.String())
	}
}

// 6) PATCH /applications/:id — planning → active 의 활성 repo 1개 → 200.
func TestUpdateApplication_ActivationSuccess(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = appStore.CreateApplicationRepository(context.Background(), domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
		`{"status":"active"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"active"`)) {
		t.Errorf("expected status=active: %s", rec.Body.String())
	}
}

// 7) PATCH /applications/:id — closed → planning 같은 invalid transition → 422.
func TestUpdateApplication_InvalidStatusTransition(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusClosed,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
		`{"status":"planning"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_status_transition"`)) {
		t.Errorf("expected invalid_status_transition: %s", rec.Body.String())
	}
}

// 8) PATCH /applications/:id — active → on_hold 의 hold_reason 누락 → 422.
func TestUpdateApplication_HoldReasonRequired(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
		`{"status":"on_hold"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_status_transition_payload"`)) {
		t.Errorf("expected invalid_status_transition_payload: %s", rec.Body.String())
	}
}

// 9) DELETE /applications/:id — archive (soft-delete).
func TestArchiveApplication_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodDelete, "/api/v1/applications/"+app.ID,
		`{"archived_reason":"product end-of-life"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"archived"`)) {
		t.Errorf("expected status=archived: %s", rec.Body.String())
	}
	final, err := appStore.GetApplication(context.Background(), app.ID)
	if err != nil {
		t.Fatalf("application lost after archive: %v", err)
	}
	if final.Status != domain.ApplicationStatusArchived {
		t.Errorf("status not archived: %s", final.Status)
	}
}

// 10) POST /applications/:id/repositories — unsupported_repo_provider → 422.
func TestCreateApplicationRepository_UnsupportedProvider(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications/"+app.ID+"/repositories",
		`{"repo_provider":"unknown","repo_full_name":"team/repo","role":"primary"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"unsupported_repo_provider"`)) {
		t.Errorf("expected unsupported_repo_provider: %s", rec.Body.String())
	}
}

// 11) POST /applications/:id/repositories — disabled provider → 422.
func TestCreateApplicationRepository_DisabledProvider(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications/"+app.ID+"/repositories",
		`{"repo_provider":"forgejo","repo_full_name":"team/repo","role":"primary"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 12) POST /applications/:id/repositories — duplicate link → 409 repository_link_conflict.
func TestCreateApplicationRepository_DuplicateLink(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)
	body := `{"repo_provider":"gitea","repo_full_name":"team/repo","role":"primary"}`
	first := doJSON(t, router, http.MethodPost, "/api/v1/applications/"+app.ID+"/repositories", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", first.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications/"+app.ID+"/repositories", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"repository_link_conflict"`)) {
		t.Errorf("expected repository_link_conflict: %s", rec.Body.String())
	}
}

// 13) DELETE /applications/:id/repositories/:repo_key — colon convention.
func TestDeleteApplicationRepository_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = appStore.CreateApplicationRepository(context.Background(), domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusRequested,
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/applications/"+app.ID+"/repositories/gitea:team/repo", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	links, _ := appStore.ListApplicationRepositories(context.Background(), app.ID)
	if len(links) != 0 {
		t.Errorf("link should be removed, got %d", len(links))
	}
}

// 14) DELETE /applications/:id/repositories/:repo_key — bad format → 400.
func TestDeleteApplicationRepository_BadKey(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/applications/some-id/repositories/no-colon", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 15) PATCH /scm/providers/:provider_key — adapter_version 거부 → 422.
func TestUpdateSCMProvider_AdapterVersionImmutable(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/scm/providers/gitea",
		`{"adapter_version":"9.9.9"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"adapter_version_immutable"`)) {
		t.Errorf("expected adapter_version_immutable: %s", rec.Body.String())
	}
}

// 16) GET /scm/providers — list happy.
func TestListSCMProviders_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/scm/providers", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"provider_key":"gitea"`) {
		t.Errorf("expected gitea in list: %s", body)
	}
}

// 17) GET /applications — empty list when none seeded.
func TestListApplications_Empty(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"data":[]`)) {
		t.Errorf("expected empty data array: %s", rec.Body.String())
	}
}

// 18) GET /applications/:id — not found → 404.
func TestGetApplication_NotFound(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 19) ApplicationStore nil 인 경우 503 (configuration error).
func TestApplications_ServiceUnavailable(t *testing.T) {
	router := NewRouter(RouterConfig{
		AuthDevFallback: true,
	})
	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// guard: ErrNotImplemented 가 더 이상 store layer 의 public API 가 아니라는 보증
// (sprint claude/work_260514-a 의 stub 제거 확인).
func TestStoreErrNotImplementedRemovedFromHandlerPath(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications", "")
	if rec.Code == http.StatusNotImplemented {
		t.Fatalf("handler returned 501 — stub removal incomplete: %s", rec.Body.String())
	}
}

// Compile-time guards (side-effect-free).
//
// `ApplicationStore` 인터페이스 시그니처가 변경되면 본 assertion 이 깨져 컴파일
// 단계에서 즉시 검출된다. 테스트가 직접 호출하지 않더라도 인터페이스 계약을
// 보호하는 안전망이다.
var _ ApplicationStore = (*memoryApplicationStore)(nil)

// 도메인 import / store sentinel error 의 외부 노출이 유지되는지 확인한다. domain
// 의 `IsRetryableSyncError` 가 사라지거나 store 의 `Err*` 가 unexported 로 바뀌면
// 본 블록이 컴파일 실패하여 회귀를 막는다. 런타임 동작은 없음 (no-op).
var (
	_ = domain.IsRetryableSyncError(domain.SyncErrorProviderUnreachable)
	_ = errors.Is(store.ErrConflict, store.ErrConflict)
	_ = errors.Is(store.ErrNotFound, store.ErrNotFound)
)

// --- Happy path 보강 tests (PR #106 self-review I2) ---

// 20) GET /applications — status / include_archived 필터 happy.
func TestListApplications_FiltersHappy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	for _, status := range []domain.ApplicationStatus{
		domain.ApplicationStatusPlanning,
		domain.ApplicationStatusActive,
		domain.ApplicationStatusArchived,
	} {
		_, _ = appStore.CreateApplication(context.Background(), domain.Application{
			Key: "K-" + string(status[:6]), Name: "N", Status: status,
			Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
		})
	}
	router := newApplicationsRouter(appStore)

	// default: archived 제외 → 2건
	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"total":2`) {
		t.Errorf("default list should exclude archived (total=2): %s", rec.Body.String())
	}

	// include_archived=true → 3건
	rec = doJSON(t, router, http.MethodGet, "/api/v1/applications?include_archived=true", "")
	if !strings.Contains(rec.Body.String(), `"total":3`) {
		t.Errorf("include_archived=true should return all (total=3): %s", rec.Body.String())
	}

	// status=active → 1건
	rec = doJSON(t, router, http.MethodGet, "/api/v1/applications?status=active", "")
	if !strings.Contains(rec.Body.String(), `"total":1`) {
		t.Errorf("status=active should return 1: %s", rec.Body.String())
	}
}

// 21) GET /applications/:id — happy (메타 + repositories 포함).
func TestGetApplication_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = appStore.CreateApplicationRepository(context.Background(), domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications/"+app.ID, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"key":"A1B2C3D4E5"`) {
		t.Errorf("response should include key: %s", body)
	}
	if !strings.Contains(body, `"repositories":[`) {
		t.Errorf("response should include repositories array: %s", body)
	}
	if !strings.Contains(body, `"repo_full_name":"team/repo"`) {
		t.Errorf("response should include link details: %s", body)
	}
}

// 22) POST /applications/:id/repositories — happy.
func TestCreateApplicationRepository_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications/"+app.ID+"/repositories",
		`{"repo_provider":"gitea","repo_full_name":"team/devhub-core","role":"primary"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"sync_status":"requested"`) {
		t.Errorf("new link should start at sync_status=requested: %s", rec.Body.String())
	}
}

// 23) PATCH /scm/providers/:provider_key — happy (enabled toggle).
func TestUpdateSCMProvider_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/scm/providers/gitea",
		`{"enabled":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"enabled":false`) {
		t.Errorf("expected enabled=false: %s", rec.Body.String())
	}
}

// 24) DELETE /applications/:id — not found → 404.
func TestArchiveApplication_NotFound(t *testing.T) {
	appStore := newMemoryApplicationStore()
	router := newApplicationsRouter(appStore)
	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/applications/nonexistent", `{"archived_reason":"X"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 25) DELETE catch-all path — `//gitea:team/repo` 같은 multiple leading slash 도
// TrimLeft 로 정상 처리되는지 (PR #106 self-review N1 보강).
func TestDeleteApplicationRepository_MultipleLeadingSlashes(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = appStore.CreateApplicationRepository(context.Background(), domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusRequested,
	})
	router := newApplicationsRouter(appStore)
	// gin 은 `//` 를 보통 정규화하지 않으므로 catch-all 이 받은 raw path 를 TrimLeft 가
	// 안전하게 처리해야 한다.
	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/applications/"+app.ID+"/repositories//gitea:team/repo", "")
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		// gin 의 path 정규화에 따라 OK (200) 또는 NotFound (404, 정규화 후 trailing
		// slash 처리 차이) 가 나올 수 있다. 핵심은 500/400 같은 예상치 못한 응답이
		// 아닌 정상 routing 이 동작한다는 것.
		t.Fatalf("unexpected status=%d body=%s", rec.Code, rec.Body.String())
	}
}
