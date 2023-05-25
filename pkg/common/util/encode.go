package util

import "encoding/base64"

// EncodeToString returns the base64 encoding of the bytes.
func EncodeToString(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeString returns the bytes represented by the base64 string.
func DecodeString(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
