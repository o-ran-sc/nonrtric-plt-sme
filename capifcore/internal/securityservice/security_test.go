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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/keycloak"
	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"
	"oransc.org/nonrtric/capifcore/internal/securityapi"

	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"
	"oransc.org/nonrtric/capifcore/internal/publishservice"

	"github.com/labstack/echo/v4"

	invokermocks "oransc.org/nonrtric/capifcore/internal/invokermanagement/mocks"
	keycloackmocks "oransc.org/nonrtric/capifcore/internal/keycloak/mocks"
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

	jwt := keycloak.Jwttoken{
		AccessToken: "eyJhbGNIn0.e3YTQ0xLjEifQ.FcqCwCy7iJiOmw",
		ExpiresIn:   300,
		Scope:       "3gpp#aefIdpath",
	}
	accessMgmMock := keycloackmocks.AccessManagement{}
	accessMgmMock.On("GetToken", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(jwt, nil)

	requestHandler := getEcho(&serviceRegisterMock, &publishRegisterMock, &invokerRegisterMock, &accessMgmMock)

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
	assert.Equal(t, securityapi.AccessTokenRspTokenTypeBearer, resultResponse.TokenType)
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", clientId)
	invokerRegisterMock.AssertCalled(t, "VerifyInvokerSecret", clientId, clientSecret)
	serviceRegisterMock.AssertCalled(t, "IsFunctionRegistered", aefId)
	publishRegisterMock.AssertCalled(t, "IsAPIPublished", aefId, path)
	accessMgmMock.AssertCalled(t, "GetToken", clientId, clientSecret, "3gpp#"+aefId+":"+path, "invokerrealm")
}

func TestPostSecurityIdTokenInvokerNotRegistered(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(false)

	requestHandler := getEcho(nil, nil, &invokerRegisterMock, nil)

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

	requestHandler := getEcho(nil, nil, &invokerRegisterMock, nil)

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

	requestHandler := getEcho(&serviceRegisterMock, nil, &invokerRegisterMock, nil)

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

	requestHandler := getEcho(&serviceRegisterMock, &publishRegisterMock, &invokerRegisterMock, nil)

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

func TestPostSecurityIdTokenInvokerInvalidCredentials(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	invokerRegisterMock.On("VerifyInvokerSecret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)
	serviceRegisterMock := servicemocks.ServiceRegister{}
	serviceRegisterMock.On("IsFunctionRegistered", mock.AnythingOfType("string")).Return(true)
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("IsAPIPublished", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true)

	jwt := keycloak.Jwttoken{}
	accessMgmMock := keycloackmocks.AccessManagement{}
	accessMgmMock.On("GetToken", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(jwt, errors.New("invalid_credentials"))

	requestHandler := getEcho(&serviceRegisterMock, &publishRegisterMock, &invokerRegisterMock, &accessMgmMock)

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

	assert.Equal(t, http.StatusBadRequest, result.Code())
	var resultResponse securityapi.AccessTokenErr
	err := result.UnmarshalBodyToObject(&resultResponse)
	assert.NoError(t, err, "error unmarshaling response")
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", clientId)
	invokerRegisterMock.AssertCalled(t, "VerifyInvokerSecret", clientId, clientSecret)
	serviceRegisterMock.AssertCalled(t, "IsFunctionRegistered", aefId)
	publishRegisterMock.AssertCalled(t, "IsAPIPublished", aefId, path)
	accessMgmMock.AssertCalled(t, "GetToken", clientId, clientSecret, "3gpp#"+aefId+":"+path, "invokerrealm")
}

func TestPutTrustedInvokerSuccessfully(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	aefId := "aefId"
	aefProfile := getAefProfile(aefId)
	aefProfile.SecurityMethods = &[]publishserviceapi.SecurityMethod{
		publishserviceapi.SecurityMethodPKI,
	}
	aefProfiles := []publishserviceapi.AefProfile{
		aefProfile,
	}
	apiId := "apiId"
	publishedServices := []publishserviceapi.ServiceAPIDescription{
		{
			ApiId:       &apiId,
			AefProfiles: &aefProfiles,
		},
	}
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("GetAllPublishedServices").Return(publishedServices)

	requestHandler := getEcho(nil, &publishRegisterMock, &invokerRegisterMock, nil)

	invokerId := "invokerId"
	serviceSecurityUnderTest := getServiceSecurity(aefId, apiId)
	serviceSecurityUnderTest.SecurityInfo[0].ApiId = &apiId

	result := testutil.NewRequest().Put("/trustedInvokers/"+invokerId).WithJsonBody(serviceSecurityUnderTest).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultResponse securityapi.ServiceSecurity
	err := result.UnmarshalBodyToObject(&resultResponse)
	assert.NoError(t, err, "error unmarshaling response")
	assert.NotEmpty(t, resultResponse.NotificationDestination)

	for _, security := range resultResponse.SecurityInfo {
		assert.Equal(t, *security.ApiId, apiId)
		assert.Equal(t, *security.SelSecurityMethod, publishserviceapi.SecurityMethodPKI)
	}
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", invokerId)

}

func TestPutTrustedInkoverNotRegistered(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(false)

	requestHandler := getEcho(nil, nil, &invokerRegisterMock, nil)

	invokerId := "invokerId"
	serviceSecurityUnderTest := getServiceSecurity("aefId", "apiId")

	result := testutil.NewRequest().Put("/trustedInvokers/"+invokerId).WithJsonBody(serviceSecurityUnderTest).Go(t, requestHandler)

	badRequest := http.StatusBadRequest
	assert.Equal(t, badRequest, result.Code())
	var problemDetails common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, &badRequest, problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "Invoker not registered")
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", invokerId)
}

func TestPutTrustedInkoverInvalidInputServiceSecurity(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)

	requestHandler := getEcho(nil, nil, &invokerRegisterMock, nil)

	invokerId := "invokerId"
	notificationUrl := "url"
	serviceSecurityUnderTest := getServiceSecurity("aefId", "apiId")
	serviceSecurityUnderTest.NotificationDestination = common29122.Uri(notificationUrl)

	result := testutil.NewRequest().Put("/trustedInvokers/"+invokerId).WithJsonBody(serviceSecurityUnderTest).Go(t, requestHandler)

	badRequest := http.StatusBadRequest
	assert.Equal(t, badRequest, result.Code())
	var problemDetails common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, &badRequest, problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "ServiceSecurity has invalid notificationDestination")
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", invokerId)
}

func TestPutTrustedInvokerInterfaceDetailsNotNil(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)
	aefId := "aefId"
	aefProfile := getAefProfile(aefId)
	aefProfile.SecurityMethods = &[]publishserviceapi.SecurityMethod{
		publishserviceapi.SecurityMethodPKI,
	}
	aefProfiles := []publishserviceapi.AefProfile{
		aefProfile,
	}
	apiId := "apiId"
	publishedServices := []publishserviceapi.ServiceAPIDescription{
		{
			ApiId:       &apiId,
			AefProfiles: &aefProfiles,
		},
	}
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("GetAllPublishedServices").Return(publishedServices)

	requestHandler := getEcho(nil, &publishRegisterMock, &invokerRegisterMock, nil)

	invokerId := "invokerId"
	serviceSecurityUnderTest := getServiceSecurity(aefId, apiId)
	serviceSecurityUnderTest.SecurityInfo[0] = securityapi.SecurityInformation{
		ApiId: &apiId,
		PrefSecurityMethods: []publishserviceapi.SecurityMethod{
			publishserviceapi.SecurityMethodOAUTH,
		},
		InterfaceDetails: &publishserviceapi.InterfaceDescription{
			SecurityMethods: &[]publishserviceapi.SecurityMethod{
				publishserviceapi.SecurityMethodPSK,
			},
		},
	}

	result := testutil.NewRequest().Put("/trustedInvokers/"+invokerId).WithJsonBody(serviceSecurityUnderTest).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultResponse securityapi.ServiceSecurity
	err := result.UnmarshalBodyToObject(&resultResponse)
	assert.NoError(t, err, "error unmarshaling response")
	assert.NotEmpty(t, resultResponse.NotificationDestination)

	for _, security := range resultResponse.SecurityInfo {
		assert.Equal(t, apiId, *security.ApiId)
		assert.Equal(t, publishserviceapi.SecurityMethodPSK, *security.SelSecurityMethod)
	}
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", invokerId)

}

func TestPutTrustedInvokerNotFoundSecurityMethod(t *testing.T) {
	invokerRegisterMock := invokermocks.InvokerRegister{}
	invokerRegisterMock.On("IsInvokerRegistered", mock.AnythingOfType("string")).Return(true)

	aefProfiles := []publishserviceapi.AefProfile{
		getAefProfile("aefId"),
	}
	apiId := "apiId"
	publishedServices := []publishserviceapi.ServiceAPIDescription{
		{
			ApiId:       &apiId,
			AefProfiles: &aefProfiles,
		},
	}
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("GetAllPublishedServices").Return(publishedServices)

	requestHandler := getEcho(nil, &publishRegisterMock, &invokerRegisterMock, nil)

	invokerId := "invokerId"
	serviceSecurityUnderTest := getServiceSecurity("aefId", "apiId")

	result := testutil.NewRequest().Put("/trustedInvokers/"+invokerId).WithJsonBody(serviceSecurityUnderTest).Go(t, requestHandler)

	badRequest := http.StatusBadRequest
	assert.Equal(t, badRequest, result.Code())
	var problemDetails common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, &badRequest, problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "not found")
	assert.Contains(t, *problemDetails.Cause, "security method")
	invokerRegisterMock.AssertCalled(t, "IsInvokerRegistered", invokerId)
}

func getEcho(serviceRegister providermanagement.ServiceRegister, publishRegister publishservice.PublishRegister, invokerRegister invokermanagement.InvokerRegister, keycloakMgm keycloak.AccessManagement) *echo.Echo {
	swagger, err := securityapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	s := NewSecurity(serviceRegister, publishRegister, invokerRegister, keycloakMgm)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	securityapi.RegisterHandlers(e, s)
	return e
}

func getServiceSecurity(aefId string, apiId string) securityapi.ServiceSecurity {
	return securityapi.ServiceSecurity{
		NotificationDestination: common29122.Uri("http://golang.cafe/"),
		SecurityInfo: []securityapi.SecurityInformation{
			{
				AefId: &aefId,
				ApiId: &apiId,
				PrefSecurityMethods: []publishserviceapi.SecurityMethod{
					publishserviceapi.SecurityMethodOAUTH,
				},
			},
		},
	}
}

func getAefProfile(aefId string) publishserviceapi.AefProfile {
	return publishserviceapi.AefProfile{
		AefId: aefId,
		Versions: []publishserviceapi.Version{
			{
				Resources: &[]publishserviceapi.Resource{
					{
						CommType: "REQUEST_RESPONSE",
					},
				},
			},
		},
	}
}
