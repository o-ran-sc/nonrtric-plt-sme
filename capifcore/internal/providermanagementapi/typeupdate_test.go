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

func TestPrepareNewProvider(t *testing.T) {
	domainInfo := "domain info"
	funcInfo := "func info"
	providerUnderTest := APIProviderEnrolmentDetails{
		ApiProvDomInfo: &domainInfo,
		ApiProvFuncs: &[]APIProviderFunctionDetails{
			{
				ApiProvFuncRole: ApiProviderFuncRoleAPF,
				ApiProvFuncInfo: &funcInfo,
			},
			{
				ApiProvFuncRole: ApiProviderFuncRoleAEF,
			},
		},
	}
	uuidFunc = func() string {
		return "1"
	}

	providerUnderTest.PrepareNewProvider()

	assert.Equal(t, "domain_id_domain_info", *providerUnderTest.ApiProvDomId)
	assert.Equal(t, "APF_id_func_info", *(*providerUnderTest.ApiProvFuncs)[0].ApiProvFuncId)
	assert.Equal(t, "AEF_id_1", *(*providerUnderTest.ApiProvFuncs)[1].ApiProvFuncId)

	providerUnderTest = APIProviderEnrolmentDetails{}

	providerUnderTest.PrepareNewProvider()

	assert.Equal(t, "domain_id_1", *providerUnderTest.ApiProvDomId)
}

func TestUpdateFuncs(t *testing.T) {
	registeredProvider := getProvider()

	funcInfo := "func info"
	updatedFuncs := []APIProviderFunctionDetails{
		(*registeredProvider.ApiProvFuncs)[0],
		(*registeredProvider.ApiProvFuncs)[2],
		{
			ApiProvFuncRole: ApiProviderFuncRoleAEF,
			ApiProvFuncInfo: &funcInfo,
		},
	}
	providerUnderTest := APIProviderEnrolmentDetails{
		ApiProvFuncs: &updatedFuncs,
	}
	err := providerUnderTest.UpdateFuncs(registeredProvider)

	assert.Nil(t, err)
	assert.Len(t, *providerUnderTest.ApiProvFuncs, 3)
	assert.Equal(t, funcIdAPF, *(*providerUnderTest.ApiProvFuncs)[0].ApiProvFuncId)
	assert.Equal(t, funcIdAEF, *(*providerUnderTest.ApiProvFuncs)[1].ApiProvFuncId)
	assert.Equal(t, "AEF_id_func_info", *(*providerUnderTest.ApiProvFuncs)[2].ApiProvFuncId)
}
