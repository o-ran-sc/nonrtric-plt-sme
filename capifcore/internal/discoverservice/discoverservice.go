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

	discoverapi "oransc.org/nonrtric/capifcore/internal/discoverserviceapi"

	"oransc.org/nonrtric/capifcore/internal/publishservice"

	"github.com/labstack/echo/v4"

	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"
)

type DiscoverService struct {
	apiRegister publishservice.APIRegister
}

func NewDiscoverService(apiRegister publishservice.APIRegister) *DiscoverService {
	return &DiscoverService{
		apiRegister: apiRegister,
	}
}

func (ds *DiscoverService) GetAllServiceAPIs(ctx echo.Context, params discoverapi.GetAllServiceAPIsParams) error {
	allApis := *ds.apiRegister.GetAPIs()
	filteredApis := []publishapi.ServiceAPIDescription{}
	gatewayDomain := "r1-expo-func-aef"
	for _, api := range allApis {
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
	profiles := *api.AefProfiles
	match := false
	for _, profile := range profiles {
		if filter.ApiVersion != nil {
			for _, version := range profile.Versions {
				if *filter.ApiVersion == version.ApiVersion {
					match = true
				}
			}
		} else {
			match = true
		}
	}
	return match
}
