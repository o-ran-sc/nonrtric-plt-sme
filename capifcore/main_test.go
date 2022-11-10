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

package main

import (
	"net/http"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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
			name: "Security path",
			args: args{
				url:          "/capif-security/v1/trustedInvokers/apiInvokerId",
				returnStatus: http.StatusNotImplemented,
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
