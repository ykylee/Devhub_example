package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"testing"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/httpapi"
)

// kratosAdminFake satisfies httpapi.KratosAdmin with per-method error
// injection. Local to main_test.go because httpapi.MockKratosAdmin lacks
// a Create-failure path and we want to exercise the seedLocalAdmin
// fallback (Create error → Find).
type kratosAdminFake struct {
	createError error
	createID    string
	findError   error
	findID      string
	findCalls   int
}

var _ httpapi.KratosAdmin = (*kratosAdminFake)(nil)

func (f *kratosAdminFake) CreateIdentity(_ context.Context, _, _, userID, _ string) (string, error) {
	if f.createError != nil {
		return "", f.createError
	}
	if f.createID != "" {
		return f.createID, nil
	}
	return "fake-create-" + userID, nil
}

func (f *kratosAdminFake) FindIdentityByUserID(_ context.Context, userID string) (string, error) {
	f.findCalls++
	if f.findError != nil {
		return "", f.findError
	}
	if f.findID != "" {
		return f.findID, nil
	}
	return "fake-find-" + userID, nil
}

func (f *kratosAdminFake) UpdateIdentityPassword(_ context.Context, _, _ string) error { return nil }
func (f *kratosAdminFake) SetIdentityState(_ context.Context, _ string, _ bool) error  { return nil }
func (f *kratosAdminFake) DeleteIdentity(_ context.Context, _ string) error            { return nil }

type orgStoreFake struct {
	createUserCalls       int
	setKratosIdentityArgs []setIdentityCall
	createError           error
}

type setIdentityCall struct {
	UserID, IdentityID string
}

var _ seedOrgStore = (*orgStoreFake)(nil)

func (f *orgStoreFake) CreateUser(_ context.Context, in domain.CreateUserInput) (domain.AppUser, error) {
	f.createUserCalls++
	if f.createError != nil {
		return domain.AppUser{}, f.createError
	}
	return domain.AppUser{UserID: in.UserID, Email: in.Email}, nil
}

func (f *orgStoreFake) SetKratosIdentityID(_ context.Context, userID, identityID string) error {
	f.setKratosIdentityArgs = append(f.setKratosIdentityArgs, setIdentityCall{userID, identityID})
	return nil
}

func TestMain(m *testing.M) {
	// seedLocalAdmin logs unconditionally; suppress to keep `go test`
	// output noise-free.
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

// Happy path: Create succeeds, Find is never consulted, DB rows + link
// both written with the freshly-created Kratos ID.
func TestSeedLocalAdmin_CreateSucceeds(t *testing.T) {
	kratos := &kratosAdminFake{createID: "kratos-fresh-1"}
	org := &orgStoreFake{}

	seedLocalAdmin(context.Background(), kratos, org)

	if kratos.findCalls != 0 {
		t.Errorf("findCalls = %d, want 0 (Find must skip on Create success)", kratos.findCalls)
	}
	if org.createUserCalls != 1 {
		t.Errorf("createUserCalls = %d, want 1", org.createUserCalls)
	}
	if len(org.setKratosIdentityArgs) != 1 || org.setKratosIdentityArgs[0].IdentityID != "kratos-fresh-1" {
		t.Errorf("SetKratosIdentityID args = %+v, want [{test kratos-fresh-1}]", org.setKratosIdentityArgs)
	}
}

// Idempotent re-run: Create fails (409 in production), Find returns the
// pre-existing identity, seed proceeds and links DB row to that ID.
func TestSeedLocalAdmin_CreateFailsFindSucceeds(t *testing.T) {
	kratos := &kratosAdminFake{
		createError: errors.New("kratos create identity status 409"),
		findID:      "kratos-existing-1",
	}
	org := &orgStoreFake{}

	seedLocalAdmin(context.Background(), kratos, org)

	if kratos.findCalls != 1 {
		t.Errorf("findCalls = %d, want 1 (Find must run as fallback)", kratos.findCalls)
	}
	if org.createUserCalls != 1 {
		t.Errorf("createUserCalls = %d, want 1", org.createUserCalls)
	}
	if len(org.setKratosIdentityArgs) != 1 || org.setKratosIdentityArgs[0].IdentityID != "kratos-existing-1" {
		t.Errorf("linked identity = %+v, want kratos-existing-1", org.setKratosIdentityArgs)
	}
}

// Operational outage: Kratos is unreachable. Both Create and Find fail.
// seed must abort *before* touching the DB — otherwise we'd leave a
// dangling AppUser row without a linked Kratos identity.
func TestSeedLocalAdmin_BothKratosCallsFail(t *testing.T) {
	kratos := &kratosAdminFake{
		createError: errors.New("connect: connection refused"),
		findError:   errors.New("connect: connection refused"),
	}
	org := &orgStoreFake{}

	seedLocalAdmin(context.Background(), kratos, org)

	if kratos.findCalls != 1 {
		t.Errorf("findCalls = %d, want 1", kratos.findCalls)
	}
	if org.createUserCalls != 0 {
		t.Errorf("createUserCalls = %d, want 0 (DB ops must skip when Kratos unreachable)", org.createUserCalls)
	}
	if len(org.setKratosIdentityArgs) != 0 {
		t.Errorf("SetKratosIdentityID must not be called, got %+v", org.setKratosIdentityArgs)
	}
}
