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
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	securityapi "oransc.org/nonrtric/capifcore/internal/securityapi"

	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"
	"oransc.org/nonrtric/capifcore/internal/publishservice"
)

type AccessTokenReq struct {
	ClientId     string                              `json:"client_id" form:"client_id"`
	ClientSecret *string                             `json:"client_secret,omitempty" form:"client_secret"`
	GrantType    securityapi.AccessTokenReqGrantType `json:"grant_type" form:"grant_type"`
	Scope        *string                             `json:"scope,omitempty" form:"scope"`
}

var jwtKey = "my-secret-key"

type Security struct {
	serviceRegister providermanagement.ServiceRegister
	publishRegister publishservice.PublishRegister
	invokerRegister invokermanagement.InvokerRegister
}

func NewSecurity(serviceRegister providermanagement.ServiceRegister, publishRegister publishservice.PublishRegister, invokerRegister invokermanagement.InvokerRegister) *Security {
	return &Security{
		serviceRegister: serviceRegister,
		publishRegister: publishRegister,
		invokerRegister: invokerRegister,
	}
}

func (s *Security) PostSecuritiesSecurityIdToken(ctx echo.Context, securityId string) error {
	var accessTokenReq AccessTokenReq

	if err := ctx.Bind(&accessTokenReq); err != nil {
		return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidRequest, "Invalid request")
	}

	if accessTokenReq.GrantType != securityapi.AccessTokenReqGrantTypeClientCredentials {
		return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorUnsupportedGrantType, "Invalid value for grant_type")
	}

	if !s.invokerRegister.IsInvokerRegistered(accessTokenReq.ClientId) {
		return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidClient, "Invoker not registered")
	}

	if accessTokenReq.ClientSecret != nil {
		if !s.invokerRegister.VerifyInvokerSecret(accessTokenReq.ClientId, *accessTokenReq.ClientSecret) {
			return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorUnauthorizedClient, "Invoker secret not valid")
		}
	}
	//3gpp#aefId1:apiName1,apiName2,…apiNameX;aefId2:apiName1,apiName2,…apiNameY;…aefIdN:apiName1,apiName2,…apiNameZ
	if accessTokenReq.Scope != nil {
		scope := strings.Split(*accessTokenReq.Scope, "#")
		if len(scope) < 2 {
			return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidScope, "Malformed scope")
		}
		if scope[0] != "3gpp" {
			return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidScope, "Scope should start with 3gpp")
		}

		aefList := strings.Split(scope[1], ";")
		for _, aef := range aefList {
			apiList := strings.Split(aef, ":")
			if len(apiList) < 2 {
				return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidScope, "Malformed scope")
			}
			if !s.serviceRegister.IsFunctionRegistered(apiList[0]) {
				return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidScope, "AEF Function not registered")
			}
			for _, api := range strings.Split(apiList[1], ",") {
				if !s.publishRegister.IsAPIPublished(apiList[0], api) {
					return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidScope, "API not published")
				}
			}
		}
	}

	expirationTime := time.Now().Add(time.Hour).Unix()

	claims := &jwt.MapClaims{
		"iss": accessTokenReq.ClientId,
		"exp": expirationTime,
		"data": map[string]interface{}{
			"scope": accessTokenReq.Scope,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		return err
	}

	accessTokenResp := securityapi.AccessTokenRsp{
		AccessToken: tokenString,
		ExpiresIn:   common29122.DurationSec(expirationTime),
		Scope:       accessTokenReq.Scope,
		TokenType:   "Bearer",
	}

	err = ctx.JSON(http.StatusCreated, accessTokenResp)
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

func sendAccessTokenError(ctx echo.Context, code int, err securityapi.AccessTokenErrError, message string) error {
	accessTokenErr := securityapi.AccessTokenErr{
		Error:            err,
		ErrorDescription: &message,
	}
	return ctx.JSON(code, accessTokenErr)
}
