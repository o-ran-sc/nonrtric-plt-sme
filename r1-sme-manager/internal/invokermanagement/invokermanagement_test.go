// -
//
//		========================LICENSE_START=================================
//		O-RAN-SC
//		%%
//	  Copyright (C) 2023-2024: OpenInfra Foundation Europe
//		%%
//		Licensed under the Apache License, Version 2.0 (the "License");
//		you may not use this file except in compliance with the License.
//		You may obtain a copy of the License at
//
//		     http://www.apache.org/licenses/LICENSE-2.0
//
//		Unless required by applicable law or agreed to in writing, software
//		distributed under the License is distributed on an "AS IS" BASIS,
//		WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//		See the License for the specific language governing permissions and
//		limitations under the License.
//		========================LICENSE_END===================================
package invokermanagement

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"oransc.org/nonrtric/r1-sme-manager/internal/common29122"
	"oransc.org/nonrtric/r1-sme-manager/internal/envreader"
	"oransc.org/nonrtric/r1-sme-manager/internal/invokermanagementapi"
	"oransc.org/nonrtric/r1-sme-manager/internal/kongclear"
	"oransc.org/nonrtric/r1-sme-manager/internal/providermanagement"
	provapi "oransc.org/nonrtric/r1-sme-manager/internal/providermanagementapi"
	"oransc.org/nonrtric/r1-sme-manager/internal/publishservice"
	publishapi "oransc.org/nonrtric/r1-sme-manager/internal/publishserviceapi"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var requestHandler *echo.Echo

func TestMain(m *testing.M) {
	err := setupTest()
	if err != nil {
		return
	}

	ret := m.Run()
	if ret == 0 {
		teardown()
	}
	os.Exit(ret)
}

func setupTest() error {
	myEnv, myPorts, err := envreader.ReadDotEnv()
	if err != nil {
		log.Fatal("error loading environment file on setupTest")
		return err
	}

	requestHandler, err = getEcho(myEnv, myPorts)
	if err != nil {
		log.Fatal("getEcho fatal error on setupTest")
		return err
	}
	err = teardown()
	if err != nil {
		log.Fatal("getEcho fatal error on teardown")
		return err
	}

	return err
}

func getProvider() provapi.APIProviderEnrolmentDetails {
	var (
		domainInfo  = "Kong"
		funcInfoAPF = "rApp Kong as APF"
		funcInfoAEF = "rApp Kong as AEF"
	)

	testFuncs := []provapi.APIProviderFunctionDetails{
		{
			ApiProvFuncInfo: &funcInfoAPF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAPF,
			RegInfo: provapi.RegistrationInformation{
				ApiProvPubKey: "APF-PublicKey",
			},
		},
		{
			ApiProvFuncInfo: &funcInfoAEF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAEF,
			RegInfo: provapi.RegistrationInformation{
				ApiProvPubKey: "AEF-PublicKey",
			},
		},
	}
	return provapi.APIProviderEnrolmentDetails{
		RegSec:         "sec",
		ApiProvDomInfo: &domainInfo,
		ApiProvFuncs:   &testFuncs,
	}
}

func teardown() error {
	log.Trace("entering teardown")

	t := new(testing.T) // Create a new testing.T instance for teardown

	// Delete the invoker
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	result := testutil.NewRequest().Delete("/api-invoker-management/v1/onboardedInvokers/"+invokerId).Go(t, requestHandler)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the original published service
	apfId := "APF_id_rApp_Kong_as_APF"
	apiName := "apiName"
	apiId := "api_id_" + apiName

	result = testutil.NewRequest().Delete("/published-apis/v1/"+apfId+"/service-apis/"+apiId).Go(t, requestHandler)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the first published service
	apfId = "APF_id_rApp_Kong_as_APF"
	apiName = "apiName1"
	apiId = "api_id_" + apiName

	result = testutil.NewRequest().Delete("/published-apis/v1/"+apfId+"/service-apis/"+apiId).Go(t, requestHandler)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the second published service
	apiName = "apiName2"
	apiId = "api_id_" + apiName

	result = testutil.NewRequest().Delete("/published-apis/v1/"+apfId+"/service-apis/"+apiId).Go(t, requestHandler)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the provider
	domainID := "domain_id_Kong"
	result = testutil.NewRequest().Delete("/api-provider-management/v1/registrations/"+domainID).Go(t, requestHandler)
	assert.Equal(t, http.StatusNoContent, result.Code())

	myEnv, myPorts, err := envreader.ReadDotEnv()
	if err != nil {
		log.Fatal("error loading environment file")
		return err
	}

	err = kongclear.KongClear(myEnv, myPorts)
	if err != nil {
		log.Fatal("error clearing Kong on teardown")
	}
	return err
}

func TestRegisterValidProvider(t *testing.T) {
	teardown()

	newProvider := getProvider()

	// Register a valid provider
	result := testutil.NewRequest().Post("/api-provider-management/v1/registrations").WithJsonBody(newProvider).Go(t, requestHandler)
	assert.Equal(t, http.StatusCreated, result.Code())

	var resultProvider provapi.APIProviderEnrolmentDetails
	err := result.UnmarshalBodyToObject(&resultProvider)
	assert.NoError(t, err, "error unmarshaling response")
}

func TestPublishUnpublishService(t *testing.T) {
	apfId := "APF_id_rApp_Kong_as_APF"
	apiName := "apiName"
	newApiId := "api_id_" + apiName

	myEnv, myPorts, err := envreader.ReadDotEnv()
	assert.Nil(t, err, "error reading env file")

	testServiceIpv4 := common29122.Ipv4Addr(myEnv["TEST_SERVICE_IPV4"])
	testServicePort := common29122.Port(myPorts["TEST_SERVICE_PORT"])

	assert.NotEmpty(t, testServiceIpv4, "TEST_SERVICE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, testServicePort, "TEST_SERVICE_PORT is required in .env file for unit testing")

	// Check no services published
	result := testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis").Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	// Parse JSON from the response body
	var resultServices []publishapi.ServiceAPIDescription
	err = result.UnmarshalJsonToObject(&resultServices)
	assert.NoError(t, err, "error unmarshaling response")

	// Check if the parsed array is empty
	assert.Zero(t, len(resultServices))

	aefId := "AEF_id_rApp_Kong_as_AEF"
	namespace := "namespace"
	repoName := "repoName"
	chartName := "chartName"
	releaseName := "releaseName"
	description := fmt.Sprintf("Description,%s,%s,%s,%s", namespace, repoName, chartName, releaseName)

	newServiceDescription := getServiceAPIDescription(aefId, apiName, description, testServiceIpv4, testServicePort)

	// Publish a service for provider
	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)
	assert.Equal(t, http.StatusCreated, result.Code())

	if result.Code() != http.StatusCreated {
		log.Fatalf("failed to publish the service with HTTP result code %d", result.Code())
		return
	}

	var resultService publishapi.ServiceAPIDescription
	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newApiId, *resultService.ApiId)

	assert.Equal(t, "http://example.com/published-apis/v1/"+apfId+"/service-apis/"+*resultService.ApiId, result.Recorder.Header().Get(echo.HeaderLocation))

	// Check that the service is published for the provider
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newApiId, *resultService.ApiId)

	aefProfile := (*resultService.AefProfiles)[0]
	interfaceDescription := (*aefProfile.InterfaceDescriptions)[0]

	resultServiceIpv4 := *interfaceDescription.Ipv4Addr
	resultServicePort := *interfaceDescription.Port

	kongIPv4 := common29122.Ipv4Addr(myEnv["KONG_IPV4"])
	kongDataPlanePort := common29122.Port(myPorts["KONG_DATA_PLANE_PORT"])

	assert.NotEmpty(t, kongIPv4, "KONG_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, kongDataPlanePort, "KONG_DATA_PLANE_PORT is required in .env file for unit testing")

	assert.Equal(t, kongIPv4, resultServiceIpv4)
	assert.Equal(t, kongDataPlanePort, resultServicePort)
}

func TestOnboardInvoker(t *testing.T) {
	invokerInfo := "invoker a"
	newInvoker := getInvoker(invokerInfo)

	// Onboard a valid invoker
	result := testutil.NewRequest().Post("/api-invoker-management/v1/onboardedInvokers").WithJsonBody(newInvoker).Go(t, requestHandler)
	assert.Equal(t, http.StatusCreated, result.Code())

	var resultInvoker invokermanagementapi.APIInvokerEnrolmentDetails

	err := result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")

	wantedInvokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	assert.Equal(t, wantedInvokerId, *resultInvoker.ApiInvokerId)
	assert.Equal(t, newInvoker.NotificationDestination, resultInvoker.NotificationDestination)
	assert.Equal(t, newInvoker.OnboardingInformation.ApiInvokerPublicKey, resultInvoker.OnboardingInformation.ApiInvokerPublicKey)

	assert.Equal(t, "http://example.com/api-invoker-management/v1/onboardedInvokers/"+*resultInvoker.ApiInvokerId, result.Recorder.Header().Get(echo.HeaderLocation))

	// Onboarding the same invoker should result in Forbidden
	result = testutil.NewRequest().Post("/api-invoker-management/v1/onboardedInvokers").WithJsonBody(newInvoker).Go(t, requestHandler)

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
	result = testutil.NewRequest().Post("/api-invoker-management/v1/onboardedInvokers").WithJsonBody(invalidInvoker).Go(t, requestHandler)
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

	result = testutil.NewRequest().Post("/api-invoker-management/v1/onboardedInvokers").WithJsonBody(invalidInvoker).Go(t, requestHandler)
	assert.Equal(t, http.StatusBadRequest, result.Code())

	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "missing")
	assert.Contains(t, *problemDetails.Cause, "OnboardingInformation.ApiInvokerPublicKey")
}

func TestDeleteInvoker(t *testing.T) {
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	// Delete the invoker
	result := testutil.NewRequest().Delete("/api-invoker-management/v1/onboardedInvokers/"+invokerId).Go(t, requestHandler)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func TestUpdateInvoker(t *testing.T) {
	invokerInfo := "invoker a"
	invoker := getInvoker(invokerInfo)
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	// Onboard a valid invoker
	result := testutil.NewRequest().Post("/api-invoker-management/v1/onboardedInvokers").WithJsonBody(invoker).Go(t, requestHandler)
	assert.Equal(t, http.StatusCreated, result.Code())

	// Update the invoker with valid invoker, should return 200 with updated invoker details
	newNotifURL := "http://golang.org/"
	invoker.NotificationDestination = common29122.Uri(newNotifURL)
	newPublicKey := "newPublicKey"
	invoker.OnboardingInformation.ApiInvokerPublicKey = newPublicKey

	invoker.ApiInvokerId = &invokerId

	result = testutil.NewRequest().Put("/api-invoker-management/v1/onboardedInvokers/"+invokerId).WithJsonBody(invoker).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultInvoker invokermanagementapi.APIInvokerEnrolmentDetails

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
	result = testutil.NewRequest().Put("/api-invoker-management/v1/onboardedInvokers/"+invokerId).WithJsonBody(invalidInvoker).Go(t, requestHandler)
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
	result = testutil.NewRequest().Put("/api-invoker-management/v1/onboardedInvokers/"+invokerId).WithJsonBody(invalidInvoker).Go(t, requestHandler)
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
	result = testutil.NewRequest().Put("/api-invoker-management/v1/onboardedInvokers/"+invokerId).WithJsonBody(invalidInvoker).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())

	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "not matching")
	assert.Contains(t, *problemDetails.Cause, "ApiInvokerId")

	// Update an invoker that has not been onboarded, should get 404 with problem details
	missingId := "1"
	invoker.ApiInvokerId = &missingId
	result = testutil.NewRequest().Put("/api-invoker-management/v1/onboardedInvokers/"+missingId).WithJsonBody(invoker).Go(t, requestHandler)
	assert.Equal(t, http.StatusNotFound, result.Code())

	err = result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Equal(t, http.StatusNotFound, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "not been onboarded")
	assert.Contains(t, *problemDetails.Cause, "invoker")
	teardown()
}

func getEcho(myEnv map[string]string, myPorts map[string]int) (*echo.Echo, error) {
	capifProtocol := myEnv["CAPIF_PROTOCOL"]
	capifIPv4 := common29122.Ipv4Addr(myEnv["CAPIF_IPV4"])
	capifPort := common29122.Port(myPorts["CAPIF_PORT"])
	kongDomain := myEnv["KONG_DOMAIN"]
	kongProtocol := myEnv["KONG_PROTOCOL"]
	kongIPv4 := common29122.Ipv4Addr(myEnv["KONG_IPV4"])
	kongDataPlanePort := common29122.Port(myPorts["KONG_DATA_PLANE_PORT"])
	kongControlPlanePort := common29122.Port(myPorts["KONG_CONTROL_PLANE_PORT"])

	e := echo.New()

	// Register ProviderManagement
	providerManagerSwagger, err := provapi.GetSwagger()
	if err != nil {
		log.Fatalf("error loading ProviderManagement swagger spec\n: %s", err)
		return nil, err
	}
	providerManagerSwagger.Servers = nil
	providerManager := providermanagement.NewProviderManager(capifProtocol, capifIPv4, capifPort)

	var group *echo.Group

	group = e.Group("/api-provider-management/v1")
	group.Use(middleware.OapiRequestValidator(providerManagerSwagger))
	provapi.RegisterHandlersWithBaseURL(e, providerManager, "/api-provider-management/v1")

	publishServiceSwagger, err := publishapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading PublishService swagger spec\n: %s", err)
		return nil, err
	}

	publishServiceSwagger.Servers = nil

	ps := publishservice.NewPublishService(kongDomain, kongProtocol, kongIPv4, kongDataPlanePort, kongControlPlanePort, capifProtocol, capifIPv4, capifPort)

	group = e.Group("/published-apis/v1")
	group.Use(echomiddleware.Logger())
	group.Use(middleware.OapiRequestValidator(publishServiceSwagger))
	publishapi.RegisterHandlersWithBaseURL(e, ps, "/published-apis/v1")

	invokerServiceSwagger, err := invokermanagementapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading InvokerManagement swagger spec\n: %s", err)
		return nil, err
	}

	invokerServiceSwagger.Servers = nil

	im := NewInvokerManager(capifProtocol, capifIPv4, capifPort)

	group = e.Group("/api-invoker-management/v1")
	group.Use(echomiddleware.Logger())
	group.Use(middleware.OapiRequestValidator(invokerServiceSwagger))
	invokermanagementapi.RegisterHandlersWithBaseURL(e, im, "api-invoker-management/v1")

	return e, err
}

func getServiceAPIDescription(aefId, apiName, description string, testServiceIpv4 common29122.Ipv4Addr, testServicePort common29122.Port) publishapi.ServiceAPIDescription {
	domainName := "Kong"
	var protocol publishapi.Protocol = "HTTP_1_1"

	return publishapi.ServiceAPIDescription{
		AefProfiles: &[]publishapi.AefProfile{
			{
				AefId: aefId,
				InterfaceDescriptions: &[]publishapi.InterfaceDescription{
					{
						Ipv4Addr: &testServiceIpv4,
						Port:     &testServicePort,
						SecurityMethods: &[]publishapi.SecurityMethod{
							"PKI",
						},
					},
				},
				DomainName: &domainName,
				Protocol:   &protocol,
				Versions: []publishapi.Version{
					{
						ApiVersion: "v1",
						Resources: &[]publishapi.Resource{
							{
								CommType: "REQUEST_RESPONSE",
								Operations: &[]publishapi.Operation{
									"GET",
								},
								ResourceName: "helloworld",
								Uri:          "/helloworld",
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
