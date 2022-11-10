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
		return sendCoreError(ctx, http.StatusNotFound, "Invoker not registered")
	}
	filteredApis := []publishapi.ServiceAPIDescription{}
	gatewayDomain := "r1-expo-func-aef"
	for _, api := range *allApis {
		if !matchesFilter(api, params) {
			continue
		}
		profiles := *api.AefProfiles
		for i, profile := range profiles {
			profile.DomainName = &gatewayDomain // Hardcoded for now. Should be provided through some other mechanism.
			profiles[i] = profile
		}
		filteredApis = append(filteredApis, api)
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
	aefIdMatch := true
	protocolMatch := true
	dataFormatMatch := true
	versionMatch := true
	for _, profile := range profiles {
		if filter.AefId != nil {
			aefIdMatch = *filter.AefId == profile.AefId
		}
		if filter.ApiVersion != nil || filter.CommType != nil {
			versionMatch = checkVersionAndCommType(profile, filter.ApiVersion, filter.CommType)
		}
		if filter.Protocol != nil {
			protocolMatch = profile.Protocol != nil && *filter.Protocol == *profile.Protocol
		}
		if filter.DataFormat != nil {
			dataFormatMatch = profile.DataFormat != nil && *filter.DataFormat == *profile.DataFormat
		}
		if aefIdMatch && versionMatch && protocolMatch && dataFormatMatch {
			return true
		}
	}
	return false
}

func checkVersionAndCommType(profile publishapi.AefProfile, wantedVersion *string, commType *publishapi.CommunicationType) bool {
	match := false
	if wantedVersion != nil {
		for _, version := range profile.Versions {
			match = checkVersion(version, wantedVersion, commType)
			if match {
				break
			}
		}
	} else if commType != nil {
		for _, version := range profile.Versions {
			match = checkCommType(version.Resources, commType)
		}
	} else {
		match = true
	}
	return match
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
	match := false
	if commType != nil {
		for _, resource := range *resources {
			if resource.CommType == *commType {
				match = true
				break
			}
		}
	} else {
		match = true
	}
	return match
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
