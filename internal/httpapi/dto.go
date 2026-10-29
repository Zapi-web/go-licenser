package httpapi

import (
	"time"

	"github.com/zapi-web/go-licenser/internal/domain"
)

type activateRequest struct {
	Key  string `json:"key"`
	HWID string `json:"hwid"`
}

type renewRequest struct {
	Key  string `json:"key"`
	HWID string `json:"hwid"`
}

type createRequest struct {
	HWID      *string    `json:"hwid,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type extendRequest struct {
	ExpiresAt time.Time `json:"expires_at"`
}

type licenseResponse struct {
	Key         string     `json:"key"`
	Status      string     `json:"status"`
	HWID        *string    `json:"hwid,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
}

func toLicenseResponse(l *domain.License) licenseResponse {
	return licenseResponse{
		Key:         l.LicenseKey,
		Status:      string(l.Status),
		HWID:        l.HWID,
		CreatedAt:   l.CreatedAt,
		ExpiresAt:   l.ExpiresAt,
		ActivatedAt: l.ActivatedAt,
	}
}
