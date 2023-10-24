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
	"os"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	log "github.com/sirupsen/logrus"

	"oransc.org/nonrtric/r1-sme-manager/internal/common29122"
	"oransc.org/nonrtric/r1-sme-manager/internal/envreader"
)

var e *echo.Echo
var myPorts map [string]int

func TestMain(m *testing.M) {
    // Init code to run before tests
	myEnv, myPorts, err := envreader.ReadDotEnv()
	if err != nil {
		log.Fatal("Error loading environment file")
		return
	}

	e, err = getEcho(myEnv, myPorts)
	if err != nil {
		log.Fatal("getEcho fatal error")
		return
	}
    
    // Run tests
    exitVal := m.Run()
    
    // Finalization code to run after tests

    // Exit with exit value from tests
    os.Exit(exitVal)
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
				result = testutil.NewRequest().Get(tt.args.url).Go(t, e)
			} else if tt.args.method == "DELETE" {
				result = testutil.NewRequest().Delete(tt.args.url).Go(t, e)
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
			result := testutil.NewRequest().Get("/swagger/"+tt.args.apiPath).Go(t, e)
			assert.Equal(t, http.StatusOK, result.Code())
			var swaggerResponse openapi3.T
			err := result.UnmarshalJsonToObject(&swaggerResponse)
			assert.Nil(t, err)
			assert.Contains(t, swaggerResponse.Info.Title, tt.args.apiName)
		})
	}
	invalidApi := "foobar"
	result := testutil.NewRequest().Get("/swagger/"+invalidApi).Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	var errorResponse common29122.ProblemDetails
	err := result.UnmarshalJsonToObject(&errorResponse)
	assert.Nil(t, err)
	assert.Contains(t, *errorResponse.Cause, "Invalid API")
	assert.Contains(t, *errorResponse.Cause, invalidApi)
}

// func TestHTTPServer(t *testing.T) {
// 	port := myPorts["R1_SME_MANAGER_PORT"]

// 	go startWebServer(e, port)
// 	log.Info("Server started and listening on port: ", port)

// 	time.Sleep(100 * time.Millisecond)

// 	client := &http.Client{}
// 	res, err := client.Get(fmt.Sprintf("http://localhost:%d", port))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	defer res.Body.Close()
// 	assert.Equal(t, res.StatusCode, res.StatusCode)

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	expected := []byte("Hello, World!")
// 	assert.Equal(t, expected, body)
// }
