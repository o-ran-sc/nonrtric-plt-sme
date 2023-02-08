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

package securityapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateClientIdNotPresent(t *testing.T) {
	accessTokenUnderTest := AccessTokenReq{}
	valid, err := accessTokenUnderTest.Validate()

	assert.Equal(t, false, valid)
	assert.Equal(t, AccessTokenErrErrorInvalidRequest, err.Error)
	assert.Equal(t, "Invalid request", *err.ErrorDescription)
}

func TestValidateGrantType(t *testing.T) {
	accessTokenUnderTest := AccessTokenReq{
		ClientId:  "clientId",
		GrantType: AccessTokenReqGrantType(""),
	}
	valid, err := accessTokenUnderTest.Validate()

	assert.Equal(t, false, valid)
	assert.Equal(t, AccessTokenErrErrorInvalidGrant, err.Error)
	assert.Equal(t, "Invalid value for grant_type", *err.ErrorDescription)

	accessTokenUnderTest.GrantType = AccessTokenReqGrantType("client_credentials")
	valid, err = accessTokenUnderTest.Validate()
	assert.Equal(t, true, valid)
}

func TestValidateScopeNotValid(t *testing.T) {
	scope := "scope#aefId:path"
	accessTokenUnderTest := AccessTokenReq{
		ClientId:  "clientId",
		GrantType: ("client_credentials"),
		Scope:     &scope,
	}
	valid, err := accessTokenUnderTest.Validate()

	assert.Equal(t, false, valid)
	assert.Equal(t, AccessTokenErrErrorInvalidScope, err.Error)
	assert.Equal(t, "Scope should start with 3gpp", *err.ErrorDescription)

	scope = "3gpp#aefId:path"
	accessTokenUnderTest.Scope = &scope
	valid, err = accessTokenUnderTest.Validate()
	assert.Equal(t, true, valid)
}

func TestValidateScopeMalformed(t *testing.T) {
	scope := "3gpp"
	accessTokenUnderTest := AccessTokenReq{
		ClientId:  "clientId",
		GrantType: ("client_credentials"),
		Scope:     &scope,
	}
	valid, err := accessTokenUnderTest.Validate()

	assert.Equal(t, false, valid)
	assert.Equal(t, AccessTokenErrErrorInvalidScope, err.Error)
	assert.Equal(t, "Malformed scope", *err.ErrorDescription)

	scope = "3gpp#aefId"
	accessTokenUnderTest.Scope = &scope
	valid, err = accessTokenUnderTest.Validate()
	assert.Equal(t, false, valid)
	assert.Equal(t, AccessTokenErrErrorInvalidScope, err.Error)
	assert.Equal(t, "Malformed scope", *err.ErrorDescription)

	scope = "3gpp#aefId:path"
	accessTokenUnderTest.Scope = &scope
	valid, err = accessTokenUnderTest.Validate()
	assert.Equal(t, true, valid)
}
