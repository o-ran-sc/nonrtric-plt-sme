// Package providermanagementapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
package providermanagementapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
)

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
	// PostRegistrations request with any body
	PostRegistrationsWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostRegistrations(ctx context.Context, body PostRegistrationsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteRegistrationsRegistrationId request
	DeleteRegistrationsRegistrationId(ctx context.Context, registrationId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ModifyIndApiProviderEnrolment request with any body
	ModifyIndApiProviderEnrolmentWithBody(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PutRegistrationsRegistrationId request with any body
	PutRegistrationsRegistrationIdWithBody(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PutRegistrationsRegistrationId(ctx context.Context, registrationId string, body PutRegistrationsRegistrationIdJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) PostRegistrationsWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostRegistrationsRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostRegistrations(ctx context.Context, body PostRegistrationsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostRegistrationsRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteRegistrationsRegistrationId(ctx context.Context, registrationId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteRegistrationsRegistrationIdRequest(c.Server, registrationId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ModifyIndApiProviderEnrolmentWithBody(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewModifyIndApiProviderEnrolmentRequestWithBody(c.Server, registrationId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PutRegistrationsRegistrationIdWithBody(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPutRegistrationsRegistrationIdRequestWithBody(c.Server, registrationId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PutRegistrationsRegistrationId(ctx context.Context, registrationId string, body PutRegistrationsRegistrationIdJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPutRegistrationsRegistrationIdRequest(c.Server, registrationId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewPostRegistrationsRequest calls the generic PostRegistrations builder with application/json body
func NewPostRegistrationsRequest(server string, body PostRegistrationsJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostRegistrationsRequestWithBody(server, "application/json", bodyReader)
}

// NewPostRegistrationsRequestWithBody generates requests for PostRegistrations with any type of body
func NewPostRegistrationsRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/registrations")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteRegistrationsRegistrationIdRequest generates requests for DeleteRegistrationsRegistrationId
func NewDeleteRegistrationsRegistrationIdRequest(server string, registrationId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "registrationId", runtime.ParamLocationPath, registrationId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/registrations/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewModifyIndApiProviderEnrolmentRequestWithBody generates requests for ModifyIndApiProviderEnrolment with any type of body
func NewModifyIndApiProviderEnrolmentRequestWithBody(server string, registrationId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "registrationId", runtime.ParamLocationPath, registrationId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/registrations/%s", pathParam0)
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

// NewPutRegistrationsRegistrationIdRequest calls the generic PutRegistrationsRegistrationId builder with application/json body
func NewPutRegistrationsRegistrationIdRequest(server string, registrationId string, body PutRegistrationsRegistrationIdJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPutRegistrationsRegistrationIdRequestWithBody(server, registrationId, "application/json", bodyReader)
}

// NewPutRegistrationsRegistrationIdRequestWithBody generates requests for PutRegistrationsRegistrationId with any type of body
func NewPutRegistrationsRegistrationIdRequestWithBody(server string, registrationId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "registrationId", runtime.ParamLocationPath, registrationId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/registrations/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
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
	// PostRegistrations request with any body
	PostRegistrationsWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostRegistrationsResponse, error)

	PostRegistrationsWithResponse(ctx context.Context, body PostRegistrationsJSONRequestBody, reqEditors ...RequestEditorFn) (*PostRegistrationsResponse, error)

	// DeleteRegistrationsRegistrationId request
	DeleteRegistrationsRegistrationIdWithResponse(ctx context.Context, registrationId string, reqEditors ...RequestEditorFn) (*DeleteRegistrationsRegistrationIdResponse, error)

	// ModifyIndApiProviderEnrolment request with any body
	ModifyIndApiProviderEnrolmentWithBodyWithResponse(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModifyIndApiProviderEnrolmentResponse, error)

	// PutRegistrationsRegistrationId request with any body
	PutRegistrationsRegistrationIdWithBodyWithResponse(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PutRegistrationsRegistrationIdResponse, error)

	PutRegistrationsRegistrationIdWithResponse(ctx context.Context, registrationId string, body PutRegistrationsRegistrationIdJSONRequestBody, reqEditors ...RequestEditorFn) (*PutRegistrationsRegistrationIdResponse, error)
}

type PostRegistrationsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *APIProviderEnrolmentDetails
}

// Status returns HTTPResponse.Status
func (r PostRegistrationsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostRegistrationsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteRegistrationsRegistrationIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r DeleteRegistrationsRegistrationIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteRegistrationsRegistrationIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ModifyIndApiProviderEnrolmentResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *APIProviderEnrolmentDetails
}

// Status returns HTTPResponse.Status
func (r ModifyIndApiProviderEnrolmentResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ModifyIndApiProviderEnrolmentResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PutRegistrationsRegistrationIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *APIProviderEnrolmentDetails
}

// Status returns HTTPResponse.Status
func (r PutRegistrationsRegistrationIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PutRegistrationsRegistrationIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// PostRegistrationsWithBodyWithResponse request with arbitrary body returning *PostRegistrationsResponse
func (c *ClientWithResponses) PostRegistrationsWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostRegistrationsResponse, error) {
	rsp, err := c.PostRegistrationsWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostRegistrationsResponse(rsp)
}

func (c *ClientWithResponses) PostRegistrationsWithResponse(ctx context.Context, body PostRegistrationsJSONRequestBody, reqEditors ...RequestEditorFn) (*PostRegistrationsResponse, error) {
	rsp, err := c.PostRegistrations(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostRegistrationsResponse(rsp)
}

// DeleteRegistrationsRegistrationIdWithResponse request returning *DeleteRegistrationsRegistrationIdResponse
func (c *ClientWithResponses) DeleteRegistrationsRegistrationIdWithResponse(ctx context.Context, registrationId string, reqEditors ...RequestEditorFn) (*DeleteRegistrationsRegistrationIdResponse, error) {
	rsp, err := c.DeleteRegistrationsRegistrationId(ctx, registrationId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteRegistrationsRegistrationIdResponse(rsp)
}

// ModifyIndApiProviderEnrolmentWithBodyWithResponse request with arbitrary body returning *ModifyIndApiProviderEnrolmentResponse
func (c *ClientWithResponses) ModifyIndApiProviderEnrolmentWithBodyWithResponse(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ModifyIndApiProviderEnrolmentResponse, error) {
	rsp, err := c.ModifyIndApiProviderEnrolmentWithBody(ctx, registrationId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseModifyIndApiProviderEnrolmentResponse(rsp)
}

// PutRegistrationsRegistrationIdWithBodyWithResponse request with arbitrary body returning *PutRegistrationsRegistrationIdResponse
func (c *ClientWithResponses) PutRegistrationsRegistrationIdWithBodyWithResponse(ctx context.Context, registrationId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PutRegistrationsRegistrationIdResponse, error) {
	rsp, err := c.PutRegistrationsRegistrationIdWithBody(ctx, registrationId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePutRegistrationsRegistrationIdResponse(rsp)
}

func (c *ClientWithResponses) PutRegistrationsRegistrationIdWithResponse(ctx context.Context, registrationId string, body PutRegistrationsRegistrationIdJSONRequestBody, reqEditors ...RequestEditorFn) (*PutRegistrationsRegistrationIdResponse, error) {
	rsp, err := c.PutRegistrationsRegistrationId(ctx, registrationId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePutRegistrationsRegistrationIdResponse(rsp)
}

// ParsePostRegistrationsResponse parses an HTTP response from a PostRegistrationsWithResponse call
func ParsePostRegistrationsResponse(rsp *http.Response) (*PostRegistrationsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostRegistrationsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest APIProviderEnrolmentDetails
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	}

	return response, nil
}

// ParseDeleteRegistrationsRegistrationIdResponse parses an HTTP response from a DeleteRegistrationsRegistrationIdWithResponse call
func ParseDeleteRegistrationsRegistrationIdResponse(rsp *http.Response) (*DeleteRegistrationsRegistrationIdResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteRegistrationsRegistrationIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}

// ParseModifyIndApiProviderEnrolmentResponse parses an HTTP response from a ModifyIndApiProviderEnrolmentWithResponse call
func ParseModifyIndApiProviderEnrolmentResponse(rsp *http.Response) (*ModifyIndApiProviderEnrolmentResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ModifyIndApiProviderEnrolmentResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest APIProviderEnrolmentDetails
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePutRegistrationsRegistrationIdResponse parses an HTTP response from a PutRegistrationsRegistrationIdWithResponse call
func ParsePutRegistrationsRegistrationIdResponse(rsp *http.Response) (*PutRegistrationsRegistrationIdResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PutRegistrationsRegistrationIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest APIProviderEnrolmentDetails
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}