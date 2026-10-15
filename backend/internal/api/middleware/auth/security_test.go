package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnimplementedAuthenticationFailsClosed(t *testing.T) {
	called := false
	handler := RequireAuth(nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	for _, cookie := range []string{"", "session=forged"} {
		req := httptest.NewRequest("GET", "/api/user/me", nil)
		req.Header.Set("Cookie", cookie)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized || called {
			t.Fatal("unverified request reached protected handler")
		}
	}
}
