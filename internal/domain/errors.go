package domain

import "errors"

var (
	ErrNotFound               = errors.New("license not found")
	ErrInvalidInput           = errors.New("invalid input")
	ErrHWIDMismatch           = errors.New("hwid mismatch")
	ErrLicenseExpired         = errors.New("license had expired")
	ErrLicenseInactive        = errors.New("license is inactive")
	ErrLicenseBanned          = errors.New("license is banned")
	ErrLicenseRevoked         = errors.New("license had revoked")
	ErrInvalidKeySize         = errors.New("invalid private key size")
	ErrCorruptedKey           = errors.New("private key is corrupted: public key mismatch")
	ErrAlreadyBound           = errors.New("hwid is already bounded")
	ErrConcurrentModification = errors.New("license state changed concurrently, retry")
	ErrAlreadyExists          = errors.New("already exists")
	ErrUnauthorized           = errors.New("user unauthorized")
)
