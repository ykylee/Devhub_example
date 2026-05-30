package view

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/store"
	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// fakeOrgUnitStore — separate OrganizationStore impl for org-unit tests
// ---------------------------------------------------------------------------

type fakeOrgUnitStore struct {
	listUsersResult        []domain.AppUser
	listUsersTotal         int
	listUsersErr           error
	getUserResult          domain.AppUser
	getUserErr             error
	createUserResult       domain.AppUser
	createUserErr          error
	updateUserResult       domain.AppUser
	updateUserErr          error
	deleteUserErr          error
	setIdPSubjectErr       error
	getUserByIdPResult     domain.AppUser
	getUserByIdPErr        error
	getHierarchyResult     domain.Hierarchy
	getHierarchyErr        error
	updateHierarchyErr     error
	getOrgUnitResult       domain.OrgUnit
	getOrgUnitErr          error
	createOrgUnitResult    domain.OrgUnit
	createOrgUnitErr       error
	updateOrgUnitResult    domain.OrgUnit
	updateOrgUnitErr       error
	deleteOrgUnitErr       error
	listUnitMembersResult  []domain.AppUser
	listUnitMembersErr     error
	replaceUnitMembersErr  error
	submitOnboardingResult domain.AppUser
	submitOnboardingErr    error
	confirmReviewResult    domain.AppUser
	confirmReviewErr       error
	searchOrgUnitsResult   []domain.OrgUnit
	searchOrgUnitsErr      error
}

func (f *fakeOrgUnitStore) ListUsers(_ context.Context, _ domain.UserListOptions) ([]domain.AppUser, int, error) {
	return f.listUsersResult, f.listUsersTotal, f.listUsersErr
}
func (f *fakeOrgUnitStore) GetUser(_ context.Context, _ string) (domain.AppUser, error) {
	return f.getUserResult, f.getUserErr
}
func (f *fakeOrgUnitStore) CreateUser(_ context.Context, _ domain.CreateUserInput) (domain.AppUser, error) {
	return f.createUserResult, f.createUserErr
}
func (f *fakeOrgUnitStore) UpdateUser(_ context.Context, _ string, _ domain.UpdateUserInput) (domain.AppUser, error) {
	return f.updateUserResult, f.updateUserErr
}
func (f *fakeOrgUnitStore) DeleteUser(_ context.Context, _ string) error {
	return f.deleteUserErr
}
func (f *fakeOrgUnitStore) SetIdPSubject(_ context.Context, _, _ string) error {
	return f.setIdPSubjectErr
}
func (f *fakeOrgUnitStore) GetUserByIdPSubject(_ context.Context, _ string) (domain.AppUser, error) {
	return f.getUserByIdPResult, f.getUserByIdPErr
}
func (f *fakeOrgUnitStore) GetHierarchy(_ context.Context) (domain.Hierarchy, error) {
	return f.getHierarchyResult, f.getHierarchyErr
}
func (f *fakeOrgUnitStore) UpdateHierarchy(_ context.Context, _ domain.Hierarchy) error {
	return f.updateHierarchyErr
}
func (f *fakeOrgUnitStore) GetOrgUnit(_ context.Context, _ string) (domain.OrgUnit, error) {
	return f.getOrgUnitResult, f.getOrgUnitErr
}
func (f *fakeOrgUnitStore) CreateOrgUnit(_ context.Context, in domain.CreateOrgUnitInput) (domain.OrgUnit, error) {
	if f.createOrgUnitErr != nil {
		return domain.OrgUnit{}, f.createOrgUnitErr
	}
	r := f.createOrgUnitResult
	if r.UnitID == "" {
		r.UnitID = in.UnitID
	}
	return r, nil
}
func (f *fakeOrgUnitStore) UpdateOrgUnit(_ context.Context, _ string, _ domain.UpdateOrgUnitInput) (domain.OrgUnit, error) {
	return f.updateOrgUnitResult, f.updateOrgUnitErr
}
func (f *fakeOrgUnitStore) DeleteOrgUnit(_ context.Context, _ string) error {
	return f.deleteOrgUnitErr
}
func (f *fakeOrgUnitStore) ListUnitMembers(_ context.Context, _ string) ([]domain.AppUser, error) {
	return f.listUnitMembersResult, f.listUnitMembersErr
}
func (f *fakeOrgUnitStore) ReplaceUnitMembers(_ context.Context, _ string, _ []string) error {
	return f.replaceUnitMembersErr
}
func (f *fakeOrgUnitStore) SubmitOnboarding(_ context.Context, _ domain.OnboardingSubmitInput) (domain.AppUser, error) {
	return f.submitOnboardingResult, f.submitOnboardingErr
}
func (f *fakeOrgUnitStore) ConfirmUserReview(_ context.Context, _ string) (domain.AppUser, error) {
	return f.confirmReviewResult, f.confirmReviewErr
}
func (f *fakeOrgUnitStore) SearchOrgUnits(_ context.Context, _ string, _ int) ([]domain.OrgUnit, error) {
	return f.searchOrgUnitsResult, f.searchOrgUnitsErr
}

// ---------------------------------------------------------------------------
// GetOrgUnit handler tests
// ---------------------------------------------------------------------------

func TestGetOrgUnit_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.GET("/units/:unit_id", h.GetOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/team-a", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetOrgUnit_EmptyUnitID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.GET("/units/:unit_id", h.GetOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/%20", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetOrgUnit_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{getOrgUnitErr: store.ErrNotFound}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.GET("/units/:unit_id", h.GetOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/nope", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetOrgUnit_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{getOrgUnitErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.GET("/units/:unit_id", h.GetOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/team-a", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetOrgUnit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{getOrgUnitResult: domain.OrgUnit{UnitID: "team-a", Label: "Team A", UnitType: "team"}}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.GET("/units/:unit_id", h.GetOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/team-a", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "team-a") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CreateOrgUnit handler tests
// ---------------------------------------------------------------------------

func TestCreateOrgUnit_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateOrgUnit_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateOrgUnit_MissingRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	body := `{"unit_id":"","label":"","unit_type":"team"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateOrgUnit_InvalidUnitType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	body := `{"unit_id":"t-1","label":"T","unit_type":"bogus"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateOrgUnit_SelfReference(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	body := `{"unit_id":"t-1","label":"T","unit_type":"team","parent_unit_id":"t-1"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateOrgUnit_StoreConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{createOrgUnitErr: store.ErrConflict}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	body := `{"unit_id":"t-1","label":"T","unit_type":"team"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 409 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateOrgUnit_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{createOrgUnitErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	body := `{"unit_id":"t-1","label":"T","unit_type":"team"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateOrgUnit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{createOrgUnitResult: domain.OrgUnit{UnitID: "t-1", Label: "T", UnitType: "team"}}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.POST("/units", h.CreateOrgUnit)

	body := `{"unit_id":"t-1","label":"T","unit_type":"team","parent_unit_id":"root"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/units", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "created") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// UpdateOrgUnit handler tests
// ---------------------------------------------------------------------------

func TestUpdateOrgUnit_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateOrgUnit_EmptyUnitID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/%20", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateOrgUnit_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateOrgUnit_InvalidUnitType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{"unit_type":"bogus"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateOrgUnit_EmptyLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{"label":"  "}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateOrgUnit_SelfCycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{"parent_unit_id":"t-1"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 422 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "cycle_detected") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestUpdateOrgUnit_AncestorCycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Hierarchy: root -> A -> B -> C
	// Try to move A under C → cycle
	s := &fakeOrgUnitStore{
		getHierarchyResult: domain.Hierarchy{
			Edges: []domain.OrgEdge{
				{SourceUnitID: "root", TargetUnitID: "A"},
				{SourceUnitID: "A", TargetUnitID: "B"},
				{SourceUnitID: "B", TargetUnitID: "C"},
			},
		},
		updateOrgUnitResult: domain.OrgUnit{UnitID: "A", Label: "A"},
	}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/A", strings.NewReader(`{"parent_unit_id":"C"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 422 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "cycle_detected") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestUpdateOrgUnit_StoreNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{updateOrgUnitErr: store.ErrNotFound}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{"label":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateOrgUnit_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{updateOrgUnitErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{"label":"New"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateOrgUnit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{updateOrgUnitResult: domain.OrgUnit{UnitID: "t-1", Label: "New", UnitType: "team"}}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{"label":"New","unit_type":"division","leader_user_id":"u-1","position_x":10,"position_y":20}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateOrgUnit_SuccessParentNoEdges(t *testing.T) {
	// When hierarchy has no edges, parent change should succeed (no cycle possible)
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{
		getHierarchyResult:  domain.Hierarchy{Edges: []domain.OrgEdge{}},
		updateOrgUnitResult: domain.OrgUnit{UnitID: "t-1", Label: "T"},
	}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PATCH("/units/:unit_id", h.UpdateOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/units/t-1", strings.NewReader(`{"parent_unit_id":"root"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// DeleteOrgUnit handler tests
// ---------------------------------------------------------------------------

func TestDeleteOrgUnit_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.DELETE("/units/:unit_id", h.DeleteOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/units/t-1", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteOrgUnit_EmptyUnitID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.DELETE("/units/:unit_id", h.DeleteOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/units/%20", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteOrgUnit_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{deleteOrgUnitErr: store.ErrNotFound}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.DELETE("/units/:unit_id", h.DeleteOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/units/nope", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteOrgUnit_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{deleteOrgUnitErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.DELETE("/units/:unit_id", h.DeleteOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/units/t-1", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestDeleteOrgUnit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.DELETE("/units/:unit_id", h.DeleteOrgUnit)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/units/t-1", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "deleted") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// ListUnitMembers handler tests
// ---------------------------------------------------------------------------

func TestListUnitMembers_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.GET("/units/:unit_id/members", h.ListUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/t-1/members", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListUnitMembers_EmptyUnitID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.GET("/units/:unit_id/members", h.ListUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/%20/members", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListUnitMembers_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{listUnitMembersErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.GET("/units/:unit_id/members", h.ListUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/t-1/members", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestListUnitMembers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{
		listUnitMembersResult: []domain.AppUser{{UserID: "u-1", DisplayName: "A"}},
	}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.GET("/units/:unit_id/members", h.ListUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/units/t-1/members", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "u-1") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// ReplaceUnitMembers handler tests
// ---------------------------------------------------------------------------

func TestReplaceUnitMembers_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.PUT("/units/:unit_id/members", h.ReplaceUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/units/t-1/members", strings.NewReader(`{"user_ids":[]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestReplaceUnitMembers_EmptyUnitID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.PUT("/units/:unit_id/members", h.ReplaceUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/units/%20/members", strings.NewReader(`{"user_ids":[]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestReplaceUnitMembers_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.PUT("/units/:unit_id/members", h.ReplaceUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/units/t-1/members", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestReplaceUnitMembers_ReplaceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{replaceUnitMembersErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PUT("/units/:unit_id/members", h.ReplaceUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/units/t-1/members", strings.NewReader(`{"user_ids":["u-1"]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestReplaceUnitMembers_ListAfterReplaceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{listUnitMembersErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PUT("/units/:unit_id/members", h.ReplaceUnitMembers)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/units/t-1/members", strings.NewReader(`{"user_ids":["u-1"]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestReplaceUnitMembers_SuccessWithDedup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{
		listUnitMembersResult: []domain.AppUser{{UserID: "u-1"}},
	}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PUT("/units/:unit_id/members", h.ReplaceUnitMembers)

	// user_ids with duplicates and blanks
	body := `{"user_ids":["u-1"," u-1 ","","u-2"]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/units/t-1/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// GetHierarchy handler tests
// ---------------------------------------------------------------------------

func TestGetHierarchy_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.GET("/hierarchy", h.GetHierarchy)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hierarchy", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetHierarchy_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{getHierarchyErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.GET("/hierarchy", h.GetHierarchy)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hierarchy", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestGetHierarchy_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{
		getHierarchyResult: domain.Hierarchy{
			Units: []domain.OrgUnit{{UnitID: "root", Label: "Root", UnitType: "company"}},
			Edges: []domain.OrgEdge{{SourceUnitID: "root", TargetUnitID: "t-1"}},
		},
	}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.GET("/hierarchy", h.GetHierarchy)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hierarchy", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "root") {
		t.Fatalf("body = %q", body)
	}
	if !strings.Contains(body, "t-1") {
		t.Fatalf("edges missing: %q", body)
	}
}

// ---------------------------------------------------------------------------
// UpdateHierarchy handler tests
// ---------------------------------------------------------------------------

func TestUpdateHierarchy_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{})
	r := gin.New()
	r.PUT("/hierarchy", h.UpdateHierarchy)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/hierarchy", strings.NewReader(`{"nodes":[],"edges":[]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateHierarchy_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.PUT("/hierarchy", h.UpdateHierarchy)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/hierarchy", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateHierarchy_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{updateHierarchyErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PUT("/hierarchy", h.UpdateHierarchy)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/hierarchy", strings.NewReader(`{"nodes":[],"edges":[]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUpdateHierarchy_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{}
	h := NewOrganizationHandler(OrganizationConfig{OrganizationStore: s})
	r := gin.New()
	r.PUT("/hierarchy", h.UpdateHierarchy)

	body := `{"nodes":[{"id":"root","position":{"x":10,"y":20},"data":{"label":"Root","type":"company","direct_count":3,"total_count":10}}],"edges":[{"source":"root","target":"t-1"}]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/hierarchy", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// SearchOrganizations handler tests
// ---------------------------------------------------------------------------

func TestSearchOrganizations_OnboardingDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: false, OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.GET("/search", h.SearchOrganizations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=team", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSearchOrganizations_NilStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: true})
	r := gin.New()
	r.GET("/search", h.SearchOrganizations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=team", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 503 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSearchOrganizations_QueryTooShort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: true, OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.GET("/search", h.SearchOrganizations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=x", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 422 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSearchOrganizations_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: true, OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.GET("/search", h.SearchOrganizations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=team&limit=abc", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 422 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSearchOrganizations_LimitTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: true, OrganizationStore: &fakeOrgUnitStore{}})
	r := gin.New()
	r.GET("/search", h.SearchOrganizations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=team&limit=100", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 422 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSearchOrganizations_StoreError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{searchOrgUnitsErr: errors.New("db down")}
	h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: true, OrganizationStore: s})
	r := gin.New()
	r.GET("/search", h.SearchOrganizations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=team", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 500 {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSearchOrganizations_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := &fakeOrgUnitStore{
		searchOrgUnitsResult: []domain.OrgUnit{
			{UnitID: "t-1", Label: "Team Alpha"},
			{UnitID: "t-2", Label: "Team Beta"},
		},
	}
	h := NewOrganizationHandler(OrganizationConfig{OnboardingGateEnabled: true, OrganizationStore: s})
	r := gin.New()
	r.GET("/search", h.SearchOrganizations)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/search?q=team&limit=10", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Team Alpha") || !strings.Contains(body, "Team Beta") {
		t.Fatalf("body = %q", body)
	}
}
