package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubAuthenticator struct{}

func (stubAuthenticator) AuthenticateSession(context.Context, string) (*Principal, error) {
	return &Principal{Role: RoleFamily, FamilyID: intPointer(7)}, nil
}

func intPointer(value int) *int { return &value }

func TestMiddlewareAddsAuthenticatedPrincipal(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "opaque-token"})
	recorder := httptest.NewRecorder()

	Middleware(stubAuthenticator{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok || principal.Role != RoleFamily || *principal.FamilyID != 7 {
			t.Fatal("authenticated principal was not attached to context")
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestTokenAndPasswordHelpers(t *testing.T) {
	first, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) < 40 {
		t.Fatal("token generator did not produce independent high-entropy tokens")
	}

	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if CheckPassword(hash, "wrong password") == nil || CheckPassword(hash, "correct horse battery staple") != nil {
		t.Fatal("password verification behaved incorrectly")
	}
}
