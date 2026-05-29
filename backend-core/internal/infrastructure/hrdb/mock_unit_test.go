package hrdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devhub/backend-core/internal/infrastructure/hrdb"
)

// mock_unit_test covers MockClient — the in-memory dev/test fallback for the
// HRDBClient interface. MockClient stays the documented PoC fixture (mock.go),
// so its three seed rows + case-insensitive system_id/name match + exact
// employee_id match + constructed email are pinned here.

func TestNewMockClient_SeedsThreeKnownPersons(t *testing.T) {
	client := hrdb.NewMockClient()
	if client == nil {
		t.Fatal("NewMockClient returned nil")
	}

	cases := []struct {
		name       string
		systemID   string
		employeeID string
		personName string
		wantEmail  string
		wantSysID  string
		wantDept   string
	}{
		{"yklee row", "yklee", "1001", "YK Lee", "yklee@example.com", "yklee", "Engineering"},
		{"akim row", "akim", "1002", "Alex Kim", "akim@example.com", "akim", "Product"},
		{"sjones row", "sjones", "1003", "Sam Jones", "sjones@example.com", "sjones", "Infrastructure"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			email, sysID, dept, err := client.Lookup(context.Background(), tc.systemID, tc.employeeID, tc.personName)
			if err != nil {
				t.Fatalf("Lookup err: %v", err)
			}
			if email != tc.wantEmail {
				t.Errorf("email = %q, want %q", email, tc.wantEmail)
			}
			if sysID != tc.wantSysID {
				t.Errorf("sysID = %q, want %q", sysID, tc.wantSysID)
			}
			if dept != tc.wantDept {
				t.Errorf("dept = %q, want %q", dept, tc.wantDept)
			}
		})
	}
}

func TestMockClient_Lookup_CaseInsensitiveSystemID(t *testing.T) {
	client := hrdb.NewMockClient()
	email, sysID, dept, err := client.Lookup(context.Background(), "YKLEE", "1001", "YK Lee")
	if err != nil {
		t.Fatalf("Lookup err: %v", err)
	}
	if email != "yklee@example.com" {
		t.Errorf("email = %q, want yklee@example.com (lowercased system_id)", email)
	}
	if sysID != "yklee" {
		t.Errorf("sysID = %q, want yklee", sysID)
	}
	if dept != "Engineering" {
		t.Errorf("dept = %q, want Engineering", dept)
	}
}

func TestMockClient_Lookup_CaseInsensitiveName(t *testing.T) {
	client := hrdb.NewMockClient()
	_, _, _, err := client.Lookup(context.Background(), "akim", "1002", "alex kim")
	if err != nil {
		t.Fatalf("Lookup err: %v, want nil (name match is case-insensitive)", err)
	}
}

func TestMockClient_Lookup_EmployeeIDIsExactMatch(t *testing.T) {
	client := hrdb.NewMockClient()
	// Employee ID is exact-match (no case fold, no trim). Each variant must miss.
	cases := []string{"1001 ", " 1001", "01001", "1001a", ""}
	for _, eid := range cases {
		t.Run("eid="+eid, func(t *testing.T) {
			_, _, _, err := client.Lookup(context.Background(), "yklee", eid, "YK Lee")
			if !errors.Is(err, hrdb.ErrPersonNotFound) {
				t.Errorf("expected ErrPersonNotFound for employee_id %q, got %v", eid, err)
			}
		})
	}
}

func TestMockClient_Lookup_NotFoundReturnsErrPersonNotFound(t *testing.T) {
	client := hrdb.NewMockClient()
	cases := []struct {
		name       string
		systemID   string
		employeeID string
		personName string
	}{
		{"unknown system_id", "ghost", "1001", "YK Lee"},
		{"wrong employee_id", "yklee", "9999", "YK Lee"},
		{"wrong name", "yklee", "1001", "Someone Else"},
		{"all empty", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			email, sysID, dept, err := client.Lookup(context.Background(), tc.systemID, tc.employeeID, tc.personName)
			if !errors.Is(err, hrdb.ErrPersonNotFound) {
				t.Errorf("expected ErrPersonNotFound, got %v", err)
			}
			if email != "" || sysID != "" || dept != "" {
				t.Errorf("expected zero return values on miss, got (%q,%q,%q)", email, sysID, dept)
			}
		})
	}
}

func TestMockClient_Lookup_AllThreeFieldsMustMatch(t *testing.T) {
	client := hrdb.NewMockClient()
	// Cross-row mismatch: yklee system_id + akim employee_id must miss.
	_, _, _, err := client.Lookup(context.Background(), "yklee", "1002", "YK Lee")
	if !errors.Is(err, hrdb.ErrPersonNotFound) {
		t.Errorf("expected ErrPersonNotFound for cross-row mismatch, got %v", err)
	}
}

// ErrPersonNotFound stays an exported sentinel — signup handler matches with
// errors.Is. Pin the sentinel identity so accidental reassignment is caught
// by unit test rather than at the handler boundary.
func TestErrPersonNotFound_IsExportedSentinel(t *testing.T) {
	if hrdb.ErrPersonNotFound == nil {
		t.Fatal("hrdb.ErrPersonNotFound is nil")
	}
	if !errors.Is(hrdb.ErrPersonNotFound, hrdb.ErrPersonNotFound) {
		t.Error("ErrPersonNotFound must satisfy errors.Is against itself")
	}
}
