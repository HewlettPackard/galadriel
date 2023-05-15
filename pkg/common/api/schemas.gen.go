// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/getkin/kin-openapi/openapi3"
)

// Defines values for ConsentStatus.
const (
	Accepted ConsentStatus = "accepted"
	Denied   ConsentStatus = "denied"
	Pending  ConsentStatus = "pending"
)

// ApiError defines model for ApiError.
type ApiError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// BundleDigest base64 encoded SHA-256 digest of the bundle
type BundleDigest = string

// Certificate X.509 certificate in PEM format
type Certificate = string

// ConsentStatus defines model for ConsentStatus.
type ConsentStatus string

// JWT defines model for JWT.
type JWT = string

// JoinToken defines model for JoinToken.
type JoinToken = UUID

// Relationship defines model for Relationship.
type Relationship struct {
	CreatedAt           time.Time     `json:"created_at"`
	Id                  UUID          `json:"id"`
	TrustDomainAConsent ConsentStatus `json:"trust_domain_a_consent"`
	TrustDomainAId      UUID          `json:"trust_domain_a_id"`
	TrustDomainBConsent ConsentStatus `json:"trust_domain_b_consent"`
	TrustDomainBId      UUID          `json:"trust_domain_b_id"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// SPIFFEID defines model for SPIFFEID.
type SPIFFEID = string

// Signature defines model for Signature.
type Signature = string

// TrustBundle X.509 certificate in PEM format
type TrustBundle = Certificate

// TrustDomain defines model for TrustDomain.
type TrustDomain struct {
	CreatedAt         time.Time       `json:"created_at"`
	Description       *string         `json:"description,omitempty"`
	HarvesterSpiffeId *SPIFFEID       `json:"harvester_spiffe_id,omitempty"`
	Id                UUID            `json:"id"`
	Name              TrustDomainName `json:"name"`

	// OnboardingBundle X.509 certificate in PEM format
	OnboardingBundle *TrustBundle `json:"onboarding_bundle,omitempty"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// TrustDomainName defines model for TrustDomainName.
type TrustDomainName = string

// UUID defines model for UUID.
type UUID = openapi_types.UUID

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
