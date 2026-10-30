package httpapi

import (
	"net/http"

	"github.com/zapi-web/go-licenser/internal/domain"
)

func (a *API) HandleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (a *API) HandleActivate(w http.ResponseWriter, r *http.Request) {
	var req activateRequest

	err := decodeJSON(w, r, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if req.Key == "" || req.HWID == "" {
		writeError(w, r, domain.ErrInvalidInput)
		return
	}

	sigLic, err := a.licenser.Activate(r.Context(), req.Key, req.HWID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, sigLic)
}

func (a *API) HandleRenew(w http.ResponseWriter, r *http.Request) {
	var req renewRequest
	err := decodeJSON(w, r, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if req.HWID == "" || req.Key == "" {
		writeError(w, r, domain.ErrInvalidInput)
		return
	}

	sigLic, err := a.licenser.Renew(r.Context(), req.Key, req.HWID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, sigLic)
}

func (a *API) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	err := decodeJSON(w, r, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	lic, err := a.licenser.Create(r.Context(), req.HWID, req.ExpiresAt)
	if err != nil {
		writeError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, toLicenseResponse(lic))
}

func (a *API) HandleGet(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeError(w, r, domain.ErrInvalidInput)
		return
	}

	lic, err := a.licenser.Get(r.Context(), key)
	if err != nil {
		writeError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, toLicenseResponse(lic))
}

func (a *API) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeError(w, r, domain.ErrInvalidInput)
		return
	}

	err := a.licenser.Revoke(r.Context(), key)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
func (a *API) HandleResetHWID(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeError(w, r, domain.ErrInvalidInput)
		return
	}

	err := a.licenser.ResetHWID(r.Context(), key)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) HandleExtend(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeError(w, r, domain.ErrInvalidInput)
		return
	}

	var req extendRequest
	err := decodeJSON(w, r, &req)
	if err != nil {
		writeError(w, r, err)
		return
	}

	err = a.licenser.Extend(r.Context(), key, req.ExpiresAt)
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
