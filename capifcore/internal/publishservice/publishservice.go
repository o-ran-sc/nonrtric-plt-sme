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

package publishservice

import (
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/helmmanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"

	log "github.com/sirupsen/logrus"
)

//go:generate mockery --name APIRegister
type APIRegister interface {
	AreAPIsRegistered(serviceDescriptions *[]publishserviceapi.ServiceAPIDescription) bool
	GetAPIs() *[]publishserviceapi.ServiceAPIDescription
	IsAPIRegistered(aefId, path string) bool
}

type PublishService struct {
	publishedServices map[string]publishserviceapi.ServiceAPIDescription
	serviceRegister   providermanagement.ServiceRegister
	helmManager       helmmanagement.HelmManager
	lock              sync.Mutex
}

func NewPublishService(serviceRegister providermanagement.ServiceRegister, hm helmmanagement.HelmManager) *PublishService {
	return &PublishService{
		helmManager:       hm,
		publishedServices: make(map[string]publishserviceapi.ServiceAPIDescription),
		serviceRegister:   serviceRegister,
	}
}

func (ps *PublishService) AreAPIsRegistered(serviceDescriptions *[]publishserviceapi.ServiceAPIDescription) bool {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	allRegistered := true
	if serviceDescriptions != nil {
	out:
		for _, newApi := range *serviceDescriptions {
			registeredApi, ok := ps.publishedServices[*newApi.ApiId]
			if ok {
				if !ps.areProfilesRegistered(newApi.AefProfiles, registeredApi.AefProfiles) {
					allRegistered = false
					break out
				}
			} else {
				allRegistered = false
				break out
			}
		}
	}
	return allRegistered
}

func (ps *PublishService) areProfilesRegistered(newProfiles *[]publishserviceapi.AefProfile, registeredProfiles *[]publishserviceapi.AefProfile) bool {
	allRegistered := true
	if newProfiles != nil && registeredProfiles != nil {
	out:
		for _, newProfile := range *newProfiles {
			for _, registeredProfile := range *registeredProfiles {
				if newProfile.AefId == registeredProfile.AefId {
					break
				}
				allRegistered = false
				break out
			}
		}
	} else if registeredProfiles == nil {
		allRegistered = false
	}
	return allRegistered
}

func (ps *PublishService) GetAPIs() *[]publishserviceapi.ServiceAPIDescription {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	apis := []publishserviceapi.ServiceAPIDescription{}
	for _, service := range ps.publishedServices {
		apis = append(apis, service)
	}
	return &apis
}

func (ps *PublishService) IsAPIRegistered(aefId, path string) bool {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	registered := false
out:
	for _, service := range ps.publishedServices {
		if service.ApiName == path {
			for _, profile := range *service.AefProfiles {
				if profile.AefId == aefId {
					registered = true
					break out
				}
			}
		}
	}
	return registered
}

func (ps *PublishService) GetApfIdServiceApis(ctx echo.Context, apfId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (ps *PublishService) PostApfIdServiceApis(ctx echo.Context, apfId string) error {
	var newServiceAPIDescription publishserviceapi.ServiceAPIDescription
	err := ctx.Bind(&newServiceAPIDescription)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, "Invalid format for service")
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, profile := range *newServiceAPIDescription.AefProfiles {
		if !ps.serviceRegister.IsFunctionRegistered(profile.AefId) {
			return sendCoreError(ctx, http.StatusNotFound, "Function not registered, "+profile.AefId)
		}
	}

	newId := "api_id_" + newServiceAPIDescription.ApiName
	newServiceAPIDescription.ApiId = &newId
	info := strings.Split(*newServiceAPIDescription.Description, ",")
	if len(info) == 5 {
		err = ps.helmManager.InstallHelmChart(info[1], info[2], info[3], info[4])
		if err != nil {
			return sendCoreError(ctx, http.StatusBadRequest, "Unable to install Helm chart due to: "+err.Error())
		}
		log.Info("Installed service: ", newId)
	}
	ps.publishedServices[*newServiceAPIDescription.ApiId] = newServiceAPIDescription

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, *newServiceAPIDescription.ApiId))
	err = ctx.JSON(http.StatusCreated, newServiceAPIDescription)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (ps *PublishService) DeleteApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	serviceDescription, ok := ps.publishedServices[string(serviceApiId)]
	if ok {
		info := strings.Split(*serviceDescription.Description, ",")
		if len(info) == 5 {
			ps.helmManager.UninstallHelmChart(info[1], info[3])
			log.Info("Deleted service: ", serviceApiId)
		}
		delete(ps.publishedServices, string(serviceApiId))
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (ps *PublishService) GetApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	serviceDescription, ok := ps.publishedServices[string(serviceApiId)]
	if ok {
		err := ctx.JSON(http.StatusOK, serviceDescription)
		if err != nil {
			// Something really bad happened, tell Echo that our handler failed
			return err
		}

		return nil
	}
	return ctx.NoContent(http.StatusNotFound)
}

func (ps *PublishService) ModifyIndAPFPubAPI(ctx echo.Context, apfId string, serviceApiId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (ps *PublishService) PutApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
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
