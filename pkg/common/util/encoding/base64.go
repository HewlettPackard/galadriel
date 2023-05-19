package encoding

import "encoding/base64"

// EncodeToBase64 encodes the given data to a base64 string
func EncodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// DecodeFromBase64 decodes the given base64 string to a byte slice
func DecodeFromBase64(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
