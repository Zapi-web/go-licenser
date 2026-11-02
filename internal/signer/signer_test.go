package signer

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zapi-web/go-licenser/internal/domain"
)

const key = "some-key"

func newTestSigner(t *testing.T, ttl time.Duration, version int) (*Ed25519Signer, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)

	s, err := New(base64.StdEncoding.EncodeToString(der), ttl, version)
	require.NoError(t, err)

	return s, pub
}

func TestEd25519SignerSign(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ttl := 24 * time.Hour
	version := 1

	tests := []struct {
		name        string
		lic         *domain.License
		wantExpires time.Time
		wantHWID    *string
	}{
		{
			name:        "no license expiresAt",
			lic:         &domain.License{LicenseKey: key},
			wantExpires: now.Add(ttl),
		},
		{
			name:        "license wins",
			lic:         &domain.License{LicenseKey: key, ExpiresAt: new(now.Add(time.Hour))},
			wantExpires: now.Add(time.Hour),
		},
		{
			name:        "ttl wins",
			lic:         &domain.License{LicenseKey: key, ExpiresAt: new(now.Add(48 * time.Hour))},
			wantExpires: now.Add(ttl),
		},
		{
			name:        "with HWID",
			lic:         &domain.License{LicenseKey: key, HWID: new("hw")},
			wantExpires: now.Add(ttl),
			wantHWID:    new("hw"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, pub := newTestSigner(t, ttl, version)
			s.now = func() time.Time { return now }

			signed, err := s.Sign(tt.lic)
			require.NoError(t, err)

			payloadBytes, err := base64.StdEncoding.DecodeString(signed.Payload)
			require.NoError(t, err)
			sigBytes, err := base64.StdEncoding.DecodeString(signed.Signature)
			require.NoError(t, err)

			require.True(t, ed25519.Verify(pub, payloadBytes, sigBytes))

			var payload domain.LicensePayload
			require.NoError(t, json.Unmarshal(payloadBytes, &payload))
			require.Equal(t, tt.lic.LicenseKey, payload.LicenseKey)
			require.Equal(t, tt.wantHWID, payload.HWID)
			require.Equal(t, version, payload.Version)
			require.True(t, tt.wantExpires.Equal(*payload.ExpiresAt))
			require.True(t, now.Equal(payload.IssuedAt))
		})
	}
}

func TestParsePrivateKey(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	validB64 := base64.StdEncoding.EncodeToString(der)

	rsaPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	rsaDer, err := x509.MarshalPKCS8PrivateKey(rsaPriv)
	require.NoError(t, err)

	tests := []struct {
		name            string
		input           string
		wantErrContains string
	}{
		{
			name:  "valid",
			input: validB64,
		},
		{
			name:            "invalid base64",
			input:           "%%%",
			wantErrContains: "decode",
		},
		{
			name:            "not pkcs8",
			input:           base64.StdEncoding.EncodeToString([]byte("not pkcs8 data")),
			wantErrContains: "parse",
		},
		{
			name:            "wrong key type",
			input:           base64.StdEncoding.EncodeToString(rsaDer),
			wantErrContains: "not an ed25519",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePrivateKey(tt.input)
			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
				return
			}
			require.NoError(t, err)
		})
	}
}
