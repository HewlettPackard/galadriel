package cryptoutil

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jmhodges/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var clk = clock.NewFake()

func TestIsSelfSigned(t *testing.T) {
	cert, _ := createRootCA(t)
	assert.True(t, IsSelfSigned(cert))

	cert, _ = createCert(t, DefaultKeyType)
	assert.False(t, IsSelfSigned(cert))
}

func TestCertificatesMatch(t *testing.T) {
	cert, _ := createRootCA(t)
	assert.True(t, CertificatesMatch(cert, cert))

	cert2, _ := createCert(t, DefaultKeyType)
	assert.False(t, CertificatesMatch(cert, cert2))
}

func TestVerifyCertificateChain(t *testing.T) {
	now := clk.Now()

	// Generate leaf, intermediate, and root certificates
	leaf, intermediate, root := createCertChain(t, DefaultKeyType)
	otherRoot, _ := createRootCA(t)

	// Test valid certificate chain
	certChain := []*x509.Certificate{leaf, intermediate}
	intermediates := []*x509.Certificate{intermediate}
	roots := []*x509.Certificate{root, otherRoot}

	err := VerifyCertificateChain(certChain, intermediates, roots, now)
	require.NoError(t, err)

	// Test certificate chain with missing root
	missingRootChain := []*x509.Certificate{leaf}
	missingRootIntermediates := []*x509.Certificate{intermediate}
	missingRoots := []*x509.Certificate{}

	err = VerifyCertificateChain(missingRootChain, missingRootIntermediates, missingRoots, now)
	require.Error(t, err)

	// Test certificate chain with wrong root
	wrongRootChain := []*x509.Certificate{leaf, intermediate}
	wrongRootIntermediates := []*x509.Certificate{intermediate}
	wrongRootRoots := []*x509.Certificate{otherRoot}

	err = VerifyCertificateChain(wrongRootChain, wrongRootIntermediates, wrongRootRoots, now)
	require.Error(t, err)
}

func TestSplitCertsIntoRootsAndIntermediates(t *testing.T) {
	_, intermediate, root := createCertChain(t, DefaultKeyType)

	bundle := []*x509.Certificate{intermediate, root, intermediate, root}

	roots, intermediates := SplitCertsIntoRootsAndIntermediates(bundle)
	require.Len(t, intermediates, 1)
	require.Len(t, roots, 1)
	require.True(t, CertificatesMatch(root, roots[0]))
	require.True(t, CertificatesMatch(intermediate, intermediates[0]))
}

func TestRemoveCertificateFromBundle(t *testing.T) {
	_, intermediate, root := createCertChain(t, DefaultKeyType)
	bundle := []*x509.Certificate{intermediate, root, intermediate, root}

	// Remove intermediate
	remaining := RemoveCertificateFromBundle(bundle, intermediate)
	require.Equal(t, len(remaining), 2)

	// Remove root
	remaining = RemoveCertificateFromBundle(remaining, root)
	require.Equal(t, len(remaining), 0)
}

func TestLoadCertificate(t *testing.T) {
	// not a certificate
	_, err := LoadCertificate(rsaKeyPath)
	require.Error(t, err)

	// success
	cert, err := LoadCertificate(certPath)
	require.NoError(t, err)
	require.NotNil(t, cert)
}

func TestLoadCertificates(t *testing.T) {
	chain, err := LoadCertificates(certChainPath)
	require.NoError(t, err)
	require.Len(t, chain, 2)
}

func TestParseCertificate(t *testing.T) {
	// not a certificate
	_, err := ParseCertificate(readFile(t, rsaKeyPath))
	require.Error(t, err)

	// success with one certificate
	cert, err := ParseCertificate(readFile(t, certPath))
	require.NoError(t, err)
	require.NotNil(t, cert)
}

func TestParseCertificates(t *testing.T) {
	chain, err := ParseCertificates(readFile(t, certChainPath))
	require.NoError(t, err)
	require.NotNil(t, chain)
	require.Len(t, chain, 2)
}

func TestEncodeCertificates(t *testing.T) {
	cert, err := LoadCertificate(certPath)
	require.NoError(t, err)
	expCertPem, err := os.ReadFile(certPath)
	require.NoError(t, err)
	require.Equal(t, expCertPem, EncodeCertificate(cert))

}

func TestCreateX509Template(t *testing.T) {
	key, err := GenerateSigner(ECP384)
	require.NoError(t, err)
	uris := []*url.URL{{Scheme: "https", Host: "domain.test"}}
	dns := []string{"test"}
	name := pkix.Name{CommonName: "test-cn"}
	twoHours := 2 * time.Hour

	tmpl, err := CreateX509Template(clk, key.Public(), name, uris, dns, twoHours)
	require.NoError(t, err)
	require.NotNil(t, tmpl)
	assert.False(t, tmpl.IsCA)
	assert.Equal(t, key.Public(), tmpl.PublicKey)
	assert.Equal(t, name, tmpl.Subject)
	assert.Equal(t, uris, tmpl.URIs)
	assert.Equal(t, dns, tmpl.DNSNames)
	assert.Equal(t, clk.Now().Add(twoHours).UTC(), tmpl.NotAfter)
	assert.Equal(t, clk.Now().Add(-NotBeforeTolerance).UTC(), tmpl.NotBefore)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}, tmpl.ExtKeyUsage)
	assert.Equal(t, x509.KeyUsageKeyEncipherment|x509.KeyUsageKeyAgreement|x509.KeyUsageDigitalSignature, tmpl.KeyUsage)
}

func TestCreateCATemplate(t *testing.T) {
	key, err := GenerateSigner(ECP384)
	require.NoError(t, err)
	name := pkix.Name{CommonName: "test-cn"}
	twoHours := 2 * time.Hour

	tmpl, err := CreateCATemplate(clk, key.Public(), name, twoHours)
	require.NoError(t, err)
	require.NotNil(t, tmpl)
	assert.True(t, tmpl.IsCA)
	assert.Equal(t, key.Public(), tmpl.PublicKey)
	assert.Equal(t, name, tmpl.Subject)
	assert.Equal(t, clk.Now().Add(twoHours).UTC(), tmpl.NotAfter)
	assert.Equal(t, clk.Now(), tmpl.NotBefore)
	assert.Equal(t, x509.KeyUsageCertSign|x509.KeyUsageCRLSign, tmpl.KeyUsage)
}

func TestCreateRootCATemplate(t *testing.T) {
	name := pkix.Name{CommonName: "test-cn"}
	twoHours := 2 * time.Hour

	tmpl, err := CreateRootCATemplate(clk, name, twoHours)
	require.NoError(t, err)
	require.NotNil(t, tmpl)
	assert.True(t, tmpl.IsCA)
	assert.Equal(t, name, tmpl.Subject)
	assert.Equal(t, clk.Now().Add(twoHours).UTC(), tmpl.NotAfter)
	assert.Equal(t, clk.Now(), tmpl.NotBefore)
	assert.Equal(t, x509.KeyUsageCertSign, tmpl.KeyUsage)
}

func TestSignX509(t *testing.T) {
	key, err := GenerateSigner(ECP384)
	require.NoError(t, err)
	uris := []*url.URL{{Scheme: "https", Host: "domain.test"}}
	dns := []string{"test"}
	name := pkix.Name{CommonName: "test-cn"}
	twoHours := 2 * time.Hour

	tmpl, err := CreateX509Template(clk, key.Public(), name, uris, dns, twoHours)
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	// create parent certificate for signing
	parentCert, signingKey := createRootCA(t)

	cert, err := SignX509(tmpl, parentCert, signingKey)
	require.NoError(t, err)
	require.NotNil(t, cert)
	// verify certificate signature was created with the signing key
	err = cert.CheckSignatureFrom(parentCert)
	require.NoError(t, err)
}

func TestSelfSignX509(t *testing.T) {
	name := pkix.Name{CommonName: "root"}
	tmpl, err := CreateRootCATemplate(clk, name, 5*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	cert, key, err := SelfSignX509(tmpl)
	require.NoError(t, err)
	require.NotNil(t, cert)
	require.NotNil(t, key)
	assert.Equal(t, 5*time.Minute, cert.NotAfter.Sub(cert.NotBefore))
	// verify certificate signature was created with the signing key
	err = cert.CheckSignatureFrom(cert)
	require.NoError(t, err)
}

func createRootCA(t *testing.T) (*x509.Certificate, crypto.PrivateKey) {
	name := pkix.Name{CommonName: "root-ca"}
	tmpl, err := CreateRootCATemplate(clk, name, 5*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	cert, key, err := SelfSignX509(tmpl)
	require.NoError(t, err)
	require.NotNil(t, cert)
	require.NotNil(t, key)
	return cert, key
}

func createCert(t *testing.T, keyType KeyType) (*x509.Certificate, crypto.PrivateKey) {
	// create parent certificate for signing
	parentCert, signingKey := createRootCA(t)

	tmpl, key := createCertTemplate(t, keyType, pkix.Name{CommonName: "leaf-cert"}, false)
	cert, err := SignX509(tmpl, parentCert, signingKey)
	require.NoError(t, err)
	require.NotNil(t, cert)

	return cert, key
}

func createCertChain(t *testing.T, keyType KeyType) (leaf *x509.Certificate, intermediate *x509.Certificate, root *x509.Certificate) {
	root, rootKey := createRootCA(t)

	intermediateTemplate, intermediateKey := createCertTemplate(t, keyType, pkix.Name{CommonName: "intermediate-cert"}, true)
	intermediate, err := SignX509(intermediateTemplate, root, rootKey)
	require.NoError(t, err)
	require.NotNil(t, intermediate)

	certTemplate, _ := createCertTemplate(t, keyType, pkix.Name{CommonName: "leaf-cert"}, false)
	leaf, _ = SignX509(certTemplate, intermediate, intermediateKey)
	require.NoError(t, err)
	require.NotNil(t, leaf)

	return
}

func createCertTemplate(t *testing.T, keyType KeyType, name pkix.Name, isCa bool) (*x509.Certificate, crypto.PrivateKey) {
	key, err := GenerateSigner(keyType)
	require.NoError(t, err)

	var tmpl *x509.Certificate
	if isCa {
		tmpl, err = CreateCATemplate(clk, key.Public(), name, 1*time.Hour)
		require.NoError(t, err)
		require.NotNil(t, tmpl)
	} else {
		tmpl, err = CreateX509Template(clk, key.Public(), name, nil, nil, 1*time.Hour)
		require.NoError(t, err)
		require.NotNil(t, tmpl)
	}

	return tmpl, key
}

func readFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}
