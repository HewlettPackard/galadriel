package certtest

import (
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"time"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/jmhodges/clock"
)

func CreateTestCACertificate(clk clock.Clock) (*x509.Certificate, crypto.PrivateKey, error) {
	name := pkix.Name{Organization: []string{"Galadriel"}}
	template, _ := cryptoutil.CreateCATemplate(clk, name, name, time.Hour)
	caCert, caKey, err := cryptoutil.SelfSign(template)
	if err != nil {
		return nil, nil, err
	}

	return caCert, caKey, nil
}
