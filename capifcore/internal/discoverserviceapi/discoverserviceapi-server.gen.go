// Package discoverserviceapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
package discoverserviceapi

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	externalRef0 "oransc.org/nonrtric/capifcore/internal/common29122"
	externalRef1 "oransc.org/nonrtric/capifcore/internal/common29571"
	externalRef2 "oransc.org/nonrtric/capifcore/internal/publishserviceapi"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /allServiceAPIs)
	GetAllServiceAPIs(ctx echo.Context, params GetAllServiceAPIsParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetAllServiceAPIs converts echo context to params.
func (w *ServerInterfaceWrapper) GetAllServiceAPIs(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetAllServiceAPIsParams
	// ------------- Required query parameter "api-invoker-id" -------------

	err = runtime.BindQueryParameter("form", true, true, "api-invoker-id", ctx.QueryParams(), &params.ApiInvokerId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter api-invoker-id: %s", err))
	}

	// ------------- Optional query parameter "api-name" -------------

	err = runtime.BindQueryParameter("form", true, false, "api-name", ctx.QueryParams(), &params.ApiName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter api-name: %s", err))
	}

	// ------------- Optional query parameter "api-version" -------------

	err = runtime.BindQueryParameter("form", true, false, "api-version", ctx.QueryParams(), &params.ApiVersion)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter api-version: %s", err))
	}

	// ------------- Optional query parameter "comm-type" -------------

	err = runtime.BindQueryParameter("form", true, false, "comm-type", ctx.QueryParams(), &params.CommType)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter comm-type: %s", err))
	}

	// ------------- Optional query parameter "protocol" -------------

	err = runtime.BindQueryParameter("form", true, false, "protocol", ctx.QueryParams(), &params.Protocol)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter protocol: %s", err))
	}

	// ------------- Optional query parameter "aef-id" -------------

	err = runtime.BindQueryParameter("form", true, false, "aef-id", ctx.QueryParams(), &params.AefId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter aef-id: %s", err))
	}

	// ------------- Optional query parameter "data-format" -------------

	err = runtime.BindQueryParameter("form", true, false, "data-format", ctx.QueryParams(), &params.DataFormat)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter data-format: %s", err))
	}

	// ------------- Optional query parameter "api-cat" -------------

	err = runtime.BindQueryParameter("form", true, false, "api-cat", ctx.QueryParams(), &params.ApiCat)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter api-cat: %s", err))
	}

	// ------------- Optional query parameter "preferred-aef-loc" -------------

	if paramValue := ctx.QueryParam("preferred-aef-loc"); paramValue != "" {

		var value externalRef2.AefLocation
		err = json.Unmarshal([]byte(paramValue), &value)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Error unmarshaling parameter 'preferred-aef-loc' as JSON")
		}
		params.PreferredAefLoc = &value

	}

	// ------------- Optional query parameter "supported-features" -------------

	err = runtime.BindQueryParameter("form", true, false, "supported-features", ctx.QueryParams(), &params.SupportedFeatures)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter supported-features: %s", err))
	}

	// ------------- Optional query parameter "api-supported-features" -------------

	err = runtime.BindQueryParameter("form", true, false, "api-supported-features", ctx.QueryParams(), &params.ApiSupportedFeatures)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter api-supported-features: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetAllServiceAPIs(ctx, params)
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

	router.GET(baseURL+"/allServiceAPIs", wrapper.GetAllServiceAPIs)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/7xX3W4bOw5+FUK7Fy1gz8STv8Z33sQpvFik3tjdm20RyBqOrVYjTSWNU2/gB9rXOE92",
	"QGn8E8cOkLo4N006oT5+JD9S1BMTpqyMRu0d6z4xJ2ZY8vDrjXTCzNFi3hsOwpccnbCy8tJo1mX3WFl0",
	"dBA4KOk8mALIFERtLWqvFmBxKp0nDJAa/Azhujcc3IIwFqGotSAs4DoHx710xULqKXDQdTlBS3iFVB4t",
	"CCs9WsmhsmYuc8xhsghwveEAhNGuLtEmXzRrscqaCq2XGONBO5cCe8PBzYb8nmC2/kpuCbo5GlxwB1U9",
	"UdLNNq6bvyfQ52L2zHoLGtyMKwVSC1XnCL3+LYVQSIUOSu7FjAImtJ1AYyzSYxnI/t1iwbpsPMqusix7",
	"CEl8GEZGD6Po+qE3HCQLXqq/pZuapk1B09G+RLBli5VSD6KbTov5RYWsy7i1fMGWy/UHM/mGwrMlfcKf",
	"Hq3m6saIPYk8/TgcwngE2VWSZRn8p3OZXCQncG3K0uiQnVvLS3w09jsUxkKwvzPWzyam1nkQEGux2irW",
	"ZTPvK9dN08fHx+R0WlWJsdO08FU6qlC4lFsxk3NMs6sHh1aiS6PXlOKSujAv6REBcps34qb0b5XOJQBf",
	"9B//h+wky1qR3Cc75Vr+jxMCVzDk1mu0Dt717gf/aEFvPBi14Pp61GtBfzwatGA8ugk/xj365/p9wOwp",
	"BVZOZ94BdY2dYx5r7KVXlOJY01XTbReVtdgcrYsBdJIs6VB8pkLNK0kpT06SE1I+97NQkZQrtal3+DRF",
	"v0f0jbMtbW/nIvSlRW8lzhE4CKMUilWLRAshjM2Dhg0ItJ5LvV/K1JQhhYOcddlH9L3nHIk+CcOjdaz7",
	"312qIx9KJXPUXsYxsWp/qefmO1rgzsmp3vTnnkmTwMADV86A3QyvYHt9u8KWaJ/NqvZFeoEgtUdbcIFN",
	"YxKnHzXaBWsxzcvQNJVsN1zaMmctZvFHLS3mrOttja1mtlIZmq5yISjqs30yJdgWSA/SgUNPM+iJV/KO",
	"l7iEilu/GlSf7wfgvK2Fry2SWY6F1HHkCsVrh3CeZMkZ2W+1ZyfLXg8m/Ppm2iX/Ziw0gl3ze4fJNIF5",
	"533yisOVyt/kkyZLraUI4gIyh9o9vyCi8/v+vz/3R+OH+/5o+Olu1D9IRZiybAe/20SOH8HPmI4Jf084",
	"Q2u8EUbtBnGIbNXY/2auKxp7q9xfNwvag/XEInbBG0p5wz2n6Vxy7w4UkeY8V800hlXw8M/Rp7tw4mBR",
	"c+55O2L/5lQR69sIvCem8c4mIbjHqbELGpiPMylmL3aNCSqjpw68ea1XxE4ce5IrjPaow9znVaUa4aXf",
	"HBF7+q056GHxLxPx45LwMgmVxQItbYKkH9VYH5Z1Y90mISkj2MvU3iKniefA1VVlrN/o5e52vRSG235z",
	"HayvvNEm44c4rGHbRePpkHLOLzsPccMhMbyyhK0QV9T3KeZwWPl6I38mGKlzqm20Ww1uWF+nCYxnkrbN",
	"BRitFjAJtaC7D2SxsQ/Bb07RrdOYvX5L/DVZ+krXqauMdnGvz05O6MdbJb6/gZ8/dA7od+UeJiZfkLxo",
	"14nbg0VXK795N9BeCkFl9P/V02jrLRQ2zS9h/T49udxJU4da8JU0rfOQ0tkA8eEIiA8EcRbT+UsQdDZA",
	"dI6A6ESI0yMgTiPE2REQZxHi4giIiwDROYJFJ7LIrn4dIrsiiPMjinoei3p+REXOY0VyLHit/C/DrM4v",
	"l6Exw9OpeSLERyItxffG+GXaDMU2r6RL5x16N3Er+UTFmdHYxXdQQ2r9xMSfvKwUJsKUbLf9m4P7V+vL",
	"5Hxnsc6yLCG6X5d/BgAA//+yTjhGXhEAAA==",
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
