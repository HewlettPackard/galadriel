package spire

import (
	"github.com/spiffe/go-spiffe/v2/bundle/spiffebundle"
	"google.golang.org/grpc/codes"
)

type Status struct {
	// TODO: see if we need to map this to our (probably simplified) own set of codes
	Code    codes.Code
	Message string
}

type BatchSetFederatedBundleStatus struct {
	Bundle *spiffebundle.Bundle
	Status *Status
}
