// Package common provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package common

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Defines values for FederationRelationshipStatus.
const (
	FederationRelationshipStatusActive   FederationRelationshipStatus = "active"
	FederationRelationshipStatusInactive FederationRelationshipStatus = "inactive"
	FederationRelationshipStatusInvited  FederationRelationshipStatus = "invited"
)

// Defines values for TrustBundleStatus.
const (
	TrustBundleStatusActive   TrustBundleStatus = "active"
	TrustBundleStatusInactive TrustBundleStatus = "inactive"
	TrustBundleStatusToDelete TrustBundleStatus = "to_delete"
)

// Error defines model for Error.
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// FederationRelationship defines model for FederationRelationship.
type FederationRelationship struct {
	FederationGroupId               int64                         `json:"federationGroupId"`
	Id                              int64                         `json:"id"`
	SpireServer                     string                        `json:"spireServer"`
	SpireServerConsent              *string                       `json:"spireServerConsent,omitempty"`
	SpireServerFederatedWith        string                        `json:"spireServerFederatedWith"`
	SpireServerFederatedWithConsent *string                       `json:"spireServerFederatedWithConsent,omitempty"`
	Status                          *FederationRelationshipStatus `json:"status,omitempty"`
}

// FederationRelationshipStatus defines model for FederationRelationship.Status.
type FederationRelationshipStatus string

// TrustBundle defines model for TrustBundle.
type TrustBundle struct {
	Bundle      string             `json:"bundle"`
	Id          int64              `json:"id"`
	Status      *TrustBundleStatus `json:"status,omitempty"`
	TrustDomain *string            `json:"trustDomain,omitempty"`
}

// TrustBundleStatus defines model for TrustBundle.Status.
type TrustBundleStatus string

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/6RSy07rMBD9l1l7de/VXWTJU2wBiQVCyI1Pm0GJbcaTSlWVf0dOSugjoFSsPLZnjs5j",
	"tlSGJgYPr4mKLaWyQmP78lokSC6ihAhRRv9cBod8LoM0Vqkg9vr3DxnSTcRwxQpCnaEGKdlV3737TCrs",
	"V9R1hgTvLQscFc8D5lf/ywgWFm8oNWPdwEGscvD3qPszVRxP6S3HvlsJbbxzx1z//5vkynMbU2TBA2QN",
	"OZjYSTPHUg8mLoNP8Hr+4E4+3BNr9cvxM0mo1ba3Fr5tclq2VF7nvNjvlWtWuL3ovomaHZmJkA59/UH6",
	"1G48Spv0ovWuxulCLMb3Uexio0hTWucvwSxTNLw61FBM2GJIM+ur0Fj2s6KYMrLHcAOG+ZR6alGeZb8M",
	"VPi2rg2FCG8jU0FkKFqt0vDTfQQAAP//WwXWvg8EAAA=",
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
