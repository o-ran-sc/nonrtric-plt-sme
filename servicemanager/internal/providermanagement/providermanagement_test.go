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

package providermanagement

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"

	"oransc.org/nonrtric/servicemanager/internal/common29122"
	"oransc.org/nonrtric/servicemanager/internal/envreader"
	"oransc.org/nonrtric/servicemanager/internal/kongclear"

	provapi "oransc.org/nonrtric/servicemanager/internal/providermanagementapi"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"

	"oransc.org/nonrtric/capifcore"
	"oransc.org/nonrtric/servicemanager/mockkong"
)

var (
	domainID    = "domain_id_Kong"
	domainInfo  = "Kong"
	funcInfoAPF = "rApp Kong as APF"
	funcIdAPF   = "APF_id_rApp_Kong_as_APF"
	funcInfoAEF = "rApp Kong as AEF"
	funcIdAEF   = "AEF_id_rApp_Kong_as_AEF"
)

var (
	eServiceManager  	 *echo.Echo
	eCapifWeb        	 *echo.Echo
	eKong            	 *echo.Echo
	mockConfigReader 	 *envreader.MockConfigReader
	serviceManagerServer *httptest.Server
	capifServer 		 *httptest.Server
	mockKongServer 		 *httptest.Server
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
			"KONG_CONTROL_PLANE_IPV4":  mockKongHost,
			"KONG_CONTROL_PLANE_PORT":  mockKongControlPlanePort,
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

	myEnv, myPorts, err := mockConfigReader.ReadDotEnv()
	if err != nil {
		log.Fatal("error loading environment file on setupTest")
		return err
	}

	eServiceManager = echo.New()
	err = registerHandlers(eServiceManager, myEnv, myPorts)
	if err != nil {
		log.Fatal("registerHandlers fatal error on setupTest")
		return err
	}
	serviceManagerServer = httptest.NewServer(eServiceManager)

	return err
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

func TestRegisterValidProvider(t *testing.T) {
	capifCleanUp()

	// Register a valid provider
	newProvider := getProvider()
	result := testutil.NewRequest().Post("/api-provider-management/v1/registrations").WithJsonBody(newProvider).Go(t, eServiceManager)
	assert.Equal(t, http.StatusCreated, result.Code())

	var resultProvider provapi.APIProviderEnrolmentDetails
	err := result.UnmarshalBodyToObject(&resultProvider)
	assert.NoError(t, err, "error unmarshaling response")

	assert.NotNil(t, resultProvider.ApiProvDomId, "error reading resultProvider")

	if resultProvider.ApiProvDomId != nil {
		assert.Equal(t, *resultProvider.ApiProvDomId, domainID)

		apiProvFuncAPF := (*resultProvider.ApiProvFuncs)[0]
		apiProvFuncIdAPF := *apiProvFuncAPF.ApiProvFuncId
		assert.Equal(t, apiProvFuncIdAPF, funcIdAPF)

		// We don't handle AMF
		apiProvFuncAEF := (*resultProvider.ApiProvFuncs)[1]
		apiProvFuncIdAEF := *apiProvFuncAEF.ApiProvFuncId
		assert.Equal(t, apiProvFuncIdAEF, funcIdAEF)

		assert.Empty(t, resultProvider.FailReason)
		assert.Equal(t, "http://example.com/api-provider-management/v1/registrations/"+*resultProvider.ApiProvDomId, result.Recorder.Header().Get(echo.HeaderLocation))

		// Register same provider again should result in Forbidden
		result = testutil.NewRequest().Post("/api-provider-management/v1/registrations").WithJsonBody(newProvider).Go(t, eServiceManager)

		assert.Equal(t, http.StatusForbidden, result.Code())

		var errorObj common29122.ProblemDetails
		err = result.UnmarshalBodyToObject(&errorObj)
		assert.NoError(t, err, "error unmarshaling response")
		assert.Equal(t, http.StatusForbidden, *errorObj.Status)
		assert.Contains(t, *errorObj.Cause, "already registered")
	}
}

func TestUpdateValidProviderWithNewFunction(t *testing.T) {
	// Modify the provider
	updatedProvider := getProvider()
	updatedProvider.ApiProvDomId = &domainID

	newDomainInfo := "New domain info"
	updatedProvider.ApiProvDomInfo = &newDomainInfo
	newFunctionInfo := "New function info"
	(*updatedProvider.ApiProvFuncs)[0].ApiProvFuncInfo = &newFunctionInfo
	newFuncInfoAEF := "new func as AEF"

	testFuncs := append(*updatedProvider.ApiProvFuncs, provapi.APIProviderFunctionDetails{
		ApiProvFuncInfo: &newFuncInfoAEF,
		ApiProvFuncRole: provapi.ApiProviderFuncRoleAEF,
		RegInfo: provapi.RegistrationInformation{
			ApiProvPubKey: "key",
		},
	})
	updatedProvider.ApiProvFuncs = &testFuncs

	result := testutil.NewRequest().Put("/api-provider-management/v1/registrations/"+domainID).WithJsonBody(updatedProvider).Go(t, eServiceManager)

	var resultProvider provapi.APIProviderEnrolmentDetails
	assert.Equal(t, http.StatusOK, result.Code())

	err := result.UnmarshalBodyToObject(&resultProvider)
	assert.NoError(t, err, "error unmarshaling response")

	assert.Equal(t, newDomainInfo, *resultProvider.ApiProvDomInfo)
	assert.Equal(t, newFunctionInfo, *(*resultProvider.ApiProvFuncs)[0].ApiProvFuncInfo)
	assert.Equal(t, *(*resultProvider.ApiProvFuncs)[2].ApiProvFuncId, "AEF_id_new_func_as_AEF")
	assert.Empty(t, resultProvider.FailReason)
}

func TestDeleteProvider(t *testing.T) {
	provider := getProvider()
	provider.ApiProvDomId = &domainID
	(*provider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	result := testutil.NewRequest().Delete("/api-provider-management/v1/registrations/"+domainID).Go(t, eServiceManager)
	assert.Equal(t, http.StatusNoContent, result.Code())
	capifCleanUp()
}

func getProvider() provapi.APIProviderEnrolmentDetails {
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

func registerHandlers(e *echo.Echo, myEnv map[string]string, myPorts map[string]int) (err error) {
	capifProtocol := myEnv["CAPIF_PROTOCOL"]
	capifIPv4 := common29122.Ipv4Addr(myEnv["CAPIF_IPV4"])
	capifPort := common29122.Port(myPorts["CAPIF_PORT"])

	// Register ProviderManagement
	providerManagerSwagger, err := provapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading ProviderManagement swagger spec\n: %s", err)
		return err
	}
	providerManagerSwagger.Servers = nil
	providerManager := NewProviderManager(capifProtocol, capifIPv4, capifPort)

	group := e.Group("/api-provider-management/v1")
	group.Use(echomiddleware.Logger())
	group.Use(middleware.OapiRequestValidator(providerManagerSwagger))
	provapi.RegisterHandlersWithBaseURL(e, providerManager, "/api-provider-management/v1")

	return err
}
