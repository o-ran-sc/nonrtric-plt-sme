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

package discoverservice

import (
	"context"
	"fmt"
	"net/http"

	"oransc.org/nonrtric/servicemanager/internal/common29122"
	discoverapi "oransc.org/nonrtric/servicemanager/internal/discoverserviceapi"

	echo "github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

type DiscoverService struct {
	CapifProtocol string
	CapifIPv4     common29122.Ipv4Addr
	CapifPort     common29122.Port
}

func NewDiscoverService(capifProtocol string, capifIPv4 common29122.Ipv4Addr, capifPort common29122.Port) *DiscoverService {
	return &DiscoverService{
		CapifProtocol: capifProtocol,
		CapifIPv4:     capifIPv4,
		CapifPort:     capifPort,
	}
}

func (ds *DiscoverService) GetAllServiceAPIs(ctx echo.Context, params discoverapi.GetAllServiceAPIsParams) error {
	log.Trace("entering GetAllServiceAPIs")

	capifcoreUrl := fmt.Sprintf("%s://%s:%d/service-apis/v1/", ds.CapifProtocol, ds.CapifIPv4, ds.CapifPort)
	client, err := discoverapi.NewClientWithResponses(capifcoreUrl)
	if err != nil {
		return err
	}

	var (
		ctxHandler context.Context
		cancel     context.CancelFunc
	)
	ctxHandler, cancel = context.WithCancel(context.Background())
	defer cancel()

	var rsp *discoverapi.GetAllServiceAPIsResponse
	rsp, err = client.GetAllServiceAPIsWithResponse(ctxHandler, &params)

	if err != nil {
		msg := err.Error()
		log.Errorf("error on GetAllServiceAPIsWithResponse %s", msg)
		return sendCoreError(ctx, http.StatusInternalServerError, msg)
	}

	if rsp.StatusCode() != http.StatusOK {
		msg := string(rsp.Body)
		log.Errorf("GetAllServiceAPIs status code %d", rsp.StatusCode())
		log.Errorf("GetAllServiceAPIs error %s", string(rsp.Body))
		return sendCoreError(ctx, rsp.StatusCode(), msg)
	}

	rspDiscoveredAPIs := *rsp.JSON200
	err = ctx.JSON(http.StatusOK, rspDiscoveredAPIs)
	if err != nil {
		return err // tell Echo that our handler failed
	}
	return nil
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
