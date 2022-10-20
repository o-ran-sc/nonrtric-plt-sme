// Package invokermanagementapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
package invokermanagementapi

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	externalRef0 "oransc.org/nonrtric/sme/internal/common29122"
	externalRef1 "oransc.org/nonrtric/sme/internal/common29571"
	externalRef2 "oransc.org/nonrtric/sme/internal/publishserviceapi"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /onboardedInvokers)
	PostOnboardedInvokers(ctx echo.Context) error

	// (DELETE /onboardedInvokers/{onboardingId})
	DeleteOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error

	// (PATCH /onboardedInvokers/{onboardingId})
	ModifyIndApiInvokeEnrolment(ctx echo.Context, onboardingId string) error

	// (PUT /onboardedInvokers/{onboardingId})
	PutOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// PostOnboardedInvokers converts echo context to params.
func (w *ServerInterfaceWrapper) PostOnboardedInvokers(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostOnboardedInvokers(ctx)
	return err
}

// DeleteOnboardedInvokersOnboardingId converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteOnboardedInvokersOnboardingId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "onboardingId" -------------
	var onboardingId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "onboardingId", runtime.ParamLocationPath, ctx.Param("onboardingId"), &onboardingId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter onboardingId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteOnboardedInvokersOnboardingId(ctx, onboardingId)
	return err
}

// ModifyIndApiInvokeEnrolment converts echo context to params.
func (w *ServerInterfaceWrapper) ModifyIndApiInvokeEnrolment(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "onboardingId" -------------
	var onboardingId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "onboardingId", runtime.ParamLocationPath, ctx.Param("onboardingId"), &onboardingId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter onboardingId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ModifyIndApiInvokeEnrolment(ctx, onboardingId)
	return err
}

// PutOnboardedInvokersOnboardingId converts echo context to params.
func (w *ServerInterfaceWrapper) PutOnboardedInvokersOnboardingId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "onboardingId" -------------
	var onboardingId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "onboardingId", runtime.ParamLocationPath, ctx.Param("onboardingId"), &onboardingId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter onboardingId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PutOnboardedInvokersOnboardingId(ctx, onboardingId)
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

	router.POST(baseURL+"/onboardedInvokers", wrapper.PostOnboardedInvokers)
	router.DELETE(baseURL+"/onboardedInvokers/:onboardingId", wrapper.DeleteOnboardedInvokersOnboardingId)
	router.PATCH(baseURL+"/onboardedInvokers/:onboardingId", wrapper.ModifyIndApiInvokeEnrolment)
	router.PUT(baseURL+"/onboardedInvokers/:onboardingId", wrapper.PutOnboardedInvokersOnboardingId)

}

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

	pathPrefix := path.Dir(pathToFile)

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(pathPrefix, "TS29122_CommonData.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	for rawPath, rawFunc := range externalRef2.PathToRawSpec(path.Join(pathPrefix, "TS29222_CAPIF_Publish_Service_API.yaml")) {
		if _, ok := res[rawPath]; ok {
			// it is not possible to compare functions in golang, so always overwrite the old value
		}
		res[rawPath] = rawFunc
	}
	for rawPath, rawFunc := range externalRef1.PathToRawSpec(path.Join(pathPrefix, "TS29571_CommonData.yaml")) {
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
