package service

import (
	"context"
	"time"

	"github.com/zapi-web/go-licenser/internal/domain"
)

type LicenserRepository interface {
	Create(ctx context.Context, lic *domain.License) error
	Get(ctx context.Context, licenseKey string) (*domain.License, error)
	BindHWID(ctx context.Context, licenseKey string, hwid string) error
	ResetHWID(ctx context.Context, licenseKey string) error
	ExtendExpiry(ctx context.Context, licenseKey string, newExpiresAt time.Time) error
	UpdateStatus(ctx context.Context, licenseKey string, newStatus domain.Status) error
}

type Signer interface {
	Sign(license *domain.License) (*domain.SignedLicense, error)
}
