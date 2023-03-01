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
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	securityapi "oransc.org/nonrtric/capifcore/internal/securityapi"

	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/keycloak"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"
	"oransc.org/nonrtric/capifcore/internal/publishservice"
)

type Security struct {
	serviceRegister providermanagement.ServiceRegister
	publishRegister publishservice.PublishRegister
	invokerRegister invokermanagement.InvokerRegister
	keycloak        keycloak.AccessManagement
	trustedInvokers map[string]securityapi.ServiceSecurity
	lock            sync.Mutex
}

func NewSecurity(serviceRegister providermanagement.ServiceRegister, publishRegister publishservice.PublishRegister, invokerRegister invokermanagement.InvokerRegister, km keycloak.AccessManagement) *Security {
	return &Security{
		serviceRegister: serviceRegister,
		publishRegister: publishRegister,
		invokerRegister: invokerRegister,
		keycloak:        km,
		trustedInvokers: make(map[string]securityapi.ServiceSecurity),
	}
}

func (s *Security) PostSecuritiesSecurityIdToken(ctx echo.Context, securityId string) error {
	var accessTokenReq securityapi.AccessTokenReq
	accessTokenReq.GetAccessTokenReq(ctx)

	if valid, err := accessTokenReq.Validate(); !valid {
		return ctx.JSON(http.StatusBadRequest, err)
	}

	if !s.invokerRegister.IsInvokerRegistered(accessTokenReq.ClientId) {
		return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorInvalidClient, "Invoker not registered")
	}

	if !s.invokerRegister.VerifyInvokerSecret(accessTokenReq.ClientId, *accessTokenReq.ClientSecret) {
		return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorUnauthorizedClient, "Invoker secret not valid")
	}

	if accessTokenReq.Scope != nil && *accessTokenReq.Scope != "" {
		scope := strings.Split(*accessTokenReq.Scope, "#")
		aefList := strings.Split(scope[1], ";")
		for _, aef := range aefList {
			apiList := strings.Split(aef, ":")
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
	jwtToken, err := s.keycloak.GetToken(accessTokenReq.ClientId, *accessTokenReq.ClientSecret, *accessTokenReq.Scope, "invokerrealm")
	if err != nil {
		return sendAccessTokenError(ctx, http.StatusBadRequest, securityapi.AccessTokenErrErrorUnauthorizedClient, err.Error())
	}

	accessTokenResp := securityapi.AccessTokenRsp{
		AccessToken: jwtToken.AccessToken,
		ExpiresIn:   common29122.DurationSec(jwtToken.ExpiresIn),
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
	errMsg := "Unable to update security context due to %s."

	if !s.invokerRegister.IsInvokerRegistered(apiInvokerId) {
		return sendCoreError(ctx, http.StatusBadRequest, "Unable to update security context due to Invoker not registered")
	}
	serviceSecurity, err := getServiceSecurityFromRequest(ctx)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	if err := serviceSecurity.Validate(); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	err = s.prepareNewSecurityContext(&serviceSecurity, apiInvokerId)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, apiInvokerId))

	err = ctx.JSON(http.StatusCreated, s.trustedInvokers[apiInvokerId])
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func getServiceSecurityFromRequest(ctx echo.Context) (securityapi.ServiceSecurity, error) {
	var serviceSecurity securityapi.ServiceSecurity
	err := ctx.Bind(&serviceSecurity)
	if err != nil {
		return securityapi.ServiceSecurity{}, fmt.Errorf("invalid format for service security")
	}
	return serviceSecurity, nil
}

func (s *Security) prepareNewSecurityContext(newContext *securityapi.ServiceSecurity, apiInvokerId string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	err := newContext.PrepareNewSecurityContext(s.publishRegister.GetAllPublishedServices())
	if err != nil {
		return err
	}

	s.trustedInvokers[apiInvokerId] = *newContext
	return nil
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

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendCoreError(ctx echo.Context, code int, message string) error {
	pd := common29122.ProblemDetails{
		Cause:  &message,
		Status: &code,
	}
	err := ctx.JSON(code, pd)
	return err
}
