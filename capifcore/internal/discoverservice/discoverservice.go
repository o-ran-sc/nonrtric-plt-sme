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

package discoverservice

import (
	"fmt"
	"net/http"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	discoverapi "oransc.org/nonrtric/capifcore/internal/discoverserviceapi"
	"oransc.org/nonrtric/capifcore/internal/invokermanagement"

	"github.com/labstack/echo/v4"

	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"
)

type DiscoverService struct {
	invokerRegister invokermanagement.InvokerRegister
}

func NewDiscoverService(invokerRegister invokermanagement.InvokerRegister) *DiscoverService {
	return &DiscoverService{
		invokerRegister: invokerRegister,
	}
}

func (ds *DiscoverService) GetAllServiceAPIs(ctx echo.Context, params discoverapi.GetAllServiceAPIsParams) error {
	allApis := ds.invokerRegister.GetInvokerApiList(params.ApiInvokerId)
	if allApis == nil {
		return sendCoreError(ctx, http.StatusNotFound, fmt.Sprintf("Invoker %s not registered", params.ApiInvokerId))
	}

	filteredApis := []publishapi.ServiceAPIDescription{}
	for _, api := range *allApis {
		if matchesFilter(api, params) {
			filteredApis = append(filteredApis, api)
		}
	}
	discoveredApis := discoverapi.DiscoveredAPIs{
		ServiceAPIDescriptions: &filteredApis,
	}
	err := ctx.JSON(http.StatusOK, discoveredApis)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func matchesFilter(api publishapi.ServiceAPIDescription, filter discoverapi.GetAllServiceAPIsParams) bool {
	if filter.ApiName != nil && *filter.ApiName != api.ApiName {
		return false
	}
	if filter.ApiCat != nil && (api.ServiceAPICategory == nil || *filter.ApiCat != *api.ServiceAPICategory) {
		return false
	}
	profiles := *api.AefProfiles
	for _, profile := range profiles {
		if checkAefId(filter, profile) && checkVersionAndCommType(profile, filter) && checkProtocol(filter, profile) && checkDataFormat(filter, profile) {
			return true
		}
	}
	return false
}

func checkAefId(filter discoverapi.GetAllServiceAPIsParams, profile publishapi.AefProfile) bool {
	if filter.AefId != nil {
		return *filter.AefId == profile.AefId
	}
	return true
}

func checkVersionAndCommType(profile publishapi.AefProfile, filter discoverapi.GetAllServiceAPIsParams) bool {
	match := false
	if filter.ApiVersion != nil {
		for _, version := range profile.Versions {
			match = checkVersion(version, filter.ApiVersion, filter.CommType)
			if match {
				break
			}
		}
	} else if filter.CommType != nil {
		for _, version := range profile.Versions {
			match = checkCommType(version.Resources, filter.CommType)
		}
	} else {
		match = true
	}
	return match
}

func checkProtocol(filter discoverapi.GetAllServiceAPIsParams, profile publishapi.AefProfile) bool {
	if filter.Protocol != nil {
		return profile.Protocol != nil && *filter.Protocol == *profile.Protocol
	}
	return true
}

func checkDataFormat(filter discoverapi.GetAllServiceAPIsParams, profile publishapi.AefProfile) bool {
	if filter.DataFormat != nil {
		return profile.DataFormat != nil && *filter.DataFormat == *profile.DataFormat
	}
	return true
}

func checkVersion(version publishapi.Version, wantedVersion *string, commType *publishapi.CommunicationType) bool {
	match := false
	if *wantedVersion == version.ApiVersion {
		if commType != nil {
			match = checkCommType(version.Resources, commType)
		} else {
			match = true
		}
	}
	return match
}

func checkCommType(resources *[]publishapi.Resource, commType *publishapi.CommunicationType) bool {
	for _, resource := range *resources {
		if resource.CommType == *commType {
			return true
		}
	}
	return false
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
