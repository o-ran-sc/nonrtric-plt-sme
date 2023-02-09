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
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var uuidFunc = getUUID

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
		idAsString = uuidFunc()
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
	}
	var id string
	if funcInfo != nil {
		id = strings.ReplaceAll(*funcInfo, " ", "_")
	} else {
		id = uuidFunc()
	}
	idAsString := idPrefix + id
	return &idAsString
}

func getUUID() string {
	return uuid.NewString()
}
