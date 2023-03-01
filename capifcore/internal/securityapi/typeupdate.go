// -
//
//	========================LICENSE_START=================================
//	O-RAN-SC
//	%%
//	Copyright (C) 2023: Nordix Foundation
//	%%
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	     http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
//	========================LICENSE_END===================================
package securityapi

import (
	"fmt"
	"strings"

	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"
)

var securityMethods []publishserviceapi.SecurityMethod

func (newContext *ServiceSecurity) PrepareNewSecurityContext(services []publishserviceapi.ServiceAPIDescription) error {
	securityMethods = []publishserviceapi.SecurityMethod{}
	for i, securityInfo := range newContext.SecurityInfo {

		if securityInfo.InterfaceDetails != nil {
			addSecurityMethodsFromInterfaceDetails(securityInfo.InterfaceDetails.SecurityMethods, &securityInfo.PrefSecurityMethods)

		} else {
			checkNil := securityInfo.ApiId != nil && securityInfo.AefId != nil
			if checkNil {
				service := getServiceByApiId(&services, securityInfo.ApiId)
				afpProfile := service.GetAefProfileById(securityInfo.AefId)

				addSecurityMethodsFromAefProfile(afpProfile)
			}
		}

		if isSecuryMethodsEmpty() {
			return fmt.Errorf("not found compatible security method")
		}
		newContext.SecurityInfo[i].SelSecurityMethod = &securityMethods[0]
	}
	return nil
}

func isSecuryMethodsEmpty() bool {
	return len(securityMethods) <= 0
}

func addSecurityMethodsFromInterfaceDetails(methodsFromInterface *[]publishserviceapi.SecurityMethod, prefMethods *[]publishserviceapi.SecurityMethod) {

	if methodsFromInterface != nil {
		securityMethods = append(securityMethods, *methodsFromInterface...)
	}
	if prefMethods != nil {
		securityMethods = append(securityMethods, *prefMethods...)
	}
}

func addSecurityMethodsFromAefProfile(afpProfile *publishserviceapi.AefProfile) {
	if afpProfile.SecurityMethods != nil {
		securityMethods = append(securityMethods, *afpProfile.SecurityMethods...)
	}
}

func getServiceByApiId(services *[]publishserviceapi.ServiceAPIDescription, apiId *string) *publishserviceapi.ServiceAPIDescription {

	for _, service := range *services {
		if apiId != nil && strings.Compare(*service.ApiId, *apiId) == 0 {
			return &service
		}
	}
	return nil
}
