package view

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/domain/application-lifecycle/repository"
	"github.com/devhub/backend-core/internal/store"
)
// fakeViewPlatformStore — view 패키지 endpoint handler test 용 in-memory store.
// view.PlatformStore (27 메서드) 만족. httpapi 패키지의 memoryPlatformStore
// 와 별도로 view 패키지 안에 둠 — httpapi 의 fake 는 IntegrationStore 메서드까지
// 포함해 cross-package import 가 불가하고, view 의 interface 는 27 메서드만 필요.
// SQL CHECK / FK 제약은 흉내내지 않고, handler 레벨 validation/매핑/audit 만 cover.
type fakeViewPlatformStore struct {
	mu sync.Mutex

	platforms                 map[string]domain.Platform
	links                map[string][]domain.PlatformRepository
	providers            map[string]domain.SCMProvider
	projects             map[string]domain.Project
	projectRepositories  map[string][]domain.ProjectRepository
	repositories         map[string][]domain.Repository // providerKey → repos
	criticalWarnings     map[string]int

	// 에러 주입 — 특정 메서드가 실패해야 하는 테스트용. nil 이면 정상 동작.
	errListPlatforms              error
	errGetPlatform                error
	errCreatePlatform             error
	errUpdatePlatform             error
	errArchivePlatform            error
	errDeletePlatform             error
	errListPlatformRepositories   error
	errCreatePlatformRepository   error
	errDeletePlatformRepository   error
	errListSCMProviders              error
	errUpdateSCMProvider             error
	errListProjects                  error
	errGetProject                    error
	errCreateProject                 error
	errUpdateProject                 error
	errArchiveProject                error
	errDeleteProject                 error
	errListProjectRepositories       error
	errCreateProjectRepository       error
	errDeleteProjectRepository       error
	errUpdateProjectRepositoryWeight error
	errCreateProjectWithRepoPayload  error
	errComputeProjectWeightedKPI     error // Sprint B
	errListRepositoriesByProvider    error
	errComputePlatformRollup      error
	errListRepositoryBuildRuns       error
	errListProjectTestResults        error // Sprint B-Tests
}

func newFakeViewPlatformStore() *fakeViewPlatformStore {
	return &fakeViewPlatformStore{
		platforms:                make(map[string]domain.Platform),
		links:               make(map[string][]domain.PlatformRepository),
		providers:           map[string]domain.SCMProvider{},
		projects:            make(map[string]domain.Project),
		projectRepositories: make(map[string][]domain.ProjectRepository),
		repositories:        make(map[string][]domain.Repository),
		criticalWarnings:    make(map[string]int),
	}
}

// seedApp 은 test setup helper — direct map insert 보다 가독성 좋게.
func (s *fakeViewPlatformStore) seedApp(app domain.Platform) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if app.ID == "" {
		app.ID = "app-" + app.Key
	}
	if app.CreatedAt.IsZero() {
		app.CreatedAt = time.Now().UTC()
	}
	if app.UpdatedAt.IsZero() {
		app.UpdatedAt = app.CreatedAt
	}
	s.platforms[app.ID] = app
}

func (s *fakeViewPlatformStore) seedProvider(p domain.SCMProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[p.ProviderKey] = p
}

func (s *fakeViewPlatformStore) seedProject(p domain.Project) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = "proj-" + p.Key
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	if p.UpdatedAt.IsZero() {
		p.UpdatedAt = p.CreatedAt
	}
	s.projects[p.ID] = p
}

func (s *fakeViewPlatformStore) seedLink(link domain.PlatformRepository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if link.LinkedAt.IsZero() {
		link.LinkedAt = time.Now().UTC()
	}
	s.links[link.PlatformID] = append(s.links[link.PlatformID], link)
}

func (s *fakeViewPlatformStore) seedProjectRepo(link domain.ProjectRepository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if link.LinkedAt.IsZero() {
		link.LinkedAt = time.Now().UTC()
	}
	s.projectRepositories[link.ProjectID] = append(s.projectRepositories[link.ProjectID], link)
}

func (s *fakeViewPlatformStore) seedRepository(providerKey string, repo domain.Repository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repositories[providerKey] = append(s.repositories[providerKey], repo)
}

// --- PlatformStore interface 구현 ---

func (s *fakeViewPlatformStore) ListPlatforms(_ context.Context, opts store.PlatformListOptions) ([]domain.Platform, int, error) {
	if s.errListPlatforms != nil {
		return nil, 0, s.errListPlatforms
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Platform, 0, len(s.platforms))
	for _, a := range s.platforms {
		if opts.Status != "" && string(a.Status) != opts.Status {
			continue
		}
		if !opts.IncludeArchived && a.Status == domain.PlatformStatusArchived {
			continue
		}
		out = append(out, a)
	}
	return out, len(out), nil
}

func (s *fakeViewPlatformStore) GetPlatform(_ context.Context, id string) (domain.Platform, error) {
	if s.errGetPlatform != nil {
		return domain.Platform{}, s.errGetPlatform
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if a, ok := s.platforms[id]; ok {
		return a, nil
	}
	return domain.Platform{}, store.ErrNotFound
}

func (s *fakeViewPlatformStore) GetPlatformByKey(_ context.Context, key string) (domain.Platform, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.platforms {
		if a.Key == key {
			return a, nil
		}
	}
	return domain.Platform{}, store.ErrNotFound
}

func (s *fakeViewPlatformStore) CreatePlatform(_ context.Context, app domain.Platform) (domain.Platform, error) {
	if s.errCreatePlatform != nil {
		return domain.Platform{}, s.errCreatePlatform
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.platforms {
		if a.Key == app.Key {
			return domain.Platform{}, store.ErrConflict
		}
	}
	if app.ID == "" {
		app.ID = "app-" + app.Key
	}
	app.CreatedAt = time.Now().UTC()
	app.UpdatedAt = app.CreatedAt
	s.platforms[app.ID] = app
	return app, nil
}

func (s *fakeViewPlatformStore) UpdatePlatform(_ context.Context, app domain.Platform) (domain.Platform, error) {
	if s.errUpdatePlatform != nil {
		return domain.Platform{}, s.errUpdatePlatform
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.platforms[app.ID]
	if !ok {
		return domain.Platform{}, store.ErrNotFound
	}
	app.CreatedAt = current.CreatedAt
	app.UpdatedAt = time.Now().UTC()
	if app.Status == domain.PlatformStatusArchived && app.ArchivedAt == nil {
		now := time.Now().UTC()
		app.ArchivedAt = &now
	} else if app.Status != domain.PlatformStatusArchived {
		app.ArchivedAt = nil
	}
	s.platforms[app.ID] = app
	return app, nil
}

// UpdatePlatformInboundSource (N-13, ADR-0028 §6 a) — fake implementation
// for application-lifecycle/view handler test. Validates inbound_source_type
// via domain.IsValidPlatformInboundSourceType and stores the field pair on
// the existing platform row.
func (s *fakeViewPlatformStore) UpdatePlatformInboundSource(_ context.Context, platformID, inboundType, inboundConfig string) (domain.Platform, error) {
	if s.errUpdatePlatform != nil {
		return domain.Platform{}, s.errUpdatePlatform
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.platforms[platformID]
	if !ok {
		return domain.Platform{}, store.ErrNotFound
	}
	if !domain.IsValidPlatformInboundSourceType(inboundType) {
		return domain.Platform{}, repository.ErrInvalidInboundSourceType
	}
	if inboundType == "" && inboundConfig != "" {
		return domain.Platform{}, repository.ErrInvalidInboundSourceConfig
	}
	if inboundConfig != "" && !json.Valid([]byte(inboundConfig)) {
		return domain.Platform{}, repository.ErrInvalidInboundSourceConfig
	}
	current.InboundSourceType = inboundType
	current.InboundSourceConfig = inboundConfig
	current.UpdatedAt = time.Now().UTC()
	s.platforms[platformID] = current
	return current, nil
}

// ListEnabledInboundSourcePlatforms (N-13, ADR-0028 §6 a) — fake implementation
// returning only platforms whose inbound_source_type is non-empty.
func (s *fakeViewPlatformStore) ListEnabledInboundSourcePlatforms(_ context.Context) ([]domain.Platform, error) {
	if s.errListPlatforms != nil {
		return nil, s.errListPlatforms
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Platform, 0, len(s.platforms))
	for _, p := range s.platforms {
		if p.InboundSourceType != "" {
			out = append(out, p)
		}
	}
	return out, nil
}
func (s *fakeViewPlatformStore) ArchivePlatform(_ context.Context, id, _ string) (domain.Platform, error) {
	if s.errArchivePlatform != nil {
		return domain.Platform{}, s.errArchivePlatform
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	app, ok := s.platforms[id]
	if !ok {
		return domain.Platform{}, store.ErrNotFound
	}
	app.Status = domain.PlatformStatusArchived
	now := time.Now().UTC()
	app.ArchivedAt = &now
	app.UpdatedAt = now
	s.platforms[id] = app
	return app, nil
}

func (s *fakeViewPlatformStore) DeletePlatform(_ context.Context, id string) error {
	if s.errDeletePlatform != nil {
		return s.errDeletePlatform
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.platforms[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.platforms, id)
	delete(s.links, id)
	return nil
}

func (s *fakeViewPlatformStore) CountActivePlatformRepositories(_ context.Context, platformID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, l := range s.links[platformID] {
		if l.SyncStatus == domain.SyncStatusActive {
			count++
		}
	}
	return count, nil
}

func (s *fakeViewPlatformStore) ListPlatformRepositories(_ context.Context, platformID string) ([]domain.PlatformRepository, error) {
	if s.errListPlatformRepositories != nil {
		return nil, s.errListPlatformRepositories
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]domain.PlatformRepository(nil), s.links[platformID]...)
	return out, nil
}

func (s *fakeViewPlatformStore) CreatePlatformRepository(_ context.Context, link domain.PlatformRepository) (domain.PlatformRepository, error) {
	if s.errCreatePlatformRepository != nil {
		return domain.PlatformRepository{}, s.errCreatePlatformRepository
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.platforms[link.PlatformID]; !ok {
		return domain.PlatformRepository{}, store.ErrConflict
	}
	for _, e := range s.links[link.PlatformID] {
		if e.RepoProvider == link.RepoProvider && e.RepoFullName == link.RepoFullName {
			return domain.PlatformRepository{}, store.ErrConflict
		}
	}
	link.LinkedAt = time.Now().UTC()
	s.links[link.PlatformID] = append(s.links[link.PlatformID], link)
	return link, nil
}

func (s *fakeViewPlatformStore) DeletePlatformRepository(_ context.Context, key store.PlatformRepositoryLinkKey) error {
	if s.errDeletePlatformRepository != nil {
		return s.errDeletePlatformRepository
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.links[key.PlatformID]
	for i, l := range links {
		if l.RepoProvider == key.RepoProvider && l.RepoFullName == key.RepoFullName {
			s.links[key.PlatformID] = append(links[:i], links[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *fakeViewPlatformStore) UpdatePlatformRepositorySync(_ context.Context, key store.PlatformRepositoryLinkKey, status domain.PlatformRepositorySyncStatus, code domain.SyncErrorCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.links[key.PlatformID]
	for i, l := range links {
		if l.RepoProvider == key.RepoProvider && l.RepoFullName == key.RepoFullName {
			links[i].SyncStatus = status
			links[i].SyncErrorCode = code
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *fakeViewPlatformStore) ListSCMProviders(_ context.Context) ([]domain.SCMProvider, error) {
	if s.errListSCMProviders != nil {
		return nil, s.errListSCMProviders
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.SCMProvider, 0, len(s.providers))
	for _, p := range s.providers {
		out = append(out, p)
	}
	return out, nil
}

func (s *fakeViewPlatformStore) UpdateSCMProvider(_ context.Context, p domain.SCMProvider) (domain.SCMProvider, error) {
	if s.errUpdateSCMProvider != nil {
		return domain.SCMProvider{}, s.errUpdateSCMProvider
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.providers[p.ProviderKey]
	if !ok {
		return domain.SCMProvider{}, store.ErrNotFound
	}
	cur.DisplayName = p.DisplayName
	cur.Enabled = p.Enabled
	cur.UpdatedAt = time.Now().UTC()
	s.providers[p.ProviderKey] = cur
	return cur, nil
}

func (s *fakeViewPlatformStore) ListProjects(_ context.Context, opts store.ProjectListOptions) ([]domain.Project, int, error) {
	if s.errListProjects != nil {
		return nil, 0, s.errListProjects
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Project, 0)
	for _, p := range s.projects {
		if opts.RepositoryID != 0 && p.RepositoryID != opts.RepositoryID {
			continue
		}
		if opts.PlatformID != "" && p.PlatformID != opts.PlatformID {
			continue
		}
		if opts.StandaloneOnly && p.PlatformID != "" {
			continue
		}
		if opts.Status != "" && string(p.Status) != opts.Status {
			continue
		}
		if !opts.IncludeArchived && p.Status == domain.PlatformStatusArchived {
			continue
		}
		out = append(out, p)
	}
	return out, len(out), nil
}

func (s *fakeViewPlatformStore) GetProject(_ context.Context, id string) (domain.Project, error) {
	if s.errGetProject != nil {
		return domain.Project{}, s.errGetProject
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.projects[id]; ok {
		return p, nil
	}
	return domain.Project{}, store.ErrNotFound
}

func (s *fakeViewPlatformStore) CreateProject(_ context.Context, p domain.Project) (domain.Project, error) {
	if s.errCreateProject != nil {
		return domain.Project{}, s.errCreateProject
	}
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

func (s *fakeViewPlatformStore) UpdateProject(_ context.Context, p domain.Project) (domain.Project, error) {
	if s.errUpdateProject != nil {
		return domain.Project{}, s.errUpdateProject
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[p.ID]; !ok {
		return domain.Project{}, store.ErrNotFound
	}
	p.UpdatedAt = time.Now().UTC()
	s.projects[p.ID] = p
	return p, nil
}

func (s *fakeViewPlatformStore) ArchiveProject(_ context.Context, id, _ string) (domain.Project, error) {
	if s.errArchiveProject != nil {
		return domain.Project{}, s.errArchiveProject
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.projects[id]
	if !ok {
		return domain.Project{}, store.ErrNotFound
	}
	p.Status = domain.PlatformStatusArchived
	now := time.Now().UTC()
	p.ArchivedAt = &now
	p.UpdatedAt = now
	s.projects[id] = p
	return p, nil
}

func (s *fakeViewPlatformStore) DeleteProject(_ context.Context, id string) error {
	if s.errDeleteProject != nil {
		return s.errDeleteProject
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.projects, id)
	return nil
}

func (s *fakeViewPlatformStore) ListProjectRepositories(_ context.Context, projectID string) ([]domain.ProjectRepository, error) {
	if s.errListProjectRepositories != nil {
		return nil, s.errListProjectRepositories
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.ProjectRepository(nil), s.projectRepositories[projectID]...), nil
}

func (s *fakeViewPlatformStore) CreateProjectRepository(_ context.Context, link domain.ProjectRepository) (domain.ProjectRepository, error) {
	if s.errCreateProjectRepository != nil {
		return domain.ProjectRepository{}, s.errCreateProjectRepository
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[link.ProjectID]; !ok {
		return domain.ProjectRepository{}, store.ErrConflict
	}
	for _, existing := range s.projectRepositories[link.ProjectID] {
		if existing.RepositoryID == link.RepositoryID {
			return domain.ProjectRepository{}, store.ErrConflict
		}
	}
	if link.Role == "" {
		link.Role = "linked"
	}
	link.LinkedAt = time.Now().UTC()
	s.projectRepositories[link.ProjectID] = append(s.projectRepositories[link.ProjectID], link)
	return link, nil
}

func (s *fakeViewPlatformStore) DeleteProjectRepository(_ context.Context, projectID string, repositoryID int64) error {
	if s.errDeleteProjectRepository != nil {
		return s.errDeleteProjectRepository
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.projectRepositories[projectID]
	for i, l := range links {
		if l.RepositoryID == repositoryID {
			s.projectRepositories[projectID] = append(links[:i], links[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *fakeViewPlatformStore) UpdateProjectRepositoryWeight(_ context.Context, projectID string, repositoryID int64, weight float64) error {
	if s.errUpdateProjectRepositoryWeight != nil {
		return s.errUpdateProjectRepositoryWeight
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.projectRepositories[projectID]
	for i, l := range links {
		if l.RepositoryID == repositoryID {
			links[i].ContributionWeight = weight
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *fakeViewPlatformStore) CreateProjectWithRepositoryPayload(_ context.Context, p domain.Project, repositoryIDs []int64, repoPayload *store.RepositoryCreatePayload) (domain.Project, error) {
	if s.errCreateProjectWithRepoPayload != nil {
		return domain.Project{}, s.errCreateProjectWithRepoPayload
	}
	if repoPayload != nil {
		fullName := strings.TrimSpace(repoPayload.Slug)
		if fullName == "" {
			return domain.Project{}, store.ErrConflict
		}
	}
	// 단순화: CreateProject 호출. handler 의 audit 분기 cover 가 목표.
	created, err := s.CreateProject(context.Background(), p)
	if err != nil {
		return domain.Project{}, err
	}
	for _, rid := range repositoryIDs {
		if rid <= 0 || rid == created.RepositoryID {
			continue
		}
		_, linkErr := s.CreateProjectRepository(context.Background(), domain.ProjectRepository{
			ProjectID:    created.ID,
			RepositoryID: rid,
			Role:         "linked",
		})
		if linkErr != nil && !errors.Is(linkErr, store.ErrConflict) {
			return domain.Project{}, linkErr
		}
	}
	return created, nil
}

func (s *fakeViewPlatformStore) UpsertRepository(_ context.Context, repo domain.Repository) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repositories[repo.ProviderKey] = append(s.repositories[repo.ProviderKey], repo)
	return nil
}

func (s *fakeViewPlatformStore) ListRepositoriesByProvider(_ context.Context, providerKey string) ([]domain.Repository, error) {
	if s.errListRepositoriesByProvider != nil {
		return nil, s.errListRepositoriesByProvider
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.Repository(nil), s.repositories[providerKey]...), nil
}

// GetRepositoryByID — sprint mvs/work_260607-h-486-ci-runs-api (N-7) 시 추가.
// codex P1 review feedback 반영. view package 의 in-memory store.
func (s *fakeViewPlatformStore) GetRepositoryByID(_ context.Context, id int64) (domain.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, repos := range s.repositories {
		for _, r := range repos {
			if r.ID == id {
				return r, nil
			}
		}
	}
	return domain.Repository{}, store.ErrNotFound
}

func (s *fakeViewPlatformStore) ListRepositoryActivity(_ context.Context, repoID int64, _ store.RepositoryActivityOptions) (domain.RepositoryActivity, error) {
	return domain.RepositoryActivity{RepositoryID: repoID}, nil
}

func (s *fakeViewPlatformStore) ListRepositoryPullRequests(_ context.Context, _ int64, _ store.PRActivityListOptions) ([]domain.PRActivity, int, error) {
	return []domain.PRActivity{}, 0, nil
}

func (s *fakeViewPlatformStore) ListRepositoryBuildRuns(_ context.Context, _ int64, _ store.BuildRunListOptions) ([]domain.BuildRun, int, error) {
	if s.errListRepositoryBuildRuns != nil {
		return nil, 0, s.errListRepositoryBuildRuns
	}
	return []domain.BuildRun{}, 0, nil
}

func (s *fakeViewPlatformStore) ListRepositoryQualitySnapshots(_ context.Context, _ int64, _ store.QualitySnapshotListOptions) ([]domain.QualitySnapshot, int, error) {
	return []domain.QualitySnapshot{}, 0, nil
}

func (s *fakeViewPlatformStore) CountOpenAndMergedPRs(_ context.Context, _ int64, _, _ time.Time) (int, int, error) {
	return 0, 0, nil
}

// Sprint B — Project 가중치 rollup fake. test 가 stub 으로 mock 가능하도록
// err + return value field 노출 (기존 ComputePlatformRollup 패턴 정합).
func (s *fakeViewPlatformStore) ComputeProjectWeightedKPI(_ context.Context, projectID string, opts store.RepositoryActivityOptions) (domain.ProjectWeightedKPI, error) {
	if s.errComputeProjectWeightedKPI != nil {
		return domain.ProjectWeightedKPI{}, s.errComputeProjectWeightedKPI
	}
	return domain.ProjectWeightedKPI{
		ProjectID:              projectID,
		WindowFrom:             opts.WindowFrom,
		WindowTo:               opts.WindowTo,
		WeightedQualityScore:   85.0,
		WeightedBuildSuccess:   0.9,
		TotalBuildRunCount:     100,
		WeightedOpenPRCount:    5,
		WeightedMergedPRCount:  15,
		ActiveContributorCount: 8,
		LinkedRepositoryCount:  3,
		WeightedAt:             time.Now().UTC(),
	}, nil
}

func (s *fakeViewPlatformStore) CountProjectOpenAndMergedPRs(_ context.Context, _ string, _, _ time.Time) (int, int, error) {
	return 5, 15, nil
}

// Sprint B-Tests — Project 가중치 적용 test results fake. 0.93 pass_rate + 5 status
// (성공 145 / 실패 8 / running 1 / cancelled 2) + recent 3 row + 156 total
// (PR #597 의 RepositoryTestsSection 정공법 — view 패키지 handler test 의
// "happy path" 의 seed 가 풍부해야 component 정합 검증 가능). handler test 가
// stub 으로 override 가능.
func (s *fakeViewPlatformStore) ListProjectTestResults(_ context.Context, projectID string, opts store.BuildRunListOptions) (domain.ProjectWeightedTestResults, int, error) {
	if s.errListProjectTestResults != nil {
		return domain.ProjectWeightedTestResults{}, 0, s.errListProjectTestResults
	}
	passRate := 0.93
	return domain.ProjectWeightedTestResults{
		ProjectID:        projectID,
		WindowFrom:       opts.WindowFrom,
		WindowTo:         opts.WindowTo,
		WeightedPassRate: &passRate,
		Totals: map[string]int{
			"success": 145, "failed": 8, "running": 1, "cancelled": 2,
			"skipped": 0, "queued": 0, "unknown": 0,
		},
		Recent: []domain.ProjectBuildRun{
			{ID: 100, RepositoryID: 1, RepositoryFullName: "org/repo-a", RunExternalID: "ext-100", Branch: "main", CommitSHA: "feedface", Status: "success", StartedAt: time.Now().UTC().Add(-2 * time.Hour), FinishedAt: ptrTime(time.Now().UTC().Add(-2 * time.Hour).Add(2 * time.Minute))},
			{ID: 99, RepositoryID: 2, RepositoryFullName: "org/repo-b", RunExternalID: "ext-99", Branch: "feat/x", CommitSHA: "badfeed", Status: "failed", StartedAt: time.Now().UTC().Add(-3 * time.Hour), FinishedAt: ptrTime(time.Now().UTC().Add(-3 * time.Hour).Add(1 * time.Minute))},
			{ID: 98, RepositoryID: 1, RepositoryFullName: "org/repo-a", RunExternalID: "ext-98", Branch: "main", CommitSHA: "cafe0001", Status: "running", StartedAt: time.Now().UTC().Add(-15 * time.Minute)},
		},
	}, 156, nil
}

func ptrTime(t time.Time) *time.Time { return &t }

func (s *fakeViewPlatformStore) ComputePlatformRollup(_ context.Context, _ string, opts domain.PlatformRollupOptions) (domain.PlatformRollup, error) {
	if s.errComputePlatformRollup != nil {
		return domain.PlatformRollup{}, s.errComputePlatformRollup
	}
	if opts.Policy == "" {
		opts.Policy = domain.WeightPolicyEqual
	}
	if opts.Policy == domain.WeightPolicyCustom {
		sum := 0.0
		for _, w := range opts.CustomWeights {
			if w < 0 {
				return domain.PlatformRollup{}, errors.New("invalid weight policy: negative weight")
			}
			sum += w
		}
		if sum < 1.0-domain.CustomWeightTolerance || sum > 1.0+domain.CustomWeightTolerance {
			return domain.PlatformRollup{}, errors.New("invalid weight policy: custom weights must sum to 1.0")
		}
	}
	return domain.PlatformRollup{
		PullRequestDistribution: map[string]int{},
		Meta: domain.PlatformRollupMeta{
			Period:         domain.RollupPeriod{From: time.Now().UTC(), To: time.Now().UTC()},
			Filters:        map[string]any{},
			WeightPolicy:   opts.Policy,
			AppliedWeights: map[string]float64{},
			Fallbacks:      []domain.RollupFallback{},
			DataGaps:       []domain.RollupDataGap{},
		},
	}, nil
}

func (s *fakeViewPlatformStore) CountPlatformCriticalWarnings(_ context.Context, platformID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.criticalWarnings[platformID]; ok {
		return v, nil
	}
	return 0, nil
}

// fakeViewDevRequestStore — DevRequestStore interface 만족 (1 메서드).
type fakeViewDevRequestStore struct {
	mu      sync.Mutex
	dreqs   []domain.DevRequest
	errList error
}

func (s *fakeViewDevRequestStore) ListDevRequests(_ context.Context, _ store.DevRequestListOptions) ([]domain.DevRequest, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.errList != nil {
		return nil, 0, s.errList
	}
	return append([]domain.DevRequest(nil), s.dreqs...), len(s.dreqs), nil
}
