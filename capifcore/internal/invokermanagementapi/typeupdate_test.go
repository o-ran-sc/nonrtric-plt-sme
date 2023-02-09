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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareNewInvoker(t *testing.T) {
	invokerUnderTest := APIInvokerEnrolmentDetails{}
	uuidFunc = func() string {
		return "1"
	}

	invokerUnderTest.PrepareNewInvoker()
	assert.Equal(t, "api_invoker_id_1", *invokerUnderTest.ApiInvokerId)
	assert.Equal(t, "onboarding_secret_api_invoker_id_1", *invokerUnderTest.OnboardingInformation.OnboardingSecret)

	invokerInfo := "invoker info"
	invokerUnderTest.ApiInvokerInformation = &invokerInfo
	invokerUnderTest.PrepareNewInvoker()
	assert.Equal(t, "api_invoker_id_invoker_info", *invokerUnderTest.ApiInvokerId)
	assert.Equal(t, "onboarding_secret_invoker_info", *invokerUnderTest.OnboardingInformation.OnboardingSecret)
}
