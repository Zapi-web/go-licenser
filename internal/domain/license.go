package domain

import "time"

type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusRevoked   Status = "revoked"
	StatusBanned    Status = "banned"
	StatusSuspended Status = "suspended"
)

type License struct {
	ID             string
	LicenseKey     string
	Status         Status
	HWID           *string
	MaxActivations uint
	CreatedAt      time.Time
	ActivatedAt    *time.Time
	ExpiresAt      *time.Time
}
