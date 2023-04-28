package common

import "fmt"

type ErrWrongSPIFFEID struct {
	Cause    error
	SPIFFEID string
}

func (e ErrWrongSPIFFEID) Error() string {
	return fmt.Errorf("malformed SPIFFE ID[%v]: %w", e.SPIFFEID, e.Cause).Error()
}

type ErrWrongTrustDomain struct {
	Cause       error
	TrustDomain string
}

func (e ErrWrongTrustDomain) Error() string {
	return fmt.Errorf("malformed trust domain[%v]: %w", e.TrustDomain, e.Cause).Error()
}
