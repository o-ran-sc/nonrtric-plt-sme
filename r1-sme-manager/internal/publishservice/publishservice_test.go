// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023-2024: OpenInfra Foundation Europe
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
	"strings"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"oransc.org/nonrtric/r1-sme-manager/internal/common29122"
	"oransc.org/nonrtric/r1-sme-manager/internal/envreader"
	"oransc.org/nonrtric/r1-sme-manager/internal/kongclear"
	"oransc.org/nonrtric/r1-sme-manager/internal/providermanagement"
	provapi "oransc.org/nonrtric/r1-sme-manager/internal/providermanagementapi"
	publishapi "oransc.org/nonrtric/r1-sme-manager/internal/publishserviceapi"
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
		log.Fatal("Error loading environment file on setupTest")
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
	log.Info("Entering teardown")

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
		log.Fatal("Error loading environment file")
		return err
	}

	err = kongclear.KongClear(myEnv, myPorts)
	if err != nil {
		log.Fatal("Error clearing Kong on teardown")
	}
	return err
}

func TestPostUnpublishedServiceWithUnregisteredPublisher(t *testing.T) {
	teardown()

	apfId := "APF_id_rApp_Kong_as_APF"
	apiName := "apiName"

	// Check no services published
	result := testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis").Go(t, requestHandler)
	assert.Equal(t, http.StatusNotFound, result.Code())

	var resultError common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&resultError)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Contains(t, *resultError.Cause, apfId)
	assert.Contains(t, *resultError.Cause, "api is only available for publishers")

	aefId := "AEF_id_rApp_Kong_as_AEF"
	namespace := "namespace"
	repoName := "repoName"
	chartName := "chartName"
	releaseName := "releaseName"
	description := fmt.Sprintf("Description,%s,%s,%s,%s", namespace, repoName, chartName, releaseName)

	myEnv, myPorts, err := envreader.ReadDotEnv()
	assert.Nil(t, err, "error reading env file")

	testServiceIpv4 := common29122.Ipv4Addr(myEnv["TEST_SERVICE_IPV4"])
	testServicePort := common29122.Port(myPorts["TEST_SERVICE_PORT"])

	assert.NotEmpty(t, testServiceIpv4, "TEST_SERVICE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, testServicePort, "TEST_SERVICE_PORT is required in .env file for unit testing")

	newServiceDescription := getServiceAPIDescription(aefId, apiName, description, testServiceIpv4, testServicePort)

	// Attempt to publish a service for provider
	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)
	assert.Equal(t, http.StatusForbidden, result.Code())

	// var resultError common29122.ProblemDetails
	err = result.UnmarshalBodyToObject(&resultError)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Contains(t, *resultError.Cause, apfId)
	assert.Contains(t, *resultError.Cause, "Unable to publish the service due to api is only available for publishers")
}

func TestRegisterValidProvider(t *testing.T) {
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

	newServiceDescription := getServiceAPIDescriptionMissingInterface(aefId, apiName, description)

	// Publish a service for provider
	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)
	assert.Equal(t, http.StatusBadRequest, result.Code())

	var resultError common29122.ProblemDetails
	err = result.UnmarshalJsonToObject(&resultError)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Contains(t, *resultError.Cause, "cannot read interfaceDescription")

	newServiceDescription = getServiceAPIDescription(aefId, apiName, description, testServiceIpv4, testServicePort)

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

	// Publish the same service again should result in Forbidden
	newServiceDescription.ApiId = &newApiId
	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)
	assert.Equal(t, http.StatusForbidden, result.Code())

	err = result.UnmarshalBodyToObject(&resultError)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Contains(t, *resultError.Cause, "already published")
	assert.Equal(t, http.StatusForbidden, *resultError.Status)

	// Delete the service
	result = testutil.NewRequest().Delete("/published-apis/v1/"+apfId+"/service-apis/"+newApiId).Go(t, requestHandler)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Check no services published
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis").Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	// Parse JSON from the response body
	err = result.UnmarshalJsonToObject(&resultServices)
	assert.NoError(t, err, "error unmarshaling response")

	// Check if the parsed array is empty
	assert.Zero(t, len(resultServices))
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
		log.Fatalf("Error loading ProviderManagement swagger spec\n: %s", err)
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

	ps := NewPublishService(kongDomain, kongProtocol, kongIPv4, kongDataPlanePort, kongControlPlanePort, capifProtocol, capifIPv4, capifPort)

	group = e.Group("/published-apis/v1")
	group.Use(echomiddleware.Logger())
	group.Use(middleware.OapiRequestValidator(publishServiceSwagger))
	publishapi.RegisterHandlersWithBaseURL(e, ps, "/published-apis/v1")

	// return ps, eventChannel, e
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

func getServiceAPIDescriptionMissingInterface(aefId, apiName, description string) publishapi.ServiceAPIDescription {
	domainName := "Kong"
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
