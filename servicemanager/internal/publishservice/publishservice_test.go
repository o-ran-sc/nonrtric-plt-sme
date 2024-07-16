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
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"oransc.org/nonrtric/servicemanager/internal/envreader"
	"oransc.org/nonrtric/servicemanager/internal/kongclear"

	"oransc.org/nonrtric/servicemanager/internal/common29122"
	"oransc.org/nonrtric/servicemanager/internal/providermanagement"
	provapi "oransc.org/nonrtric/servicemanager/internal/providermanagementapi"
	publishapi "oransc.org/nonrtric/servicemanager/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore"
	mockKong "oransc.org/nonrtric/servicemanager/mockkong"
)

var (
	eServiceManager      *echo.Echo
	eCapifWeb            *echo.Echo
	eKong                *echo.Echo
	mockConfigReader     *envreader.MockConfigReader
	serviceManagerServer *httptest.Server
	capifServer          *httptest.Server
	mockKongServer       *httptest.Server
)

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
	// Start the mock Kong server
	eKong = echo.New()
	mockKong.RegisterHandlers(eKong)
	mockKongServer = httptest.NewServer(eKong)

	// Parse the server URL
	parsedMockKongURL, err := url.Parse(mockKongServer.URL)
	if err != nil {
		log.Fatalf("error parsing mock Kong URL: %v", err)
		return err
	}

	// Extract the host and port
	mockKongHost := parsedMockKongURL.Hostname()
	mockKongControlPlanePort := parsedMockKongURL.Port()

	eCapifWeb = echo.New()
	capifcore.RegisterHandlers(eCapifWeb, nil, nil)
	capifServer = httptest.NewServer(eCapifWeb)

	// Parse the server URL
	parsedCapifURL, err := url.Parse(capifServer.URL)
	if err != nil {
		log.Fatalf("error parsing mock Kong URL: %v", err)
		return err
	}

	// Extract the host and port
	capifHost := parsedCapifURL.Hostname()
	capifPort := parsedCapifURL.Port()

	// Set up the mock config reader with the desired configuration for testing
	mockConfigReader = &envreader.MockConfigReader{
		MockedConfig: map[string]string{
			"KONG_DOMAIN":             "kong",
			"KONG_PROTOCOL":           "http",
			"KONG_CONTROL_PLANE_IPV4": mockKongHost,
			"KONG_CONTROL_PLANE_PORT": mockKongControlPlanePort,
			"KONG_DATA_PLANE_IPV4":    "10.101.1.101",
			"KONG_DATA_PLANE_PORT":    "32080",
			"CAPIF_PROTOCOL":          "http",
			"CAPIF_IPV4":              capifHost,
			"CAPIF_PORT":              capifPort,
			"LOG_LEVEL":               "Info",
			"SERVICE_MANAGER_PORT":    "8095",
			"TEST_SERVICE_IPV4":       "10.101.1.101",
			"TEST_SERVICE_PORT":       "30951",
		},
	}

	// Use the mock implementation for testing
	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	if err != nil {
		log.Fatalf("error reading mock config: %v", err)
	}

	eServiceManager = echo.New()
	err = registerHandlers(eServiceManager, myEnv, myPorts)
	if err != nil {
		log.Fatal("registerHandlers fatal error on setupTest")
		return err
	}
	serviceManagerServer = httptest.NewServer(eServiceManager)
	capifCleanUp()
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

func capifCleanUp() {
	t := new(testing.T) // Create a new testing.T instance for capifCleanUp

	// Delete the invoker
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	result := testutil.NewRequest().Delete("/api-invoker-management/v1/onboardedInvokers/"+invokerId).Go(t, eServiceManager)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the original published service
	apfId := "APF_id_rApp_Kong_as_APF"
	apiName := "apiName"
	apiId := "api_id_" + apiName

	result = testutil.NewRequest().Delete("/published-apis/v1/"+apfId+"/service-apis/"+apiId).Go(t, eServiceManager)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the first published service
	apfId = "APF_id_rApp_Kong_as_APF"
	apiName = "apiName1"
	apiId = "api_id_" + apiName

	result = testutil.NewRequest().Delete("/published-apis/v1/"+apfId+"/service-apis/"+apiId).Go(t, eServiceManager)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the second published service
	apiName = "apiName2"
	apiId = "api_id_" + apiName

	result = testutil.NewRequest().Delete("/published-apis/v1/"+apfId+"/service-apis/"+apiId).Go(t, eServiceManager)
	assert.Equal(t, http.StatusNoContent, result.Code())

	// Delete the provider
	domainID := "domain_id_Kong"
	result = testutil.NewRequest().Delete("/api-provider-management/v1/registrations/"+domainID).Go(t, eServiceManager)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func teardown() error {
	log.Trace("entering teardown")

	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	if err != nil {
		log.Fatal("error loading environment file")
		return err
	}

	err = kongclear.KongClear(myEnv, myPorts)
	if err != nil {
		log.Fatal("error clearing Kong on teardown")
	}

	mockKongServer.Close()
	capifServer.Close()
	serviceManagerServer.Close()

	return nil
}

func TestPostUnpublishedServiceWithUnregisteredPublisher(t *testing.T) {
	capifCleanUp()

	apfId := "APF_id_rApp_Kong_as_APF"

	// Check no services published
	result := testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis").Go(t, eServiceManager)
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

	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	assert.Nil(t, err, "error reading env file")

	testServiceIpv4 := common29122.Ipv4Addr(myEnv["TEST_SERVICE_IPV4"])
	testServicePort := common29122.Port(myPorts["TEST_SERVICE_PORT"])

	assert.NotEmpty(t, testServiceIpv4, "TEST_SERVICE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, testServicePort, "TEST_SERVICE_PORT is required in .env file for unit testing")

	apiName := "apiName"
	apiVersion := "v1"
	resourceName := "helloworld"
	uri := "/helloworld"

	newServiceDescription := getServiceAPIDescription(aefId, apiName, description, testServiceIpv4, testServicePort, apiVersion, resourceName, uri)

	// Attempt to publish a service for provider
	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, eServiceManager)
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
	result := testutil.NewRequest().Post("/api-provider-management/v1/registrations").WithJsonBody(newProvider).Go(t, eServiceManager)
	assert.Equal(t, http.StatusCreated, result.Code())

	var resultProvider provapi.APIProviderEnrolmentDetails
	err := result.UnmarshalBodyToObject(&resultProvider)
	assert.NoError(t, err, "error unmarshaling response")
}

func TestPublishUnpublishServiceMissingInterface(t *testing.T) {
	apfId := "APF_id_rApp_Kong_as_APF"
	apiName := "apiName"

	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	assert.Nil(t, err, "error reading env file")

	testServiceIpv4 := common29122.Ipv4Addr(myEnv["TEST_SERVICE_IPV4"])
	testServicePort := common29122.Port(myPorts["TEST_SERVICE_PORT"])

	assert.NotEmpty(t, testServiceIpv4, "TEST_SERVICE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, testServicePort, "TEST_SERVICE_PORT is required in .env file for unit testing")

	// Check no services published
	result := testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis").Go(t, eServiceManager)
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
	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, eServiceManager)
	assert.Equal(t, http.StatusBadRequest, result.Code())

	var resultError common29122.ProblemDetails
	err = result.UnmarshalJsonToObject(&resultError)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Contains(t, *resultError.Cause, "cannot read interfaceDescription")
}


func TestPublishUnpublishWithoutVersionId(t *testing.T) {
	apfId := "APF_id_rApp_Kong_as_APF"

	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	assert.Nil(t, err, "error reading env file")

	testServiceIpv4 := common29122.Ipv4Addr(myEnv["TEST_SERVICE_IPV4"])
	testServicePort := common29122.Port(myPorts["TEST_SERVICE_PORT"])

	assert.NotEmpty(t, testServiceIpv4, "TEST_SERVICE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, testServicePort, "TEST_SERVICE_PORT is required in .env file for unit testing")

	apiVersion := "v1"
	resourceName := "helloworld"
	uri := "/helloworld"
	apiName := "helloworld-v1"

	aefId := "AEF_id_rApp_Kong_as_AEF"
	namespace := "namespace"
	repoName := "repoName"
	chartName := "chartName"
	releaseName := "releaseName"
	description := fmt.Sprintf("Description,%s,%s,%s,%s", namespace, repoName, chartName, releaseName)

	newServiceDescription := getServiceAPIDescription(aefId, apiName, description, testServiceIpv4, testServicePort, apiVersion, resourceName, uri)

	// Publish a service for provider
	result := testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, eServiceManager)
	assert.Equal(t, http.StatusCreated, result.Code())

	if result.Code() != http.StatusCreated {
		log.Fatalf("failed to publish the service with HTTP result code %d", result.Code())
		return
	}

	var resultService publishapi.ServiceAPIDescription
	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	newApiId := "api_id_" + apiName
	assert.Equal(t, newApiId, *resultService.ApiId)

	assert.Equal(t, "http://example.com/published-apis/v1/"+apfId+"/service-apis/"+*resultService.ApiId, result.Recorder.Header().Get(echo.HeaderLocation))

	// Check that the service is published for the provider
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis/"+newApiId).Go(t, eServiceManager)
	assert.Equal(t, http.StatusOK, result.Code())

	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newApiId, *resultService.ApiId)

	aefProfile := (*resultService.AefProfiles)[0]
	interfaceDescription := (*aefProfile.InterfaceDescriptions)[0]

	resultServiceIpv4 := *interfaceDescription.Ipv4Addr
	resultServicePort := *interfaceDescription.Port

	kongDataPlaneIPv4 := common29122.Ipv4Addr(myEnv["KONG_DATA_PLANE_IPV4"])
	kongDataPlanePort := common29122.Port(myPorts["KONG_DATA_PLANE_PORT"])

	assert.NotEmpty(t, kongDataPlaneIPv4, "KONG_DATA_PLANE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, kongDataPlanePort, "KONG_DATA_PLANE_PORT is required in .env file for unit testing")

	assert.Equal(t, kongDataPlaneIPv4, resultServiceIpv4)
	assert.Equal(t, kongDataPlanePort, resultServicePort)

	// Check Versions structure
	version := aefProfile.Versions[0]
	assert.Equal(t, "v1", version.ApiVersion)

	resource := (*version.Resources)[0]
	communicationType := publishapi.CommunicationType("REQUEST_RESPONSE")
	assert.Equal(t, communicationType, resource.CommType)

	assert.Equal(t, 1, len(*resource.Operations))
	var operation publishapi.Operation = "GET"
	assert.Equal(t, operation, (*resource.Operations)[0])
	assert.Equal(t, "helloworld", resource.ResourceName)
	assert.Equal(t, "/helloworld-v1/helloworld", resource.Uri)
}

func TestPublishUnpublishVersionId(t *testing.T) {
	apfId := "APF_id_rApp_Kong_as_APF"

	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	assert.Nil(t, err, "error reading env file")

	testServiceIpv4 := common29122.Ipv4Addr(myEnv["TEST_SERVICE_IPV4"])
	testServicePort := common29122.Port(myPorts["TEST_SERVICE_PORT"])

	assert.NotEmpty(t, testServiceIpv4, "TEST_SERVICE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, testServicePort, "TEST_SERVICE_PORT is required in .env file for unit testing")

	apiVersion := "v1"
	resourceName := "helloworld-id"
	uri := "~/helloworld/(?<helloworld-id>[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*)"
	apiName := "helloworld-v1-id"

	aefId := "AEF_id_rApp_Kong_as_AEF"
	namespace := "namespace"
	repoName := "repoName"
	chartName := "chartName"
	releaseName := "releaseName"
	description := fmt.Sprintf("Description,%s,%s,%s,%s", namespace, repoName, chartName, releaseName)

	newServiceDescription := getServiceAPIDescription(aefId, apiName, description, testServiceIpv4, testServicePort, apiVersion, resourceName, uri)

	// Publish a service for provider
	result := testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, eServiceManager)
	assert.Equal(t, http.StatusCreated, result.Code())

	if result.Code() != http.StatusCreated {
		log.Fatalf("failed to publish the service with HTTP result code %d", result.Code())
		return
	}

	var resultService publishapi.ServiceAPIDescription
	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	newApiId := "api_id_" + apiName
	assert.Equal(t, newApiId, *resultService.ApiId)

	assert.Equal(t, "http://example.com/published-apis/v1/"+apfId+"/service-apis/"+*resultService.ApiId, result.Recorder.Header().Get(echo.HeaderLocation))

	// Check that the service is published for the provider
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis/"+newApiId).Go(t, eServiceManager)
	assert.Equal(t, http.StatusOK, result.Code())

	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newApiId, *resultService.ApiId)

	aefProfile := (*resultService.AefProfiles)[0]
	interfaceDescription := (*aefProfile.InterfaceDescriptions)[0]

	resultServiceIpv4 := *interfaceDescription.Ipv4Addr
	resultServicePort := *interfaceDescription.Port

	kongDataPlaneIPv4 := common29122.Ipv4Addr(myEnv["KONG_DATA_PLANE_IPV4"])
	kongDataPlanePort := common29122.Port(myPorts["KONG_DATA_PLANE_PORT"])

	assert.NotEmpty(t, kongDataPlaneIPv4, "KONG_DATA_PLANE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, kongDataPlanePort, "KONG_DATA_PLANE_PORT is required in .env file for unit testing")

	assert.Equal(t, kongDataPlaneIPv4, resultServiceIpv4)
	assert.Equal(t, kongDataPlanePort, resultServicePort)

	// Check Versions structure
	version := aefProfile.Versions[0]
	assert.Equal(t, "v1", version.ApiVersion)

	resource := (*version.Resources)[0]
	communicationType := publishapi.CommunicationType("REQUEST_RESPONSE")
	assert.Equal(t, communicationType, resource.CommType)

	assert.Equal(t, 1, len(*resource.Operations))
	var operation publishapi.Operation = "GET"
	assert.Equal(t, operation, (*resource.Operations)[0])

	assert.Equal(t, "helloworld-id", resource.ResourceName)
	assert.Equal(t, "~/helloworld-v1-id/helloworld/v1/(?<helloworld-id>[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*)", resource.Uri)
}

func TestPublishUnpublishServiceNoVersionWithId(t *testing.T) {
	apfId := "APF_id_rApp_Kong_as_APF"

	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	assert.Nil(t, err, "error reading env file")

	testServiceIpv4 := common29122.Ipv4Addr(myEnv["TEST_SERVICE_IPV4"])
	testServicePort := common29122.Port(myPorts["TEST_SERVICE_PORT"])

	assert.NotEmpty(t, testServiceIpv4, "TEST_SERVICE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, testServicePort, "TEST_SERVICE_PORT is required in .env file for unit testing")

	apiVersion := ""
	resourceName := "helloworld-no-version"
	uri := "~/helloworld/(?<helloworld-id>[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*)"
	apiName := "helloworld-no-version"

	aefId := "AEF_id_rApp_Kong_as_AEF"
	namespace := "namespace"
	repoName := "repoName"
	chartName := "chartName"
	releaseName := "releaseName"
	description := fmt.Sprintf("Description,%s,%s,%s,%s", namespace, repoName, chartName, releaseName)

	newServiceDescription := getServiceAPIDescription(aefId, apiName, description, testServiceIpv4, testServicePort, apiVersion, resourceName, uri)

	// Publish a service for provider
	result := testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, eServiceManager)
	assert.Equal(t, http.StatusCreated, result.Code())

	if result.Code() != http.StatusCreated {
		log.Fatalf("failed to publish the service with HTTP result code %d", result.Code())
		return
	}

	var resultService publishapi.ServiceAPIDescription
	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	newApiId := "api_id_" + apiName
	assert.Equal(t, newApiId, *resultService.ApiId)

	assert.Equal(t, "http://example.com/published-apis/v1/"+apfId+"/service-apis/"+*resultService.ApiId, result.Recorder.Header().Get(echo.HeaderLocation))

	// Check that the service is published for the provider
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis/"+newApiId).Go(t, eServiceManager)
	assert.Equal(t, http.StatusOK, result.Code())

	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newApiId, *resultService.ApiId)

	aefProfile := (*resultService.AefProfiles)[0]
	interfaceDescription := (*aefProfile.InterfaceDescriptions)[0]

	resultServiceIpv4 := *interfaceDescription.Ipv4Addr
	resultServicePort := *interfaceDescription.Port

	kongDataPlaneIPv4 := common29122.Ipv4Addr(myEnv["KONG_DATA_PLANE_IPV4"])
	kongDataPlanePort := common29122.Port(myPorts["KONG_DATA_PLANE_PORT"])

	assert.NotEmpty(t, kongDataPlaneIPv4, "KONG_DATA_PLANE_IPV4 is required in .env file for unit testing")
	assert.NotZero(t, kongDataPlanePort, "KONG_DATA_PLANE_PORT is required in .env file for unit testing")

	assert.Equal(t, kongDataPlaneIPv4, resultServiceIpv4)
	assert.Equal(t, kongDataPlanePort, resultServicePort)

	// Check Versions structure
	version := aefProfile.Versions[0]
	assert.Equal(t, "", version.ApiVersion)

	resource := (*version.Resources)[0]
	communicationType := publishapi.CommunicationType("REQUEST_RESPONSE")
	assert.Equal(t, communicationType, resource.CommType)

	assert.Equal(t, 1, len(*resource.Operations))
	var operation publishapi.Operation = "GET"
	assert.Equal(t, operation, (*resource.Operations)[0])

	assert.Equal(t, "helloworld-no-version", resource.ResourceName)
	assert.Equal(t, "~/helloworld-no-version/helloworld/(?<helloworld-id>[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*)", resource.Uri)

	capifCleanUp()
}

func registerHandlers(e *echo.Echo, myEnv map[string]string, myPorts map[string]int) (err error) {
	capifProtocol := myEnv["CAPIF_PROTOCOL"]
	capifIPv4 := common29122.Ipv4Addr(myEnv["CAPIF_IPV4"])
	capifPort := common29122.Port(myPorts["CAPIF_PORT"])
	kongDomain := myEnv["KONG_DOMAIN"]
	kongProtocol := myEnv["KONG_PROTOCOL"]
	kongControlPlaneIPv4 := common29122.Ipv4Addr(myEnv["KONG_CONTROL_PLANE_IPV4"])
	kongControlPlanePort := common29122.Port(myPorts["KONG_CONTROL_PLANE_PORT"])
	kongDataPlaneIPv4 := common29122.Ipv4Addr(myEnv["KONG_DATA_PLANE_IPV4"])
	kongDataPlanePort := common29122.Port(myPorts["KONG_DATA_PLANE_PORT"])

	// Register ProviderManagement
	providerManagerSwagger, err := provapi.GetSwagger()
	if err != nil {
		log.Fatalf("error loading ProviderManagement swagger spec\n: %s", err)
		return err
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
		return err
	}

	publishServiceSwagger.Servers = nil

	ps := NewPublishService(
		kongDomain, kongProtocol,
		kongControlPlaneIPv4, kongControlPlanePort,
		kongDataPlaneIPv4, kongDataPlanePort,
		capifProtocol, capifIPv4, capifPort)

	group = e.Group("/published-apis/v1")
	group.Use(echomiddleware.Logger())
	group.Use(middleware.OapiRequestValidator(publishServiceSwagger))
	publishapi.RegisterHandlersWithBaseURL(e, ps, "/published-apis/v1")

	return err
}

func getServiceAPIDescription(
	aefId string,
	apiName string,
	description string,
	testServiceIpv4 common29122.Ipv4Addr,
	testServicePort common29122.Port,
	apiVersion string,
	resourceName string,
	uri string) publishapi.ServiceAPIDescription {

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
						ApiVersion: apiVersion,
						Resources: &[]publishapi.Resource{
							{
								CommType: "REQUEST_RESPONSE",
								Operations: &[]publishapi.Operation{
									"GET",
								},
								ResourceName: resourceName,
								Uri:          uri,
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
