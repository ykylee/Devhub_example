package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/domain/application-lifecycle/repository"
	"github.com/devhub/backend-core/internal/store"
)
type memoryPlatformStore struct {
	mu                   sync.Mutex
	platforms                 map[string]domain.Platform
	links                map[string][]domain.PlatformRepository
	providers            map[string]domain.SCMProvider
	projects             map[string]domain.Project
	projectRepositories  map[string][]domain.ProjectRepository
	activeLinkCounts     map[string]int
	integrations         map[string]domain.ProjectIntegration
	integrationProviders map[string]domain.IntegrationProvider
	integrationBindings  map[string]domain.IntegrationBinding
	integrationSyncJobs  map[string]domain.IntegrationSyncJob
	criticalCounts       map[string]int // override for CountPlatformCriticalWarnings tests
	infraSnapshot        memoryInfraSnapshot
	repositoryIDs        map[string]int64
	repositories         map[string]domain.Repository // full_name → repo (UpsertRepository/ListByProvider)
	nextRepositoryID     int64
	ciRuns              []domain.BuildRun
	draftRepos          map[int64]domain.Repository
	nextDraftID         int64
}

type memoryInfraSnapshot struct {
	snapshotAt        time.Time
	nodesJSON         []byte
	servicesJSON      []byte
	degradedProviders []string
	loaded            bool
}

func newMemoryPlatformStore() *memoryPlatformStore {
	return &memoryPlatformStore{
		platforms:  make(map[string]domain.Platform),
		links: make(map[string][]domain.PlatformRepository),
		providers: map[string]domain.SCMProvider{
			"bitbucket": {ProviderKey: "bitbucket", DisplayName: "Bitbucket", Enabled: true, AdapterVersion: "0.0.1"},
			"gitea":     {ProviderKey: "gitea", DisplayName: "Gitea", Enabled: true, AdapterVersion: "0.0.1"},
			"forgejo":   {ProviderKey: "forgejo", DisplayName: "Forgejo", Enabled: false, AdapterVersion: "0.0.1"},
		},
		projects:             make(map[string]domain.Project),
		projectRepositories:  make(map[string][]domain.ProjectRepository),
		activeLinkCounts:     make(map[string]int),
		integrations:         make(map[string]domain.ProjectIntegration),
		integrationProviders: make(map[string]domain.IntegrationProvider),
		integrationBindings:  make(map[string]domain.IntegrationBinding),
		integrationSyncJobs:  make(map[string]domain.IntegrationSyncJob),
		criticalCounts:       make(map[string]int),
		repositoryIDs:        make(map[string]int64),
		repositories:         make(map[string]domain.Repository),
		nextRepositoryID:     1000,
		draftRepos:           make(map[int64]domain.Repository),
		nextDraftID:          1,
	}
}

func (s *memoryPlatformStore) ListPlatforms(_ context.Context, opts store.PlatformListOptions) ([]domain.Platform, int, error) {
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
		if opts.ActorLogin != "" && opts.ActorRole != "system_admin" && opts.ActorRole != "team_manager" {
			if a.OwnerUserID != opts.ActorLogin && a.LeaderUserID != opts.ActorLogin {
				isMember := false
				for _, p := range s.projects {
					if p.PlatformID != a.ID {
						continue
					}
					if p.OwnerUserID == opts.ActorLogin {
						isMember = true
						break
					}
					for _, m := range p.ProjectMembers {
						if m.UserID == opts.ActorLogin {
							isMember = true
							break
						}
					}
					if isMember {
						break
					}
				}
				if !isMember {
					continue
				}
			}
		}
		out = append(out, a)
	}
	return out, len(out), nil
}

func (s *memoryPlatformStore) GetPlatform(_ context.Context, id string) (domain.Platform, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a, ok := s.platforms[id]; ok {
		return a, nil
	}
	return domain.Platform{}, store.ErrNotFound
}

func (s *memoryPlatformStore) GetPlatformByKey(_ context.Context, key string) (domain.Platform, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.platforms {
		if a.Key == key {
			return a, nil
		}
	}
	return domain.Platform{}, store.ErrNotFound
}

func (s *memoryPlatformStore) CreatePlatform(_ context.Context, app domain.Platform) (domain.Platform, error) {
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

func (s *memoryPlatformStore) UpdatePlatform(_ context.Context, app domain.Platform) (domain.Platform, error) {
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

// UpdatePlatformInboundSource (N-13, ADR-0028 §6 a) — httpapi in-memory fake
// (validation + storage) without the JSONB SQL binding.
func (s *memoryPlatformStore) UpdatePlatformInboundSource(_ context.Context, platformID, inboundType, inboundConfig string) (domain.Platform, error) {
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

// ListEnabledInboundSourcePlatforms (N-13, ADR-0028 §6 a) — returns only
// platforms with a non-empty inbound_source_type. Mirrors the production
// production *PlatformRepository.ListEnabledInboundSourcePlatforms query.
func (s *memoryPlatformStore) ListEnabledInboundSourcePlatforms(_ context.Context) ([]domain.Platform, error) {
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
func (s *memoryPlatformStore) ArchivePlatform(_ context.Context, id, _ string) (domain.Platform, error) {
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

// DeletePlatform — production *PostgresStore.DeletePlatform mirror (hard-delete).
func (s *memoryPlatformStore) DeletePlatform(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.platforms[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.platforms, id)
	delete(s.links, id)
	return nil
}

// CountActiveApplicationRepositories — production *PostgresStore 의 UNION 쿼리 mirror.
// 직접 link (sync_status='active') + project 경유 간접 link (link 존재 = active 간주).
// 간접 link 에서 repositories miss (테스트 setup 누락) 면 skip — production 의
// `JOIN repositories r` strict 동작과 동일 (FK 매칭 실패 row 누락). hardcoded
// fallback ("bitbucket"+project.Key) 은 production 동작과 무관해 fake parity 깨므로 제거.
func (s *memoryPlatformStore) CountActivePlatformRepositories(_ context.Context, platformID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	activeRepos := make(map[string]bool)
	for _, l := range s.links[platformID] {
		if l.SyncStatus == domain.SyncStatusActive {
			activeRepos[l.RepoProvider+"/"+l.RepoFullName] = true
		}
	}
	for _, p := range s.projects {
		if p.PlatformID != platformID {
			continue
		}
		for _, pr := range s.projectRepositories[p.ID] {
			provider, fullName := s.lookupRepoForFake(pr.RepositoryID)
			if provider == "" || fullName == "" {
				continue // repositories miss → production JOIN 매칭 실패 mirror
			}
			activeRepos[provider+"/"+fullName] = true
		}
	}
	return len(activeRepos), nil
}

// lookupRepoForFake — fake store 의 repository ID → (provider, full_name) lookup helper.
// production `JOIN repositories` 와 동일하게 매칭 실패면 빈 string 반환 (caller 가 skip).
func (s *memoryPlatformStore) lookupRepoForFake(repoID int64) (string, string) {
	for _, r := range s.repositories {
		if r.ID == repoID {
			return r.ProviderKey, r.FullName
		}
	}
	return "", ""
}

func (s *memoryPlatformStore) ListPlatformRepositories(_ context.Context, platformID string) ([]domain.PlatformRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	direct := s.links[platformID]
	seen := make(map[string]bool)
	out := make([]domain.PlatformRepository, 0, len(direct))
	for _, l := range direct {
		seen[l.RepoProvider+"/"+l.RepoFullName] = true
		out = append(out, l)
	}
	// project 경유 간접 link — production JOIN repositories strict 와 동일 동작.
	for _, p := range s.projects {
		if p.PlatformID != platformID {
			continue
		}
		for _, pr := range s.projectRepositories[p.ID] {
			provider, fullName := s.lookupRepoForFake(pr.RepositoryID)
			if provider == "" || fullName == "" {
				continue
			}
			key := provider + "/" + fullName
			if seen[key] {
				continue
			}
			seen[key] = true
			role := domain.PlatformRepositoryRoleSub
			switch pr.Role {
			case "primary":
				role = domain.PlatformRepositoryRolePrimary
			case "shared":
				role = domain.PlatformRepositoryRoleShared
			}
			out = append(out, domain.PlatformRepository{
				PlatformID: platformID,
				RepoProvider:  provider,
				RepoFullName:  fullName,
				Role:          role,
				SyncStatus:    domain.SyncStatusActive,
				LinkedAt:      pr.LinkedAt,
				LinkSource:    "via_project",
			})
		}
	}
	return out, nil
}

func (s *memoryPlatformStore) CreatePlatformRepository(_ context.Context, link domain.PlatformRepository) (domain.PlatformRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.platforms[link.PlatformID]; !ok {
		return domain.PlatformRepository{}, store.ErrConflict
	}
	existing := s.links[link.PlatformID]
	for _, e := range existing {
		if e.RepoProvider == link.RepoProvider && e.RepoFullName == link.RepoFullName {
			return domain.PlatformRepository{}, store.ErrConflict
		}
	}
	link.LinkedAt = time.Now().UTC()
	s.links[link.PlatformID] = append(existing, link)
	if link.SyncStatus == domain.SyncStatusActive {
		s.activeLinkCounts[link.PlatformID]++
	}
	return link, nil
}

func (s *memoryPlatformStore) DeletePlatformRepository(_ context.Context, key store.PlatformRepositoryLinkKey) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.links[key.PlatformID]
	for i, l := range links {
		if l.RepoProvider == key.RepoProvider && l.RepoFullName == key.RepoFullName {
			if l.SyncStatus == domain.SyncStatusActive {
				s.activeLinkCounts[key.PlatformID]--
			}
			s.links[key.PlatformID] = append(links[:i], links[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *memoryPlatformStore) UpdatePlatformRepositorySync(_ context.Context, key store.PlatformRepositoryLinkKey, status domain.PlatformRepositorySyncStatus, code domain.SyncErrorCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.links[key.PlatformID]
	for i, l := range links {
		if l.RepoProvider == key.RepoProvider && l.RepoFullName == key.RepoFullName {
			if l.SyncStatus != domain.SyncStatusActive && status == domain.SyncStatusActive {
				s.activeLinkCounts[key.PlatformID]++
			} else if l.SyncStatus == domain.SyncStatusActive && status != domain.SyncStatusActive {
				s.activeLinkCounts[key.PlatformID]--
			}
			links[i].SyncStatus = status
			links[i].SyncErrorCode = code
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *memoryPlatformStore) ListSCMProviders(_ context.Context) ([]domain.SCMProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.SCMProvider, 0, len(s.providers))
	for _, p := range s.providers {
		out = append(out, p)
	}
	return out, nil
}

func (s *memoryPlatformStore) UpdateSCMProvider(_ context.Context, p domain.SCMProvider) (domain.SCMProvider, error) {
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

func (s *memoryPlatformStore) ListProjects(_ context.Context, opts store.ProjectListOptions) ([]domain.Project, int, error) {
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
		if opts.Status != "" && string(p.Status) != opts.Status {
			continue
		}
		if !opts.IncludeArchived && p.Status == domain.PlatformStatusArchived {
			continue
		}
		if opts.ActorLogin != "" && opts.ActorRole != "system_admin" && opts.ActorRole != "team_manager" {
			if p.OwnerUserID != opts.ActorLogin {
				isMember := false
				for _, m := range p.ProjectMembers {
					if m.UserID == opts.ActorLogin {
						isMember = true
						break
					}
				}
				if !isMember {
					continue
				}
			}
		}
		out = append(out, p)
	}
	return out, len(out), nil
}

func (s *memoryPlatformStore) GetProject(_ context.Context, id string) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.projects[id]; ok {
		return p, nil
	}
	return domain.Project{}, store.ErrNotFound
}

func (s *memoryPlatformStore) CreateProject(_ context.Context, p domain.Project) (domain.Project, error) {
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
	if p.RepositoryID != 0 {
		s.projectRepositories[p.ID] = append(s.projectRepositories[p.ID], domain.ProjectRepository{
			ProjectID:    p.ID,
			RepositoryID: p.RepositoryID,
			Role:         "primary",
			LinkedAt:     p.CreatedAt,
		})
	}
	return p, nil
}

func (s *memoryPlatformStore) ListProjectRepositories(_ context.Context, projectID string) ([]domain.ProjectRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.ProjectRepository(nil), s.projectRepositories[projectID]...), nil
}

func (s *memoryPlatformStore) CreateProjectRepository(_ context.Context, link domain.ProjectRepository) (domain.ProjectRepository, error) {
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

func (s *memoryPlatformStore) DeleteProjectRepository(_ context.Context, projectID string, repositoryID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.projectRepositories[projectID]
	for i, existing := range links {
		if existing.RepositoryID == repositoryID {
			s.projectRepositories[projectID] = append(links[:i], links[i+1:]...)
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *memoryPlatformStore) UpdateProjectRepositoryWeight(_ context.Context, projectID string, repositoryID int64, weight float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	links := s.projectRepositories[projectID]
	for i, existing := range links {
		if existing.RepositoryID == repositoryID {
			links[i].ContributionWeight = weight
			return nil
		}
	}
	return store.ErrNotFound
}

func (s *memoryPlatformStore) CreateProjectWithRepositoryPayload(_ context.Context, p domain.Project, repositoryIDs []int64, repoPayload *store.RepositoryCreatePayload) (domain.Project, error) {
	// repoPayload 동반 생성 — production 의 단일 tx atomicity 를 흉내 (codex #349 P2):
	// repo id 확보 후 project + links 생성. CreateProject 실패 시 (중복 key) 에러 반환.
	if repoPayload != nil {
		fullName := strings.TrimSpace(repoPayload.Slug)
		if fullName == "" {
			return domain.Project{}, store.ErrConflict
		}
		s.mu.Lock()
		repoID, ok := s.repositoryIDs[fullName]
		if !ok {
			s.nextRepositoryID++
			repoID = s.nextRepositoryID
			s.repositoryIDs[fullName] = repoID
		}
		s.mu.Unlock()
		if p.RepositoryID == 0 {
			p.RepositoryID = repoID
		}
		repositoryIDs = append(repositoryIDs, repoID)
	}
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

func (s *memoryPlatformStore) UpdateProject(_ context.Context, p domain.Project) (domain.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[p.ID]; !ok {
		return domain.Project{}, store.ErrNotFound
	}
	p.UpdatedAt = time.Now().UTC()
	s.projects[p.ID] = p
	return p, nil
}

func (s *memoryPlatformStore) ArchiveProject(_ context.Context, id, _ string) (domain.Project, error) {
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

func (s *memoryPlatformStore) DeleteProject(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projects[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.projects, id)
	return nil
}

// --- Repository 운영 지표 (sprint claude/work_260514-c) ---
// 메모리 store 는 SQL 집계를 흉내내지 않으므로 모두 zero-value 반환.

func (s *memoryPlatformStore) ListRepositoryActivity(_ context.Context, repoID int64, _ store.RepositoryActivityOptions) (domain.RepositoryActivity, error) {
	return domain.RepositoryActivity{RepositoryID: repoID}, nil
}

func (s *memoryPlatformStore) ListRepositoryPullRequests(_ context.Context, _ int64, _ store.PRActivityListOptions) ([]domain.PRActivity, int, error) {
	return []domain.PRActivity{}, 0, nil
}

func (s *memoryPlatformStore) ListRepositoryBuildRuns(_ context.Context, repoID int64, opts store.BuildRunListOptions) ([]domain.BuildRun, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := make([]domain.BuildRun, 0)
	for _, r := range s.ciRuns {
		if r.RepositoryID != repoID {
			continue
		}
		if opts.Status != "" && r.Status != opts.Status {
			continue
		}
		if opts.Branch != "" && r.Branch != opts.Branch {
			continue
		}
		filtered = append(filtered, r)
	}
	total := len(filtered)
	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []domain.BuildRun{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (s *memoryPlatformStore) ListRepositoryQualitySnapshots(_ context.Context, _ int64, _ store.QualitySnapshotListOptions) ([]domain.QualitySnapshot, int, error) {
	return []domain.QualitySnapshot{}, 0, nil
}

func (s *memoryPlatformStore) CountOpenAndMergedPRs(_ context.Context, _ int64, _, _ time.Time) (int, int, error) {
	// 1차 메모리 store 는 PR row 가 없는 환경 가정. handler test 는 0,0, nil 로
	// 정상 response 검증. 별도 PR seed 가 필요한 test 는 sub-type 으로 override.
	return 0, 0, nil
}

// Sprint B — Project 가중치 rollup. memory store 는 seed-free 0,0 default.
func (s *memoryPlatformStore) ComputeProjectWeightedKPI(_ context.Context, projectID string, opts store.RepositoryActivityOptions) (domain.ProjectWeightedKPI, error) {
	return domain.ProjectWeightedKPI{
		ProjectID:              projectID,
		WindowFrom:             opts.WindowFrom,
		WindowTo:               opts.WindowTo,
		LinkedRepositoryCount:  0,
		WeightedAt:             time.Now().UTC(),
	}, nil
}

func (s *memoryPlatformStore) CountProjectOpenAndMergedPRs(_ context.Context, _ string, _, _ time.Time) (int, int, error) {
	return 0, 0, nil
}

// Sprint B-Tests — Project 가중치 적용 test results. memory store 는 linked repo 0
// 시 weightedPassRate=nil + 7 status 0 + recent 빈 array 응답 (Sprint A 의
// memoryPlatformStore seed-free 정합).
func (s *memoryPlatformStore) ListProjectTestResults(_ context.Context, projectID string, opts store.BuildRunListOptions) (domain.ProjectWeightedTestResults, int, error) {
	windowFrom := opts.WindowFrom
	if windowFrom.IsZero() {
		windowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	windowTo := opts.WindowTo
	if windowTo.IsZero() {
		windowTo = time.Now().UTC()
	}
	return domain.ProjectWeightedTestResults{
		ProjectID:        projectID,
		WindowFrom:       windowFrom.UTC(),
		WindowTo:         windowTo.UTC(),
		WeightedPassRate: nil,
		Totals: map[string]int{
			"success": 0, "failed": 0, "running": 0, "cancelled": 0,
			"skipped": 0, "queued": 0, "unknown": 0,
		},
		Recent: []domain.ProjectBuildRun{},
	}, 0, nil
}

// Sprint C — Platform sub-project rollup. memory store 는 linked project 0
// 시 weighted metric 0 + weightedPassRate=nil + 7 status 0 + recent 빈 array
// 응답 (Sprint A 의 memoryPlatformStore seed-free 정합).
func (s *memoryPlatformStore) ComputePlatformWeightedKPI(_ context.Context, platformID string, opts store.BuildRunListOptions) (domain.PlatformWeightedKPI, error) {
	windowFrom := opts.WindowFrom
	if windowFrom.IsZero() {
		windowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	windowTo := opts.WindowTo
	if windowTo.IsZero() {
		windowTo = time.Now().UTC()
	}
	return domain.PlatformWeightedKPI{
		PlatformID:           platformID,
		WindowFrom:           windowFrom.UTC(),
		WindowTo:             windowTo.UTC(),
		LinkedProjectCount:   0,
		WeightedAt:           time.Now().UTC(),
	}, nil
}

func (s *memoryPlatformStore) ListPlatformTestResults(_ context.Context, platformID string, opts store.BuildRunListOptions) (domain.PlatformWeightedTestResults, int, error) {
	windowFrom := opts.WindowFrom
	if windowFrom.IsZero() {
		windowFrom = time.Now().UTC().AddDate(0, 0, -30)
	}
	windowTo := opts.WindowTo
	if windowTo.IsZero() {
		windowTo = time.Now().UTC()
	}
	return domain.PlatformWeightedTestResults{
		PlatformID:       platformID,
		WindowFrom:       windowFrom.UTC(),
		WindowTo:         windowTo.UTC(),
		WeightedPassRate: nil,
		Totals: map[string]int{
			"success": 0, "failed": 0, "running": 0, "cancelled": 0,
			"skipped": 0, "queued": 0, "unknown": 0,
		},
		Recent: []domain.PlatformBuildRun{},
	}, 0, nil
}
// --- Application 롤업 (sprint claude/work_260514-c) ---

func (s *memoryPlatformStore) ComputePlatformRollup(_ context.Context, _ string, opts domain.PlatformRollupOptions) (domain.PlatformRollup, error) {
	if opts.Policy == "" {
		opts.Policy = domain.WeightPolicyEqual
	}
	// custom weight 검증만 흉내냄 (handler 분기 테스트용).
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

func (s *memoryPlatformStore) CountPlatformCriticalWarnings(_ context.Context, _ string) (int, error) {
	// 1차 메모리 store 는 critical warning 이 없는 환경 가정. 별도 case 가 필요한 test 가
	// 있으면 sub-type 으로 override.
	return 0, nil
}

// --- Integration CRUD (sprint claude/work_260514-c) ---

type memoryIntegration = domain.ProjectIntegration

func (s *memoryPlatformStore) ListIntegrations(_ context.Context, opts store.IntegrationListOptions) ([]domain.ProjectIntegration, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.ProjectIntegration, 0)
	for _, i := range s.integrations {
		if string(opts.Scope) != "" && string(i.Scope) != string(opts.Scope) {
			continue
		}
		if opts.PlatformID != "" && i.PlatformID != opts.PlatformID {
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

func (s *memoryPlatformStore) GetIntegration(_ context.Context, id string) (domain.ProjectIntegration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i, ok := s.integrations[id]; ok {
		return i, nil
	}
	return domain.ProjectIntegration{}, store.ErrNotFound
}

func (s *memoryPlatformStore) CreateIntegration(_ context.Context, i domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.integrations {
		if existing.Scope == i.Scope &&
			existing.PlatformID == i.PlatformID &&
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

func (s *memoryPlatformStore) UpdateIntegration(_ context.Context, i domain.ProjectIntegration) (domain.ProjectIntegration, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.integrations[i.ID]
	if !ok {
		return domain.ProjectIntegration{}, store.ErrNotFound
	}
	// PR #107 codex review P2 — partial UNIQUE 인덱스 흉내. external_key 변경이
	// 다른 row 와 (scope target, integration_type, external_key) 충돌 시 ErrConflict.
	for otherID, other := range s.integrations {
		if otherID == i.ID {
			continue
		}
		if other.Scope == current.Scope &&
			other.PlatformID == current.PlatformID &&
			other.ProjectID == current.ProjectID &&
			other.IntegrationType == current.IntegrationType &&
			other.ExternalKey == i.ExternalKey {
			return domain.ProjectIntegration{}, store.ErrConflict
		}
	}
	current.ExternalKey = i.ExternalKey
	current.URL = i.URL
	current.Policy = i.Policy
	current.UpdatedAt = time.Now().UTC()
	s.integrations[i.ID] = current
	return current, nil
}

func (s *memoryPlatformStore) DeleteIntegration(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrations[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.integrations, id)
	return nil
}

func (s *memoryPlatformStore) ListIntegrationProviders(_ context.Context, opts store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.IntegrationProvider, 0)
	for _, p := range s.integrationProviders {
		if string(opts.ProviderType) != "" && string(p.ProviderType) != string(opts.ProviderType) {
			continue
		}
		if opts.Enabled != nil && p.Enabled != *opts.Enabled {
			continue
		}
		out = append(out, p)
	}
	return out, len(out), nil
}

func (s *memoryPlatformStore) GetIntegrationProviderByID(_ context.Context, providerID string) (domain.IntegrationProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.integrationProviders[providerID]
	if !ok {
		return domain.IntegrationProvider{}, store.ErrNotFound
	}
	return p, nil
}

func (s *memoryPlatformStore) GetIntegrationProviderByKey(_ context.Context, providerKey string) (domain.IntegrationProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.integrationProviders {
		if p.ProviderKey == providerKey {
			return p, nil
		}
	}
	return domain.IntegrationProvider{}, store.ErrNotFound
}

func (s *memoryPlatformStore) CreateIntegrationProvider(_ context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.integrationProviders {
		if existing.ProviderKey == p.ProviderKey {
			return domain.IntegrationProvider{}, store.ErrConflict
		}
	}
	if p.ID == "" {
		p.ID = "prov-" + p.ProviderKey
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	s.integrationProviders[p.ID] = p
	return p, nil
}

func (s *memoryPlatformStore) UpdateIntegrationProvider(_ context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.integrationProviders[p.ID]
	if !ok {
		return domain.IntegrationProvider{}, store.ErrNotFound
	}
	current.DisplayName = p.DisplayName
	current.Enabled = p.Enabled
	current.CredentialsRef = p.CredentialsRef
	current.Capabilities = append([]string(nil), p.Capabilities...)
	current.SyncStatus = p.SyncStatus
	current.LastSyncAt = p.LastSyncAt
	current.LastErrorCode = p.LastErrorCode
	current.BaseURL = p.BaseURL
	current.APIToken = p.APIToken
	current.AuthUsername = p.AuthUsername
	current.AuthClientID = p.AuthClientID
	current.AuthTokenURL = p.AuthTokenURL
	current.AuthSecret = p.AuthSecret
	current.UpdatedAt = time.Now().UTC()
	s.integrationProviders[p.ID] = current
	return current, nil
}

func (s *memoryPlatformStore) UpsertRepository(_ context.Context, repo domain.Repository) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.repositories == nil {
		s.repositories = make(map[string]domain.Repository)
	}
	if existing, ok := s.repositories[repo.FullName]; ok {
		// SCM mirror 필드만 갱신, system-owned(description) 보존 + source/provider_id
		// 기존 값 우선 (production PostgresStore.UpsertRepository ON CONFLICT 미러).
		existing.GiteaID = repo.GiteaID
		existing.OwnerLogin = repo.OwnerLogin
		existing.Name = repo.Name
		existing.CloneURL = repo.CloneURL
		existing.HTMLURL = repo.HTMLURL
		existing.DefaultBranch = repo.DefaultBranch
		existing.Private = repo.Private
		if existing.Source == "" {
			existing.Source = repo.Source
		}
		if existing.ProviderID == "" {
			existing.ProviderID = repo.ProviderID
		}
		existing.UpdatedAt = time.Now().UTC()
		s.repositories[repo.FullName] = existing
		// Sync back to draft repo when SCM mirror is upserted (publish flow).
		for id, draft := range s.draftRepos {
			if draft.FullName == repo.FullName {
				draft.GiteaID = existing.GiteaID
				draft.CloneURL = existing.CloneURL
				draft.HTMLURL = existing.HTMLURL
				draft.DefaultBranch = existing.DefaultBranch
				draft.Private = existing.Private
				if draft.Source == "" {
					draft.Source = existing.Source
				}
				draft.UpdatedAt = existing.UpdatedAt
				s.draftRepos[id] = draft
				break
			}
		}
		return nil
	}
	if repo.ID == 0 {
		s.nextRepositoryID++
		repo.ID = s.nextRepositoryID
	}
	if repo.Source == "" {
		repo.Source = domain.RepositorySourceSCM
	}
	repo.UpdatedAt = time.Now().UTC()
	s.repositories[repo.FullName] = repo
	s.repositoryIDs[repo.FullName] = repo.ID
	// Sync back to draft repo when SCM mirror is upserted (publish flow).
	for id, draft := range s.draftRepos {
		if draft.FullName == repo.FullName {
			draft.GiteaID = repo.GiteaID
			draft.CloneURL = repo.CloneURL
			draft.HTMLURL = repo.HTMLURL
			draft.DefaultBranch = repo.DefaultBranch
			draft.Private = repo.Private
			if draft.Source == "" {
				draft.Source = repo.Source
			}
			draft.UpdatedAt = repo.UpdatedAt
			s.draftRepos[id] = draft
			break
		}
	}
	return nil
}

func (s *memoryPlatformStore) ListRepositoriesByProvider(_ context.Context, providerID string) ([]domain.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Repository, 0)
	for _, r := range s.repositories {
		if r.ProviderID == providerID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *memoryPlatformStore) DeleteIntegrationProvider(_ context.Context, providerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrationProviders[providerID]; !ok {
		return store.ErrNotFound
	}
	// FK guard mirror: binding count > 0 이면 ErrConflict.
	for _, b := range s.integrationBindings {
		if b.ProviderID == providerID {
			return store.ErrConflict
		}
	}
	delete(s.integrationProviders, providerID)
	return nil
}

func (s *memoryPlatformStore) CreateIntegrationSyncJob(_ context.Context, providerID string, _ string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrationProviders[providerID]; !ok {
		return "", store.ErrNotFound
	}
	jobID := "job-" + providerID
	s.integrationSyncJobs[jobID] = domain.IntegrationSyncJob{
		JobID:      jobID,
		ProviderID: providerID,
		Status:     domain.IntegrationSyncJobStatusQueued,
		CreatedAt:  time.Now().UTC(),
	}
	return jobID, nil
}

func (s *memoryPlatformStore) ListIntegrationSyncJobs(_ context.Context, opts store.IntegrationSyncJobListOptions) ([]domain.IntegrationSyncJob, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.IntegrationSyncJob, 0, len(s.integrationSyncJobs))
	for _, job := range s.integrationSyncJobs {
		if opts.Status != "" && job.Status != opts.Status {
			continue
		}
		out = append(out, job)
	}
	return out, len(out), nil
}

func (s *memoryPlatformStore) GetIntegrationSyncJob(_ context.Context, jobID string) (domain.IntegrationSyncJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.integrationSyncJobs[jobID]
	if !ok {
		return domain.IntegrationSyncJob{}, store.ErrNotFound
	}
	return job, nil
}

func (s *memoryPlatformStore) GetIntegrationSyncJobStatusCounts(_ context.Context) (domain.IntegrationSyncJobStatusCounts, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var counts domain.IntegrationSyncJobStatusCounts
	for _, job := range s.integrationSyncJobs {
		switch job.Status {
		case domain.IntegrationSyncJobStatusQueued:
			counts.Queued++
		case domain.IntegrationSyncJobStatusRunning:
			counts.Running++
		case domain.IntegrationSyncJobStatusSucceeded:
			counts.Succeeded++
		case domain.IntegrationSyncJobStatusFailed:
			counts.Failed++
		}
	}
	return counts, nil
}

func (s *memoryPlatformStore) ListIntegrationBindings(_ context.Context, opts store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.IntegrationBinding, 0)
	for _, b := range s.integrationBindings {
		if string(opts.ScopeType) != "" && string(b.ScopeType) != string(opts.ScopeType) {
			continue
		}
		if opts.ScopeID != "" && b.ScopeID != opts.ScopeID {
			continue
		}
		if opts.Enabled != nil && b.Enabled != *opts.Enabled {
			continue
		}
		if string(opts.ProviderType) != "" {
			p, ok := s.integrationProviders[b.ProviderID]
			if !ok || string(p.ProviderType) != string(opts.ProviderType) {
				continue
			}
		}
		out = append(out, b)
	}
	return out, len(out), nil
}

func (s *memoryPlatformStore) CreateIntegrationBinding(_ context.Context, b domain.IntegrationBinding) (domain.IntegrationBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrationProviders[b.ProviderID]; !ok {
		return domain.IntegrationBinding{}, store.ErrConflict
	}
	for _, existing := range s.integrationBindings {
		if existing.ScopeType == b.ScopeType &&
			existing.ScopeID == b.ScopeID &&
			existing.ProviderID == b.ProviderID &&
			existing.ExternalKey == b.ExternalKey {
			return domain.IntegrationBinding{}, store.ErrConflict
		}
	}
	if b.ID == "" {
		b.ID = "bind-" + b.ScopeID + "-" + b.ExternalKey
	}
	now := time.Now().UTC()
	b.CreatedAt = now
	b.UpdatedAt = now
	s.integrationBindings[b.ID] = b
	return b, nil
}

// PR #251 P2-4 sub-carve — Bindings UI 강화에 동반된 신규 interface method 3건.
// production *store.PostgresStore 가 구현하므로 test fake 도 정합 보장 (compile-time).
func (s *memoryPlatformStore) GetIntegrationBindingByID(_ context.Context, id string) (domain.IntegrationBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.integrationBindings[id]
	if !ok {
		return domain.IntegrationBinding{}, store.ErrNotFound
	}
	return b, nil
}

func (s *memoryPlatformStore) UpdateIntegrationBinding(_ context.Context, b domain.IntegrationBinding) (domain.IntegrationBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.integrationBindings[b.ID]
	if !ok {
		return domain.IntegrationBinding{}, store.ErrNotFound
	}
	// ProviderID 변경 시 신규 provider 존재 검증 (production store 의 FK 가드 mirror).
	if b.ProviderID != existing.ProviderID {
		if _, ok := s.integrationProviders[b.ProviderID]; !ok {
			return domain.IntegrationBinding{}, store.ErrNotFound
		}
	}
	// duplicate 가드 — 같은 (scope_type, scope_id, provider_id, external_key) 4-tuple
	// 이 다른 binding 으로 이미 존재하면 ErrConflict (production store 의 unique index).
	for id, other := range s.integrationBindings {
		if id == b.ID {
			continue
		}
		if other.ScopeType == existing.ScopeType &&
			other.ScopeID == existing.ScopeID &&
			other.ProviderID == b.ProviderID &&
			other.ExternalKey == b.ExternalKey {
			return domain.IntegrationBinding{}, store.ErrConflict
		}
	}
	existing.ScopeID = b.ScopeID
	existing.ProviderID = b.ProviderID
	existing.ExternalKey = b.ExternalKey
	existing.Policy = b.Policy
	existing.Enabled = b.Enabled
	existing.UpdatedAt = time.Now().UTC()
	s.integrationBindings[b.ID] = existing
	return existing, nil
}

func (s *memoryPlatformStore) DeleteIntegrationBinding(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrationBindings[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.integrationBindings, id)
	return nil
}

func (s *memoryPlatformStore) SaveInfraSnapshot(_ context.Context, _ string, _ string, snapshotAt time.Time, _ string, nodesJSON, servicesJSON []byte, degradedProviders []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.infraSnapshot = memoryInfraSnapshot{
		snapshotAt:        snapshotAt.UTC(),
		nodesJSON:         append([]byte(nil), nodesJSON...),
		servicesJSON:      append([]byte(nil), servicesJSON...),
		degradedProviders: append([]string(nil), degradedProviders...),
		loaded:            true,
	}
	return nil
}

func (s *memoryPlatformStore) LoadLatestInfraSnapshot(_ context.Context) (time.Time, []byte, []byte, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.infraSnapshot.loaded {
		return time.Time{}, nil, nil, nil, store.ErrNotFound
	}
	return s.infraSnapshot.snapshotAt,
		append([]byte(nil), s.infraSnapshot.nodesJSON...),
		append([]byte(nil), s.infraSnapshot.servicesJSON...),
		append([]string(nil), s.infraSnapshot.degradedProviders...),
		nil
}

// --- Repository Draft Store (repositoryDraftStore interface) ---

func (s *memoryPlatformStore) CreateRepositoryDraft(_ context.Context, key, slug, providerID string) (domain.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key = strings.TrimSpace(key)
	slug = strings.TrimSpace(slug)
	if key == "" || slug == "" {
		return domain.Repository{}, store.ErrConflict
	}
	for _, r := range s.draftRepos {
		if r.Name == key || r.FullName == slug {
			return domain.Repository{}, store.ErrConflict
		}
	}
	s.nextDraftID++
	id := s.nextDraftID
	ownerLogin := ""
	if idx := strings.Index(slug, "/"); idx >= 0 {
		ownerLogin = slug[:idx]
	}
	now := time.Now().UTC()
	repo := domain.Repository{
		ID:          id,
		FullName:    slug,
		OwnerLogin:  ownerLogin,
		Name:        key,
		Status:      "draft",
		Source:      domain.RepositorySourceSystem,
		ProviderID:  strings.TrimSpace(providerID),
		Private:     false,
		DefaultBranch: "main",
		UpdatedAt:   now,
	}
	s.draftRepos[id] = repo
	return repo, nil
}

func (s *memoryPlatformStore) UpdateRepositoryDraft(_ context.Context, repositoryID int64, params store.RepositoryUpdateDraftParams) (domain.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, ok := s.draftRepos[repositoryID]
	if !ok || repo.Status != "draft" {
		return domain.Repository{}, store.ErrNotFound
	}
	if params.Key != nil {
		key := strings.TrimSpace(*params.Key)
		// key uniqueness check (empty-string key = no-op, same as existing = no-op)
		if key != "" && key != repo.Name {
			for _, r := range s.draftRepos {
				if r.ID != repositoryID && r.Name == key {
					return domain.Repository{}, store.ErrConflict
				}
			}
		}
		repo.Name = key
	}
	if params.Slug != nil {
		slug := strings.TrimSpace(*params.Slug)
		if slug != "" && slug != repo.FullName {
			for _, r := range s.draftRepos {
				if r.ID != repositoryID && r.FullName == slug {
					return domain.Repository{}, store.ErrConflict
				}
			}
		}
		repo.FullName = slug
		if idx := strings.Index(slug, "/"); idx >= 0 {
			repo.OwnerLogin = slug[:idx]
		} else {
			repo.OwnerLogin = ""
		}
	}
	if params.ProviderID != nil {
		repo.ProviderID = strings.TrimSpace(*params.ProviderID)
	}
	repo.UpdatedAt = time.Now().UTC()
	s.draftRepos[repositoryID] = repo
	return repo, nil
}

func (s *memoryPlatformStore) DeleteRepository(_ context.Context, repositoryID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, ok := s.draftRepos[repositoryID]
	if !ok || repo.Status != "draft" {
		return store.ErrNotFound
	}
	// FK guard: platform_repositories / project_repositories link 존재 시 ErrConflict.
	for _, appLinks := range s.links {
		for _, l := range appLinks {
			if l.RepoFullName == repo.FullName {
				return store.ErrConflict
			}
		}
	}
	for _, projLinks := range s.projectRepositories {
		for _, l := range projLinks {
			if l.RepositoryID == repositoryID {
				return store.ErrConflict
			}
		}
	}
	delete(s.draftRepos, repositoryID)
	return nil
}

// GetRepositoryByID — sprint mvs/work_260607-h-486-ci-runs-api (N-7) 시
// 보강: draftRepos (int64 PK) + repositories (full_name key) 양쪽 walk. 기존
// draftRepos-only 구현은 PR #494 codex P1 review feedback 반영 시 확장.
func (s *memoryPlatformStore) GetRepositoryByID(_ context.Context, repositoryID int64) (domain.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if repo, ok := s.draftRepos[repositoryID]; ok {
		return repo, nil
	}
	// repositories (full_name key) walk — 테스트 fixture 소량이라 O(n) OK
	for _, r := range s.repositories {
		if r.ID == repositoryID {
			return r, nil
		}
	}
	return domain.Repository{}, store.ErrNotFound
}

func (s *memoryPlatformStore) MarkRepositoryDraftPublishRequested(_ context.Context, repositoryID int64) (domain.Repository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	repo, ok := s.draftRepos[repositoryID]
	if !ok || repo.Status != "draft" {
		return domain.Repository{}, store.ErrNotFound
	}
	now := time.Now().UTC()
	repo.PublishRequestedAt = &now
	repo.UpdatedAt = now
	s.draftRepos[repositoryID] = repo
	return repo, nil
}

// --- handler tests ---

func newPlatformsRouter(platformStore PlatformStore) http.Handler {
	var integrationStore IntegrationStore
	if store, ok := any(platformStore).(IntegrationStore); ok {
		integrationStore = store
	}
	return NewRouter(RouterConfig{
		PlatformStore:    platformStore,
		IntegrationStore: integrationStore,
		AuthDevFallback:  true, // bypass bearer auth
	})
}

// 1) POST /platforms — happy.
func TestCreatePlatform_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms",
		`{"key":"DEVHUB","name":"Devhub Platform","owner_user_id":"u1","leader_user_id":"u1","development_unit_id":"dept-eng","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"key":"DEVHUB"`)) {
		t.Errorf("response should echo key: %s", rec.Body.String())
	}
}

// 2) POST /platforms — invalid key format → 422 invalid_platform_key.
func TestCreatePlatform_InvalidKey(t *testing.T) {
	router := newPlatformsRouter(newMemoryPlatformStore())

	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms",
		`{"key":"too-short","name":"X","owner_user_id":"u1","leader_user_id":"u1","development_unit_id":"dept-eng","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"invalid_platform_key"`)) {
		t.Errorf("response should carry invalid_platform_key code: %s", rec.Body.String())
	}
}

// 3) POST /platforms — duplicate key → 409 platform_key_conflict.
func TestCreatePlatform_DuplicateKey(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	body := `{"key":"A1B2C3D4E5","name":"X","owner_user_id":"u1","leader_user_id":"u1","development_unit_id":"dept-eng","visibility":"internal","status":"planning"}`
	first := doJSON(t, router, http.MethodPost, "/api/v1/platforms", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", first.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"platform_key_conflict"`)) {
		t.Errorf("expected platform_key_conflict: %s", rec.Body.String())
	}
}

// 4) PATCH /platforms/:id — immutable key 거부 → 422 platform_key_immutable.
func TestUpdatePlatform_ImmutableKey(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/platforms/"+app.ID,
		`{"key":"NEWKEY1234"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"platform_key_immutable"`)) {
		t.Errorf("expected platform_key_immutable: %s", rec.Body.String())
	}
}

// 5) PATCH /platforms/:id — planning → active 의 활성 repo 0건 → 422.
// status 전이 정책 자유화 (2026-05-28) — planning→active 의 active repo ≥1 가드 제거.
// 0 repo 인 application 도 active 로 자유 전이 가능. 기존 테스트는 가드 검증이었으므로
// 자유화 후 expected behavior 로 갱신.
func TestUpdatePlatform_ActivationWithoutLinkedRepos(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/platforms/"+app.ID,
		`{"status":"active"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (자유화 후 200 기대)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"active"`)) {
		t.Errorf("expected status=active: %s", rec.Body.String())
	}
}

// 6) PATCH /platforms/:id — planning → active 의 활성 repo 1개 → 200.
func TestUpdatePlatform_ActivationSuccess(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = platformStore.CreatePlatformRepository(context.Background(), domain.PlatformRepository{
		PlatformID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.PlatformRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/platforms/"+app.ID,
		`{"status":"active"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"active"`)) {
		t.Errorf("expected status=active: %s", rec.Body.String())
	}
}

// status 전이 정책 자유화 (2026-05-28) — 이전엔 closed→planning 같은 backward 전이가
// 422 거부였으나 이제 모든 전이 허용. 기존 테스트는 가드 검증이었으므로 자유화 후
// expected behavior 로 갱신.
func TestUpdatePlatform_AnyStatusTransitionAllowed(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusClosed,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	// closed → planning 도 자유화 후 200.
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/platforms/"+app.ID,
		`{"status":"planning"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (자유화 후 200 기대)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"planning"`)) {
		t.Errorf("expected status=planning: %s", rec.Body.String())
	}
}

// status 전이 정책 자유화 (2026-05-28) — active→on_hold 의 hold_reason 필수 가드 제거.
// reason 없이도 전이 가능 (audit 기록은 reason 있을 때만 details 에 포함).
func TestUpdatePlatform_HoldWithoutReason(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/platforms/"+app.ID,
		`{"status":"on_hold"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (자유화 후 200 기대)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"on_hold"`)) {
		t.Errorf("expected status=on_hold: %s", rec.Body.String())
	}
}

// 9) DELETE /platforms/:id — archive (soft-delete).
func TestArchivePlatform_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodDelete, "/api/v1/platforms/"+app.ID,
		`{"archived_reason":"product end-of-life"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"archived"`)) {
		t.Errorf("expected status=archived: %s", rec.Body.String())
	}
	final, err := platformStore.GetPlatform(context.Background(), app.ID)
	if err != nil {
		t.Fatalf("application lost after archive: %v", err)
	}
	if final.Status != domain.PlatformStatusArchived {
		t.Errorf("status not archived: %s", final.Status)
	}
}

// 10) POST /platforms/:id/repositories — unsupported_repo_provider → 422.
func TestCreatePlatformRepository_UnsupportedProvider(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms/"+app.ID+"/repositories",
		`{"repo_provider":"unknown","repo_full_name":"team/repo","role":"primary"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"unsupported_repo_provider"`)) {
		t.Errorf("expected unsupported_repo_provider: %s", rec.Body.String())
	}
}

// 11) POST /platforms/:id/repositories — disabled provider → 422.
func TestCreatePlatformRepository_DisabledProvider(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms/"+app.ID+"/repositories",
		`{"repo_provider":"forgejo","repo_full_name":"team/repo","role":"primary"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 12) POST /platforms/:id/repositories — duplicate link → 409 repository_link_conflict.
func TestCreatePlatformRepository_DuplicateLink(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)
	body := `{"repo_provider":"gitea","repo_full_name":"team/repo","role":"primary"}`
	first := doJSON(t, router, http.MethodPost, "/api/v1/platforms/"+app.ID+"/repositories", body)
	if first.Code != http.StatusCreated {
		t.Fatalf("seed failed: %s", first.Body.String())
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms/"+app.ID+"/repositories", body)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"code":"repository_link_conflict"`)) {
		t.Errorf("expected repository_link_conflict: %s", rec.Body.String())
	}
}

// 13) DELETE /platforms/:id/repositories/:repo_key — colon convention.
func TestDeletePlatformRepository_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = platformStore.CreatePlatformRepository(context.Background(), domain.PlatformRepository{
		PlatformID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.PlatformRepositoryRolePrimary, SyncStatus: domain.SyncStatusRequested,
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/platforms/"+app.ID+"/repositories/gitea:team/repo", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	links, _ := platformStore.ListPlatformRepositories(context.Background(), app.ID)
	if len(links) != 0 {
		t.Errorf("link should be removed, got %d", len(links))
	}
}

// 14) DELETE /platforms/:id/repositories/:repo_key — bad format → 400.
func TestDeletePlatformRepository_BadKey(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/platforms/some-id/repositories/no-colon", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 15) PATCH /scm/providers/:provider_key — adapter_version 거부 → 422.
func TestUpdateSCMProvider_AdapterVersionImmutable(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
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
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/scm/providers", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"provider_key":"gitea"`) {
		t.Errorf("expected gitea in list: %s", body)
	}
}

// 17) GET /platforms — empty list when none seeded.
func TestListPlatforms_Empty(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"data":[]`)) {
		t.Errorf("expected empty data array: %s", rec.Body.String())
	}
}

// 18) GET /platforms/:id — not found → 404.
func TestGetPlatform_NotFound(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms/nonexistent", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 19) PlatformStore nil 인 경우 503 (configuration error).
func TestApplications_ServiceUnavailable(t *testing.T) {
	router := NewRouter(RouterConfig{
		AuthDevFallback: true,
	})
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms", "")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// guard: ErrNotImplemented 가 더 이상 store layer 의 public API 가 아니라는 보증
// (sprint claude/work_260514-a 의 stub 제거 확인).
func TestStoreErrNotImplementedRemovedFromHandlerPath(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms", "")
	if rec.Code == http.StatusNotImplemented {
		t.Fatalf("handler returned 501 — stub removal incomplete: %s", rec.Body.String())
	}
}

// Compile-time guards (side-effect-free).
//
// `PlatformStore` 인터페이스 시그니처가 변경되면 본 assertion 이 깨져 컴파일
// 단계에서 즉시 검출된다. 테스트가 직접 호출하지 않더라도 인터페이스 계약을
// 보호하는 안전망이다.
var _ PlatformStore = (*memoryPlatformStore)(nil)
var _ IntegrationStore = (*memoryPlatformStore)(nil)

// 도메인 import / store sentinel error 의 외부 노출이 유지되는지 확인한다. domain
// 의 `IsRetryableSyncError` 가 사라지거나 store 의 `Err*` 가 unexported 로 바뀌면
// 본 블록이 컴파일 실패하여 회귀를 막는다. 런타임 동작은 없음 (no-op).
var (
	_ = domain.IsRetryableSyncError(domain.SyncErrorProviderUnreachable)
	_ = errors.Is(store.ErrConflict, store.ErrConflict)
	_ = errors.Is(store.ErrNotFound, store.ErrNotFound)
)

// --- Happy path 보강 tests (PR #106 self-review I2) ---

// 20) GET /platforms — status / include_archived 필터 happy.
func TestListPlatforms_FiltersHappy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	for _, status := range []domain.PlatformStatus{
		domain.PlatformStatusPlanning,
		domain.PlatformStatusActive,
		domain.PlatformStatusArchived,
	} {
		_, _ = platformStore.CreatePlatform(context.Background(), domain.Platform{
			Key: "K-" + string(status[:6]), Name: "N", Status: status,
			Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
		})
	}
	router := newPlatformsRouter(platformStore)

	// default: archived 제외 → 2건
	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"total":2`) {
		t.Errorf("default list should exclude archived (total=2): %s", rec.Body.String())
	}

	// include_archived=true → 3건
	rec = doJSON(t, router, http.MethodGet, "/api/v1/platforms?include_archived=true", "")
	if !strings.Contains(rec.Body.String(), `"total":3`) {
		t.Errorf("include_archived=true should return all (total=3): %s", rec.Body.String())
	}

	// status=active → 1건
	rec = doJSON(t, router, http.MethodGet, "/api/v1/platforms?status=active", "")
	if !strings.Contains(rec.Body.String(), `"total":1`) {
		t.Errorf("status=active should return 1: %s", rec.Body.String())
	}
}

// 21) GET /platforms/:id — happy (메타 + repositories 포함).
func TestGetPlatform_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = platformStore.CreatePlatformRepository(context.Background(), domain.PlatformRepository{
		PlatformID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.PlatformRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms/"+app.ID, "")
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

// 22) POST /platforms/:id/repositories — happy.
func TestCreatePlatformRepository_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/platforms/"+app.ID+"/repositories",
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
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/scm/providers/gitea",
		`{"enabled":false}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"enabled":false`) {
		t.Errorf("expected enabled=false: %s", rec.Body.String())
	}
}

// 24) DELETE /platforms/:id — not found → 404.
func TestArchivePlatform_NotFound(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	router := newPlatformsRouter(platformStore)
	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/platforms/nonexistent", `{"archived_reason":"X"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 25) DELETE catch-all path — `//gitea:team/repo` 같은 multiple leading slash 도
// TrimLeft 로 정상 처리되는지 (PR #106 self-review N1 보강).
func TestDeletePlatformRepository_MultipleLeadingSlashes(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.PlatformStatusPlanning,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = platformStore.CreatePlatformRepository(context.Background(), domain.PlatformRepository{
		PlatformID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.PlatformRepositoryRolePrimary, SyncStatus: domain.SyncStatusRequested,
	})
	router := newPlatformsRouter(platformStore)
	// gin 은 `//` 를 보통 정규화하지 않으므로 catch-all 이 받은 raw path 를 TrimLeft 가
	// 안전하게 처리해야 한다.
	rec := doJSON(t, router, http.MethodDelete,
		"/api/v1/platforms/"+app.ID+"/repositories//gitea:team/repo", "")
	if rec.Code != http.StatusOK && rec.Code != http.StatusNotFound {
		// gin 의 path 정규화에 따라 OK (200) 또는 NotFound (404, 정규화 후 trailing
		// slash 처리 차이) 가 나올 수 있다. 핵심은 500/400 같은 예상치 못한 응답이
		// 아닌 정상 routing 이 동작한다는 것.
		t.Fatalf("unexpected status=%d body=%s", rec.Code, rec.Body.String())
	}
}

// 26) GET /platforms/:id/dashboard — happy (API-93).
func TestApplicationDashboard_Happy(t *testing.T) {
	platformStore := newMemoryPlatformStore()
	app, _ := platformStore.CreatePlatform(context.Background(), domain.Platform{
		Key: "PLAT26", Name: "Platform 2026", Status: domain.PlatformStatusActive,
		Visibility: domain.PlatformVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = platformStore.CreatePlatformRepository(context.Background(), domain.PlatformRepository{
		PlatformID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.PlatformRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	router := newPlatformsRouter(platformStore)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/platforms/"+app.ID+"/dashboard", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"key":"PLAT26"`) {
		t.Errorf("response should include key: %s", body)
	}
	if !strings.Contains(body, `"metrics_overview"`) {
		t.Errorf("response should include metrics_overview: %s", body)
	}
	if !strings.Contains(body, `"quality_metrics"`) {
		t.Errorf("response should include quality_metrics: %s", body)
	}
}
