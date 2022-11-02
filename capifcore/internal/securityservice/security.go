// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022: Nordix Foundation
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

package security

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	securityapi "oransc.org/nonrtric/capifcore/internal/securityapi"

	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"
	"oransc.org/nonrtric/capifcore/internal/publishservice"
)

type Security struct {
	serviceRegister providermanagement.ServiceRegister
	apiRegister     publishservice.APIRegister
	invokerRegister invokermanagement.InvokerRegister
}

func NewSecurity(serviceRegister providermanagement.ServiceRegister, apiRegister publishservice.APIRegister, invokerRegister invokermanagement.InvokerRegister) *Security {
	return &Security{
		serviceRegister: serviceRegister,
		apiRegister:     apiRegister,
		invokerRegister: invokerRegister,
	}
}

func (s *Security) PostSecuritiesSecurityIdToken(ctx echo.Context, securityId string) error {
	clientId := ctx.FormValue("client_id")
	clientSecret := ctx.FormValue("client_secret")
	scope := ctx.FormValue("scope")

	if !s.invokerRegister.IsInvokerRegistered(clientId) {
		return sendCoreError(ctx, http.StatusBadRequest, "Invoker not registered")
	}
	if !s.invokerRegister.VerifyInvokerSecret(clientId, clientSecret) {
		return sendCoreError(ctx, http.StatusBadRequest, "Invoker secret not valid")
	}
	if scope != "" {
		scopeData := strings.Split(strings.Split(scope, "#")[1], ":")
		if !s.serviceRegister.IsFunctionRegistered(scopeData[0]) {
			return sendCoreError(ctx, http.StatusBadRequest, "Function not registered")
		}
		if !s.apiRegister.IsAPIRegistered(scopeData[0], scopeData[1]) {
			return sendCoreError(ctx, http.StatusBadRequest, "API not published")
		}
	}

	accessTokenResp := securityapi.AccessTokenRsp{
		AccessToken: "asdadfsrt dsr t5",
		ExpiresIn:   0,
		Scope:       &scope,
		TokenType:   "Bearer",
	}

	err := ctx.JSON(http.StatusCreated, accessTokenResp)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (s *Security) DeleteTrustedInvokersApiInvokerId(ctx echo.Context, apiInvokerId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (s *Security) GetTrustedInvokersApiInvokerId(ctx echo.Context, apiInvokerId string, params securityapi.GetTrustedInvokersApiInvokerIdParams) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (s *Security) PutTrustedInvokersApiInvokerId(ctx echo.Context, apiInvokerId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (s *Security) PostTrustedInvokersApiInvokerIdDelete(ctx echo.Context, apiInvokerId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (s *Security) PostTrustedInvokersApiInvokerIdUpdate(ctx echo.Context, apiInvokerId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func sendCoreError(ctx echo.Context, code int, message string) error {
	pd := common29122.ProblemDetails{
		Cause:  &message,
		Status: &code,
	}
	err := ctx.JSON(code, pd)
	return err
}
