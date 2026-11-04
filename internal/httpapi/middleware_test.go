package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPI_RequireAdmin(t *testing.T) {
	adminToken := "secret-token"
	api := New(nil, nil, adminToken)

	nextCalled := false
	dummyNext := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := api.requireAdmin(dummyNext)

	tests := []struct {
		name         string
		authHeader   string
		expectStatus int
		expectNext   bool
	}{
		{
			name:         "missing authorization header",
			authHeader:   "",
			expectStatus: http.StatusUnauthorized,
			expectNext:   false,
		},
		{
			name:         "invalid schema",
			authHeader:   "Basic " + adminToken,
			expectStatus: http.StatusUnauthorized,
			expectNext:   false,
		},
		{
			name:         "empty bearer token",
			authHeader:   "Bearer ",
			expectStatus: http.StatusUnauthorized,
			expectNext:   false,
		},
		{
			name:         "wrong token",
			authHeader:   "Bearer wrong-token",
			expectStatus: http.StatusUnauthorized,
			expectNext:   false,
		},
		{
			name:         "valid token",
			authHeader:   "Bearer " + adminToken,
			expectStatus: http.StatusOK,
			expectNext:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled = false

			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.Equal(t, tt.expectStatus, rec.Code)
			require.Equal(t, tt.expectNext, nextCalled)

			if !tt.expectNext {
				require.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
			}
		})
	}
}

func TestMiddleware_Recovery(t *testing.T) {
	panickingHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("manual panic")
	})

	handler := recovery(panickingHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	require.NotPanics(t, func() {
		handler.ServeHTTP(rec, req)
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
