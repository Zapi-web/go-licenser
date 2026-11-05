package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zapi-web/go-licenser/internal/domain"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
		checkHeader    func(t *testing.T, h http.Header)
	}{
		{
			name:           "unknown error",
			err:            errors.New("unknown errro"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "internal_error",
		},
		{
			name:           "not found wrapper",
			err:            domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "not_found",
		},
		{
			name:           "retry-after on concurent",
			err:            domain.ErrConcurrentModification,
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "retry",
			checkHeader: func(t *testing.T, h http.Header) {
				t.Helper()
				require.Equal(t, "1", h.Get("Retry-After"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			writeError(rec, req, tt.err)

			require.Equal(t, tt.expectedStatus, rec.Code)
			require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

			var resp errorResponse
			err := json.NewDecoder(rec.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, tt.expectedCode, resp.Code)

			if tt.checkHeader != nil {
				tt.checkHeader(t, rec.Header())
			}
		})
	}
}

func TestDecodeJSON(t *testing.T) {
	t.Run("disallow unknown fields", func(t *testing.T) {
		body := strings.NewReader(`{"key":"123","unknown_field":true}`)
		req := httptest.NewRequest(http.MethodPost, "/", body)
		rec := httptest.NewRecorder()

		var dst activateRequest
		err := decodeJSON(rec, req, &dst)

		require.Error(t, err)
		require.True(t, errors.Is(err, domain.ErrInvalidInput))
	})

	t.Run("payload exceeding 1MB", func(t *testing.T) {
		largeBody := strings.NewReader(`{"key":"` + strings.Repeat("A", 1<<20) + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/", largeBody)
		rec := httptest.NewRecorder()

		var dst activateRequest
		err := decodeJSON(rec, req, &dst)

		require.Error(t, err)
		require.True(t, errors.Is(err, domain.ErrInvalidInput))
	})
}
