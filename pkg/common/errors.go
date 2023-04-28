package common

import "fmt"

type ErrWrongSPIFFEID struct {
	Cause error
}

func (e ErrWrongSPIFFEID) Error() string {
	return fmt.Errorf("malformed trust domain name: %w", e.Cause).Error()
}

type ErrWrongTrustDomain struct {
	Cause error
}

func (e ErrWrongTrustDomain) Error() string {
	return fmt.Errorf("malformed trust domain name: %w", e.Cause).Error()
}
