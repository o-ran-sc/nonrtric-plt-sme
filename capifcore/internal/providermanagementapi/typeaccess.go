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

func (ed APIProviderEnrolmentDetails) GetExposingFunctionIdsForPublisher(apfId string) []string {
	for _, registeredFunc := range *ed.ApiProvFuncs {
		if *registeredFunc.ApiProvFuncId == apfId {
			return ed.getExposingFunctionIds()
		}
	}
	return nil
}

func (ed APIProviderEnrolmentDetails) getExposingFunctionIds() []string {
	exposedFuncs := []string{}
	for _, registeredFunc := range *ed.ApiProvFuncs {
		if registeredFunc.isExposingFunction() {
			exposedFuncs = append(exposedFuncs, *registeredFunc.ApiProvFuncId)
		}
	}
	return exposedFuncs
}

func (ed APIProviderEnrolmentDetails) IsFunctionRegistered(functionId string) bool {
	for _, registeredFunc := range *ed.ApiProvFuncs {
		if *registeredFunc.ApiProvFuncId == functionId {
			return true
		}
	}
	return false
}

func (ed APIProviderEnrolmentDetails) IsPublishingFunctionRegistered(functionId string) bool {
	for _, registeredFunc := range *ed.ApiProvFuncs {
		if *registeredFunc.ApiProvFuncId == functionId && registeredFunc.isPublishingFunction() {
			return true
		}
	}
	return false
}

func (fd APIProviderFunctionDetails) isPublishingFunction() bool {
	return fd.ApiProvFuncRole == ApiProviderFuncRoleAPF
}

func (fd APIProviderFunctionDetails) isExposingFunction() bool {
	return fd.ApiProvFuncRole == ApiProviderFuncRoleAEF
}
