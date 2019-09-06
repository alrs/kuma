package tls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

var (
	DefaultEllipticCurve  = elliptic.P256()
	DefaultValidityPeriod = 365 * 24 * time.Hour
)

type KeyPair struct {
	CertPEM []byte
	KeyPEM  []byte
}

func NewSelfSignedCert(commonName string) (KeyPair, error) {
	key, err := ecdsa.GenerateKey(DefaultEllipticCurve, rand.Reader)
	if err != nil {
		return KeyPair{}, errors.Wrap(err, "failed to generate TLS key")
	}

	certBytes, err := generateCert(commonName, key)
	if err != nil {
		return KeyPair{}, err
	}

	keyBytes, err := marshalKey(key)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		CertPEM: certBytes,
		KeyPEM:  keyBytes,
	}, nil
}

func generateCert(commonName string, key *ecdsa.PrivateKey) ([]byte, error) {
	csr, err := newCert(commonName)
	if err != nil {
		return nil, err
	}
	certDerBytes, err := x509.CreateCertificate(rand.Reader, &csr, &csr, &key.PublicKey, key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate TLS certificate")
	}
	var certBuf bytes.Buffer
	if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: certDerBytes}); err != nil {
		return nil, errors.Wrap(err, "failed to PEM encode TLS certificate")
	}
	return certBuf.Bytes(), nil
}

func newCert(commonName string) (x509.Certificate, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(DefaultValidityPeriod)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return x509.Certificate{}, errors.Wrap(err, "failed to generate serial number")
	}
	csr := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	return csr, nil
}

func marshalKey(key *ecdsa.PrivateKey) ([]byte, error) {
	keyDerBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal TLS key")
	}
	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDerBytes}); err != nil {
		return nil, errors.Wrap(err, "failed to PEM encode TLS key")
	}
	return keyBuf.Bytes(), nil
}