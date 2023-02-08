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
package publishserviceapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	serviceDescriptionUnderTest := ServiceAPIDescription{}
	err := serviceDescriptionUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "missing")
		assert.Contains(t, err.Error(), "apiName")
	}

	serviceDescriptionUnderTest.ApiName = "apiName"
	err = serviceDescriptionUnderTest.Validate()
	assert.Nil(t, err)

}

func TestValidateAlreadyPublished(t *testing.T) {
	apiName := "apiName"
	serviceUnderTest := ServiceAPIDescription{
		ApiName: apiName,
	}

	otherService := ServiceAPIDescription{
		ApiName: "otherApiName",
	}
	assert.Nil(t, serviceUnderTest.ValidateAlreadyPublished(otherService))

	otherService.ApiName = apiName
	assert.NotNil(t, serviceUnderTest.ValidateAlreadyPublished(otherService))
}
