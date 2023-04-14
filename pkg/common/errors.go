package common

import "fmt"

type ErrWrongSPIFFEID struct {
	Cause error
}

func (e ErrWrongSPIFFEID) Error() string {
	return fmt.Sprintf("malformed spiffe ID: %v", e.Cause)
}

type ErrWrongTrustDomain struct {
	Cause error
}

func (e ErrWrongTrustDomain) Error() string {
	return fmt.Sprintf("malformed trust domain name: %v", e.Cause)
}
