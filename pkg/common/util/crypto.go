package util

import "golang.org/x/crypto/sha3"

func GetDigest(input []byte) []byte {
	d := sha3.Sum256(input)
	return d[:]
}
