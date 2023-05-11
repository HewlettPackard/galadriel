package keytest

import (
	"crypto"

	"github.com/HewlettPackard/galadriel/pkg/common/cryptoutil"
)

func MustSignerRSA2048() crypto.Signer {
	signer, err := cryptoutil.GenerateSigner(cryptoutil.RSA2048)
	if err != nil {
		panic(err)
	}
	return signer
}
