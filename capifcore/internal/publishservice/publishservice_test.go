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
	"time"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/eventsapi"
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
	serviceRegisterMock := serviceMocks.ServiceRegister{}
	serviceRegisterMock.On("GetAefsForPublisher", apfId).Return([]string{aefId, "otherAefId"})
	helmManagerMock := helmMocks.HelmManager{}
	helmManagerMock.On("InstallHelmChart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	serviceUnderTest, eventChannel, requestHandler := getEcho(&serviceRegisterMock, &helmManagerMock)

	// Check no services published for provider
	result := testutil.NewRequest().Get("/"+apfId+"/service-apis").Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())

	apiName := "app-management"
	namespace := "namespace"
	repoName := "repoName"
	chartName := "chartName"
	releaseName := "releaseName"
	description := fmt.Sprintf("Description,%s,%s,%s,%s", namespace, repoName, chartName, releaseName)
	newServiceDescription := getServiceAPIDescription(aefId, apiName, description)

	// Publish a service for provider
	result = testutil.NewRequest().Post("/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultService publishapi.ServiceAPIDescription
	err := result.UnmarshalBodyToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	newApiId := "api_id_" + apiName
	assert.Equal(t, *resultService.ApiId, newApiId)
	assert.Equal(t, "http://example.com/"+apfId+"/service-apis/"+*resultService.ApiId, result.Recorder.Header().Get(echo.HeaderLocation))
	newServiceDescription.ApiId = &newApiId
	wantedAPILIst := []publishapi.ServiceAPIDescription{newServiceDescription}
	assert.True(t, serviceUnderTest.AreAPIsPublished(&wantedAPILIst))
	assert.True(t, serviceUnderTest.IsAPIPublished(aefId, apiName))
	serviceRegisterMock.AssertCalled(t, "GetAefsForPublisher", apfId)
	helmManagerMock.AssertCalled(t, "InstallHelmChart", namespace, repoName, chartName, releaseName)
	assert.ElementsMatch(t, []string{aefId}, serviceUnderTest.getAllAefIds())
	if publishEvent, ok := waitForEvent(eventChannel, 1*time.Second); ok {
		assert.Fail(t, "No event sent")
	} else {
		assert.Equal(t, *resultService.ApiId, (*publishEvent.EventDetail.ApiIds)[0])
		assert.Equal(t, eventsapi.CAPIFEventSERVICEAPIAVAILABLE, publishEvent.Events)
	}

	// Check that the service is published for the provider
	result = testutil.NewRequest().Get("/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	err = result.UnmarshalBodyToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, *resultService.ApiId, newApiId)

	// Delete the service
	helmManagerMock.On("UninstallHelmChart", mock.Anything, mock.Anything).Return(nil)

	result = testutil.NewRequest().Delete("/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)

	assert.Equal(t, http.StatusNoContent, result.Code())
	helmManagerMock.AssertCalled(t, "UninstallHelmChart", namespace, chartName)
	assert.Empty(t, serviceUnderTest.getAllAefIds())

	// Check no services published
	result = testutil.NewRequest().Get("/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)

	if publishEvent, ok := waitForEvent(eventChannel, 1*time.Second); ok {
		assert.Fail(t, "No event sent")
	} else {
		assert.Equal(t, *resultService.ApiId, (*publishEvent.EventDetail.ApiIds)[0])
		assert.Equal(t, eventsapi.CAPIFEventSERVICEAPIUNAVAILABLE, publishEvent.Events)
	}

	assert.Equal(t, http.StatusNotFound, result.Code())
}

func TestPostUnpublishedServiceWithUnregisteredFunction(t *testing.T) {
	apfId := "apfId"
	aefId := "aefId"
	serviceRegisterMock := serviceMocks.ServiceRegister{}
	serviceRegisterMock.On("GetAefsForPublisher", apfId).Return([]string{"otherAefId"})
	_, _, requestHandler := getEcho(&serviceRegisterMock, nil)

	newServiceDescription := getServiceAPIDescription(aefId, "apiName", "description")

	// Publish a service
	result := testutil.NewRequest().Post("/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())
	var resultError common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&resultError)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Contains(t, *resultError.Cause, aefId)
	assert.Contains(t, *resultError.Cause, "not registered")
	notFound := http.StatusNotFound
	assert.Equal(t, &notFound, resultError.Status)
}

func TestGetServices(t *testing.T) {
	apfId := "apfId"
	aefId := "aefId"
	serviceRegisterMock := serviceMocks.ServiceRegister{}
	serviceRegisterMock.On("GetAefsForPublisher", apfId).Return([]string{aefId})
	_, _, requestHandler := getEcho(&serviceRegisterMock, nil)

	// Check no services published for provider
	result := testutil.NewRequest().Get("/"+apfId+"/service-apis").Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())

	serviceDescription1 := getServiceAPIDescription(aefId, "api1", "Description")
	serviceDescription2 := getServiceAPIDescription(aefId, "api2", "Description")

	// Publish a service for provider
	testutil.NewRequest().Post("/"+apfId+"/service-apis").WithJsonBody(serviceDescription1).Go(t, requestHandler)
	testutil.NewRequest().Post("/"+apfId+"/service-apis").WithJsonBody(serviceDescription2).Go(t, requestHandler)

	// Get all services for provider
	result = testutil.NewRequest().Get("/"+apfId+"/service-apis").Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())
	var resultServices []publishapi.ServiceAPIDescription
	err := result.UnmarshalBodyToObject(&resultServices)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Len(t, resultServices, 2)
	apiId1 := "api_id_api1"
	serviceDescription1.ApiId = &apiId1
	apiId2 := "api_id_api2"
	serviceDescription2.ApiId = &apiId2
	assert.Contains(t, resultServices, serviceDescription1)
	assert.Contains(t, resultServices, serviceDescription2)
}

func TestGetPublishedServices(t *testing.T) {
	serviceUnderTest := NewPublishService(nil, nil, nil)

	profiles := make([]publishapi.AefProfile, 1)
	serviceDescription := publishapi.ServiceAPIDescription{
		AefProfiles: &profiles,
	}
	serviceUnderTest.publishedServices["publisher1"] = []publishapi.ServiceAPIDescription{
		serviceDescription,
	}
	serviceUnderTest.publishedServices["publisher2"] = []publishapi.ServiceAPIDescription{
		serviceDescription,
	}
	result := serviceUnderTest.GetAllPublishedServices()
	assert.Len(t, result, 2)
}

func TestUpdateDescription(t *testing.T) {
	apfId := "apfId"
	serviceApiId := "serviceApiId"
	aefId := "aefId"
	apiName := "apiName"
	description := "description"

	serviceRegisterMock := serviceMocks.ServiceRegister{}
	serviceRegisterMock.On("GetAefsForPublisher", apfId).Return([]string{aefId, "otherAefId", "aefIdNew"})
	helmManagerMock := helmMocks.HelmManager{}
	helmManagerMock.On("InstallHelmChart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	serviceUnderTest, eventChannel, requestHandler := getEcho(&serviceRegisterMock, &helmManagerMock)
	serviceDescription := getServiceAPIDescription(aefId, apiName, description)
	serviceDescription.ApiId = &serviceApiId
	serviceUnderTest.publishedServices[apfId] = []publishapi.ServiceAPIDescription{serviceDescription}
	(*serviceDescription.AefProfiles)[0].AefId = aefId

	//Modify the service
	updatedServiceDescription := getServiceAPIDescription(aefId, apiName, description)
	updatedServiceDescription.ApiId = &serviceApiId
	(*updatedServiceDescription.AefProfiles)[0].AefId = aefId
	newDescription := "new description"
	updatedServiceDescription.Description = &newDescription
	newDomainName := "new domainName"
	(*updatedServiceDescription.AefProfiles)[0].DomainName = &newDomainName

	newProfileDomain := "new profile Domain name"
	var protocol publishapi.Protocol = "HTTP_1_1"
	test := make([]publishapi.AefProfile, 1)
	test = *updatedServiceDescription.AefProfiles
	test = append(test, publishapi.AefProfile{

		AefId:      "aefIdNew",
		DomainName: &newProfileDomain,
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
	)

	updatedServiceDescription.AefProfiles = &test

	result := testutil.NewRequest().Put("/"+apfId+"/service-apis/"+serviceApiId).WithJsonBody(updatedServiceDescription).Go(t, requestHandler)

	var resultService publishapi.ServiceAPIDescription
	assert.Equal(t, http.StatusOK, result.Code())
	err := result.UnmarshalBodyToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newDescription, *resultService.Description)
	assert.Equal(t, newDomainName, *(*resultService.AefProfiles)[0].DomainName)
	assert.Equal(t, "aefIdNew", (*resultService.AefProfiles)[1].AefId)
	assert.True(t, serviceUnderTest.IsAPIPublished("aefIdNew", "path"))

	if publishEvent, ok := waitForEvent(eventChannel, 1*time.Second); ok {
		assert.Fail(t, "No event sent")
	} else {
		assert.Equal(t, *resultService.ApiId, (*publishEvent.EventDetail.ApiIds)[0])
		assert.Equal(t, eventsapi.CAPIFEventSERVICEAPIUPDATE, publishEvent.Events)
	}
}

func TestUpdateValidServiceWithDeletedFunction(t *testing.T) {
	apfId := "apfId"
	serviceApiId := "serviceApiId"
	aefId := "aefId"
	apiName := "apiName"
	description := "description"

	serviceRegisterMock := serviceMocks.ServiceRegister{}
	serviceRegisterMock.On("GetAefsForPublisher", apfId).Return([]string{aefId, "otherAefId", "aefIdNew"})
	helmManagerMock := helmMocks.HelmManager{}
	helmManagerMock.On("InstallHelmChart", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	serviceUnderTest, _, requestHandler := getEcho(&serviceRegisterMock, &helmManagerMock)

	serviceDescription := getServiceAPIDescription(aefId, apiName, description)
	serviceDescription.ApiId = &serviceApiId
	(*serviceDescription.AefProfiles)[0].AefId = aefId

	newProfileDomain := "new profile Domain name"
	var protocol publishapi.Protocol = "HTTP_1_1"
	test := make([]publishapi.AefProfile, 1)
	test = *serviceDescription.AefProfiles
	test = append(test, publishapi.AefProfile{

		AefId:      "aefIdNew",
		DomainName: &newProfileDomain,
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
	)
	serviceDescription.AefProfiles = &test
	serviceUnderTest.publishedServices[apfId] = []publishapi.ServiceAPIDescription{serviceDescription}

	//Modify the service
	updatedServiceDescription := getServiceAPIDescription(aefId, apiName, description)
	updatedServiceDescription.ApiId = &serviceApiId
	test1 := make([]publishapi.AefProfile, 1)
	test1 = *updatedServiceDescription.AefProfiles
	test1 = append(test1, publishapi.AefProfile{

		AefId:      "aefIdNew",
		DomainName: &newProfileDomain,
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
	)
	updatedServiceDescription.AefProfiles = &test1
	testFunc := []publishapi.AefProfile{
		(*updatedServiceDescription.AefProfiles)[1],
	}

	updatedServiceDescription.AefProfiles = &testFunc
	result := testutil.NewRequest().Put("/"+apfId+"/service-apis/"+serviceApiId).WithJsonBody(updatedServiceDescription).Go(t, requestHandler)
	var resultService publishapi.ServiceAPIDescription
	assert.Equal(t, http.StatusOK, result.Code())
	err := result.UnmarshalBodyToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Len(t, (*resultService.AefProfiles), 1)
	assert.False(t, serviceUnderTest.IsAPIPublished("aefId", "path"))

}
func getEcho(serviceRegister providermanagement.ServiceRegister, helmManager helmmanagement.HelmManager) (*PublishService, chan eventsapi.EventNotification, *echo.Echo) {
	swagger, err := publishapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	eventChannel := make(chan eventsapi.EventNotification)
	ps := NewPublishService(serviceRegister, helmManager, eventChannel)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	publishapi.RegisterHandlers(e, ps)
	return ps, eventChannel, e
}

func getServiceAPIDescription(aefId, apiName, description string) publishapi.ServiceAPIDescription {
	domainName := "domainName"
	var protocol publishapi.Protocol = "HTTP_1_1"
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
		ApiName:     apiName,
		Description: &description,
	}
}

// waitForEvent waits for the channel to receive an event for the specified max timeout.
// Returns true if waiting timed out.
func waitForEvent(ch chan eventsapi.EventNotification, timeout time.Duration) (*eventsapi.EventNotification, bool) {
	select {
	case event := <-ch:
		return &event, false // completed normally
	case <-time.After(timeout):
		return nil, true // timed out
	}
}
