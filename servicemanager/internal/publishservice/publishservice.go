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

package publishservice

import (
	"context"
	"fmt"
	"net/http"
	"path"

	echo "github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"oransc.org/nonrtric/servicemanager/internal/common29122"
	publishapi "oransc.org/nonrtric/servicemanager/internal/publishserviceapi"
)

type PublishService struct {
	KongDomain				string;
	KongProtocol			string;
	KongControlPlanePort	common29122.Port;
	KongControlPlaneIPv4	common29122.Ipv4Addr;
	KongDataPlaneIPv4       common29122.Ipv4Addr;
	KongDataPlanePort 		common29122.Port;
	CapifProtocol			string;
	CapifIPv4        		common29122.Ipv4Addr;
	CapifPort		 		common29122.Port;
}

// Creates a service that implements both the PublishRegister and the publishserviceapi.ServerInterface interfaces.
func NewPublishService(
		kongDomain 				string,
		kongProtocol 			string,
		kongControlPlaneIPv4 	common29122.Ipv4Addr,
		kongControlPlanePort 	common29122.Port,
		kongDataPlaneIPv4 		common29122.Ipv4Addr,
		kongDataPlanePort 		common29122.Port,
		capifProtocol 			string,
		capifIPv4 				common29122.Ipv4Addr,
		capifPort 				common29122.Port) *PublishService {
	return &PublishService{
		KongDomain 				: kongDomain,
		KongProtocol			: kongProtocol,
		KongControlPlaneIPv4	: kongControlPlaneIPv4,
		KongControlPlanePort	: kongControlPlanePort,
		KongDataPlaneIPv4		: kongDataPlaneIPv4,
		KongDataPlanePort 		: kongDataPlanePort,
		CapifProtocol			: capifProtocol,
		CapifIPv4				: capifIPv4,
		CapifPort				: capifPort,
	}
}

// Publish a new API.
func (ps *PublishService) PostApfIdServiceApis(ctx echo.Context, apfId string) error {
	log.Tracef("entering PostApfIdServiceApis apfId %s", apfId)
	log.Debugf("PostApfIdServiceApis KongControlPlaneIPv4 %s", ps.KongControlPlaneIPv4)
	log.Debugf("PostApfIdServiceApis KongDataPlaneIPv4 %s", ps.KongDataPlaneIPv4)

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/published-apis/v1/", ps.CapifProtocol, ps.CapifIPv4, ps.CapifPort)
	client, err := publishapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	newServiceAPIDescription, err := getServiceFromRequest(ctx)
	if err != nil {
		return err
	}

	newServiceAPIDescription.PrepareNewService()

	statusCode, err := newServiceAPIDescription.RegisterKong(
			ps.KongDomain,
			ps.KongProtocol,
			ps.KongControlPlaneIPv4,
			ps.KongControlPlanePort,
			ps.KongDataPlaneIPv4,
			ps.KongDataPlanePort,
			apfId)
	if (err != nil) || (statusCode != http.StatusCreated) {
		// We can return with http.StatusForbidden if there is a http.StatusConflict detected by Kong
		msg := err.Error()
		log.Errorf("error on RegisterKong %s", msg)
		return sendCoreError(ctx, statusCode, msg)
	}

	bodyServiceAPIDescription := publishapi.PostApfIdServiceApisJSONRequestBody(newServiceAPIDescription)
	var rsp *publishapi.PostApfIdServiceApisResponse

	log.Trace("calling PostApfIdServiceApisWithResponse")
	rsp, err = client.PostApfIdServiceApisWithResponse(ctxHandler, apfId, bodyServiceAPIDescription)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on PostApfIdServiceApisWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	if rsp.StatusCode() != http.StatusCreated {
		msg := string(rsp.Body)
		log.Debugf("PostApfIdServiceApisWithResponse status code %d", rsp.StatusCode())
		log.Debugf("PostApfIdServiceApisWithResponse error %s", msg)
		if rsp.StatusCode() == http.StatusForbidden || rsp.StatusCode() == http.StatusBadRequest {
			newServiceAPIDescription.UnregisterKong(ps.KongDomain, ps.KongProtocol, ps.KongControlPlaneIPv4, ps.KongControlPlanePort)
		}
		return sendCoreError(ctx, rsp.StatusCode(), msg)
	}

	rspServiceAPIDescription := *rsp.JSON201
	apiId := *rspServiceAPIDescription.ApiId

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, apiId))

	err = ctx.JSON(http.StatusCreated, rspServiceAPIDescription)
	if err != nil {
		return err // Tell Echo that our handler failed
	}

	return nil
}


// Unpublish a published service API.
func (ps *PublishService) DeleteApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	log.Tracef("entering DeleteApfIdServiceApisServiceApiId apfId %s serviceApiId %s", apfId, serviceApiId)

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/published-apis/v1/", ps.CapifProtocol, ps.CapifIPv4, ps.CapifPort)
	client, err := publishapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	log.Debugf("call GetApfIdServiceApisServiceApiIdWithResponse before delete apfId %s serviceApiId %s", apfId, serviceApiId)
	var rsp *publishapi.GetApfIdServiceApisServiceApiIdResponse
	rsp, err = client.GetApfIdServiceApisServiceApiIdWithResponse(ctxHandler, apfId, serviceApiId)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on GetApfIdServiceApisServiceApiIdWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	statusCode := rsp.StatusCode()
	if statusCode != http.StatusOK {
		log.Debugf("GetApfIdServiceApisServiceApiIdWithResponse status %d", statusCode)
		return ctx.NoContent(statusCode)
	}

	rspServiceAPIDescription := *rsp.JSON200

	statusCode, err = rspServiceAPIDescription.UnregisterKong(ps.KongDomain, ps.KongProtocol, ps.KongControlPlaneIPv4, ps.KongControlPlanePort)
	if (err != nil) || (statusCode != http.StatusNoContent) {
		msg := err.Error()
		log.Errorf("error on UnregisterKong %s", msg)
		return sendCoreError(ctx, statusCode, msg)
	}

	log.Trace("call DeleteApfIdServiceApisServiceApiIdWithResponse")
	_, err = client.DeleteApfIdServiceApisServiceApiIdWithResponse(ctxHandler, apfId, serviceApiId)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on DeleteApfIdServiceApisServiceApiIdWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Retrieve all published APIs.
func (ps *PublishService) GetApfIdServiceApis(ctx echo.Context, apfId string) error {
	log.Tracef("entering GetApfIdServiceApis apfId %s", apfId)

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/published-apis/v1/", ps.CapifProtocol, ps.CapifIPv4, ps.CapifPort)
	client, err := publishapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	var rsp *publishapi.GetApfIdServiceApisResponse
	rsp, err = client.GetApfIdServiceApisWithResponse(ctxHandler, apfId)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on GetApfIdServiceApisWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	if rsp.StatusCode() != http.StatusOK {
		msg := string(rsp.Body)
		log.Errorf("GetApfIdServiceApisWithResponse status %d", rsp.StatusCode())
		log.Errorf("GetApfIdServiceApisWithResponse error %s", msg)
		return sendCoreError(ctx, rsp.StatusCode(), msg)
	}

	rspServiceAPIDescriptions := *rsp.JSON200
	err = ctx.JSON(rsp.StatusCode(), rspServiceAPIDescriptions)
	if err != nil {
		return err // tell Echo that our handler failed
	}
	return nil
}

// Retrieve a published service API.
func (ps *PublishService) GetApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	log.Tracef("entering GetApfIdServiceApisServiceApiId apfId %s", apfId)

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/published-apis/v1/", ps.CapifProtocol, ps.CapifIPv4, ps.CapifPort)
	client, err := publishapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	var rsp *publishapi.GetApfIdServiceApisServiceApiIdResponse
	rsp, err = client.GetApfIdServiceApisServiceApiIdWithResponse(ctxHandler, apfId, serviceApiId)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on GetApfIdServiceApisServiceApiIdWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	statusCode := rsp.StatusCode()
	if statusCode != http.StatusOK {
		return ctx.NoContent(statusCode)
	}

	rspServiceAPIDescription := *rsp.JSON200

	err = ctx.JSON(http.StatusOK, rspServiceAPIDescription)
	if err != nil {
		return err // tell Echo that our handler failed
	}
	return nil
}

// Modify an existing published service API.
func (ps *PublishService) ModifyIndAPFPubAPI(ctx echo.Context, apfId string, serviceApiId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Update a published service API.
func (ps *PublishService) PutApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	log.Tracef("entering PutApfIdServiceApisServiceApiId apfId %s", apfId)

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/published-apis/v1/", ps.CapifProtocol, ps.CapifIPv4, ps.CapifPort)
	client, err := publishapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	updatedServiceDescription, err := getServiceFromRequest(ctx)
	if err != nil {
		return err
	}

	var rsp *publishapi.PutApfIdServiceApisServiceApiIdResponse
	bodyServiceAPIDescription := publishapi.PutApfIdServiceApisServiceApiIdJSONRequestBody(updatedServiceDescription)

	rsp, err = client.PutApfIdServiceApisServiceApiIdWithResponse(ctxHandler, apfId, serviceApiId, bodyServiceAPIDescription)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on PutApfIdServiceApisServiceApiIdWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	if rsp.StatusCode() != http.StatusOK {
		log.Errorf("PutApfIdServiceApisServiceApiIdWithResponse status code %d", rsp.StatusCode())
		if rsp.StatusCode() == http.StatusBadRequest {
			updatedServiceDescription.UnregisterKong(ps.KongDomain, ps.KongProtocol,ps.KongControlPlaneIPv4, ps.KongControlPlanePort)
		}
		msg := string(rsp.Body)
		return sendCoreError(ctx, rsp.StatusCode(), msg)
	}

	rspServiceAPIDescription := *rsp.JSON200
	apiId := *rspServiceAPIDescription.ApiId

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, apiId))

	err = ctx.JSON(http.StatusOK, rspServiceAPIDescription)
	if err != nil {
		return err // Tell Echo that our handler failed
	}
	return nil
}


func getServiceFromRequest(ctx echo.Context) (publishapi.ServiceAPIDescription, error) {
	var updatedServiceDescription publishapi.ServiceAPIDescription
	err := ctx.Bind(&updatedServiceDescription)
	if err != nil {
		return publishapi.ServiceAPIDescription{}, fmt.Errorf("invalid format for service")
	}
	return updatedServiceDescription, nil
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