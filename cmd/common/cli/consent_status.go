package cli

import (
	"fmt"
	"strings"
)

var ValidConsentStatusValues = []string{"approved", "denied", "pending"}

func ValidateConsentStatusValue(status string) error {
	for _, validValue := range ValidConsentStatusValues {
		if status == validValue {
			return nil
		}
	}
	return fmt.Errorf("invalid value for status. Valid values: %s", strings.Join(ValidConsentStatusValues, ", "))
}
