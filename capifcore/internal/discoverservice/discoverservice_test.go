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

	"oransc.org/nonrtric/capifcore/internal/discoverserviceapi"

	"oransc.org/nonrtric/capifcore/internal/publishservice"

	"github.com/labstack/echo/v4"

	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/publishservice/mocks"

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
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get all APIs, without any filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id=api_invoker_id").Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 2, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
	assert.Equal(t, "apiName2", (*resultInvoker.ServiceAPIDescriptions)[1].ApiName)
	apiRegisterMock.AssertCalled(t, "GetAPIs")
	assert.Equal(t, 2, len(*resultInvoker.ServiceAPIDescriptions))
}

func TestFilterApiName(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", "", nil, nil, ""),
		getAPI("apiName2", "", "", "", nil, nil, ""),
	}
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-name=apiName1&api-invoker-id=api_invoker_id").Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterAefId(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "aefId", "", "", nil, nil, ""),
		getAPI("apiName2", "otherAefId", "", "", nil, nil, ""),
	}
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id=api_invoker_id&aef-id=aefId").Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterVersion(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", "v1", nil, nil, ""),
		getAPI("apiName2", "", "", "v2", nil, nil, ""),
	}
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-version=v1&api-invoker-id=api_invoker_id").Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterCommType(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", "", nil, nil, publishapi.CommunicationTypeREQUESTRESPONSE),
		getAPI("apiName2", "", "", "", nil, nil, publishapi.CommunicationTypeSUBSCRIBENOTIFY),
	}
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id=api_invoker_id&comm-type=REQUEST_RESPONSE").Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterVersionAndCommType(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "", "v1", nil, nil, publishapi.CommunicationTypeREQUESTRESPONSE),
		getAPI("apiName2", "", "", "v1", nil, nil, publishapi.CommunicationTypeSUBSCRIBENOTIFY),
		getAPI("apiName3", "", "", "v2", nil, nil, publishapi.CommunicationTypeSUBSCRIBENOTIFY),
	}
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id=api_invoker_id&api-version=v1&comm-type=SUBSCRIBE_NOTIFY").Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName2", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func TestFilterAPICategory(t *testing.T) {
	var err error

	apiList := []publishapi.ServiceAPIDescription{
		getAPI("apiName1", "", "apiCategory", "", nil, nil, ""),
		getAPI("apiName2", "", "", "", nil, nil, ""),
	}
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id=api_invoker_id&api-cat=apiCategory").Go(t, requestHandler)

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
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id=api_invoker_id&protocol=HTTP_1_1").Go(t, requestHandler)

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
	apiRegisterMock := mocks.APIRegister{}
	apiRegisterMock.On("GetAPIs").Return(&apiList)
	requestHandler := getEcho(&apiRegisterMock)

	// Get APIs with filter
	result := testutil.NewRequest().Get("/allServiceAPIs?api-invoker-id=api_invoker_id&data-format=JSON").Go(t, requestHandler)

	assert.Equal(t, http.StatusOK, result.Code())
	var resultInvoker discoverserviceapi.DiscoveredAPIs
	err = result.UnmarshalBodyToObject(&resultInvoker)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, 1, len(*resultInvoker.ServiceAPIDescriptions))
	assert.Equal(t, "apiName1", (*resultInvoker.ServiceAPIDescriptions)[0].ApiName)
}

func getEcho(apiRegister publishservice.APIRegister) *echo.Echo {
	swagger, err := discoverserviceapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	ds := NewDiscoverService(apiRegister)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	discoverserviceapi.RegisterHandlers(e, ds)
	return e
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
