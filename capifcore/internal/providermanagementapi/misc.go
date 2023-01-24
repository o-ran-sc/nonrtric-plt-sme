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
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (ed APIProviderEnrolmentDetails) GetExposedFunctionIdsForPublisher(apfId string) []string {
	for _, registeredFunc := range *ed.ApiProvFuncs {
		if *registeredFunc.ApiProvFuncId == apfId && registeredFunc.isProvidingFunction() {
			return ed.getExposedFunctionIds()
		}
	}
	return nil
}

func (ed APIProviderEnrolmentDetails) getExposedFunctionIds() []string {
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

func (ed *APIProviderEnrolmentDetails) UpdateFuncs(registeredProvider APIProviderEnrolmentDetails) error {
	for pos, function := range *ed.ApiProvFuncs {
		if function.ApiProvFuncId == nil {
			(*ed.ApiProvFuncs)[pos].ApiProvFuncId = getFuncId(function.ApiProvFuncRole, function.ApiProvFuncInfo)
		} else {
			if !registeredProvider.IsFunctionRegistered(*function.ApiProvFuncId) {
				return fmt.Errorf("function with ID %s is not registered for the provider", *function.ApiProvFuncId)
			}
		}
	}
	return nil
}

func (ed *APIProviderEnrolmentDetails) PrepareNewProvider() {
	ed.ApiProvDomId = ed.getDomainId()

	ed.registerFunctions()

}

func (ed *APIProviderEnrolmentDetails) getDomainId() *string {
	var idAsString string
	if ed.ApiProvDomInfo != nil {
		idAsString = strings.ReplaceAll(*ed.ApiProvDomInfo, " ", "_")
	} else {
		idAsString = uuid.New().String()
	}
	newId := "domain_id_" + idAsString
	return &newId
}

func (ed *APIProviderEnrolmentDetails) registerFunctions() {
	if ed.ApiProvFuncs == nil {
		return
	}
	for i, provFunc := range *ed.ApiProvFuncs {
		(*ed.ApiProvFuncs)[i].ApiProvFuncId = getFuncId(provFunc.ApiProvFuncRole, provFunc.ApiProvFuncInfo)
	}
}

func getFuncId(role ApiProviderFuncRole, funcInfo *string) *string {
	var idPrefix string
	switch role {
	case ApiProviderFuncRoleAPF:
		idPrefix = "APF_id_"
	case ApiProviderFuncRoleAMF:
		idPrefix = "AMF_id_"
	case ApiProviderFuncRoleAEF:
		idPrefix = "AEF_id_"
	default:
		idPrefix = "function_id_"
	}
	idAsString := idPrefix + strings.ReplaceAll(*funcInfo, " ", "_")
	return &idAsString
}

func (fd APIProviderFunctionDetails) isProvidingFunction() bool {
	return fd.ApiProvFuncRole == ApiProviderFuncRoleAPF
}

func (fd APIProviderFunctionDetails) isExposingFunction() bool {
	return fd.ApiProvFuncRole == ApiProviderFuncRoleAEF
}

func (ri RegistrationInformation) Validate() error {
	if len(strings.TrimSpace(ri.ApiProvPubKey)) == 0 {
		return errors.New("RegistrationInformation missing required apiProvPubKey")
	}
	return nil
}

func (fd APIProviderFunctionDetails) Validate() error {
	switch role := fd.ApiProvFuncRole; role {
	case ApiProviderFuncRoleAEF:
	case ApiProviderFuncRoleAPF:
	case ApiProviderFuncRoleAMF:
	default:
		return errors.New("APIProviderFunctionDetails missing required apiProvFuncRole")
	}

	return fd.RegInfo.Validate()
}

func (pd APIProviderEnrolmentDetails) Validate() error {
	if len(strings.TrimSpace(pd.RegSec)) == 0 {
		return errors.New("APIProviderEnrolmentDetails missing required regSec")
	}
	if pd.ApiProvFuncs != nil {
		return pd.validateFunctions()
	}
	return nil
}

func (pd APIProviderEnrolmentDetails) validateFunctions() error {
	for _, function := range *pd.ApiProvFuncs {
		err := function.Validate()
		if err != nil {
			return fmt.Errorf("apiProvFuncs contains invalid function: %s", err)
		}
	}
	return nil
}
