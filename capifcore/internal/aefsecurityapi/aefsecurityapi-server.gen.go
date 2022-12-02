// Package aefsecurityapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
package aefsecurityapi

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	externalRef0 "oransc.org/nonrtric/capifcore/internal/common29571"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Check authentication.
	// (POST /check-authentication)
	PostCheckAuthentication(ctx echo.Context) error
	// Revoke authorization.
	// (POST /revoke-authorization)
	PostRevokeAuthorization(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// PostCheckAuthentication converts echo context to params.
func (w *ServerInterfaceWrapper) PostCheckAuthentication(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostCheckAuthentication(ctx)
	return err
}

// PostRevokeAuthorization converts echo context to params.
func (w *ServerInterfaceWrapper) PostRevokeAuthorization(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostRevokeAuthorization(ctx)
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

	router.POST(baseURL+"/check-authentication", wrapper.PostCheckAuthentication)
	router.POST(baseURL+"/revoke-authorization", wrapper.PostRevokeAuthorization)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xY3W7bNhR+lQNuFxugSJZSI43vVCcZdNMZsbGbtgho6chiI5EsSdn1Aj/QXmNPNpBU",
	"vNixgyBCAxTITSyI53w8Px8/HuWO5KKRgiM3mozuiM4rbKh7HFeY36atqZAbllPDBL/Gb3alQJ0rJu0b",
	"MiLXKBVqCwB0xxpyiwAKv7WoDRTU0JAERCohURmGbhcqWcaX4hZVVjzGTicZML8M2QVQrdmCYwHzNZgK",
	"YZxOsivIhUIoW567TY1wSw89VxWrEQQ/mQuqCsYX+xbhZ04CYtYSyYhooxhfkE1AdCulUAaLK6SmVT7g",
	"XxWWZERm0+R8eBbfjEXTCH5hc1vTpv4l+r+eUVfMaPoIZ7MJiK0LU1iQ0afdKhza+Ms2PDH/irmx4R1q",
	"kJYvaZCWgms80qHXq8Lz0r5GWyibt1Ds72fzcmsMCpciv398ipvK7ZTxUuylnCTJjePezRTzVjGzvkkn",
	"2ROpd1YfhWFlV/zXJdiDVIIedX4GvY7U+SeimPXB7wYVp/WFyPXjlE//mExgNoXkPEySBP6Kz8JhOAAf",
	"ohOWK0UbXAl1C6VQ4Ow/CmWquWh5YS00CUirajIilTFSj6JotVqFpwspQ6EWUWlkNJWY64iqvGJLjJLz",
	"G42KoY78rpHtEeu4+Vg17bbp5RXojnrQUE4X2CA3IcBn/u8/kAySJPCx/akWlHd9ozVMqDIclYbf0uvs",
	"QwDpLJsGMB5P0wAuZ9MsgNn0wv3MUvtn/LvDTOsaFFtURtuGo1pi0SkrM7WtcHq5e2JIQJaotI86DpNw",
	"YJMSEjmVzNY5HIQDyxZqKteGyCnWya6M2QUptLG/llXupb1NyERoc0AliecEavNBFGvrlgtukBt/Icm6",
	"s4u+ao/u+fWAlAfZd+TC3Oxy0KgW3Qt/IlxayWDwQ6PQ0kexS5NZhVsBXFENus1z1Lps69C24XRwtncM",
	"Yyt7TxzDbU6R9XUQ73tAvLcQ73xpXgRhfR1E3AMi9hCnPSBOPcS7HhDvHETcI5HYJxL3SCT2icTDHhBD",
	"B5GcvxwiObcQwx68GHpeDHs0deibWmBJ29q8GObef+OGgaahak1GfrDbm9VCZxP5e/xk56J9Wv8OXOM/",
	"SP+ODGavrH9HxpY3/XvTvzf9+zn0zx/h3a+J0J9gN1UqTUaf7rrx+Y5Kdi2E2UQUy5P7cTdaxna4pIrR",
	"eb39T4O18+NyF/Z2+MbvtJE1hrloyL5OdI5ANRRYMo4FMA55TVuNcBYOQZSw+0Xggv2y+S8AAP//WFej",
	"JF0RAAA=",
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

	for rawPath, rawFunc := range externalRef0.PathToRawSpec(path.Join(pathPrefix, "TS29571_CommonData.yaml")) {
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