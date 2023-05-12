// -
//
//	========================LICENSE_START=================================
//	O-RAN-SC
//	%%
//	Copyright (C) 2022: Nordix Foundation
//	%%
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	     http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
//	========================LICENSE_END===================================
package invokermanagement

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"oransc.org/nonrtric/capifcore/internal/eventsapi"
	"oransc.org/nonrtric/capifcore/internal/invokermanagementapi"
	"oransc.org/nonrtric/capifcore/internal/keycloak"

	"github.com/labstack/echo/v4"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	keycloackmocks "oransc.org/nonrtric/capifcore/internal/keycloak/mocks"
	"oransc.org/nonrtric/capifcore/internal/publishservice"
	publishmocks "oransc.org/nonrtric/capifcore/internal/publishservice/mocks"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOnboardInvoker(t *testing.T) {
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

	invokerInfo := "invoker a"
	wantedInvokerSecret := "onboarding_secret_" + strings.Replace(invokerInfo, " ", "_", 1)
	var client keycloak.Client
	client.Secret = &wantedInvokerSecret
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("GetAllPublishedServices").Return(publishedServices)

	accessMgmMock := keycloackmocks.AccessManagement{}
	accessMgmMock.On("AddClient", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	accessMgmMock.On("GetClientRepresentation", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&client, nil)

	invokerUnderTest, eventChannel, requestHandler := getEcho(&publishRegisterMock, &accessMgmMock)

	newInvoker := getInvoker(invokerInfo)

	// Onboard a valid invoker
	result := testutil.NewRequest().Post("/onboardedInvokers").WithJsonBody(newInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultInvoker invokermanagementapi.APIInvokerEnrolmentDetails
	err := result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	wantedInvokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)
	assert.Equal(t, wantedInvokerId, *resultInvoker.ApiInvokerId)
	assert.Equal(t, newInvoker.NotificationDestination, resultInvoker.NotificationDestination)
	assert.Equal(t, newInvoker.OnboardingInformation.ApiInvokerPublicKey, resultInvoker.OnboardingInformation.ApiInvokerPublicKey)

	assert.Equal(t, wantedInvokerSecret, *resultInvoker.OnboardingInformation.OnboardingSecret)
	assert.Equal(t, "http://example.com/onboardedInvokers/"+*resultInvoker.ApiInvokerId, result.Recorder.Header().Get(echo.HeaderLocation))
	assert.True(t, invokerUnderTest.IsInvokerRegistered(wantedInvokerId))
	assert.True(t, invokerUnderTest.VerifyInvokerSecret(wantedInvokerId, wantedInvokerSecret))
	publishRegisterMock.AssertCalled(t, "GetAllPublishedServices")
	assert.Equal(t, invokermanagementapi.APIList(publishedServices), *resultInvoker.ApiList)
	if invokerEvent, timeout := waitForEvent(eventChannel, 1*time.Second); timeout {
		assert.Fail(t, "No event sent")
	} else {
		assert.Equal(t, *resultInvoker.ApiInvokerId, (*invokerEvent.EventDetail.ApiInvokerIds)[0])
		assert.Equal(t, eventsapi.CAPIFEventAPIINVOKERONBOARDED, invokerEvent.Events)
	}

	// Onboarding the same invoker should result in Forbidden
	result = testutil.NewRequest().Post("/onboardedInvokers").WithJsonBody(newInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusForbidden, result.Code())
	var problemDetails common29122.ProblemDetails
	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusForbidden, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "already onboarded")

	// Onboard an invoker missing required NotificationDestination, should get 400 with problem details
	invalidInvoker := invokermanagementapi.APIInvokerEnrolmentDetails{
		OnboardingInformation: invokermanagementapi.OnboardingInformation{
			ApiInvokerPublicKey: "newKey",
		},
	}
	result = testutil.NewRequest().Post("/onboardedInvokers").WithJsonBody(invalidInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "missing")
	assert.Contains(t, *problemDetails.Cause, "NotificationDestination")

	// Onboard an invoker missing required OnboardingInformation.ApiInvokerPublicKey, should get 400 with problem details
	invalidInvoker = invokermanagementapi.APIInvokerEnrolmentDetails{
		NotificationDestination: "http://golang.cafe/",
	}

	result = testutil.NewRequest().Post("/onboardedInvokers").WithJsonBody(invalidInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "missing")
	assert.Contains(t, *problemDetails.Cause, "OnboardingInformation.ApiInvokerPublicKey")
}

func TestDeleteInvoker(t *testing.T) {
	invokerUnderTest, eventChannel, requestHandler := getEcho(nil, nil)

	invokerId := "invokerId"
	newInvoker := invokermanagementapi.APIInvokerEnrolmentDetails{
		ApiInvokerId:            &invokerId,
		NotificationDestination: "url",
		OnboardingInformation: invokermanagementapi.OnboardingInformation{
			ApiInvokerPublicKey: "key",
		},
	}
	invokerUnderTest.onboardedInvokers[invokerId] = newInvoker
	assert.True(t, invokerUnderTest.IsInvokerRegistered(invokerId))

	// Delete the invoker
	result := testutil.NewRequest().Delete("/onboardedInvokers/"+invokerId).Go(t, requestHandler)

	assert.Equal(t, http.StatusNoContent, result.Code())
	assert.False(t, invokerUnderTest.IsInvokerRegistered(invokerId))
	if invokerEvent, timeout := waitForEvent(eventChannel, 1*time.Second); timeout {
		assert.Fail(t, "No event sent")
	} else {
		assert.Equal(t, invokerId, (*invokerEvent.EventDetail.ApiInvokerIds)[0])
		assert.Equal(t, eventsapi.CAPIFEventAPIINVOKEROFFBOARDED, invokerEvent.Events)
	}
}

func TestUpdateInvoker(t *testing.T) {
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("GetAllPublishedServices").Return([]publishserviceapi.ServiceAPIDescription{})
	serviceUnderTest, _, requestHandler := getEcho(&publishRegisterMock, nil)

	invokerId := "invokerId"
	invoker := invokermanagementapi.APIInvokerEnrolmentDetails{
		ApiInvokerId:            &invokerId,
		NotificationDestination: "http://golang.cafe/",
		OnboardingInformation: invokermanagementapi.OnboardingInformation{
			ApiInvokerPublicKey: "key",
		},
	}
	serviceUnderTest.onboardedInvokers[invokerId] = invoker

	// Update the invoker with valid invoker, should return 200 with updated invoker details
	newNotifURL := "http://golang.org/"
	invoker.NotificationDestination = common29122.Uri(newNotifURL)
	newPublicKey := "newPublicKey"
	invoker.OnboardingInformation.ApiInvokerPublicKey = newPublicKey
	result := testutil.NewRequest().Put("/onboardedInvokers/"+invokerId).WithJsonBody(invoker).Go(t, requestHandler)

	var resultInvoker invokermanagementapi.APIInvokerEnrolmentDetails
	assert.Equal(t, http.StatusOK, result.Code())
	err := result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, invokerId, *resultInvoker.ApiInvokerId)
	assert.Equal(t, newNotifURL, string(resultInvoker.NotificationDestination))
	assert.Equal(t, newPublicKey, resultInvoker.OnboardingInformation.ApiInvokerPublicKey)

	// Update with an invoker missing required NotificationDestination, should get 400 with problem details
	validOnboardingInfo := invokermanagementapi.OnboardingInformation{
		ApiInvokerPublicKey: "key",
	}
	invalidInvoker := invokermanagementapi.APIInvokerEnrolmentDetails{
		ApiInvokerId:          &invokerId,
		OnboardingInformation: validOnboardingInfo,
	}
	result = testutil.NewRequest().Put("/onboardedInvokers/"+invokerId).WithJsonBody(invalidInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	var problemDetails common29122.ProblemDetails
	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "missing")
	assert.Contains(t, *problemDetails.Cause, "NotificationDestination")

	// Update with an invoker missing required OnboardingInformation.ApiInvokerPublicKey, should get 400 with problem details
	invalidInvoker.NotificationDestination = "http://golang.org/"
	invalidInvoker.OnboardingInformation = invokermanagementapi.OnboardingInformation{}
	result = testutil.NewRequest().Put("/onboardedInvokers/"+invokerId).WithJsonBody(invalidInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "missing")
	assert.Contains(t, *problemDetails.Cause, "OnboardingInformation.ApiInvokerPublicKey")

	// Update with an invoker with other ApiInvokerId than the one provided in the URL, should get 400 with problem details
	invalidId := "1"
	invalidInvoker.ApiInvokerId = &invalidId
	invalidInvoker.OnboardingInformation = validOnboardingInfo
	result = testutil.NewRequest().Put("/onboardedInvokers/"+invokerId).WithJsonBody(invalidInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "not matching")
	assert.Contains(t, *problemDetails.Cause, "ApiInvokerId")

	// Update an invoker that has not been onboarded, should get 404 with problem details
	missingId := "1"
	invoker.ApiInvokerId = &missingId
	result = testutil.NewRequest().Put("/onboardedInvokers/"+missingId).WithJsonBody(invoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())
	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusNotFound, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "not been onboarded")
	assert.Contains(t, *problemDetails.Cause, "invoker")
}

func TestGetInvokerApiList(t *testing.T) {
	aefProfiles1 := []publishserviceapi.AefProfile{
		getAefProfile("aefId"),
	}
	apiId := "apiId"
	apiList := []publishserviceapi.ServiceAPIDescription{
		{
			ApiId:       &apiId,
			AefProfiles: &aefProfiles1,
		},
	}
	aefProfiles2 := []publishserviceapi.AefProfile{
		getAefProfile("aefId2"),
	}
	apiId2 := "apiId2"
	apiList = append(apiList, publishserviceapi.ServiceAPIDescription{
		ApiId:       &apiId2,
		AefProfiles: &aefProfiles2,
	})
	publishRegisterMock := publishmocks.PublishRegister{}
	publishRegisterMock.On("GetAllPublishedServices").Return(apiList)
	invokerUnderTest, _, _ := getEcho(&publishRegisterMock, nil)

	invokerInfo := "invoker a"
	newInvoker := getInvoker(invokerInfo)
	invokerAId := "api_invoker_id_" + strings.ReplaceAll(invokerInfo, " ", "_")
	newInvoker.ApiInvokerId = &invokerAId
	invokerUnderTest.onboardedInvokers[invokerAId] = newInvoker
	invokerInfo = "invoker b"
	newInvoker = getInvoker(invokerInfo)
	invokerId := "api_invoker_id_" + strings.ReplaceAll(invokerInfo, " ", "_")
	newInvoker.ApiInvokerId = &invokerId
	invokerUnderTest.onboardedInvokers[invokerId] = newInvoker

	wantedApiList := invokerUnderTest.GetInvokerApiList(invokerAId)
	assert.NotNil(t, wantedApiList)
	assert.Len(t, *wantedApiList, 2)
	assert.Equal(t, apiId, *(*wantedApiList)[0].ApiId)
}

func getEcho(publishRegister publishservice.PublishRegister, keycloakMgm keycloak.AccessManagement) (*InvokerManager, chan eventsapi.EventNotification, *echo.Echo) {
	swagger, err := invokermanagementapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	eventChannel := make(chan eventsapi.EventNotification)
	im := NewInvokerManager(publishRegister, keycloakMgm, eventChannel)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	invokermanagementapi.RegisterHandlers(e, im)
	return im, eventChannel, e
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

func getInvoker(invokerInfo string) invokermanagementapi.APIInvokerEnrolmentDetails {
	newInvoker := invokermanagementapi.APIInvokerEnrolmentDetails{
		ApiInvokerInformation:   &invokerInfo,
		NotificationDestination: "http://golang.cafe/",
		OnboardingInformation: invokermanagementapi.OnboardingInformation{
			ApiInvokerPublicKey: "key",
		},
		ApiList: nil,
	}
	return newInvoker
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
