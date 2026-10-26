package signer

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/zapi-web/go-licenser/internal/domain"
)

type Ed25519Signer struct {
	privateKey ed25519.PrivateKey
	version    int
}

func New(key ed25519.PrivateKey, v int) (*Ed25519Signer, error) {
	newSigner := &Ed25519Signer{privateKey: key, version: v}

	err := newSigner.Validate()
	if err != nil {
		return nil, err
	}

	return newSigner, nil
}

func (s *Ed25519Signer) Sign(lic *domain.License) (*domain.SignedLicense, error) {
	if lic == nil {
		return nil, errors.New("license can not be nil")
	}

	payload := domain.LicensePayload{
		LicenseKey: lic.LicenseKey,
		HWID:       lic.HWID,
		IssuedAt:   time.Now().UTC(),
		ExpiresAt:  lic.ExpiresAt,
		Version:    s.version,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload struct: %w", err)
	}

	signatureBytes := ed25519.Sign(s.privateKey, payloadBytes)

	signLicense := &domain.SignedLicense{
		Payload:   base64.StdEncoding.EncodeToString(payloadBytes),
		Signature: base64.StdEncoding.EncodeToString(signatureBytes),
	}

	return signLicense, nil
}

func (s *Ed25519Signer) Validate() error {
	if len(s.privateKey) != ed25519.PrivateKeySize {
		return domain.ErrInvalidKeySize
	}

	expectedKey := ed25519.NewKeyFromSeed(s.privateKey.Seed())

	if !bytes.Equal(s.privateKey, expectedKey) {
		return domain.ErrCorruptedKey
	}

	return nil
}
