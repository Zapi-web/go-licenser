package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/zapi-web/go-licenser/internal/domain"
	"github.com/zapi-web/go-licenser/internal/httpapi/mocks"
)

func TestAPI_HandleActivate(t *testing.T) {
	t.Run("empty required fields", func(t *testing.T) {
		api := New(nil, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/activate", strings.NewReader(`{"key":"","hwid":""}`))
		rec := httptest.NewRecorder()

		api.HandleActivate(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("licenser returns domain error", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("Activate", mock.Anything, "key-1", "hwid-1").Return(nil, domain.ErrHWIDMismatch).Once()

		api := New(lic, nil, "")

		body := strings.NewReader(`{"key":"key-1","hwid":"hwid-1"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/activate", body)
		rec := httptest.NewRecorder()

		api.HandleActivate(rec, req)

		require.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		expected := &domain.SignedLicense{Signature: "sig"}

		lic := mocks.NewLicenser(t)
		lic.On("Activate", mock.Anything, "key-1", "hwid-1").Return(expected, nil).Once()

		api := New(lic, nil, "")

		body := strings.NewReader(`{"key":"key-1","hwid":"hwid-1"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/activate", body)
		rec := httptest.NewRecorder()

		api.HandleActivate(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp domain.SignedLicense
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		require.Equal(t, expected.Signature, resp.Signature)
	})
}

func TestAPI_HandleRenew(t *testing.T) {
	t.Run("empty required fields", func(t *testing.T) {
		api := New(nil, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/renew", strings.NewReader(`{"key":"","hwid":""}`))
		rec := httptest.NewRecorder()

		api.HandleRenew(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("licenser returns domain error", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("Renew", mock.Anything, "key-1", "hwid-1").Return(nil, domain.ErrHWIDMismatch).Once()

		api := New(lic, nil, "")

		body := strings.NewReader(`{"key":"key-1","hwid":"hwid-1"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/renew", body)
		rec := httptest.NewRecorder()

		api.HandleRenew(rec, req)

		require.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		expected := &domain.SignedLicense{Signature: "sig"}

		lic := mocks.NewLicenser(t)
		lic.On("Renew", mock.Anything, "key-1", "hwid-1").Return(expected, nil).Once()

		api := New(lic, nil, "")

		body := strings.NewReader(`{"key":"key-1","hwid":"hwid-1"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/renew", body)
		rec := httptest.NewRecorder()

		api.HandleRenew(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		var resp domain.SignedLicense
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		require.Equal(t, expected.Signature, resp.Signature)
	})
}

func TestAPI_HandleCreate(t *testing.T) {
	t.Run("licenser returns domain error", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		hwid := ""
		lic.On("Create", mock.Anything, &hwid, (*time.Time)(nil)).Return(nil, domain.ErrInvalidInput).Once()

		api := New(lic, nil, "")

		body := strings.NewReader(`{"hwid":""}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/licenses", body)
		rec := httptest.NewRecorder()

		api.HandleCreate(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		expires := now.Add(time.Hour)
		hwid := "hwid"

		expected := &domain.License{
			LicenseKey: "key",
			HWID:       &hwid,
			ExpiresAt:  &expires,
			Status:     domain.StatusActive,
		}

		lic := mocks.NewLicenser(t)
		lic.On("Create", mock.Anything, &hwid, mock.Anything).Return(expected, nil).Once()

		api := New(lic, nil, "")

		reqBody := fmt.Sprintf(`{"hwid":"%s","expires_at":"%s"}`, hwid, expires.Format(time.RFC3339))
		req := httptest.NewRequest(http.MethodPost, "/v1/licenses", strings.NewReader(reqBody))
		rec := httptest.NewRecorder()

		api.HandleCreate(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)

		var resp licenseResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		require.Equal(t, expected.LicenseKey, resp.Key)
	})
}

func TestAPI_HandleGet(t *testing.T) {
	t.Run("empty key", func(t *testing.T) {
		api := New(nil, nil, "")

		req := httptest.NewRequest(http.MethodGet, "/v1/licenses/", nil)
		rec := httptest.NewRecorder()

		api.HandleGet(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("licenser returns not found", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("Get", mock.Anything, "key-123").
			Return(nil, domain.ErrNotFound).
			Once()

		api := New(lic, nil, "")

		req := httptest.NewRequest(http.MethodGet, "/v1/licenses/key-123", nil)
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleGet(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		hwid := "hwid-1"
		expected := &domain.License{
			LicenseKey: "key-123",
			HWID:       &hwid,
			ExpiresAt:  &now,
			Status:     domain.StatusActive,
		}

		lic := mocks.NewLicenser(t)
		lic.On("Get", mock.Anything, "key-123").
			Return(expected, nil).
			Once()

		api := New(lic, nil, "")

		req := httptest.NewRequest(http.MethodGet, "/v1/licenses/key-123", nil)
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleGet(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp licenseResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		require.Equal(t, expected.LicenseKey, resp.Key)
	})
}

func TestAPI_HandleRevoke(t *testing.T) {
	t.Run("empty key", func(t *testing.T) {
		api := New(nil, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/licenses//revoke", nil)
		rec := httptest.NewRecorder()

		api.HandleRevoke(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("licenser returns error", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("Revoke", mock.Anything, "key-123").
			Return(domain.ErrNotFound).
			Once()

		api := New(lic, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/licenses/key-123/revoke", nil)
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleRevoke(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("Revoke", mock.Anything, "key-123").
			Return(nil).
			Once()

		api := New(lic, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/licenses/key-123/revoke", nil)
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleRevoke(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}

func TestAPI_HandleResetHWID(t *testing.T) {
	t.Run("empty key", func(t *testing.T) {
		api := New(nil, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/licenses//reset-hwid", nil)
		rec := httptest.NewRecorder()

		api.HandleResetHWID(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("licenser returns error", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("ResetHWID", mock.Anything, "key-123").
			Return(domain.ErrNotFound).
			Once()

		api := New(lic, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/licenses/key-123/reset-hwid", nil)
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleResetHWID(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("ResetHWID", mock.Anything, "key-123").
			Return(nil).
			Once()

		api := New(lic, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/licenses/key-123/reset-hwid", nil)
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleResetHWID(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}

func TestAPI_HandleExtend(t *testing.T) {
	t.Run("empty key", func(t *testing.T) {
		api := New(nil, nil, "")

		body := strings.NewReader(`{"expires_at":"2030-01-01T00:00:00Z"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/licenses//extend", body)
		rec := httptest.NewRecorder()

		api.HandleExtend(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		api := New(nil, nil, "")

		req := httptest.NewRequest(http.MethodPost, "/v1/licenses/key-123/extend", strings.NewReader(`{invalid json}`))
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleExtend(rec, req)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("licenser returns domain error", func(t *testing.T) {
		lic := mocks.NewLicenser(t)
		lic.On("Extend", mock.Anything, "key-123", mock.Anything).
			Return(domain.ErrNotFound).
			Once()

		api := New(lic, nil, "")

		body := strings.NewReader(`{"expires_at":"2030-01-01T00:00:00Z"}`)
		req := httptest.NewRequest(http.MethodPost, "/v1/licenses/key-123/extend", body)
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleExtend(rec, req)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		targetDate := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)

		lic := mocks.NewLicenser(t)
		lic.On("Extend", mock.Anything, "key-123", mock.MatchedBy(func(t time.Time) bool {
			return t.Equal(targetDate)
		})).Return(nil).Once()

		api := New(lic, nil, "")

		body := fmt.Sprintf(`{"expires_at":"%s"}`, targetDate.Format(time.RFC3339))
		req := httptest.NewRequest(http.MethodPost, "/v1/licenses/key-123/extend", strings.NewReader(body))
		req.SetPathValue("key", "key-123")
		rec := httptest.NewRecorder()

		api.HandleExtend(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}
