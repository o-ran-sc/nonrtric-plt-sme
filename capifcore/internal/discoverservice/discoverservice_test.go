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

package discoverservice

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/discoverserviceapi"
	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/invokermanagementapi"

	"github.com/labstack/echo/v4"

	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/invokermanagement/mocks"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

var protocolHTTP11 = publishapi.ProtocolHTTP11
var dataFormatJSON = publishapi.DataFormatJSON

func TestGetAllServiceAPIs(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "aefId", "apiCategory", "v1", nil, nil, ""),
		getAPI("apiName2", "aefId", "apiCategory", "v1", nil, nil, ""),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get all APIs, without any filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 2, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
	assert.Equal(t, "apiName2", (*resultInvoker.ServiceAPIDescriptions)[1].ApiName)
	assert.Equal(t, 2, len(*resultInvoker.ServiceAPIDescriptions))
}

func TestGetAllServiceAPIsWhenMissingProvider(t *testing.T) {
	invokerId := "unregistered"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, nil)

	requestHandler := getEcho(invokerRegisterrMock)

	// Get all APIs, without any filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId).Go(t, requestHandler)

	assert.Equal(t, http.StatusNotFound, result.Code())
	var problemDetails common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	notFound := http.StatusNotFound
	assert.Equal(t, &notFound, problemDetails.Status)
	errMsg := "Invoker not registered"
	assert.Equal(t, &errMsg, problemDetails.Cause)
}

func TestFilterApiName(t *testing.T) {
	var err error

	apiName := "apiName1"
	apiList := []publishapi.ServiceAPIDescription{
		getAPI(apiName, "", "", "", nil, nil, ""),
		getAPI("apiName2", "", "", "", nil, nil, ""),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&api-name="+apiName).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterAefId(t *testing.T) {
	var err error

	aefId := "aefId"
	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", aefId, "", "", nil, nil, ""),
		getAPI("apiName2", "otherAefId", "", "", nil, nil, ""),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&aef-id="+aefId).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterVersion(t *testing.T) {
	var err error

	version := "v1"
	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", version, nil, nil, ""),
		getAPI("apiName2", "", "", "v2", nil, nil, ""),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&api-version="+version).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterCommType(t *testing.T) {
	var err error

	commType := publishapi.CommunicationTypeREQUESTRESPONSE
	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", "", nil, nil, commType),
		getAPI("apiName2", "", "", "", nil, nil, publishapi.CommunicationTypeSUBSCRIBENOTIFY),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&comm-type="+string(commType)).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterVersionAndCommType(t *testing.T) {
	var err error

	version := "v1"
	commType := publishapi.CommunicationTypeSUBSCRIBENOTIFY
	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", version, nil, nil, publishapi.CommunicationTypeREQUESTRESPONSE),
		getAPI("apiName2", "", "", version, nil, nil, commType),
		getAPI("apiName3", "", "", "v2", nil, nil, commType),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&api-version="+version+"&comm-type="+string(commType)).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName2", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterAPICategory(t *testing.T) {
	var err error

	apiCategory := "apiCategory"
	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", apiCategory, "", nil, nil, ""),
		getAPI("apiName2", "", "", "", nil, nil, ""),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&api-cat="+apiCategory).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterProtocol(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", "", &protocolHTTP11, nil, ""),
		getAPI("apiName2", "", "", "", nil, nil, ""),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&protocol="+string(protocolHTTP11)).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

var DataFormatOther publishapi.DataFormat = "OTHER"

func TestFilterDataFormat(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", "", nil, &dataFormatJSON, ""),
		getAPI("apiName2", "", "", "", nil, nil, ""),
	}
	invokerId := "api_invoker_id"
	invokerRegisterrMock := getInvokerRegisterMock(invokerId, apiList)
	requestHandler := getEcho(invokerRegisterrMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id="+invokerId+"&data-format="+string(dataFormatJSON)).Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func getEcho(invokerManager invokermanagement.InvokerRegister) *echo.Echo {
	swagger, err := discoverserviceapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	ds := NewDiscoverService(invokerManager)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	discoverserviceapi.RegisterHandlers(e, ds)
	return e
}

func getInvokerRegisterMock(invokerId string, apisToReturn []publishapi.ServiceAPIDescription) *mocks.InvokerRegister {
	apiList := invokermanagementapi.APIList(apisToReturn)
	invokerRegisterrMock := mocks.InvokerRegister{}
	if apisToReturn != nil {
		invokerRegisterrMock.On("GetInvokerApiList", invokerId).Return(&apiList)
	} else {
		invokerRegisterrMock.On("GetInvokerApiList", invokerId).Return(nil)
	}
	return &invokerRegisterrMock
}

func getAPI(apiName, aefId, apiCategory, apiVersion string, protocol *publishapi.Protocol, dataFormat *publishapi.DataFormat, commType publishapi.CommunicationType) publishapi.ServiceAPIDescription {
	apiId := "apiId_" + apiName
	description := "description"
	domainName := "domain"
	otherDomainName := "otherDomain"
	var otherProtocol publishapi.Protocol = "HTTP_2"
	categoryPointer := &apiCategory
	if apiCategory == "" {
		categoryPointer = nil
	}
	return publishapi.ServiceAPIDescription{
		ApiId:              &apiId,
		ApiName:            apiName,
		Description:        &description,
		ServiceAPICategory: categoryPointer,
		AefProfiles: &[]publishapi.AefProfile{
			{
				AefId:      aefId,
				DomainName: &domainName,
				Protocol:   protocol,
				DataFormat: dataFormat,
				Versions: []publishapi.Version{
					{
						ApiVersion: apiVersion,
						Resources: &[]publishapi.Resource{
							{
								ResourceName: "app",
								CommType:     commType,
								Uri:          "uri",
								Operations: &[]publishapi.Operation{
									"POST",
								},
							},
						},
					},
					{
						ApiVersion: "v2",
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
			{
				AefId:      "otherAefId",
				DomainName: &otherDomainName,
				Protocol:   &otherProtocol,
				DataFormat: &DataFormatOther,
				Versions: []publishapi.Version{
					{
						ApiVersion: "v3",
						Resources: &[]publishapi.Resource{
							{
								ResourceName: "app",
								CommType:     commType,
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
	}
}
