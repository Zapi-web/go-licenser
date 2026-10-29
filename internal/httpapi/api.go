package httpapi

import (
	"context"
	"net/http"
	"time"

	"github.com/zapi-web/go-licenser/internal/domain"
)

type API struct {
	licenser   Licenser
	adminToken string
}

type Licenser interface {
	Activate(ctx context.Context, key, hwid string) (*domain.SignedLicense, error)
	Create(ctx context.Context, hwid *string, expiresAt *time.Time) (*domain.License, error)
	Renew(ctx context.Context, key, hwid string) (*domain.SignedLicense, error)
	Get(ctx context.Context, key string) (*domain.License, error)
	Revoke(ctx context.Context, key string) error
	ResetHWID(ctx context.Context, key string) error
	Extend(ctx context.Context, key string, expiresAt time.Time) error
}

func New(lic Licenser, adminToken string) *API {
	return &API{
		licenser:   lic,
		adminToken: adminToken,
	}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	admin := a.requireAdmin

	mux.HandleFunc("GET /healthz", a.HandleHealthz)
	mux.HandleFunc("POST /v1/activate", a.HandleActivate)
	mux.HandleFunc("POST /v1/renew", a.HandleRenew)
	mux.Handle("POST /v1/licenses", admin(http.HandlerFunc(a.HandleCreate)))
	mux.Handle("GET /v1/licenses/{key}", admin(http.HandlerFunc(a.HandleGet)))
	mux.Handle("POST /v1/licenses/{key}/revoke", admin(http.HandlerFunc(a.HandleRevoke)))
	mux.Handle("POST /v1/licenses/{key}/reset-hwid", admin(http.HandlerFunc(a.HandleResetHWID)))
	mux.Handle("POST /v1/licenses/{key}/extend", admin(http.HandlerFunc(a.HandleExtend)))

	return recovery(logging(mux))
}
