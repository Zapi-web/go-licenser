package domain

import "errors"

var (
	ErrNotFound       = errors.New("license not found")
	ErrLicenseExpired = errors.New("license had expired")
	ErrInvalidKeySize = errors.New("invalid private key size")
	ErrCorruptedKey   = errors.New("private key is corrupted: public key mismatch")
)
