package signer

import (
	"crypto/ed25519"
	"crypto/x509"
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
	now        func() time.Time
}

func New(key string, v int) (*Ed25519Signer, error) {
	keyParsed, err := parsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	newSigner := &Ed25519Signer{privateKey: keyParsed, version: v, now: time.Now}

	return newSigner, nil
}

func (s *Ed25519Signer) Sign(lic *domain.License) (*domain.SignedLicense, error) {
	if lic == nil {
		return nil, errors.New("license can not be nil")
	}

	payload := domain.LicensePayload{
		LicenseKey: lic.LicenseKey,
		HWID:       lic.HWID,
		IssuedAt:   s.now().UTC(),
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

func parsePrivateKey(b64 string) (ed25519.PrivateKey, error) {
	der, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ED25519_PRIVATE_KEY: %w", err)
	}

	k, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ED25519_PRIVATE_KEY as pkcs8: %w", err)
	}

	priv, ok := k.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("ED25519_PRIVATE_KEY is not an ed25519 key")
	}

	return priv, nil
}
