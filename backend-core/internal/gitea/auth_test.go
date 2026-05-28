package gitea_test

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/devhub/backend-core/internal/domain"
	"github.com/devhub/backend-core/internal/gitea"
)

func TestAuthorizationHeader_Token(t *testing.T) {
	h, ok, err := gitea.AuthorizationHeader(context.Background(),
		domain.OutboundAuth{Mode: domain.IntegrationAuthModeToken, Token: "pat-123"}, http.DefaultClient)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if h != "token pat-123" {
		t.Fatalf("header=%q", h)
	}
}

func TestAuthorizationHeader_TokenMissing(t *testing.T) {
	_, ok, err := gitea.AuthorizationHeader(context.Background(),
		domain.OutboundAuth{Mode: domain.IntegrationAuthModeToken}, http.DefaultClient)
	if err != nil || ok {
		t.Fatalf("expected ok=false err=nil, got ok=%v err=%v", ok, err)
	}
}

func TestAuthorizationHeader_Basic(t *testing.T) {
	for _, mode := range []domain.IntegrationAuthMode{domain.IntegrationAuthModeBasic, domain.IntegrationAuthModeAppPassword} {
		h, ok, err := gitea.AuthorizationHeader(context.Background(),
			domain.OutboundAuth{Mode: mode, Username: "alice", Secret: "s3cret"}, http.DefaultClient)
		if err != nil || !ok {
			t.Fatalf("mode=%s ok=%v err=%v", mode, ok, err)
		}
		want := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:s3cret"))
		if h != want {
			t.Fatalf("mode=%s header=%q want=%q", mode, h, want)
		}
	}
}

func TestAuthorizationHeader_BasicMissing(t *testing.T) {
	_, ok, _ := gitea.AuthorizationHeader(context.Background(),
		domain.OutboundAuth{Mode: domain.IntegrationAuthModeBasic, Username: "alice"}, http.DefaultClient)
	if ok {
		t.Fatal("expected ok=false when secret missing")
	}
}

func TestAuthorizationHeader_Agent(t *testing.T) {
	_, ok, err := gitea.AuthorizationHeader(context.Background(),
		domain.OutboundAuth{Mode: domain.IntegrationAuthModeAgent, Username: "agent-1"}, http.DefaultClient)
	if err != nil || ok {
		t.Fatalf("agent mode unsupported for direct sync: ok=%v err=%v", ok, err)
	}
}

func TestAuthorizationHeader_OAuth2(t *testing.T) {
	var gotGrant, gotClientID, gotClientSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotGrant = r.Form.Get("grant_type")
		gotClientID = r.Form.Get("client_id")
		gotClientSecret = r.Form.Get("client_secret")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"at-xyz","token_type":"bearer"}`))
	}))
	defer srv.Close()

	h, ok, err := gitea.AuthorizationHeader(context.Background(), domain.OutboundAuth{
		Mode:     domain.IntegrationAuthModeOAuth2,
		ClientID: "cid",
		Secret:   "csecret",
		TokenURL: srv.URL,
	}, &http.Client{Timeout: 5 * time.Second})
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if h != "Bearer at-xyz" {
		t.Fatalf("header=%q", h)
	}
	if gotGrant != "client_credentials" || gotClientID != "cid" || gotClientSecret != "csecret" {
		t.Fatalf("token request form mismatch: grant=%q client_id=%q secret=%q", gotGrant, gotClientID, gotClientSecret)
	}
}

func TestAuthorizationHeader_OAuth2EndpointError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, ok, err := gitea.AuthorizationHeader(context.Background(), domain.OutboundAuth{
		Mode:     domain.IntegrationAuthModeOAuth2,
		ClientID: "cid",
		Secret:   "csecret",
		TokenURL: srv.URL,
	}, &http.Client{Timeout: 5 * time.Second})
	if ok || err == nil {
		t.Fatalf("expected error on 401 token endpoint, got ok=%v err=%v", ok, err)
	}
}
