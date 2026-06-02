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
	mu                   sync.Mutex
	apps                 map[string]domain.Application
	links                map[string][]domain.ApplicationRepository
	providers            map[string]domain.SCMProvider
	projects             map[string]domain.Project
	projectRepositories  map[string][]domain.ProjectRepository
	activeLinkCounts     map[string]int
	integrations         map[string]domain.ProjectIntegration
	integrationProviders map[string]domain.IntegrationProvider
	integrationBindings  map[string]domain.IntegrationBinding
	criticalCounts       map[string]int // override for CountApplicationCriticalWarnings tests
	infraSnapshot        memoryInfraSnapshot
	repositoryIDs        map[string]int64
	repositories         map[string]domain.Repository // full_name → repo (UpsertRepository/ListByProvider)
	nextRepositoryID     int64
	ciRuns              []domain.BuildRun
}

type memoryInfraSnapshot struct {
	snapshotAt        time.Time
	nodesJSON         []byte
	servicesJSON      []byte
	degradedProviders []string
	loaded            bool
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
		projects:             make(map[string]domain.Project),
		projectRepositories:  make(map[string][]domain.ProjectRepository),
		activeLinkCounts:     make(map[string]int),
		integrations:         make(map[string]domain.ProjectIntegration),
		integrationProviders: make(map[string]domain.IntegrationProvider),
		integrationBindings:  make(map[string]domain.IntegrationBinding),
		criticalCounts:       make(map[string]int),
		repositoryIDs:        make(map[string]int64),
		repositories:         make(map[string]domain.Repository),
		nextRepositoryID:     1000,
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
		if opts.ActorLogin != "" && opts.ActorRole != "system_admin" && opts.ActorRole != "team_manager" {
			if a.OwnerUserID != opts.ActorLogin && a.LeaderUserID != opts.ActorLogin {
				isMember := false
				for _, p := range s.projects {
					if p.ApplicationID != a.ID {
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

// DeleteApplication — production *PostgresStore.DeleteApplication mirror (hard-delete).
func (s *memoryApplicationStore) DeleteApplication(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.apps[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.apps, id)
	delete(s.links, id)
	return nil
}

// CountActiveApplicationRepositories — production *PostgresStore 의 UNION 쿼리 mirror.
// 직접 link (sync_status='active') + project 경유 간접 link (link 존재 = active 간주).
// 간접 link 에서 repositories miss (테스트 setup 누락) 면 skip — production 의
// `JOIN repositories r` strict 동작과 동일 (FK 매칭 실패 row 누락). hardcoded
// fallback ("bitbucket"+project.Key) 은 production 동작과 무관해 fake parity 깨므로 제거.
func (s *memoryApplicationStore) CountActiveApplicationRepositories(_ context.Context, applicationID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	activeRepos := make(map[string]bool)
	for _, l := range s.links[applicationID] {
		if l.SyncStatus == domain.SyncStatusActive {
			activeRepos[l.RepoProvider+"/"+l.RepoFullName] = true
		}
	}
	for _, p := range s.projects {
		if p.ApplicationID != applicationID {
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
func (s *memoryApplicationStore) lookupRepoForFake(repoID int64) (string, string) {
	for _, r := range s.repositories {
		if r.ID == repoID {
			return r.ProviderKey, r.FullName
		}
	}
	return "", ""
}

func (s *memoryApplicationStore) ListApplicationRepositories(_ context.Context, applicationID string) ([]domain.ApplicationRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	direct := s.links[applicationID]
	seen := make(map[string]bool)
	out := make([]domain.ApplicationRepository, 0, len(direct))
	for _, l := range direct {
		seen[l.RepoProvider+"/"+l.RepoFullName] = true
		out = append(out, l)
	}
	// project 경유 간접 link — production JOIN repositories strict 와 동일 동작.
	for _, p := range s.projects {
		if p.ApplicationID != applicationID {
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
			role := domain.ApplicationRepositoryRoleSub
			switch pr.Role {
			case "primary":
				role = domain.ApplicationRepositoryRolePrimary
			case "shared":
				role = domain.ApplicationRepositoryRoleShared
			}
			out = append(out, domain.ApplicationRepository{
				ApplicationID: applicationID,
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
		if opts.RepositoryID != 0 && p.RepositoryID != opts.RepositoryID {
			continue
		}
		if opts.ApplicationID != "" && p.ApplicationID != opts.ApplicationID {
			continue
		}
		if opts.Status != "" && string(p.Status) != opts.Status {
			continue
		}
		if !opts.IncludeArchived && p.Status == domain.ApplicationStatusArchived {
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

func (s *memoryApplicationStore) ListProjectRepositories(_ context.Context, projectID string) ([]domain.ProjectRepository, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.ProjectRepository(nil), s.projectRepositories[projectID]...), nil
}

func (s *memoryApplicationStore) CreateProjectRepository(_ context.Context, link domain.ProjectRepository) (domain.ProjectRepository, error) {
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

func (s *memoryApplicationStore) DeleteProjectRepository(_ context.Context, projectID string, repositoryID int64) error {
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

func (s *memoryApplicationStore) CreateProjectWithRepositoryPayload(_ context.Context, p domain.Project, repositoryIDs []int64, repoPayload *store.RepositoryCreatePayload) (domain.Project, error) {
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

func (s *memoryApplicationStore) DeleteProject(_ context.Context, id string) error {
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

func (s *memoryApplicationStore) ListRepositoryActivity(_ context.Context, repoID int64, _ store.RepositoryActivityOptions) (domain.RepositoryActivity, error) {
	return domain.RepositoryActivity{RepositoryID: repoID}, nil
}

func (s *memoryApplicationStore) ListRepositoryPullRequests(_ context.Context, _ int64, _ store.PRActivityListOptions) ([]domain.PRActivity, int, error) {
	return []domain.PRActivity{}, 0, nil
}

func (s *memoryApplicationStore) ListRepositoryBuildRuns(_ context.Context, repoID int64, opts store.BuildRunListOptions) ([]domain.BuildRun, int, error) {
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
	// PR #107 codex review P2 — partial UNIQUE 인덱스 흉내. external_key 변경이
	// 다른 row 와 (scope target, integration_type, external_key) 충돌 시 ErrConflict.
	for otherID, other := range s.integrations {
		if otherID == i.ID {
			continue
		}
		if other.Scope == current.Scope &&
			other.ApplicationID == current.ApplicationID &&
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

func (s *memoryApplicationStore) DeleteIntegration(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrations[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.integrations, id)
	return nil
}

func (s *memoryApplicationStore) ListIntegrationProviders(_ context.Context, opts store.IntegrationProviderListOptions) ([]domain.IntegrationProvider, int, error) {
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

func (s *memoryApplicationStore) GetIntegrationProviderByID(_ context.Context, providerID string) (domain.IntegrationProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.integrationProviders[providerID]
	if !ok {
		return domain.IntegrationProvider{}, store.ErrNotFound
	}
	return p, nil
}

func (s *memoryApplicationStore) GetIntegrationProviderByKey(_ context.Context, providerKey string) (domain.IntegrationProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.integrationProviders {
		if p.ProviderKey == providerKey {
			return p, nil
		}
	}
	return domain.IntegrationProvider{}, store.ErrNotFound
}

func (s *memoryApplicationStore) CreateIntegrationProvider(_ context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
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

func (s *memoryApplicationStore) UpdateIntegrationProvider(_ context.Context, p domain.IntegrationProvider) (domain.IntegrationProvider, error) {
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

func (s *memoryApplicationStore) UpsertRepository(_ context.Context, repo domain.Repository) error {
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
	return nil
}

func (s *memoryApplicationStore) ListRepositoriesByProvider(_ context.Context, providerID string) ([]domain.Repository, error) {
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

func (s *memoryApplicationStore) DeleteIntegrationProvider(_ context.Context, providerID string) error {
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

func (s *memoryApplicationStore) CreateIntegrationSyncJob(_ context.Context, providerID string, _ string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrationProviders[providerID]; !ok {
		return "", store.ErrNotFound
	}
	return "job-" + providerID, nil
}

func (s *memoryApplicationStore) ListIntegrationBindings(_ context.Context, opts store.IntegrationBindingListOptions) ([]domain.IntegrationBinding, int, error) {
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

func (s *memoryApplicationStore) CreateIntegrationBinding(_ context.Context, b domain.IntegrationBinding) (domain.IntegrationBinding, error) {
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
func (s *memoryApplicationStore) GetIntegrationBindingByID(_ context.Context, id string) (domain.IntegrationBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.integrationBindings[id]
	if !ok {
		return domain.IntegrationBinding{}, store.ErrNotFound
	}
	return b, nil
}

func (s *memoryApplicationStore) UpdateIntegrationBinding(_ context.Context, b domain.IntegrationBinding) (domain.IntegrationBinding, error) {
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

func (s *memoryApplicationStore) DeleteIntegrationBinding(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.integrationBindings[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.integrationBindings, id)
	return nil
}

func (s *memoryApplicationStore) SaveInfraSnapshot(_ context.Context, _ string, _ string, snapshotAt time.Time, _ string, nodesJSON, servicesJSON []byte, degradedProviders []string) error {
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

func (s *memoryApplicationStore) LoadLatestInfraSnapshot(_ context.Context) (time.Time, []byte, []byte, []string, error) {
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
		`{"key":"DEVHUB","name":"Devhub Platform","owner_user_id":"u1","leader_user_id":"u1","development_unit_id":"dept-eng","visibility":"internal","status":"planning"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"key":"DEVHUB"`)) {
		t.Errorf("response should echo key: %s", rec.Body.String())
	}
}

// 2) POST /applications — invalid key format → 422 invalid_application_key.
func TestCreateApplication_InvalidKey(t *testing.T) {
	router := newApplicationsRouter(newMemoryApplicationStore())

	rec := doJSON(t, router, http.MethodPost, "/api/v1/applications",
		`{"key":"too-short","name":"X","owner_user_id":"u1","leader_user_id":"u1","development_unit_id":"dept-eng","visibility":"internal","status":"planning"}`)
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
	body := `{"key":"A1B2C3D4E5","name":"X","owner_user_id":"u1","leader_user_id":"u1","development_unit_id":"dept-eng","visibility":"internal","status":"planning"}`
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
// status 전이 정책 자유화 (2026-05-28) — planning→active 의 active repo ≥1 가드 제거.
// 0 repo 인 application 도 active 로 자유 전이 가능. 기존 테스트는 가드 검증이었으므로
// 자유화 후 expected behavior 로 갱신.
func TestUpdateApplication_ActivationWithoutLinkedRepos(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusPlanning,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
		`{"status":"active"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (자유화 후 200 기대)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"active"`)) {
		t.Errorf("expected status=active: %s", rec.Body.String())
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

// status 전이 정책 자유화 (2026-05-28) — 이전엔 closed→planning 같은 backward 전이가
// 422 거부였으나 이제 모든 전이 허용. 기존 테스트는 가드 검증이었으므로 자유화 후
// expected behavior 로 갱신.
func TestUpdateApplication_AnyStatusTransitionAllowed(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusClosed,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	// closed → planning 도 자유화 후 200.
	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
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
func TestUpdateApplication_HoldWithoutReason(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "A1B2C3D4E5", Name: "X", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodPatch, "/api/v1/applications/"+app.ID,
		`{"status":"on_hold"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s (자유화 후 200 기대)", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"status":"on_hold"`)) {
		t.Errorf("expected status=on_hold: %s", rec.Body.String())
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

// 26) GET /applications/:id/dashboard — happy (API-93).
func TestApplicationDashboard_Happy(t *testing.T) {
	appStore := newMemoryApplicationStore()
	app, _ := appStore.CreateApplication(context.Background(), domain.Application{
		Key: "PLAT26", Name: "Platform 2026", Status: domain.ApplicationStatusActive,
		Visibility: domain.ApplicationVisibilityInternal, OwnerUserID: "u1",
	})
	_, _ = appStore.CreateApplicationRepository(context.Background(), domain.ApplicationRepository{
		ApplicationID: app.ID, RepoProvider: "gitea", RepoFullName: "team/repo",
		Role: domain.ApplicationRepositoryRolePrimary, SyncStatus: domain.SyncStatusActive,
	})
	router := newApplicationsRouter(appStore)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/applications/"+app.ID+"/dashboard", "")
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
