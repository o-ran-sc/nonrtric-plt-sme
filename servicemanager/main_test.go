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

package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	log "github.com/sirupsen/logrus"

	"oransc.org/nonrtric/servicemanager/internal/common29122"
	"oransc.org/nonrtric/servicemanager/internal/envreader"
	"oransc.org/nonrtric/servicemanager/internal/kongclear"

	"oransc.org/nonrtric/capifcore"
	"oransc.org/nonrtric/servicemanager/mockkong"
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

// Init code to run before tests
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
			"KONG_IPV4":               mockKongHost,
			"KONG_DATA_PLANE_PORT":    "32080",
			"KONG_CONTROL_PLANE_PORT": mockKongControlPlanePort,
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

	return err
}


func Test_routing(t *testing.T) {
	type args struct {
		url          string
		returnStatus int
		method       string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Default path",
			args: args{
				url:          "/",
				returnStatus: http.StatusOK,
				method:       "GET",
			},
		},
		{
			name: "Provider path",
			args: args{
				url:          "/api-provider-management/v1/registrations/provider",
				returnStatus: http.StatusNoContent,
				method:       "DELETE",
			},
		},
		{
			name: "Publish path",
			args: args{
				url:          "/published-apis/v1/apfId/service-apis/serviceId",
				returnStatus: http.StatusNotFound,
				method:       "GET",
			},
		},
		{
			name: "Discover path",
			args: args{
				url:          "/service-apis/v1/allServiceAPIs?api-invoker-id=api_invoker_id",
				returnStatus: http.StatusNotFound,
				method:       "GET",
			},
		},
		{
			name: "Invoker path",
			args: args{
				url:          "/api-invoker-management/v1/onboardedInvokers/invoker",
				returnStatus: http.StatusNoContent,
				method:       "DELETE",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result *testutil.CompletedRequest
			if tt.args.method == "GET" {
				result = testutil.NewRequest().Get(tt.args.url).Go(t, eServiceManager)
			} else if tt.args.method == "DELETE" {
				result = testutil.NewRequest().Delete(tt.args.url).Go(t, eServiceManager)
			}

			assert.Equal(t, tt.args.returnStatus, result.Code(), tt.name)
		})
	}
}

func TestGetSwagger(t *testing.T) {
	type args struct {
		apiPath string
		apiName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Provider api",
			args: args{
				apiPath: "provider",
				apiName: "Provider",
			},
		},
		{
			name: "Publish api",
			args: args{
				apiPath: "publish",
				apiName: "Publish",
			},
		},
		{
			name: "Invoker api",
			args: args{
				apiPath: "invoker",
				apiName: "Invoker",
			},
		},
		{
			name: "Discover api",
			args: args{
				apiPath: "discover",
				apiName: "Discover",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testutil.NewRequest().Get("/swagger/"+tt.args.apiPath).Go(t, eServiceManager)
			assert.Equal(t, http.StatusOK, result.Code())
			var swaggerResponse openapi3.T
			err := result.UnmarshalJsonToObject(&swaggerResponse)
			assert.Nil(t, err)
			assert.Contains(t, swaggerResponse.Info.Title, tt.args.apiName)
		})
	}
	invalidApi := "foobar"
	result := testutil.NewRequest().Get("/swagger/"+invalidApi).Go(t, eServiceManager)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	var errorResponse common29122.ProblemDetails
	err := result.UnmarshalJsonToObject(&errorResponse)
	assert.Nil(t, err)
	assert.Contains(t, *errorResponse.Cause, "Invalid API")
	assert.Contains(t, *errorResponse.Cause, invalidApi)
}