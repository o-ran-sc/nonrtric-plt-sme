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
	"strings"
)

func (tokenReq AccessTokenReq) Validate() (bool, AccessTokenErr) {

	if tokenReq.ClientId == "" {
		return false, createAccessTokenError(AccessTokenErrErrorInvalidRequest, "Invalid request")
	}

	if tokenReq.GrantType != AccessTokenReqGrantTypeClientCredentials {
		return false, createAccessTokenError(AccessTokenErrErrorInvalidGrant, "Invalid value for grant_type")
	}

	//3gpp#aefId1:apiName1,apiName2,…apiNameX;aefId2:apiName1,apiName2,…apiNameY;…aefIdN:apiName1,apiName2,…apiNameZ
	if tokenReq.Scope != nil && *tokenReq.Scope != "" {
		scope := strings.Split(*tokenReq.Scope, "#")
		if len(scope) < 2 {
			return false, createAccessTokenError(AccessTokenErrErrorInvalidScope, "Malformed scope")
		}
		if scope[0] != "3gpp" {
			return false, createAccessTokenError(AccessTokenErrErrorInvalidScope, "Scope should start with 3gpp")
		}
		aefList := strings.Split(scope[1], ";")
		for _, aef := range aefList {
			apiList := strings.Split(aef, ":")
			if len(apiList) < 2 {
				return false, createAccessTokenError(AccessTokenErrErrorInvalidScope, "Malformed scope")
			}
		}
	}
	return true, AccessTokenErr{}
}

func createAccessTokenError(err AccessTokenErrError, message string) AccessTokenErr {
	return AccessTokenErr{
		Error:            err,
		ErrorDescription: &message,
	}
}
