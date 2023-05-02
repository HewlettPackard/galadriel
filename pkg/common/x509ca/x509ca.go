package x509ca

import (
	"context"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
	"time"
)

// X509CA is the interface used to sign X509 certificates.
type X509CA interface {
	// IssueX509Certificate issues an X509 certificate and returns the leaf certificate and the certificate chain.
	IssueX509Certificate(context.Context, *X509CertificateParams) ([]*x509.Certificate, error)
}

// X509CertificateParams holds the parameters for issuing an X509 certificate.
type X509CertificateParams struct {
	// PublicKey to be set in the certificate
	PublicKey crypto.PublicKey
	Subject   pkix.Name
	URIs      []*url.URL
	DNSNames  []string
	TTL       time.Duration
}
