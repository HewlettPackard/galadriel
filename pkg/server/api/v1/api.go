package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// Defines values for SpireServerStatus.
const (
	SpireServerStatusActive   SpireServerStatus = "active"
	SpireServerStatusDisabled SpireServerStatus = "disabled"
	SpireServerStatusInactive SpireServerStatus = "inactive"
	SpireServerStatusInvited  SpireServerStatus = "invited"
	SpireServerStatusToDelete SpireServerStatus = "to_delete"
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

// SpireServer defines model for SpireServer.
type SpireServer struct {
	CreatedAt   *time.Time        `json:"createdAt,omitempty"`
	Id          int64             `json:"id"`
	Orgid       string            `json:"orgid"`
	Status      SpireServerStatus `json:"status"`
	TrustDomain *string           `json:"trustDomain,omitempty"`
}

// SpireServerStatus defines model for SpireServer.Status.
type SpireServerStatus string

// TrustBundle defines model for TrustBundle.
type TrustBundle struct {
	Bundle      *string            `json:"bundle,omitempty"`
	CreatedAt   *time.Time         `json:"createdAt,omitempty"`
	Id          *int64             `json:"id,omitempty"`
	SourceUser  *string            `json:"sourceUser,omitempty"`
	Status      *TrustBundleStatus `json:"status,omitempty"`
	TrustDomain *string            `json:"trustDomain,omitempty"`
}

// TrustBundleStatus defines model for TrustBundle.Status.
type TrustBundleStatus string

// GetSpireServersParams defines parameters for GetSpireServers.
type GetSpireServersParams struct {
	// filter SpireServers by org
	Org *string `form:"org,omitempty" json:"org,omitempty"`

	// filter SpireServers by status
	Status *GetSpireServersParamsStatus `form:"status,omitempty" json:"status,omitempty"`
}

// GetSpireServersParamsStatus defines parameters for GetSpireServers.
type GetSpireServersParamsStatus string

// CreateSpireServerJSONBody defines parameters for CreateSpireServer.
type CreateSpireServerJSONBody = SpireServer

// CreateSpireServerJSONRequestBody defines body for CreateSpireServer for application/json ContentType.
type CreateSpireServerJSONRequestBody = CreateSpireServerJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /spireServers)
	GetSpireServers(ctx echo.Context, params GetSpireServersParams) error

	// (POST /spireServers)
	CreateSpireServer(ctx echo.Context) error

	// (DELETE /spireServers/{spireServerId})
	DeleteSpireServer(ctx echo.Context, spireServerId int64) error

	// (PUT /spireServers/{spireServerId})
	UpdateSpireServer(ctx echo.Context, spireServerId int64) error

	// (PUT /trustBundles/{trustBundleId})
	UpdateTrustBundle(ctx echo.Context, trustBundleId int64) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetSpireServers converts echo context to params.
func (w *ServerInterfaceWrapper) GetSpireServers(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetSpireServersParams
	// ------------- Optional query parameter "org" -------------

	err = runtime.BindQueryParameter("form", true, false, "org", ctx.QueryParams(), &params.Org)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter org: %s", err))
	}

	// ------------- Optional query parameter "status" -------------

	err = runtime.BindQueryParameter("form", true, false, "status", ctx.QueryParams(), &params.Status)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter status: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetSpireServers(ctx, params)
	return err
}

// CreateSpireServer converts echo context to params.
func (w *ServerInterfaceWrapper) CreateSpireServer(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateSpireServer(ctx)
	return err
}

// DeleteSpireServer converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteSpireServer(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "spireServerId" -------------
	var spireServerId int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "spireServerId", runtime.ParamLocationPath, ctx.Param("spireServerId"), &spireServerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter spireServerId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteSpireServer(ctx, spireServerId)
	return err
}

// UpdateSpireServer converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateSpireServer(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "spireServerId" -------------
	var spireServerId int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "spireServerId", runtime.ParamLocationPath, ctx.Param("spireServerId"), &spireServerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter spireServerId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateSpireServer(ctx, spireServerId)
	return err
}

// UpdateTrustBundle converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateTrustBundle(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "trustBundleId" -------------
	var trustBundleId int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "trustBundleId", runtime.ParamLocationPath, ctx.Param("trustBundleId"), &trustBundleId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter trustBundleId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateTrustBundle(ctx, trustBundleId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/spireServers", wrapper.GetSpireServers)
	router.POST(baseURL+"/spireServers", wrapper.CreateSpireServer)
	router.DELETE(baseURL+"/spireServers/:spireServerId", wrapper.DeleteSpireServer)
	router.PUT(baseURL+"/spireServers/:spireServerId", wrapper.UpdateSpireServer)
	router.PUT(baseURL+"/trustBundles/:trustBundleId", wrapper.UpdateTrustBundle)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9RXUW/bNhD+K8RtwF6Uyk2KYdBbswaD34pmfQqCgRbPNguKZMljUsPQfx9I2pZkqY6L",
	"pc36ZIk8Hr/77uMnegu1aazRqMlDtQVfr7Hh6fHGOePig3XGoiOJabg2AuPv0riGE1QgNV1dQgG0sZhf",
	"cYUO2gIa9J6vUvRu0pOTegVtW4DDz0E6FFDd5Zxd/P0hmVl8wppirlsrHd6ie8ApTA45oXhLA2CCE16Q",
	"bLADt9+/ACmOi/j9zWQRxq1Ox3Y5PXEKCRDq0MTCpH6QhAIK4DXJh4hE6sOjkJ4vVJom849AhdQvvktM",
	"Lnh6Zxou9QDJLqJ4gl6Zdog5RM5xgLovb4rxv+OC66CFwjHji8P4AcxiQ+inaPme3fEmuBo/+qyKJ4mZ",
	"6tFUZ569HUfsxiGplyYuF+hrJy1Jo6GC2/fzDzfs2kmxQvb2/TxmkxSpHkxBAQ/ofF7z+tXs1Sxp1aLm",
	"VkIFV2moAMtpnaotfXeA0sAKabz7B6TgtGdcKUZrZLf9RSm/4zF0LqCCv5CO5i13vEFKO9wd515KRegG",
	"Kdliw4xbJe6hgs8B3QYK0LxJZKWZ7EjnHMC2OHPLg/indj1MdhuPT3NPKuPTPJZMex8PpLdG+3x8Lmez",
	"7KWaUKc+cGuVrBO55SdvdGfG8UkSNmnhrw6XUMEvZWfb5c6zy75FdpLjzvFNVtyQnBXSgJnfPNtjhBS8",
	"5EHRN8E8hS5/TiZwBI1fLNaEgvVirPET+vwzWQnjTONjH/xImzlwGBEtET1dG7F5tqoGnI9ru1HYxHhm",
	"liwfX59iGRnGhYC+T5ML2I508vpHQe1Ns+TY0ugXFwTuY9pi6GDltvc2F22WSnLskWjyOOMnBfMuBQ0j",
	"TtrZXMSmJpM8auwOx85eogP33KWPetT+c9zu8PGbMJU34+K1YftuvXQPC7Bh4kx/tPEq4BOV2Xwjsae7",
	"ldc8S7dCSvWDunWOAX25eHx8vIgJL4JTqOPlWDzbMd9t6p+k4z99r74B4MurslepDkolr6Hu8uvLbe9t",
	"5zVfUbIyXDDO+lfnaekOI86UboLB8tX7KekOIP8k0u2TcoZ0v0rHd5PuEcD/oXTjP5z9Hf9uC8EpqGBN",
	"ZKuyVKbmam08VVeXl7M/Smjv238DAAD//3/547kAEAAA",
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
