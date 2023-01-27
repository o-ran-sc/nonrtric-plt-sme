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

package invokermanagementapi

import (
	"strings"

	"github.com/google/uuid"
)

var uuidFunc = getUUID

func (ied *APIInvokerEnrolmentDetails) PrepareNewInvoker() {
	ied.createId()
	ied.getOnboardingSecret()

}

func (ied *APIInvokerEnrolmentDetails) createId() {
	idAsString := "api_invoker_id_"
	if ied.ApiInvokerInformation != nil {
		idAsString = idAsString + strings.ReplaceAll(*ied.ApiInvokerInformation, " ", "_")
	} else {
		idAsString = idAsString + uuidFunc()
	}
	ied.ApiInvokerId = &idAsString
}

func getUUID() string {
	return uuid.NewString()
}

func (ied *APIInvokerEnrolmentDetails) getOnboardingSecret() {
	onboardingSecret := "onboarding_secret_"
	if ied.ApiInvokerInformation != nil {
		onboardingSecret = onboardingSecret + strings.ReplaceAll(*ied.ApiInvokerInformation, " ", "_")
	} else {
		onboardingSecret = onboardingSecret + *ied.ApiInvokerId
	}
	ied.OnboardingInformation.OnboardingSecret = &onboardingSecret
}
