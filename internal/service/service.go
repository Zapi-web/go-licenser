package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zapi-web/go-licenser/internal/domain"
)

type Licenser struct {
	repo   LicenserRepository
	signer Signer
	kgs    KGS
	now    func() time.Time
}

func New(repo LicenserRepository, signer Signer, kgs KGS) *Licenser {
	return &Licenser{
		repo:   repo,
		signer: signer,
		kgs:    kgs,
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

func (l *Licenser) Create(ctx context.Context, hwid *string, expiresAt *time.Time) (*domain.License, error) {
	if hwid != nil && *hwid == "" {
		return nil, fmt.Errorf("%w: hwid is empty", domain.ErrInvalidInput)
	}
	if expiresAt != nil && expiresAt.Before(l.now()) {
		return nil, fmt.Errorf("%w: expires_at is in the past", domain.ErrInvalidInput)
	}

	for range 5 {
		key, err := l.kgs.GenerateKey()
		if err != nil {
			return nil, fmt.Errorf("generate key: %w", err)
		}

		lic := &domain.License{
			LicenseKey: key,
			HWID:       hwid,
			ExpiresAt:  expiresAt,
			Status:     domain.StatusActive,
		}

		err = l.repo.Create(ctx, lic)
		if errors.Is(err, domain.ErrAlreadyExists) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return lic, nil
	}

	return nil, errors.New("failed to generate unique license key")
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
