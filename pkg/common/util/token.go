package util

import (
	"fmt"
	"github.com/google/uuid"
)

func GenerateToken() (string, error) {
	token, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID token: %v", err)
	}
	return token.String(), nil
}
