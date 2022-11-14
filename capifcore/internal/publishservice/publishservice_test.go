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

package publishservice

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"

	"github.com/labstack/echo/v4"

	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/helmmanagement"
	helmMocks "oransc.org/nonrtric/capifcore/internal/helmmanagement/mocks"
	serviceMocks "oransc.org/nonrtric/capifcore/internal/providermanagement/mocks"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPublishUnpublishService(t *testing.T) {
	apfId := "apfId"
	aefId := "aefId"
	newApiId := "api_id_app-management"
	serviceRegisterMock := serviceMocks.ServiceRegister{}
	serviceRegisterMock.On("GetAefsForPublisher", apfId).Return([]string{aefId, "otherAefId"})
	helmManagerMock := helmMocks.HelmManager{}
	helmManagerMock.On("InstallHelmChart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	serviceUnderTest, requestHandler := getEcho(&serviceRegisterMock, &helmManagerMock)

	// Check no services published
	result := testutil.NewRequest().Get("/aefId/service-apis/"+newApiId).Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())

	domainName := "domain"
	var protocol publishapi.Protocol = "HTTP_1_1"
	description := "Description,namespace,repoName,chartName,releaseName"
	newServiceDescription := getServiceAPIDescription(aefId, domainName, description, protocol)

	// Publish a service
	result = testutil.NewRequest().Post("/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultService publishapi.ServiceAPIDescription
	err := result.UnmarshalBodyToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, *resultService.ApiId, newApiId)
	assert.Equal(t, "http://example.com/"+apfId+"/service-apis/"+*resultService.ApiId, result.Recorder.Header().Get(echo.HeaderLocation))
	newServiceDescription.ApiId = &newApiId
	wantedAPILIst := []publishapi.ServiceAPIDescription{newServiceDescription}
	assert.True(t, serviceUnderTest.AreAPIsRegistered(&wantedAPILIst))
	assert.True(t, serviceUnderTest.IsAPIRegistered("aefId", "app-management"))
	serviceRegisterMock.AssertCalled(t, "GetAefsForPublisher", apfId)
	helmManagerMock.AssertCalled(t, "InstallHelmChart", "namespace", "repoName", "chartName", "releaseName")
	assert.ElementsMatch(t, []string{aefId}, serviceUnderTest.getAllAefIds())

	// Check that service is published
	result = testutil.NewRequest().Get("/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	err = result.UnmarshalBodyToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, *resultService.ApiId, newApiId)

	// Delete a service
	helmManagerMock.On("UninstallHelmChart", mock.Anything, mock.Anything).Return(nil)
	result = testutil.NewRequest().Delete("/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)

	assert.Equal(t, http.StatusNoContent, result.Code())
	helmManagerMock.AssertCalled(t, "UninstallHelmChart", "namespace", "chartName")
	assert.Empty(t, serviceUnderTest.getAllAefIds())

	// Check no services published
	result = testutil.NewRequest().Get("/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())
}

func TestPostUnpublishedServiceWithUnregisteredFunction(t *testing.T) {
	apfId := "apfId"
	aefId := "aefId"
	serviceRegisterMock := serviceMocks.ServiceRegister{}
	serviceRegisterMock.On("GetAefsForPublisher", apfId).Return([]string{"otherAefId"})
	_, requestHandler := getEcho(&serviceRegisterMock, nil)

	domainName := "domain"
	var protocol publishapi.Protocol = "HTTP_1_1"
	description := "Description"
	newServiceDescription := getServiceAPIDescription(aefId, domainName, description, protocol)

	// Publish a service
	result := testutil.NewRequest().Post("/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())
	var resultError common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&resultError)
	assert.NoError(t, err, "error unmarshaling response")
	errMsg := "Function not registered, aefId"
	assert.Equal(t, &errMsg, resultError.Cause)
	notFound := http.StatusNotFound
	assert.Equal(t, &notFound, resultError.Status)
}

func getEcho(serviceRegister providermanagement.ServiceRegister, helmManager helmmanagement.HelmManager) (*PublishService, *echo.Echo) {
	swagger, err := publishapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	ps := NewPublishService(serviceRegister, helmManager)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	publishapi.RegisterHandlers(e, ps)
	return ps, e
}

func getServiceAPIDescription(aefId, domainName, description string, protocol publishapi.Protocol) publishapi.ServiceAPIDescription {
	return publishapi.ServiceAPIDescription{
		AefProfiles: &[]publishapi.AefProfile{
			{
				AefId:      aefId,
				DomainName: &domainName,
				Protocol:   &protocol,
				Versions: []publishapi.Version{
					{
						ApiVersion: "v1",
						Resources: &[]publishapi.Resource{
							{
								CommType: "REQUEST_RESPONSE",
								Operations: &[]publishapi.Operation{
									"POST",
								},
								ResourceName: "app",
								Uri:          "app",
							},
						},
					},
				},
			},
		},
		ApiName:     "app-management",
		Description: &description,
	}
}
