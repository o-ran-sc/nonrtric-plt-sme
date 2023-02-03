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

package providermanagement

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/labstack/echo/v4"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	provapi "oransc.org/nonrtric/capifcore/internal/providermanagementapi"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

var (
	domainID      = "domain_id_rApp_domain"
	otherDomainID = "domain_id_other_domain"
	domainInfo    = "rApp domain"
	funcInfoAPF   = "rApp as APF"
	funcIdAPF     = "APF_id_rApp_as_APF"
	funcInfoAMF   = "rApp as AMF"
	funcIdAMF     = "AMF_id_rApp_as_AMF"
	funcInfoAEF   = "rApp as AEF"
	funcIdAEF     = "AEF_id_rApp_as_AEF"
)

func TestRegisterValidProvider(t *testing.T) {
	managerUnderTest, requestHandler := getEcho()

	newProvider := getProvider()

	// Register a valid provider
	result := testutil.NewRequest().Post("/registrations").WithJsonBody(newProvider).Go(t, requestHandler)

	assert.Equal(t, http.StatusCreated, result.Code())
	var resultProvider provapi.APIProviderEnrolmentDetails
	err := result.UnmarshalBodyToObject(&resultProvider)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, *resultProvider.ApiProvDomId, domainID)
	assert.Equal(t, *(*resultProvider.ApiProvFuncs)[0].ApiProvFuncId, funcIdAPF)
	assert.Equal(t, *(*resultProvider.ApiProvFuncs)[1].ApiProvFuncId, funcIdAMF)
	assert.Equal(t, *(*resultProvider.ApiProvFuncs)[2].ApiProvFuncId, funcIdAEF)
	assert.Empty(t, resultProvider.FailReason)
	assert.Equal(t, "http://example.com/registrations/"+*resultProvider.ApiProvDomId, result.Recorder.Header().Get(echo.HeaderLocation))
	assert.True(t, managerUnderTest.IsFunctionRegistered("APF_id_rApp_as_APF"))

	// Register same provider again should result in BadRequest
	result = testutil.NewRequest().Post("/registrations").WithJsonBody(newProvider).Go(t, requestHandler)
	var errorObj common29122.ProblemDetails
	assert.Equal(t, http.StatusForbidden, result.Code())
	err = result.UnmarshalBodyToObject(&errorObj)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusForbidden, *errorObj.Status)
	assert.Contains(t, *errorObj.Cause, "already registered")
}

func TestUpdateValidProviderWithNewFunction(t *testing.T) {
	managerUnderTest, requestHandler := getEcho()

	provider := getProvider()
	provider.ApiProvDomId = &domainID
	(*provider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	(*provider.ApiProvFuncs)[1].ApiProvFuncId = &funcIdAMF
	(*provider.ApiProvFuncs)[2].ApiProvFuncId = &funcIdAEF
	managerUnderTest.registeredProviders[domainID] = provider

	// Modify the provider
	updatedProvider := getProvider()
	updatedProvider.ApiProvDomId = &domainID
	(*updatedProvider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	(*updatedProvider.ApiProvFuncs)[1].ApiProvFuncId = &funcIdAMF
	(*updatedProvider.ApiProvFuncs)[2].ApiProvFuncId = &funcIdAEF
	newDomainInfo := "New domain info"
	updatedProvider.ApiProvDomInfo = &newDomainInfo
	newFunctionInfo := "New function info"
	(*updatedProvider.ApiProvFuncs)[0].ApiProvFuncInfo = &newFunctionInfo
	newFuncInfoAEF := "new func as AEF"
	testFuncs := *updatedProvider.ApiProvFuncs
	testFuncs = append(testFuncs, provapi.APIProviderFunctionDetails{
		ApiProvFuncInfo: &newFuncInfoAEF,
		ApiProvFuncRole: provapi.ApiProviderFuncRoleAEF,
		RegInfo: provapi.RegistrationInformation{
			ApiProvPubKey: "key",
		},
	})
	updatedProvider.ApiProvFuncs = &testFuncs

	result := testutil.NewRequest().Put("/registrations/"+domainID).WithJsonBody(updatedProvider).Go(t, requestHandler)

	var resultProvider provapi.APIProviderEnrolmentDetails
	assert.Equal(t, http.StatusOK, result.Code())
	err := result.UnmarshalBodyToObject(&resultProvider)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, newDomainInfo, *resultProvider.ApiProvDomInfo)
	assert.Equal(t, newFunctionInfo, *(*resultProvider.ApiProvFuncs)[0].ApiProvFuncInfo)
	assert.Equal(t, *(*resultProvider.ApiProvFuncs)[3].ApiProvFuncId, "AEF_id_new_func_as_AEF")
	assert.Empty(t, resultProvider.FailReason)
	assert.True(t, managerUnderTest.IsFunctionRegistered("AEF_id_new_func_as_AEF"))
}

func TestUpdateValidProviderWithDeletedFunction(t *testing.T) {
	managerUnderTest, requestHandler := getEcho()

	provider := getProvider()
	provider.ApiProvDomId = &domainID
	(*provider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	(*provider.ApiProvFuncs)[1].ApiProvFuncId = &funcIdAMF
	(*provider.ApiProvFuncs)[2].ApiProvFuncId = &funcIdAEF
	managerUnderTest.registeredProviders[domainID] = provider

	// Modify the provider
	updatedProvider := getProvider()
	updatedProvider.ApiProvDomId = &domainID
	(*updatedProvider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	(*updatedProvider.ApiProvFuncs)[2].ApiProvFuncId = &funcIdAEF
	testFuncs := []provapi.APIProviderFunctionDetails{
		(*updatedProvider.ApiProvFuncs)[0],
		(*updatedProvider.ApiProvFuncs)[2],
	}
	updatedProvider.ApiProvFuncs = &testFuncs

	result := testutil.NewRequest().Put("/registrations/"+domainID).WithJsonBody(updatedProvider).Go(t, requestHandler)

	var resultProvider provapi.APIProviderEnrolmentDetails
	assert.Equal(t, http.StatusOK, result.Code())
	err := result.UnmarshalBodyToObject(&resultProvider)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Len(t, (*resultProvider.ApiProvFuncs), 2)
	assert.Empty(t, resultProvider.FailReason)
	assert.False(t, managerUnderTest.IsFunctionRegistered(funcIdAMF))
}

func TestUpdateMissingFunction(t *testing.T) {
	managerUnderTest, requestHandler := getEcho()

	provider := getProvider()
	provider.ApiProvDomId = &domainID
	otherId := "otherId"
	(*provider.ApiProvFuncs)[0].ApiProvFuncId = &otherId
	(*provider.ApiProvFuncs)[1].ApiProvFuncId = &funcIdAMF
	(*provider.ApiProvFuncs)[2].ApiProvFuncId = &funcIdAEF
	managerUnderTest.registeredProviders[domainID] = provider

	// Modify the provider
	updatedProvider := getProvider()
	updatedProvider.ApiProvDomId = &domainID
	(*updatedProvider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	newFunctionInfo := "New function info"
	(*updatedProvider.ApiProvFuncs)[0].ApiProvFuncInfo = &newFunctionInfo

	result := testutil.NewRequest().Put("/registrations/"+domainID).WithJsonBody(updatedProvider).Go(t, requestHandler)

	var errorObj common29122.ProblemDetails
	assert.Equal(t, http.StatusBadRequest, result.Code())
	err := result.UnmarshalBodyToObject(&errorObj)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusBadRequest, *errorObj.Status)
	assert.Contains(t, *errorObj.Cause, funcIdAPF)
	assert.Contains(t, *errorObj.Cause, "not registered")
}

func TestDeleteProvider(t *testing.T) {
	managerUnderTest, requestHandler := getEcho()

	provider := getProvider()
	provider.ApiProvDomId = &domainID
	(*provider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	managerUnderTest.registeredProviders[domainID] = provider
	assert.True(t, managerUnderTest.IsFunctionRegistered(funcIdAPF))

	result := testutil.NewRequest().Delete("/registrations/"+domainID).Go(t, requestHandler)

	assert.Equal(t, http.StatusNoContent, result.Code())
	assert.False(t, managerUnderTest.IsFunctionRegistered(funcIdAPF))
}
func TestProviderHandlingValidation(t *testing.T) {
	_, requestHandler := getEcho()

	newProvider := provapi.APIProviderEnrolmentDetails{}

	// Register an invalid provider
	result := testutil.NewRequest().Post("/registrations").WithJsonBody(newProvider).Go(t, requestHandler)

	assert.Equal(t, http.StatusBadRequest, result.Code())
	var problemDetails common29122.ProblemDetails
	err := result.UnmarshalBodyToObject(&problemDetails)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, http.StatusBadRequest, *problemDetails.Status)
	assert.Contains(t, *problemDetails.Cause, "Provider not valid")
	assert.Contains(t, *problemDetails.Cause, "regSec")
}

func TestGetExposedFunctionsForPublishingFunction(t *testing.T) {
	managerUnderTest := NewProviderManager()

	provider := getProvider()
	provider.ApiProvDomId = &domainID
	(*provider.ApiProvFuncs)[0].ApiProvFuncId = &funcIdAPF
	(*provider.ApiProvFuncs)[1].ApiProvFuncId = &funcIdAMF
	(*provider.ApiProvFuncs)[2].ApiProvFuncId = &funcIdAEF
	managerUnderTest.registeredProviders[domainID] = provider
	managerUnderTest.registeredProviders[otherDomainID] = getOtherProvider()

	exposedFuncs := managerUnderTest.GetAefsForPublisher(funcIdAPF)
	assert.Equal(t, 1, len(exposedFuncs))
	assert.Equal(t, funcIdAEF, exposedFuncs[0])
}

func getProvider() provapi.APIProviderEnrolmentDetails {
	testFuncs := []provapi.APIProviderFunctionDetails{
		{
			ApiProvFuncInfo: &funcInfoAPF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAPF,
			RegInfo: provapi.RegistrationInformation{
				ApiProvPubKey: "key",
			},
		},
		{
			ApiProvFuncInfo: &funcInfoAMF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAMF,
			RegInfo: provapi.RegistrationInformation{
				ApiProvPubKey: "key",
			},
		},
		{
			ApiProvFuncInfo: &funcInfoAEF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAEF,
			RegInfo: provapi.RegistrationInformation{
				ApiProvPubKey: "key",
			},
		},
	}
	return provapi.APIProviderEnrolmentDetails{
		RegSec:         "sec",
		ApiProvDomInfo: &domainInfo,
		ApiProvFuncs:   &testFuncs,
	}

}

func getOtherProvider() provapi.APIProviderEnrolmentDetails {
	otherDomainInfo := "other domain"
	otherFuncInfoAPF := "other as APF"
	otherApfId := "APF_id_other_as_APF"
	otherFuncInfoAMF := "other as AMF"
	otherAmfId := "AMF_id_other_as_AMF"
	otherFuncInfoAEF := "other as AEF"
	otherAefId := "AEF_id_other_as_AEF"
	testFuncs := []provapi.APIProviderFunctionDetails{
		{
			ApiProvFuncId:   &otherApfId,
			ApiProvFuncInfo: &otherFuncInfoAPF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAPF,
		},
		{
			ApiProvFuncId:   &otherAmfId,
			ApiProvFuncInfo: &otherFuncInfoAMF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAMF,
		},
		{
			ApiProvFuncId:   &otherAefId,
			ApiProvFuncInfo: &otherFuncInfoAEF,
			ApiProvFuncRole: provapi.ApiProviderFuncRoleAEF,
		},
	}
	return provapi.APIProviderEnrolmentDetails{
		ApiProvDomId:   &otherDomainID,
		ApiProvDomInfo: &otherDomainInfo,
		ApiProvFuncs:   &testFuncs,
	}

}

func getEcho() (*ProviderManager, *echo.Echo) {
	swagger, err := provapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	pm := NewProviderManager()

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	provapi.RegisterHandlers(e, pm)
	return pm, e
}
