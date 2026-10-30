package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/zapi-web/go-licenser/internal/domain"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		status int
		code   string
		msg    string
	)

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, msg = http.StatusNotFound, "not_found", "license not found"
	case errors.Is(err, domain.ErrInvalidInput):
		status, code, msg = http.StatusBadRequest, "invalid_input", "invalid input"
	case errors.Is(err, domain.ErrAlreadyBound):
		status, code, msg = http.StatusConflict, "already_bound", "license is already bound to another device"
	case errors.Is(err, domain.ErrHWIDMismatch):
		status, code, msg = http.StatusConflict, "hwid_mismatch", "license is bound to another device"
	case errors.Is(err, domain.ErrAlreadyExists):
		status, code, msg = http.StatusConflict, "already_exists", "license already exists"
	case errors.Is(err, domain.ErrLicenseExpired):
		status, code, msg = http.StatusForbidden, "license_expired", "license has expired"
	case errors.Is(err, domain.ErrLicenseRevoked):
		status, code, msg = http.StatusForbidden, "license_revoked", "license has been revoked"
	case errors.Is(err, domain.ErrLicenseBanned):
		status, code, msg = http.StatusForbidden, "license_banned", "license is banned"
	case errors.Is(err, domain.ErrLicenseInactive):
		status, code, msg = http.StatusForbidden, "license_inactive", "license is inactive"
	case errors.Is(err, domain.ErrConcurrentModification):
		w.Header().Set("Retry-After", "1")
		status, code, msg = http.StatusServiceUnavailable, "retry", "please try again"
	case errors.Is(err, domain.ErrUnauthorized):
		status, code, msg = http.StatusUnauthorized, "unauthorized", "user is unauthorized"
	default:
		// #nosec G706 -- r.URL.Path is escaped by slog.NewJSONHandler
		slog.Error("internal error", "err", err, "method", r.Method, "path", r.URL.Path)
		status, code, msg = http.StatusInternalServerError, "internal_error", "internal error"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{Code: code, Message: msg}); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w: %w", domain.ErrInvalidInput, err)
	}
	return nil
}

func respondJSON(w http.ResponseWriter, code int, dst any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(dst); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}
