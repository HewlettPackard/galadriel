// Package admin provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.12.4 DO NOT EDIT.
package admin

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	externalRef0 "github.com/HewlettPackard/galadriel/pkg/common/api"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

const (
	Harvester_authScopes = "harvester_auth.Scopes"
)

// PatchRelationship defines model for PatchRelationship.
type PatchRelationship struct {
	ConsentStatus externalRef0.ConsentStatus `json:"consent_status"`
}

// Default defines model for Default.
type Default = externalRef0.ApiError

// GetRelationshipsParams defines parameters for GetRelationships.
type GetRelationshipsParams struct {
	ConsentStatus *externalRef0.ConsentStatus `form:"consentStatus,omitempty" json:"consentStatus,omitempty"`
}

// PatchRelationshipJSONRequestBody defines body for PatchRelationship for application/json ContentType.
type PatchRelationshipJSONRequestBody = PatchRelationship

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetRelationships request
	GetRelationships(ctx context.Context, params *GetRelationshipsParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PatchRelationship request with any body
	PatchRelationshipWithBody(ctx context.Context, relationshipID externalRef0.UUID, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PatchRelationship(ctx context.Context, relationshipID externalRef0.UUID, body PatchRelationshipJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetRelationships(ctx context.Context, params *GetRelationshipsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetRelationshipsRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PatchRelationshipWithBody(ctx context.Context, relationshipID externalRef0.UUID, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPatchRelationshipRequestWithBody(c.Server, relationshipID, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PatchRelationship(ctx context.Context, relationshipID externalRef0.UUID, body PatchRelationshipJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPatchRelationshipRequest(c.Server, relationshipID, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetRelationshipsRequest generates requests for GetRelationships
func NewGetRelationshipsRequest(server string, params *GetRelationshipsParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/relationships")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.ConsentStatus != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "consentStatus", runtime.ParamLocationQuery, *params.ConsentStatus); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPatchRelationshipRequest calls the generic PatchRelationship builder with application/json body
func NewPatchRelationshipRequest(server string, relationshipID externalRef0.UUID, body PatchRelationshipJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPatchRelationshipRequestWithBody(server, relationshipID, "application/json", bodyReader)
}

// NewPatchRelationshipRequestWithBody generates requests for PatchRelationship with any type of body
func NewPatchRelationshipRequestWithBody(server string, relationshipID externalRef0.UUID, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "relationshipID", runtime.ParamLocationPath, relationshipID)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/relationships/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetRelationships request
	GetRelationshipsWithResponse(ctx context.Context, params *GetRelationshipsParams, reqEditors ...RequestEditorFn) (*GetRelationshipsResponse, error)

	// PatchRelationship request with any body
	PatchRelationshipWithBodyWithResponse(ctx context.Context, relationshipID externalRef0.UUID, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PatchRelationshipResponse, error)

	PatchRelationshipWithResponse(ctx context.Context, relationshipID externalRef0.UUID, body PatchRelationshipJSONRequestBody, reqEditors ...RequestEditorFn) (*PatchRelationshipResponse, error)
}

type GetRelationshipsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]externalRef0.Relationship
	JSONDefault  *externalRef0.ApiError
}

// Status returns HTTPResponse.Status
func (r GetRelationshipsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetRelationshipsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PatchRelationshipResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *externalRef0.Relationship
	JSONDefault  *externalRef0.ApiError
}

// Status returns HTTPResponse.Status
func (r PatchRelationshipResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PatchRelationshipResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetRelationshipsWithResponse request returning *GetRelationshipsResponse
func (c *ClientWithResponses) GetRelationshipsWithResponse(ctx context.Context, params *GetRelationshipsParams, reqEditors ...RequestEditorFn) (*GetRelationshipsResponse, error) {
	rsp, err := c.GetRelationships(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetRelationshipsResponse(rsp)
}

// PatchRelationshipWithBodyWithResponse request with arbitrary body returning *PatchRelationshipResponse
func (c *ClientWithResponses) PatchRelationshipWithBodyWithResponse(ctx context.Context, relationshipID externalRef0.UUID, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PatchRelationshipResponse, error) {
	rsp, err := c.PatchRelationshipWithBody(ctx, relationshipID, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePatchRelationshipResponse(rsp)
}

func (c *ClientWithResponses) PatchRelationshipWithResponse(ctx context.Context, relationshipID externalRef0.UUID, body PatchRelationshipJSONRequestBody, reqEditors ...RequestEditorFn) (*PatchRelationshipResponse, error) {
	rsp, err := c.PatchRelationship(ctx, relationshipID, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePatchRelationshipResponse(rsp)
}

// ParseGetRelationshipsResponse parses an HTTP response from a GetRelationshipsWithResponse call
func ParseGetRelationshipsResponse(rsp *http.Response) (*GetRelationshipsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetRelationshipsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []externalRef0.Relationship
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest externalRef0.ApiError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParsePatchRelationshipResponse parses an HTTP response from a PatchRelationshipWithResponse call
func ParsePatchRelationshipResponse(rsp *http.Response) (*PatchRelationshipResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PatchRelationshipResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest externalRef0.Relationship
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest externalRef0.ApiError
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// List the relationships.
	// (GET /relationships)
	GetRelationships(ctx echo.Context, params GetRelationshipsParams) error
	// Accept/Denies relationship requests
	// (PATCH /relationships/{relationshipID})
	PatchRelationship(ctx echo.Context, relationshipID externalRef0.UUID) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetRelationships converts echo context to params.
func (w *ServerInterfaceWrapper) GetRelationships(ctx echo.Context) error {
	var err error

	ctx.Set(Harvester_authScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params GetRelationshipsParams
	// ------------- Optional query parameter "consentStatus" -------------

	err = runtime.BindQueryParameter("form", true, false, "consentStatus", ctx.QueryParams(), &params.ConsentStatus)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter consentStatus: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetRelationships(ctx, params)
	return err
}

// PatchRelationship converts echo context to params.
func (w *ServerInterfaceWrapper) PatchRelationship(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "relationshipID" -------------
	var relationshipID externalRef0.UUID

	err = runtime.BindStyledParameterWithLocation("simple", false, "relationshipID", runtime.ParamLocationPath, ctx.Param("relationshipID"), &relationshipID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter relationshipID: %s", err))
	}

	ctx.Set(Harvester_authScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PatchRelationship(ctx, relationshipID)
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

	router.GET(baseURL+"/relationships", wrapper.GetRelationships)
	router.PATCH(baseURL+"/relationships/:relationshipID", wrapper.PatchRelationship)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/8xW227bOBD9FYHbR91sJ26iN2+dtAay3aBpXzbwGrQ0ktiVKIYcxTUC/ftiKN8Uuc2l",
	"XWCfTIvkzJlzDsl5YHFVqkqCRMOiB6bBqEoasH+mkPK6QBrGlUSQdsiVKkTMUVQy+GoqSd9MnEPJafRG",
	"Q8oi9luwjxu0syaYKHGhdaVZ0zQuS8DEWiiKwyJmJ5zJ9czZQ6BVm70UeredQCSJoJ28uNaVAo2CIKe8",
	"MOAydfCJoCdAv2mlS44sYkLi+IS5rOTfRFmXLDo9P3dZKWT7bxCGLsO1gnYpZKBZ47ISjOGZjQTfeKkK",
	"mp84S+A1irQuHLAVbJe5+3wGtZBZm/AKZIY5i4YHSTbzVK2Gu1poSFh02+Le553v1lfLrxAjYXpHPEm8",
	"QY61rRUkVXBLGunqHhJGNEthBwpkQnnmvcQuu+YY55+gsKqaXKgXc2yBLMwOyY980IXdr7sT61jdP4NU",
	"A0dIFhy7Sg7D4cALB94o/ByeRaMwCsO/DkVMOIKHooRHOg6OsCmSpxj48mU2pZWoa4OLpCq5kAu+2JT+",
	"Qv56YV6dX/ISntr6mbZM7Y6PtPxxlOWvqWL52iqWr62iVsl/7IxHRhd0Kg/82IFwTNRjFH3XQ9+V5diB",
	"skx2qh6l/Ow0HZ94p28Hb72T0/HQW47S2BvG5+NROh7zlI8PWahrC+aAgNHYZYojgqb7/e/b0DvnXjp/",
	"OGu83fjkGePBsHnD+lzSOZNptX2ceGxla7Vn7wXm9ZIY1QWLWI6oTBQEmf3sx1UZfIBVAYjXPP6H6yTI",
	"eMETLaBgvZfp/XbK+cD1PRgE7fzBJc+gBIn2yTIKYpFuHkWfuawQMUgDB4gmisc5OEM/7KCKgmC1Wvnc",
	"zvqVzoLNVhNczd5dfLy58IZ+6OdYWmQo0KpzDBMB8Zw/FUgajWyie9CmrWLgh/5gQDEqBZIrQRr7oT9i",
	"VqXcXo6BPrhY7ZcMLK10g9qJWULZAT91FlIIzUtA0IZFtw9MUMq7GvSauVsG4s55d5/ZMzx+K+Zut0cZ",
	"huGL+hOBUD75QHXel2bnPa41Xx9rXm7qOAZjqAvYMdUaaddAHUu3KyTYdlq244G41gLXlsh8K++C13So",
	"bufEgKnLkus1i9iVMOhgDk5HObIg8oy0YF2l5pShK3PwcPh3Nm0IrqKGoK98v0/oSd9lZjZ1qrQHkLmt",
	"Q8h4e4N0YbDDyxJ1Dc91TPsqtEa5q8Hg71Wy/mU9bJ+BI344nHfaPsbBylmCs7nhe8U1P2nr57v5/+Te",
	"SRyDwmBKLarpOMTZaGd+5GSbTt9vnde964sq5kVeGfTNimcZaF9UAVciuB8xQrGJ+tiwE+fmenZ5eeHY",
	"DsFpW4S9RztfG7e/u1OEMBvrKw10i9kZOhB8m+USkg3jTsc0S8AVgHRwVXWQmD2ULh3NvPk3AAD//w4x",
	"IU3MDQAA",
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

	pathPrefix := path.Dir(pathToFile)

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(pathPrefix, "../../../common/api/schemas.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
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