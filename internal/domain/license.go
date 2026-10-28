package domain

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusRevoked  Status = "revoked"
	StatusBanned   Status = "banned"
)

type License struct {
	ID          string
	LicenseKey  string
	Status      Status
	HWID        *string
	CreatedAt   time.Time
	ActivatedAt *time.Time
	ExpiresAt   *time.Time
}

type LicensePayload struct {
	LicenseKey string     `json:"license_key"`
	HWID       *string    `json:"hwid,omitempty"`
	IssuedAt   time.Time  `json:"issued_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Version    int        `json:"v"`
}

type SignedLicense struct {
	Payload   string `json:"payload"`   // base64
	Signature string `json:"signature"` // base64
}
