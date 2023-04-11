package endpoints

import "fmt"

type ErrWrongSPIFFEID struct {
	cause error
}

func (e ErrWrongSPIFFEID) Error() string {
	return fmt.Sprintf("malformed spiffe ID: %v", e.cause)
}

type ErrWrongTrustDomain struct {
	cause error
}

func (e ErrWrongTrustDomain) Error() string {
	return fmt.Sprintf("malformed trust domain name: %v", e.cause)
}
