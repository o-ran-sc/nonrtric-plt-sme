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
	"github.com/labstack/echo/v4"
)

func (tokenReq *AccessTokenReq) GetAccessTokenReq(ctx echo.Context) {
	clientId := ctx.FormValue("client_id")
	clientSecret := ctx.FormValue("client_secret")
	scope := ctx.FormValue("scope")
	grantType := ctx.FormValue("grant_type")

	tokenReq.ClientId = clientId
	tokenReq.ClientSecret = &clientSecret
	tokenReq.Scope = &scope
	tokenReq.GrantType = AccessTokenReqGrantType(grantType)

}