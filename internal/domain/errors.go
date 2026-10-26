package domain

import "errors"

var (
	ErrNotFound       = errors.New("license not found")
	ErrLicenseExpired = errors.New("license had expired")
)
