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

	"oransc.org/nonrtric/capifcore/internal/common29122"

	invokermocks "oransc.org/nonrtric/capifcore/internal/invokermanagement/mocks"
	servicemocks "oransc.org/nonrtric/capifcore/internal/providermanagement/mocks"
	publishmocks "oransc.org/nonrtric/capifcore/internal/publishservice/mocks"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostSecurityIdToken(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	invokerRegisterMock.On("VerifyInvokerSecret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)
	serviceRegisterMock := servicemocks.ServiceRegister{}
	serviceRegisterMock.On("IsFunctionRegistered", mock.AnythingOfType("string")).Return(true)
	apiRegisterMock := publishmocks.APIRegister{}
	apiRegisterMock.On("IsAPIRegistered", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)

	requestHandler := getEcho(&serviceRegisterMock, &apiRegisterMock, &invokerRegisterMock)

	data := url.Values{}
	data.Set("client_id", "id")
	data.Add("client_secret", "secret")
	data.Add("grant_type", "client_credentials")
	data.Add("scope", "scope#aefId:path")
	encodedData := data.Encode()

	result := testutil.NewRequest().Post("/securities/invokerId/token").WithContentType("application/x-www-form-urlencoded").WithBody([]byte(encodedData)).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultResponse securityapi.AccessTokenRsp
	err := result.UnmarshalBodyToObject(&resultResponse)
	assert.NoError(t, err, "error unmarshaling response")
	assert.NotEmpty(t, resultResponse.AccessToken)
	assert.Equal(t, "scope#aefId:path", *resultResponse.Scope)
	assert.Equal(t, securityapi.AccessTokenRspTokenTypeBearer, resultResponse.TokenType)
	assert.Equal(t, common29122.DurationSec(0), resultResponse.ExpiresIn)
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", "id")
	invokerRegisterMock.AssertCalled(t, "VerifyInvokerSecret", "id", "secret")
	serviceRegisterMock.AssertCalled(t, "IsFunctionRegistered", "aefId")
	apiRegisterMock.AssertCalled(t, "IsAPIRegistered", "aefId", "path")
}

func getEcho(serviceRegister providermanagement.ServiceRegister, apiRegister publishservice.APIRegister, invokerRegister invokermanagement.InvokerRegister) *echo.Echo {
	swagger, err := securityapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	s := NewSecurity(serviceRegister, apiRegister, invokerRegister)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	securityapi.RegisterHandlers(e, s)
	return e
}
