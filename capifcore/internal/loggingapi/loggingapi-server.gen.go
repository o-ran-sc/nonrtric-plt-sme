// Package loggingapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
package loggingapi

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
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /{aefId}/logs)
	PostAefIdLogs(ctx echo.Context, aefId string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// PostAefIdLogs converts echo context to params.
func (w *ServerInterfaceWrapper) PostAefIdLogs(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "aefId" -------------
	var aefId string

	err = runtime.BindStyledParameterWithLocation("simple", false, "aefId", runtime.ParamLocationPath, ctx.Param("aefId"), &aefId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter aefId: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostAefIdLogs(ctx, aefId)
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

	router.POST(baseURL+"/:aefId/logs", wrapper.PostAefIdLogs)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"bwAAAP//cfEXzl0UAAA=",
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
