package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/zapi-web/go-licenser/internal/domain"
	"github.com/zapi-web/go-licenser/internal/service/mocks"
)

const (
	key  = "test-key"
	hwid = "test-hwid"
)

func TestLicenserActivate(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-1 * time.Hour)
	signed := &domain.SignedLicense{Payload: "p", Signature: "s"}

	activeLic := func() *domain.License {
		return &domain.License{LicenseKey: key, Status: domain.StatusActive, ExpiresAt: &future}
	}

	tests := []struct {
		name    string
		key     string
		hwid    string
		setup   func(repo *mocks.LicenserRepository, signer *mocks.Signer)
		wantErr error
	}{
		{
			name: "success",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, signer *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(activeLic(), nil).Once()
				repo.On("BindHWID", mock.Anything, key, hwid).Return(nil).Once()
				signer.On("Sign", mock.Anything).Return(signed, nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "revoked",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, _ *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(
					&domain.License{LicenseKey: key, Status: domain.StatusRevoked}, nil,
				).Once()
			},
			wantErr: domain.ErrLicenseRevoked,
		},
		{
			name: "not-found",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, _ *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(nil, domain.ErrNotFound).Once()
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "expired",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, _ *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(
					&domain.License{LicenseKey: key, Status: domain.StatusActive, ExpiresAt: &past}, nil,
				).Once()
			},
			wantErr: domain.ErrLicenseExpired,
		},
		{
			name:    "empty-input",
			key:     "",
			hwid:    "",
			wantErr: domain.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewLicenserRepository(t)
			signer := mocks.NewSigner(t)
			kgs := mocks.NewKGS(t)

			if tt.setup != nil {
				tt.setup(repo, signer)
			}

			s := New(repo, signer, kgs)
			got, err := s.Activate(t.Context(), tt.key, tt.hwid)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			require.Equal(t, signed, got)
		})
	}
}

func TestLicenserRenew(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-1 * time.Hour)
	signed := &domain.SignedLicense{Payload: "p", Signature: "s"}

	activeLic := func() *domain.License {
		return &domain.License{LicenseKey: key, Status: domain.StatusActive, ExpiresAt: &future}
	}

	tests := []struct {
		name    string
		key     string
		hwid    string
		setup   func(repo *mocks.LicenserRepository, signer *mocks.Signer)
		wantErr error
	}{
		{
			name: "success",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, signer *mocks.Signer) {
				lic := activeLic()
				lic.HWID = new(hwid)
				repo.On("Get", mock.Anything, key).Return(lic, nil).Once()
				signer.On("Sign", mock.Anything).Return(signed, nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "revoked",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, _ *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(
					&domain.License{LicenseKey: key, Status: domain.StatusRevoked}, nil,
				).Once()
			},
			wantErr: domain.ErrLicenseRevoked,
		},
		{
			name: "not-found",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, _ *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(nil, domain.ErrNotFound).Once()
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "expired",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, _ *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(
					&domain.License{LicenseKey: key, HWID: new(hwid), Status: domain.StatusActive, ExpiresAt: &past}, nil,
				).Once()
			},
			wantErr: domain.ErrLicenseExpired,
		},
		{
			name:    "empty-input",
			key:     "",
			hwid:    "",
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "hwid-mismatch",
			key:  key,
			hwid: hwid,
			setup: func(repo *mocks.LicenserRepository, _ *mocks.Signer) {
				repo.On("Get", mock.Anything, key).Return(
					&domain.License{LicenseKey: key, Status: domain.StatusActive, HWID: new("another-hwid")}, nil,
				).Once()
			},
			wantErr: domain.ErrHWIDMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewLicenserRepository(t)
			signer := mocks.NewSigner(t)
			kgs := mocks.NewKGS(t)

			if tt.setup != nil {
				tt.setup(repo, signer)
			}

			s := New(repo, signer, kgs)
			got, err := s.Renew(t.Context(), tt.key, tt.hwid)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			require.Equal(t, signed, got)
		})
	}
}

func TestLicenserCreate(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name            string
		setup           func(repo *mocks.LicenserRepository, kgs *mocks.KGS)
		hwid            *string
		expiresAt       *time.Time
		wantErr         error
		wantErrContains string
	}{
		{
			name: "success",
			setup: func(repo *mocks.LicenserRepository, kgs *mocks.KGS) {
				kgs.On("GenerateKey").Return(key, nil).Once()
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.License")).Return(nil).Once()
			},
			hwid:      new(hwid),
			expiresAt: &future,
		},
		{
			name:    "no-hwid",
			hwid:    new(""),
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:      "invalid-expires-at",
			hwid:      new(hwid),
			expiresAt: &past,
			wantErr:   domain.ErrInvalidInput,
		},
		{
			name: "unique-generation-fail",
			setup: func(repo *mocks.LicenserRepository, kgs *mocks.KGS) {
				kgs.On("GenerateKey").Return(key, nil).Times(5)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.License")).Return(domain.ErrAlreadyExists).Times(5)
			},
			hwid:      new(hwid),
			expiresAt: &future,
			wantErr:   errUniqueGeneration,
		},
		{
			name: "retry-on-collision",
			setup: func(repo *mocks.LicenserRepository, kgs *mocks.KGS) {
				kgs.On("GenerateKey").Return(key, nil).Twice()
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.License")).Return(domain.ErrAlreadyExists).Once()
				repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.License")).Return(nil).Once()
			},
			hwid:      new(hwid),
			expiresAt: &future,
		},
		{
			name: "generate-key-fail",
			setup: func(_ *mocks.LicenserRepository, kgs *mocks.KGS) {
				kgs.On("GenerateKey").Return("", errors.New("fail")).Once()
			},
			hwid:            new(hwid),
			expiresAt:       &future,
			wantErrContains: "generate key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewLicenserRepository(t)
			signer := mocks.NewSigner(t)
			kgs := mocks.NewKGS(t)

			if tt.setup != nil {
				tt.setup(repo, kgs)
			}

			s := New(repo, signer, kgs)
			got, err := s.Create(t.Context(), tt.hwid, tt.expiresAt)

			if tt.wantErrContains != "" {
				assert.ErrorContains(t, err, tt.wantErrContains)
				assert.Nil(t, got)
				return
			}
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			assert.NoError(t, err)
			require.Equal(t, *tt.hwid, *got.HWID)
			require.Equal(t, *tt.expiresAt, *got.ExpiresAt)
			require.Equal(t, key, got.LicenseKey)
			require.Equal(t, domain.StatusActive, got.Status)
		})
	}
}

func TestLicenserExtend(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-1 * time.Hour)

	activeLic := func() *domain.License {
		return &domain.License{LicenseKey: key, Status: domain.StatusActive, ExpiresAt: &future}
	}

	tests := []struct {
		name      string
		setup     func(repo *mocks.LicenserRepository)
		key       string
		expiresAt time.Time
		wantErr   error
	}{
		{
			name: "success",
			setup: func(repo *mocks.LicenserRepository) {
				repo.On("Get", mock.Anything, key).Return(activeLic(), nil).Once()
				repo.On("ExtendExpiry", mock.Anything, key, future).Return(nil).Once()
			},
			key:       key,
			expiresAt: future,
		},
		{
			name:    "no-key",
			key:     "",
			wantErr: domain.ErrInvalidInput,
		},
		{
			name:      "invalid-expires-at",
			key:       key,
			expiresAt: past,
			wantErr:   domain.ErrInvalidInput,
		},
		{
			name: "bad-license",
			setup: func(repo *mocks.LicenserRepository) {
				repo.On("Get", mock.Anything, key).Return(&domain.License{
					Status: domain.StatusBanned,
				}, nil).Once()
			},
			key:       key,
			expiresAt: future,
			wantErr:   domain.ErrLicenseBanned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewLicenserRepository(t)
			signer := mocks.NewSigner(t)
			kgs := mocks.NewKGS(t)

			if tt.setup != nil {
				tt.setup(repo)
			}

			s := New(repo, signer, kgs)
			err := s.Extend(t.Context(), tt.key, tt.expiresAt)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
