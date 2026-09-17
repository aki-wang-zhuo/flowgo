package endpoint

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestValidateTLS_Disabled(t *testing.T) {
	cfg := HttpConfig{Https: false}
	if err := cfg.ValidateTLS(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTLS_MissingPem(t *testing.T) {
	cfg := HttpConfig{Https: true}
	if err := cfg.ValidateTLS(); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateTLS_ValidSelfSigned(t *testing.T) {
	certPEM, keyPEM := mustSelfSignedPEM(t)
	cfg := HttpConfig{Https: true, CertPem: certPEM, KeyPem: keyPEM}
	if err := cfg.ValidateTLS(); err != nil {
		t.Fatal(err)
	}
	if cfg.TLSFingerprint() == "" {
		t.Fatal("empty fingerprint")
	}
}

func mustSelfSignedPEM(t *testing.T) (certPEM, keyPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "flowgo-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}))
	return certPEM, keyPEM
}
