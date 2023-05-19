package harvester

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundlePutToEntity(t *testing.T) {
	t.Run("Does not allow wrong trust domain names", func(t *testing.T) {
		bundlePut := BundlePut{
			TrustDomain: "a wrong trust domain",
		}

		bundle, err := bundlePut.ToEntity()
		assert.ErrorContains(t, err, "malformed trust domain[a wrong trust domain]")
		assert.Nil(t, bundle)
	})

	t.Run("Full fill correctly the bundle entity model", func(t *testing.T) {
		sig := "test-signature"
		cert := "test-certificate"
		bundlePut := BundlePut{
			Signature:          &sig,
			TrustBundle:        "a really big bundle",
			TrustDomain:        "test.com",
			SigningCertificate: &cert,
		}

		bundle, err := bundlePut.ToEntity()
		assert.NoError(t, err)
		assert.NotNil(t, bundle)

		assert.Equal(t, bundlePut.TrustDomain, bundle.TrustDomainName.String())
		assert.Equal(t, []byte(bundlePut.TrustBundle), bundle.Data)
		assert.Equal(t, []byte(*bundlePut.Signature), bundle.Signature)
		assert.Equal(t, []byte(*bundlePut.SigningCertificate), bundle.SigningCertificate)
	})
}
