// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022: Nordix Foundation
//   %%
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//   ========================LICENSE_END===================================
//

package security

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"oransc.org/nonrtric/capifcore/internal/securityapi"

	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"
	"oransc.org/nonrtric/capifcore/internal/publishservice"

	"github.com/labstack/echo/v4"

	invokermocks "oransc.org/nonrtric/capifcore/internal/invokermanagement/mocks"
	servicemocks "oransc.org/nonrtric/capifcore/internal/providermanagement/mocks"
	publishmocks "oransc.org/nonrtric/capifcore/internal/publishservice/mocks"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostSecurityIdTokenInvokerRegistered(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	invokerRegisterMock.On("VerifyInvokerSecret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)
	serviceRegisterMock := servicemocks.ServiceRegister{}
	serviceRegisterMock.On("IsFunctionRegistered", mock.AnythingOfType("string")).Return(true)
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("IsAPIPublished", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)

	requestHandler := getEcho(&serviceRegisterMock, &publishRegisterMock, &invokerRegisterMock)

	data := url.Values{}
	clientId := "id"
	clientSecret := "secret"
	aefId := "aefId"
	path := "path"
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "3gpp#"+aefId+":"+path)

	encodedData := data.Encode()

	result := testutil.NewRequest().Post("/securities/invokerId/token").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(encodedData)).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultResponse securityapi.AccessTokenRsp
	err := result.UnmarshalBodyToObject(&resultResponse)
	assert.NoError(t, err, "error unmarshaling response")
	assert.NotEmpty(t, resultResponse.AccessToken)
	assert.Equal(t, "3gpp#"+aefId+":"+path, *resultResponse.Scope)
	assert.Equal(t, securityapi.AccessTokenRspTokenTypeBearer, resultResponse.TokenType)
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", clientId)
	invokerRegisterMock.AssertCalled(t, "VerifyInvokerSecret", clientId, clientSecret)
	serviceRegisterMock.AssertCalled(t, "IsFunctionRegistered", aefId)
	publishRegisterMock.AssertCalled(t, "IsAPIPublished", aefId, path)
}

func TestPostSecurityIdTokenInvokerNotRegistered(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(false)

	requestHandler := getEcho(nil, nil, &invokerRegisterMock)

	data := url.Values{}
	data.Set("client_id", "id")
	data.Add("client_secret", "secret")
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "3gpp#aefId:path")
	encodedData := data.Encode()

	result := testutil.NewRequest().Post("/securities/invokerId/token").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(encodedData)).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	var errDetails securityapi.AccessTokenErr
	err := result.UnmarshalBodyToObject(&errDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, securityapi.AccessTokenErrErrorInvalidClient, errDetails.Error)
	errMsg := "Invoker not registered"
	assert.Equal(t, &errMsg, errDetails.ErrorDescription)
}

func TestPostSecurityIdTokenInvokerSecretNotValid(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	invokerRegisterMock.On("VerifyInvokerSecret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false)

	requestHandler := getEcho(nil, nil, &invokerRegisterMock)

	data := url.Values{}
	data.Set("client_id", "id")
	data.Add("client_secret", "secret")
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "3gpp#aefId:path")
	encodedData := data.Encode()

	result := testutil.NewRequest().Post("/securities/invokerId/token").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(encodedData)).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	var errDetails securityapi.AccessTokenErr
	err := result.UnmarshalBodyToObject(&errDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, securityapi.AccessTokenErrErrorUnauthorizedClient, errDetails.Error)
	errMsg := "Invoker secret not valid"
	assert.Equal(t, &errMsg, errDetails.ErrorDescription)
}

func TestPostSecurityIdTokenFunctionNotRegistered(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	invokerRegisterMock.On("VerifyInvokerSecret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)
	serviceRegisterMock := servicemocks.ServiceRegister{}
	serviceRegisterMock.On("IsFunctionRegistered", mock.AnythingOfType("string")).Return(false)

	requestHandler := getEcho(&serviceRegisterMock, nil, &invokerRegisterMock)

	data := url.Values{}
	data.Set("client_id", "id")
	data.Add("client_secret", "secret")
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "3gpp#aefId:path")
	encodedData := data.Encode()

	result := testutil.NewRequest().Post("/securities/invokerId/token").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(encodedData)).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	var errDetails securityapi.AccessTokenErr
	err := result.UnmarshalBodyToObject(&errDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, securityapi.AccessTokenErrErrorInvalidScope, errDetails.Error)
	errMsg := "AEF Function not registered"
	assert.Equal(t, &errMsg, errDetails.ErrorDescription)
}

func TestPostSecurityIdTokenAPINotPublished(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	invokerRegisterMock.On("VerifyInvokerSecret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)
	serviceRegisterMock := servicemocks.ServiceRegister{}
	serviceRegisterMock.On("IsFunctionRegistered", mock.AnythingOfType("string")).Return(true)
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("IsAPIPublished", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false)

	requestHandler := getEcho(&serviceRegisterMock, &publishRegisterMock, &invokerRegisterMock)

	data := url.Values{}
	data.Set("client_id", "id")
	data.Add("client_secret", "secret")
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "3gpp#aefId:path")
	encodedData := data.Encode()

	result := testutil.NewRequest().Post("/securities/invokerId/token").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(encodedData)).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	var errDetails securityapi.AccessTokenErr
	err := result.UnmarshalBodyToObject(&errDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, securityapi.AccessTokenErrErrorInvalidScope, errDetails.Error)
	errMsg := "API not published"
	assert.Equal(t, &errMsg, errDetails.ErrorDescription)
}

func getEcho(serviceRegister providermanagement.ServiceRegister, publishRegister publishservice.PublishRegister, invokerRegister invokermanagement.InvokerRegister) *echo.Echo {
	swagger, err := securityapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	s := NewSecurity(serviceRegister, publishRegister, invokerRegister)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	securityapi.RegisterHandlers(e, s)
	return e
}
