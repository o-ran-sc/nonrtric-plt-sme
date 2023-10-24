// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023-2024: OpenInfra Foundation Europe
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

package invokermanagement

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"oransc.org/nonrtric/servicemanager/internal/common29122"
	invokerapi "oransc.org/nonrtric/servicemanager/internal/invokermanagementapi"

	echo "github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type InvokerManager struct {
	CapifProtocol string
	CapifIPv4     common29122.Ipv4Addr
	CapifPort     common29122.Port
}

// Creates a manager that implements both the InvokerRegister and the invokermanagementapi.ServerInterface interfaces.
func NewInvokerManager(capifProtocol string, capifIPv4 common29122.Ipv4Addr, capifPort common29122.Port) *InvokerManager {
	return &InvokerManager{
		CapifProtocol: capifProtocol,
		CapifIPv4:     capifIPv4,
		CapifPort:     capifPort,
	}
}

// Creates a new individual API Invoker profile.
func (im *InvokerManager) PostOnboardedInvokers(ctx echo.Context) error {
	log.Trace("entering PostOnboardedInvokers")

	var newInvoker invokerapi.APIInvokerEnrolmentDetails
	errMsg := "Unable to onboard invoker due to %s"
	if err := ctx.Bind(&newInvoker); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, "invalid format for invoker"))
	}

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/api-invoker-management/v1/", im.CapifProtocol, im.CapifIPv4, im.CapifPort)
	client, err := invokerapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	postRegistrationsJSONRequestBody := invokerapi.PostOnboardedInvokersJSONRequestBody(newInvoker)

	var rspInvoker *invokerapi.PostOnboardedInvokersResponse
	rspInvoker, err = client.PostOnboardedInvokersWithResponse(ctxHandler, postRegistrationsJSONRequestBody)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on PostOnboardedInvokersWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	if rspInvoker.StatusCode() != http.StatusCreated {
		msg := string(rspInvoker.Body)
		log.Errorf("error on PostOnboardedInvokersWithResponse %s", msg)
		return sendCoreError(ctx, rspInvoker.StatusCode(), msg)
	}

	rspAPIProviderEnrolmentDetails := *rspInvoker.JSON201
	apiInvokerId := *rspAPIProviderEnrolmentDetails.ApiInvokerId

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, apiInvokerId))
	err = ctx.JSON(http.StatusCreated, rspAPIProviderEnrolmentDetails)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

// Deletes an individual API Invoker.
func (im *InvokerManager) DeleteOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error {
	log.Tracef("entering DeleteOnboardedInvokersOnboardingId onboardingId %s", onboardingId)

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/api-invoker-management/v1/", im.CapifProtocol, im.CapifIPv4, im.CapifPort)
	client, err := invokerapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	_, err = client.DeleteOnboardedInvokersOnboardingId(ctxHandler, onboardingId)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on DeleteOnboardedInvokersOnboardingId %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	return ctx.NoContent(http.StatusNoContent)
}


// Updates an individual API invoker details.
func (im *InvokerManager) PutOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error {
	log.Tracef("entering DeleteOnboardedInvokersOnboardingId onboardingId %s", onboardingId)

	var invoker invokerapi.APIInvokerEnrolmentDetails
	errMsg := "Unable to update invoker due to %s"
	if err := ctx.Bind(&invoker); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, "invalid format for invoker"))
	}

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/api-invoker-management/v1/", im.CapifProtocol, im.CapifIPv4, im.CapifPort)
	client, err := invokerapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	putRegistrationsJSONRequestBody := invokerapi.PutOnboardedInvokersOnboardingIdJSONRequestBody(invoker)

	var rspInvoker *invokerapi.PutOnboardedInvokersOnboardingIdResponse
	rspInvoker, err = client.PutOnboardedInvokersOnboardingIdWithResponse(ctxHandler, onboardingId, putRegistrationsJSONRequestBody)

	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	if rspInvoker.StatusCode() != http.StatusOK {
		msg := string(rspInvoker.Body)
		return sendCoreError(ctx, rspInvoker.StatusCode(), msg)
	}

	rspAPIProviderEnrolmentDetails := *rspInvoker.JSON200
	apiInvokerId := *rspAPIProviderEnrolmentDetails.ApiInvokerId

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, apiInvokerId))
	err = ctx.JSON(http.StatusOK, rspAPIProviderEnrolmentDetails)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (im *InvokerManager) ModifyIndApiInvokeEnrolment(ctx echo.Context, onboardingId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
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
