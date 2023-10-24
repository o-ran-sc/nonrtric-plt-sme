// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023: OpenInfra Foundation Europe
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

package providermanagement

import (
	"context"
	"fmt"
	"net/http"
	"path"

	echo "github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"oransc.org/nonrtric/r1-sme-manager/internal/common29122"
	provapi "oransc.org/nonrtric/r1-sme-manager/internal/providermanagementapi"
)

type ProviderManager struct {
	registeredProviders map[string]provapi.APIProviderEnrolmentDetails
	CapifProtocol       string
	CapifIPv4           common29122.Ipv4Addr
	CapifPort           common29122.Port
}

func NewProviderManager(capifProtocol string, capifIPv4 common29122.Ipv4Addr, capifPort common29122.Port) *ProviderManager {
	return &ProviderManager{
		registeredProviders: make(map[string]provapi.APIProviderEnrolmentDetails),
		CapifProtocol:       capifProtocol,
		CapifIPv4:           capifIPv4,
		CapifPort:           capifPort,
	}
}

func (pm *ProviderManager) PostRegistrations(ctx echo.Context) error {
	log.Info("Entering PostRegistrations")

	newProvider, err := getProviderFromRequest(ctx)
	if err != nil {
		errMsg := "Unable to register provider due to %s"
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/api-provider-management/v1/", pm.CapifProtocol, pm.CapifIPv4, pm.CapifPort)
	client, err := provapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	postRegistrationsJSONRequestBody := provapi.PostRegistrationsJSONRequestBody(newProvider)

	var rspProvider *provapi.PostRegistrationsResponse
	rspProvider, err = client.PostRegistrationsWithResponse(ctxHandler, postRegistrationsJSONRequestBody)

	if (err != nil) || (rspProvider.StatusCode() != http.StatusCreated) {
		msg := string(rspProvider.Body)
		log.Errorf("Error on PostRegistrationsWithResponse %s", msg)
		return sendCoreError(ctx, rspProvider.StatusCode(), msg)
	}

	rspAPIProviderEnrolmentDetails := *rspProvider.JSON201
	apiProvDomId := *rspAPIProviderEnrolmentDetails.ApiProvDomId

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://` + path.Join(uri, apiProvDomId))
	if err := ctx.JSON(http.StatusCreated, rspAPIProviderEnrolmentDetails); err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}
	return nil
}

func (pm *ProviderManager) DeleteRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {
	log.Infof("Entering DeleteRegistrationsRegistrationId registrationId %s", registrationId)

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/api-provider-management/v1/", pm.CapifProtocol, pm.CapifIPv4, pm.CapifPort)
	client, err := provapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	var rspProvider *provapi.DeleteRegistrationsRegistrationIdResponse
	rspProvider, err = client.DeleteRegistrationsRegistrationIdWithResponse(ctxHandler, registrationId)

	if (err != nil) || (rspProvider.StatusCode() != http.StatusNoContent) {
		msg := string(rspProvider.Body)
		log.Errorf("Error on DeleteRegistrationsRegistrationIdWithResponse %s", msg)
		return sendCoreError(ctx, rspProvider.StatusCode(), msg)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (pm *ProviderManager) PutRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {
	log.Info("Entering PutRegistrationsRegistrationId")

	updatedProvider, err := getProviderFromRequest(ctx)
	if err != nil {
		msg := "Unable to register provider due to %s"
		log.Errorf("Error on getProviderFromRequest %s", msg)
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(msg, err))
	}

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/api-provider-management/v1/", pm.CapifProtocol, pm.CapifIPv4, pm.CapifPort)
	client, err := provapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	putRegistrationsRegistrationIdJSONRequestBody := provapi.PutRegistrationsRegistrationIdJSONRequestBody(updatedProvider)

	var rspProvider *provapi.PutRegistrationsRegistrationIdResponse
	rspProvider, err = client.PutRegistrationsRegistrationIdWithResponse(ctxHandler, registrationId, putRegistrationsRegistrationIdJSONRequestBody)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on PutRegistrationsRegistrationIdWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	if rspProvider.StatusCode() != http.StatusOK {
		msg := string(rspProvider.Body)
		log.Errorf("Error on PutRegistrationsRegistrationIdWithResponse %s", msg)
		return sendCoreError(ctx, rspProvider.StatusCode(), msg)
	}

	rspAPIProviderEnrolmentDetails := *rspProvider.JSON200
	apiProvDomId := *rspAPIProviderEnrolmentDetails.ApiProvDomId

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://` + path.Join(uri, apiProvDomId))
	if err := ctx.JSON(http.StatusOK, rspAPIProviderEnrolmentDetails); err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}
	return nil
}

func (pm *ProviderManager) ModifyIndApiProviderEnrolment(ctx echo.Context, registrationId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func getProviderFromRequest(ctx echo.Context) (provapi.APIProviderEnrolmentDetails, error) {
	var updatedProvider provapi.APIProviderEnrolmentDetails
	err := ctx.Bind(&updatedProvider)
	if err != nil {
		return provapi.APIProviderEnrolmentDetails{}, fmt.Errorf("invalid format for provider")
	}
	return updatedProvider, nil
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
