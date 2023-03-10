package cryptoutil

import (
	"crypto/rsa"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func Test(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
}

func (s *Suite) TestParsePrivateKeyPEM() {
	// not a private key
	_, err := ParseRSAPrivateKeyPEM(s.readFile("testdata/cert.pem"))
	require.Error(s.T(), err)

	// success with RSA
	key, err := ParseRSAPrivateKeyPEM(s.readFile("testdata/rsa-key.pem"))
	s.Require().NoError(err)
	s.Require().NotNil(key)
	_, ok := key.(*rsa.PrivateKey)
	s.Require().True(ok)
}

func (s *Suite) TestLoadRSAPrivateKey() {
	// not a private key
	_, err := LoadRSAPrivateKey("testdata/cert.pem")
	require.Error(s.T(), err)

	// success with RSA
	key, err := LoadRSAPrivateKey("testdata/rsa-key.pem")
	s.Require().NoError(err)
	s.Require().NotNil(key)
	_, ok := key.(*rsa.PrivateKey)
	s.Require().True(ok)
}

func (s *Suite) readFile(path string) []byte {
	data, err := os.ReadFile(path)
	s.Require().NoError(err)
	return data
}

func (s *Suite) TestLoadCertificate() {
	// not a certificate
	_, err := LoadCertificate("testdata/rsa-key.pem")
	require.Error(s.T(), err)

	// success
	cert, err := LoadCertificate("testdata/cert.pem")
	s.Require().NoError(err)
	s.Require().NotNil(cert)
}

func (s *Suite) TestParseCertificate() {
	// not a certificate
	_, err := ParseCertificate(s.readFile("testdata/rsa-key.pem"))
	require.Error(s.T(), err)

	// success with one certificate
	cert, err := ParseCertificate(s.readFile("testdata/cert.pem"))
	s.Require().NoError(err)
	s.Require().NotNil(cert)
}

func (s *Suite) TestEncodeCertificates() {
	cert, err := LoadCertificate("testdata/cert.pem")
	s.Require().NoError(err)
	expCertPem, err := os.ReadFile("testdata/cert.pem")
	s.Require().NoError(err)
	s.Require().Equal(expCertPem, EncodeCertificate(cert))

}
