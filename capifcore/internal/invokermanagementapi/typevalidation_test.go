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

func TestValidateInvoker(t *testing.T) {
	invokerUnderTest := APIInvokerEnrolmentDetails{}

	err := invokerUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "NotificationDestination")
	}

	invokerUnderTest.NotificationDestination = "invalid dest"
	err = invokerUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid")
		assert.Contains(t, err.Error(), "NotificationDestination")
	}

	invokerUnderTest.NotificationDestination = "http://golang.cafe/"
	err = invokerUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "OnboardingInformation.ApiInvokerPublicKey")
	}

	invokerUnderTest.OnboardingInformation.ApiInvokerPublicKey = "key"
	err = invokerUnderTest.Validate()
	assert.Nil(t, err)
}

func TestIsOnboarded(t *testing.T) {
	publicKey := "publicKey"
	invokerUnderTest := APIInvokerEnrolmentDetails{
		OnboardingInformation: OnboardingInformation{
			ApiInvokerPublicKey: publicKey,
		},
	}

	otherInvoker := APIInvokerEnrolmentDetails{
		OnboardingInformation: OnboardingInformation{
			ApiInvokerPublicKey: "otherPublicKey",
		},
	}
	assert.False(t, invokerUnderTest.IsOnboarded(otherInvoker))

	otherInvoker.OnboardingInformation.ApiInvokerPublicKey = publicKey
	assert.True(t, invokerUnderTest.IsOnboarded(otherInvoker))
}
