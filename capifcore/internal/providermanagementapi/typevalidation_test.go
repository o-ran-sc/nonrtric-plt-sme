// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023: Nordix Foundation
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

package providermanagementapi

import (
	"testing"

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

func TestValidateRegistrationInformation(t *testing.T) {
	regInfoUnderTest := RegistrationInformation{}
	err := regInfoUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "apiProvPubKey")
	}

	regInfoUnderTest.ApiProvPubKey = "key"
	err = regInfoUnderTest.Validate()
	assert.Nil(t, err)
}

func TestValidateAPIProviderFunctionDetails(t *testing.T) {
	funcDetailsUnderTest := APIProviderFunctionDetails{}
	err := funcDetailsUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "apiProvFuncRole")
	}

	funcDetailsUnderTest.ApiProvFuncRole = ApiProviderFuncRoleAEF
	err = funcDetailsUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "apiProvPubKey")
	}

	funcDetailsUnderTest.RegInfo = RegistrationInformation{
		ApiProvPubKey: "key",
	}
	assert.Nil(t, funcDetailsUnderTest.Validate())
}

func TestValidateAPIProviderEnrolmentDetails(t *testing.T) {
	providerDetailsUnderTest := APIProviderEnrolmentDetails{}
	err := providerDetailsUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "regSec")
	}

	providerDetailsUnderTest.RegSec = "sec"
	funcs := []APIProviderFunctionDetails{{}}
	providerDetailsUnderTest.ApiProvFuncs = &funcs
	err = providerDetailsUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "apiProvFuncs")
		assert.Contains(t, err.Error(), "contains invalid")
	}
}

func TestUpdateFuncs_addNewFunction(t *testing.T) {
	providerUnderTest := getProvider()

	newFuncInfoAEF := "new func as AEF"
	newFuncs := append(*providerUnderTest.ApiProvFuncs, APIProviderFunctionDetails{
		ApiProvFuncInfo: &newFuncInfoAEF,
		ApiProvFuncRole: ApiProviderFuncRoleAEF,
	})
	providerUnderTest.ApiProvFuncs = &newFuncs

	err := providerUnderTest.UpdateFuncs(getProvider())

	assert.Nil(t, err)
	assert.Len(t, *providerUnderTest.ApiProvFuncs, 4)
	assert.True(t, providerUnderTest.IsFunctionRegistered("AEF_id_new_func_as_AEF"))
}

func TestUpdateFuncs_deleteFunction(t *testing.T) {
	providerUnderTest := getProvider()

	modFuncs := []APIProviderFunctionDetails{(*providerUnderTest.ApiProvFuncs)[0], (*providerUnderTest.ApiProvFuncs)[1]}
	providerUnderTest.ApiProvFuncs = &modFuncs

	err := providerUnderTest.UpdateFuncs(getProvider())

	assert.Nil(t, err)
	assert.Len(t, *providerUnderTest.ApiProvFuncs, 2)
	assert.True(t, providerUnderTest.IsFunctionRegistered(funcIdAPF))
	assert.True(t, providerUnderTest.IsFunctionRegistered(funcIdAMF))
}

func TestUpdateFuncs_unregisteredFunction(t *testing.T) {
	providerUnderTest := getProvider()

	unRegId := "unRegId"
	modFuncs := []APIProviderFunctionDetails{
		{
			ApiProvFuncId: &unRegId,
		},
	}
	providerUnderTest.ApiProvFuncs = &modFuncs

	err := providerUnderTest.UpdateFuncs(getProvider())
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), unRegId)
		assert.Contains(t, err.Error(), "not registered")
	}
}

func getProvider() APIProviderEnrolmentDetails {
	testFuncs := []APIProviderFunctionDetails{
		{
			ApiProvFuncId:   &funcIdAPF,
			ApiProvFuncInfo: &funcInfoAPF,
			ApiProvFuncRole: ApiProviderFuncRoleAPF,
		},
		{
			ApiProvFuncId:   &funcIdAMF,
			ApiProvFuncInfo: &funcInfoAMF,
			ApiProvFuncRole: ApiProviderFuncRoleAMF,
		},
		{
			ApiProvFuncId:   &funcIdAEF,
			ApiProvFuncInfo: &funcInfoAEF,
			ApiProvFuncRole: ApiProviderFuncRoleAEF,
		},
	}
	return APIProviderEnrolmentDetails{
		ApiProvDomId:   &domainID,
		ApiProvDomInfo: &domainInfo,
		ApiProvFuncs:   &testFuncs,
	}

}
