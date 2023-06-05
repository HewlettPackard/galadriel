package harvester

import (
	"testing"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
	"github.com/HewlettPackard/galadriel/pkg/common/util/encoding"
	"github.com/stretchr/testify/assert"
)

func TestBundlePutToEntity(t *testing.T) {
	t.Run("Does not allow wrong trust domain names", func(t *testing.T) {
		bundlePut := PutBundleRequest{
			TrustDomain: "a wrong trust domain",
		}

		bundle, err := bundlePut.ToEntity()
		assert.ErrorContains(t, err, "malformed trust domain[a wrong trust domain]")
		assert.Nil(t, bundle)
	})

	t.Run("Full fill correctly the bundle entity model", func(t *testing.T) {
		bundleData := "test-bundle"
		digest := cryptoutil.CalculateDigest([]byte(bundleData))
		sig := encoding.EncodeToBase64([]byte("test-signature"))
		cert := encoding.EncodeToBase64([]byte("test-certificate"))
		bundlePut := PutBundleRequest{
			TrustBundle:        bundleData,
			Digest:             encoding.EncodeToBase64(digest),
			Signature:          &sig,
			TrustDomain:        "test.com",
			SigningCertificate: &cert,
		}

		bundle, err := bundlePut.ToEntity()
		assert.NoError(t, err)
		assert.NotNil(t, bundle)

		assert.Equal(t, bundlePut.TrustDomain, bundle.TrustDomainName.String())
		assert.Equal(t, []byte(bundlePut.TrustBundle), bundle.Data)
		assert.Equal(t, digest, bundle.Digest)
		assert.Equal(t, []byte("test-signature"), bundle.Signature)
		assert.Equal(t, []byte("test-certificate"), bundle.SigningCertificate)
	})
}
