// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2024: OpenInfra Foundation Europe
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

package discoverservice

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"oransc.org/nonrtric/r1-sme-manager/internal/common29122"
	"oransc.org/nonrtric/r1-sme-manager/internal/discoverserviceapi"
	"oransc.org/nonrtric/r1-sme-manager/internal/invokermanagement"
	"oransc.org/nonrtric/r1-sme-manager/internal/invokermanagementapi"

	publishapi "oransc.org/nonrtric/r1-sme-manager/internal/publishserviceapi"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"

	"oransc.org/nonrtric/r1-sme-manager/internal/envreader"
	"oransc.org/nonrtric/r1-sme-manager/internal/kongclear"
	"oransc.org/nonrtric/r1-sme-manager/internal/providermanagement"

	provapi "oransc.org/nonrtric/r1-sme-manager/internal/providermanagementapi"
	"oransc.org/nonrtric/r1-sme-manager/internal/publishservice"
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
	apiName := "apiName1"
	apiId := "api_id_" + apiName

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
	assert.True (t, len(resultServices) == 0)

	aefId := "AEF_id_rApp_Kong_as_AEF"
	namespace := "namespace"
	repoName := "repoName"
	chartName := "chartName"
	releaseName := "releaseName"
	description := fmt.Sprintf("Description,%s,%s,%s,%s", namespace, repoName, chartName, releaseName)

	apiCategory := "apiCategory"
	apiVersion := "v1"
	var protocolHTTP11 = publishapi.ProtocolHTTP11
	var dataFormatJSON = publishapi.DataFormatJSON

	newServiceDescription := getServiceAPIDescription(aefId, apiName, apiCategory, apiVersion, &protocolHTTP11, &dataFormatJSON, description, testServiceIpv4, testServicePort, publishapi.CommunicationTypeREQUESTRESPONSE)

	// Publish a service for provider
	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription).Go(t, requestHandler)
	assert.Equal(t, http.StatusCreated, result.Code())

	if result.Code() != http.StatusCreated {
		log.Fatalf("failed to publish the service with HTTP result code %d", result.Code())
	}

	var resultService publishapi.ServiceAPIDescription
	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, apiId, *resultService.ApiId)

	assert.Equal(t, "http://example.com/published-apis/v1/"+apfId+"/service-apis/"+*resultService.ApiId, result.Recorder.Header().Get(echo.HeaderLocation))

	// Check that the service is published for the provider
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis/"+apiId).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, apiId, *resultService.ApiId)

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

	// Check one service published
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis").Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	// Parse JSON from the response body
	err = result.UnmarshalJsonToObject(&resultServices)
	assert.NoError(t, err, "error unmarshaling response")

	// Check if the parsed array has one item
	assert.True (t, len(resultServices) == 1)

	// Publish a second service for provider
	apiName2 := "apiName2"
	apiId2 := "api_id_" + apiName2
	apiVersion = "v2"
	apiCategory = ""
	protocolHTTP1 := publishapi.ProtocolHTTP2
	var dataFormatOther publishapi.DataFormat = "OTHER"

	newServiceDescription2 := getServiceAPIDescription(aefId, apiName2, apiCategory, apiVersion, &protocolHTTP1, &dataFormatOther, description, testServiceIpv4, testServicePort, publishapi.CommunicationTypeSUBSCRIBENOTIFY)

	result = testutil.NewRequest().Post("/published-apis/v1/"+apfId+"/service-apis").WithJsonBody(newServiceDescription2).Go(t, requestHandler)
	assert.Equal(t, http.StatusCreated, result.Code())

	if result.Code() != http.StatusCreated {
		log.Fatalf("failed to publish the service with HTTP result code %d", result.Code())
		return
	}

	err = result.UnmarshalJsonToObject(&resultService)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, apiId2, *resultService.ApiId)

	// Check no services published
	result = testutil.NewRequest().Get("/published-apis/v1/"+apfId+"/service-apis").Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	// Parse JSON from the response body
	err = result.UnmarshalJsonToObject(&resultServices)
	assert.NoError(t, err, "error unmarshaling response")

	// Check if the parsed array has two items
	assert.True (t, len(resultServices) == 2)
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

	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	assert.Equal(t, invokerId, *resultInvoker.ApiInvokerId)
	assert.Equal(t, newInvoker.NotificationDestination, resultInvoker.NotificationDestination)
	assert.Equal(t, newInvoker.OnboardingInformation.ApiInvokerPublicKey, resultInvoker.OnboardingInformation.ApiInvokerPublicKey)
	assert.Equal(t, "http://example.com/api-invoker-management/v1/onboardedInvokers/"+*resultInvoker.ApiInvokerId, result.Recorder.Header().Get(echo.HeaderLocation))
}

func TestGetAllServiceAPIs(t *testing.T) {
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)
	apiName1 := "apiName1"
	apiName2 := "apiName2"

	// Get all APIs, without any filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 2, len(*resultDiscovery.ServiceAPIDescriptions), "incorrect count of ServiceAPIDescriptions")
		if len(*resultDiscovery.ServiceAPIDescriptions) == 2 {
			// The order of the results is inconsistent.
			resultApiName1 := (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName
			resultApiName2 := (*resultDiscovery.ServiceAPIDescriptions)[1].ApiName
			resultApiNames := []string{resultApiName1, resultApiName2}
			sort.Strings(resultApiNames)
			expectedApiNames := []string{apiName1, apiName2}
			sort.Strings(expectedApiNames)
			assert.True(t, reflect.DeepEqual(resultApiNames, expectedApiNames))
		}

	}
}

func TestGetAllServiceAPIsWhenMissingProvider(t *testing.T) {
	invokerId := "unregistered"

	// Get all APIs, without any filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId).Go(t, requestHandler)
	assert.Equal(t, http.StatusNotFound, result.Code())

	var problemDetails common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")

	notFound := http.StatusNotFound
	assert.Equal(t, &notFound, problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, invokerId)
	assert.Contains(t, *problemDetails.Cause, "not registered")
}

func TestFilterApiName(t *testing.T) {
	apiName := "apiName1"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&api-name="+apiName).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 1, len(*resultDiscovery.ServiceAPIDescriptions))
		if len(*resultDiscovery.ServiceAPIDescriptions) == 1 {
			assert.Equal(t, apiName, (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName)
		}
	}
}

func TestFilterAefId(t *testing.T) {
	apiName1 := "apiName1"
	apiName2 := "apiName2"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	aefId := "AEF_id_rApp_Kong_as_AEF"

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&aef-id="+aefId).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 2, len(*resultDiscovery.ServiceAPIDescriptions), "incorrect count of ServiceAPIDescriptions")
		if len(*resultDiscovery.ServiceAPIDescriptions) == 2 {
			// The order of the results is inconsistent.
			resultApiName1 := (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName
			resultApiName2 := (*resultDiscovery.ServiceAPIDescriptions)[1].ApiName
			resultApiNames := []string{resultApiName1, resultApiName2}
			sort.Strings(resultApiNames)
			expectedApiNames := []string{apiName1, apiName2}
			sort.Strings(expectedApiNames)
			assert.True(t, reflect.DeepEqual(resultApiNames, expectedApiNames))
		}
	}
}

func TestFilterVersion(t *testing.T) {
	apiName := "apiName1"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)
	apiVersion := "v1"

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&api-version="+apiVersion).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)

	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 1, len(*resultDiscovery.ServiceAPIDescriptions))
		if len(*resultDiscovery.ServiceAPIDescriptions) == 1 {
			assert.Equal(t, apiName, (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName)
		}
	}
}

func TestFilterCommType(t *testing.T) {
	apiName := "apiName1"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	commType := publishapi.CommunicationTypeREQUESTRESPONSE

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&comm-type="+string(commType)).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 1, len(*resultDiscovery.ServiceAPIDescriptions))
		if len(*resultDiscovery.ServiceAPIDescriptions) == 1 {
			assert.Equal(t, apiName, (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName)
		}
	}
}

func TestFilterVersionAndCommType(t *testing.T) {
	apiName := "apiName1"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	apiVersion := "v1"
	commType := publishapi.CommunicationTypeREQUESTRESPONSE

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&api-version="+apiVersion+"&comm-type="+string(commType)).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 1, len(*resultDiscovery.ServiceAPIDescriptions))
		if len(*resultDiscovery.ServiceAPIDescriptions) == 1 {
			assert.Equal(t, apiName, (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName)
		}
	}
}

func TestFilterAPICategory(t *testing.T) {
	apiName := "apiName1"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	apiCategory := "apiCategory"

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&api-cat="+apiCategory).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 1, len(*resultDiscovery.ServiceAPIDescriptions))
		if len(*resultDiscovery.ServiceAPIDescriptions) == 1 {
			assert.Equal(t, apiName, (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName)
		}
	}
}

func TestFilterProtocol(t *testing.T) {
	apiName := "apiName1"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	var protocolHTTP11 = publishapi.ProtocolHTTP11

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&protocol="+string(protocolHTTP11)).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")

	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 1, len(*resultDiscovery.ServiceAPIDescriptions))
		if len(*resultDiscovery.ServiceAPIDescriptions) == 1 {
			assert.Equal(t, apiName, (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName)
		}
	}
}

func TestFilterDataFormat(t *testing.T) {
	apiName := "apiName1"
	invokerInfo := "invoker a"
	invokerId := "api_invoker_id_" + strings.Replace(invokerInfo, " ", "_", 1)

	var dataFormatJSON = publishapi.DataFormatJSON

	// Get APIs with filter
	result := testutil.NewRequest().Get("/service-apis/v1/allServiceAPIs?api-invoker-id="+invokerId+"&data-format="+string(dataFormatJSON)).Go(t, requestHandler)
	assert.Equal(t, http.StatusOK, result.Code())

	var resultDiscovery discoverserviceapi.DiscoveredAPIs
	err := result.UnmarshalBodyToObject(&resultDiscovery)

	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultDiscovery.ServiceAPIDescriptions, "error reading ServiceAPIDescriptions")
	if resultDiscovery.ServiceAPIDescriptions != nil {
		assert.Equal(t, 1, len(*resultDiscovery.ServiceAPIDescriptions))
		if len(*resultDiscovery.ServiceAPIDescriptions) == 1 {
			assert.Equal(t, apiName, (*resultDiscovery.ServiceAPIDescriptions)[0].ApiName)
		}
	}
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

	// Register PublishService
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

	// Register InvokerService
	invokerServiceSwagger, err := invokermanagementapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading InvokerManagement swagger spec\n: %s", err)
		return nil, err
	}

	invokerServiceSwagger.Servers = nil

	im := invokermanagement.NewInvokerManager(capifProtocol, capifIPv4, capifPort)

	group = e.Group("/api-invoker-management/v1")
	group.Use(echomiddleware.Logger())
	group.Use(middleware.OapiRequestValidator(invokerServiceSwagger))
	invokermanagementapi.RegisterHandlersWithBaseURL(e, im, "api-invoker-management/v1")

	// Register DiscoveryService
	discoverySeviceSwagger, err := discoverserviceapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	discoverySeviceSwagger.Servers = nil

	ds := NewDiscoverService(capifProtocol, capifIPv4, capifPort)

	group = e.Group("/service-apis/v1")
	group.Use(echomiddleware.Logger())
	group.Use(middleware.OapiRequestValidator(discoverySeviceSwagger))
	discoverserviceapi.RegisterHandlersWithBaseURL(e, ds, "service-apis/v1")

	return e, err
}

func getServiceAPIDescription(aefId string, apiName string, apiCategory string, apiVersion string, protocol *publishapi.Protocol, dataFormat *publishapi.DataFormat, description string, testServiceIpv4 common29122.Ipv4Addr, testServicePort common29122.Port, commType publishapi.CommunicationType) publishapi.ServiceAPIDescription {
	domainName := "Kong"
	otherDomainName := "otherDomain"

	var otherProtocol publishapi.Protocol = "HTTP_2"

	categoryPointer := &apiCategory
	if apiCategory == "" {
		categoryPointer = nil
	}

	var DataFormatOther publishapi.DataFormat = "OTHER"

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
				Protocol:   protocol,
				DataFormat: dataFormat,
				Versions: []publishapi.Version{
					{
						ApiVersion: apiVersion,
						Resources: &[]publishapi.Resource{
							{
								CommType: commType,
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
			{
				AefId:      aefId, // "otherAefId"
				DomainName: &otherDomainName,
				Protocol:   &otherProtocol,
				DataFormat: &DataFormatOther,
				Versions: []publishapi.Version{
					{
						ApiVersion: "v3",
						Resources: &[]publishapi.Resource{
							{
								ResourceName: "app",
								CommType:     publishapi.CommunicationTypeSUBSCRIBENOTIFY,
								Uri:          "uri",
								Operations: &[]publishapi.Operation{
									"POST",
								},
							},
						},
					},
				},
			},
		},
		ApiName:            apiName,
		Description:        &description,
		ServiceAPICategory: categoryPointer,
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
