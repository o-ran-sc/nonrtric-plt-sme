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

func (newContext *ServiceSecurity) PrepareNewSecurityContext(services []publishserviceapi.ServiceAPIDescription) error {
	for i, securityInfo := range newContext.SecurityInfo {
		var securityMethods []publishserviceapi.SecurityMethod

		if securityInfo.InterfaceDetails != nil {
			interfaceSecurityMethods := securityInfo.InterfaceDetails.SecurityMethods
			prefSecurityMethods := securityInfo.PrefSecurityMethods

			if interfaceSecurityMethods != nil {
				securityMethods = append(securityMethods, *interfaceSecurityMethods...)
			}
			if prefSecurityMethods != nil {
				securityMethods = append(securityMethods, prefSecurityMethods...)
			}

		} else {
			//check aefProfile for securityMethods
			for _, service := range services {
				if service.ApiId != nil && securityInfo.ApiId != nil && strings.Compare(*service.ApiId, *securityInfo.ApiId) == 0 {
					aefProfiles := service.AefProfiles
					for _, profile := range *aefProfiles {
						if profile.SecurityMethods != nil {
							securityMethods = append(securityMethods, *profile.SecurityMethods...)
						}
					}
				}
			}

		}
		if len(securityMethods) == 0 {
			return fmt.Errorf("not found compatible security method")
		} else {
			newContext.SecurityInfo[i].SelSecurityMethod = &securityMethods[0]
		}
	}
	return nil
}
