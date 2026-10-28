package service

import (
	"context"
	"fmt"
	"time"

	"github.com/zapi-web/go-licenser/internal/domain"
)

type Licenser struct {
	repo   LicenserRepository
	signer Signer
	now    func() time.Time
}

func New(repo LicenserRepository, signer Signer) *Licenser {
	return &Licenser{
		repo:   repo,
		signer: signer,
		now:    time.Now,
	}
}

func (l *Licenser) Activate(ctx context.Context, key, hwid string) (*domain.SignedLicense, error) {
	if key == "" || hwid == "" {
		return nil, domain.ErrInvalidInput
	}

	lic, err := l.repo.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	err = checkActive(lic)
	if err != nil {
		return nil, err
	}

	if lic.ExpiresAt != nil && lic.ExpiresAt.Before(l.now()) {
		return nil, domain.ErrLicenseExpired
	}

	err = l.repo.BindHWID(ctx, key, hwid)
	if err != nil {
		return nil, err
	}

	lic.HWID = &hwid

	signedLic, err := l.signer.Sign(lic)
	if err != nil {
		return nil, fmt.Errorf("failed to sign license: %w", err)
	}

	return signedLic, nil
}

func (l *Licenser) Renew(ctx context.Context, key, hwid string) (*domain.SignedLicense, error) {
	if key == "" || hwid == "" {
		return nil, domain.ErrInvalidInput
	}

	lic, err := l.repo.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	err = checkActive(lic)
	if err != nil {
		return nil, err
	}

	if lic.HWID == nil {
		return nil, domain.ErrLicenseInactive
	}

	if lic.ExpiresAt != nil && lic.ExpiresAt.Before(l.now()) {
		return nil, domain.ErrLicenseExpired
	}

	if *lic.HWID != hwid {
		return nil, domain.ErrHWIDMismatch
	}

	signedLic, err := l.signer.Sign(lic)
	if err != nil {
		return nil, fmt.Errorf("failed to sign license: %w", err)
	}
	return signedLic, nil
}

func (l *Licenser) Get(ctx context.Context, key string) (*domain.License, error) {
	if key == "" {
		return nil, domain.ErrInvalidInput
	}

	return l.repo.Get(ctx, key)
}

func (l *Licenser) Create(ctx context.Context, key string, hwid *string, expiresAt *time.Time) error {
	if key == "" {
		return domain.ErrInvalidInput
	}
	if expiresAt != nil {
		if expiresAt.Before(l.now()) {
			return domain.ErrInvalidInput
		}
	}

	license := &domain.License{
		LicenseKey: key,
		HWID:       hwid,
		ExpiresAt:  expiresAt,
		Status:     domain.StatusActive,
	}

	return l.repo.Create(ctx, license)
}

func (l *Licenser) Revoke(ctx context.Context, key string) error {
	if key == "" {
		return domain.ErrInvalidInput
	}
	return l.repo.UpdateStatus(ctx, key, domain.StatusRevoked)
}

func (l *Licenser) ResetHWID(ctx context.Context, key string) error {
	if key == "" {
		return domain.ErrInvalidInput
	}

	return l.repo.ResetHWID(ctx, key)
}

func (l *Licenser) Extend(ctx context.Context, key string, expiresAt time.Time) error {
	if key == "" {
		return domain.ErrInvalidInput
	}

	if expiresAt.Before(l.now()) {
		return domain.ErrInvalidInput
	}

	lic, err := l.repo.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := checkActive(lic); err != nil {
		return err
	}

	return l.repo.ExtendExpiry(ctx, key, expiresAt)
}

func checkActive(lic *domain.License) error {
	switch lic.Status {
	case domain.StatusActive:
		return nil
	case domain.StatusRevoked:
		return domain.ErrLicenseRevoked
	case domain.StatusBanned:
		return domain.ErrLicenseBanned
	case domain.StatusInactive:
		return domain.ErrLicenseInactive
	default:
		return domain.ErrLicenseInactive
	}
}
