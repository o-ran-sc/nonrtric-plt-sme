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

func TestGetExposedFunctionIds(t *testing.T) {
	providerUnderTest := getProvider()

	exposedFuncs := providerUnderTest.GetExposingFunctionIdsForPublisher(funcIdAPF)

	assert.Len(t, exposedFuncs, 1)
	assert.Equal(t, funcIdAEF, exposedFuncs[0])

	exposedFuncs = providerUnderTest.GetExposingFunctionIdsForPublisher("anyId")

	assert.Len(t, exposedFuncs, 0)
}

func TestIsFunctionRegistered(t *testing.T) {
	providerUnderTest := getProvider()

	registered := providerUnderTest.IsFunctionRegistered(funcIdAPF)

	assert.True(t, registered)

	registered = providerUnderTest.IsFunctionRegistered("anyID")

	assert.False(t, registered)
}
