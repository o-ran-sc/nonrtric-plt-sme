// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022: Nordix Foundation. All rights reserved.
//   Copyright (C) 2023 OpenInfra Foundation Europe. All rights reserved.
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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"oransc.org/nonrtric/capifcore/internal/common29122"
)

var e *echo.Echo

func Test_routing(t *testing.T) {
	e = getEcho()

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
		{
			name: "Event path",
			args: args{
				url:          "/capif-events/v1/subscriberId/subscriptions/subId",
				returnStatus: http.StatusNoContent,
				method:       "DELETE",
			},
		},
		{
			name: "Security path",
			args: args{
				url:          "/capif-security/v1/trustedInvokers/apiInvokerId",
				returnStatus: http.StatusNotFound,
				method:       "GET",
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
	e = getEcho()

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
			name: "Events api",
			args: args{
				apiPath: "events",
				apiName: "Events",
			},
		},
		{
			name: "Discover api",
			args: args{
				apiPath: "discover",
				apiName: "Discover",
			},
		},
		{
			name: "Security api",
			args: args{
				apiPath: "security",
				apiName: "Security",
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

func TestHTTPSServer(t *testing.T) {
	e = getEcho()
	var port = 44333
	go startHttpsWebServer(e, 44333, "certs/cert.pem", "certs/key.pem") //"certs/test/cert.pem", "certs/test/key.pem"

	time.Sleep(100 * time.Millisecond)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	res, err := client.Get(fmt.Sprintf("https://localhost:%d", port))
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte("Hello, World!")
	assert.Equal(t, expected, body)
}
