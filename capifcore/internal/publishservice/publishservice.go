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
	"k8s.io/utils/strings/slices"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/helmmanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"

	log "github.com/sirupsen/logrus"
)

//go:generate mockery --name APIRegister
type APIRegister interface {
	AreAPIsRegistered(serviceDescriptions *[]publishserviceapi.ServiceAPIDescription) bool
	IsAPIRegistered(aefId, path string) bool
}

type PublishService struct {
	publishedServices map[string][]*publishserviceapi.ServiceAPIDescription
	serviceRegister   providermanagement.ServiceRegister
	helmManager       helmmanagement.HelmManager
	lock              sync.Mutex
}

func NewPublishService(serviceRegister providermanagement.ServiceRegister, hm helmmanagement.HelmManager) *PublishService {
	return &PublishService{
		helmManager:       hm,
		publishedServices: make(map[string][]*publishserviceapi.ServiceAPIDescription),
		serviceRegister:   serviceRegister,
	}
}

func (ps *PublishService) AreAPIsRegistered(serviceDescriptions *[]publishserviceapi.ServiceAPIDescription) bool {

	if serviceDescriptions != nil {
		registeredApis := ps.getAllAefIds()
		return checkNewDescriptions(*serviceDescriptions, registeredApis)
	}
	return true
}

func (ps *PublishService) getAllAefIds() []string {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	allIds := []string{}
	for _, descriptions := range ps.publishedServices {
		for _, description := range descriptions {
			allIds = append(allIds, getIdsFromDescription(*description)...)
		}
	}
	return allIds
}

func getIdsFromDescription(description publishserviceapi.ServiceAPIDescription) []string {
	allIds := []string{}
	if description.AefProfiles != nil {
		for _, aefProfile := range *description.AefProfiles {
			allIds = append(allIds, aefProfile.AefId)
		}
	}
	return allIds
}

func checkNewDescriptions(newDescriptions []publishserviceapi.ServiceAPIDescription, registeredAefIds []string) bool {
	registered := true
	for _, newApi := range newDescriptions {
		if !checkProfiles(newApi.AefProfiles, registeredAefIds) {
			registered = false
			break
		}
	}
	return registered
}

func checkProfiles(newProfiles *[]publishserviceapi.AefProfile, registeredAefIds []string) bool {
	allRegistered := true
	if newProfiles != nil {
		for _, profile := range *newProfiles {
			if !slices.Contains(registeredAefIds, profile.AefId) {
				allRegistered = false
				break
			}
		}
	}
	return allRegistered
}

func (ps *PublishService) IsAPIRegistered(aefId, path string) bool {
	return slices.Contains(ps.getAllAefIds(), aefId)
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

	registeredFuncs := ps.serviceRegister.GetAefsForPublisher(apfId)
	for _, profile := range *newServiceAPIDescription.AefProfiles {
		if !slices.Contains(registeredFuncs, profile.AefId) {
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
	_, ok := ps.publishedServices[apfId]
	if ok {
		ps.publishedServices[apfId] = append(ps.publishedServices[apfId], &newServiceAPIDescription)
	} else {
		ps.publishedServices[apfId] = append([]*publishserviceapi.ServiceAPIDescription{}, &newServiceAPIDescription)
	}

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
	serviceDescriptions, ok := ps.publishedServices[string(apfId)]
	if ok {
		pos, description := getServiceDescription(serviceApiId, serviceDescriptions)
		if description != nil {
			info := strings.Split(*description.Description, ",")
			if len(info) == 5 {
				ps.helmManager.UninstallHelmChart(info[1], info[3])
				log.Info("Deleted service: ", serviceApiId)
			}
			ps.lock.Lock()
			defer ps.lock.Unlock()
			ps.publishedServices[string(apfId)] = removeServiceDescription(pos, serviceDescriptions)
		}
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (ps *PublishService) GetApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	serviceDescriptions, ok := ps.publishedServices[apfId]
	if ok {
		_, serviceDescription := getServiceDescription(serviceApiId, serviceDescriptions)
		if serviceDescription == nil {
			return ctx.NoContent(http.StatusNotFound)
		}
		err := ctx.JSON(http.StatusOK, serviceDescription)
		if err != nil {
			// Something really bad happened, tell Echo that our handler failed
			return err
		}

		return nil
	}
	return ctx.NoContent(http.StatusNotFound)
}

func getServiceDescription(serviceApiId string, descriptions []*publishserviceapi.ServiceAPIDescription) (int, *publishserviceapi.ServiceAPIDescription) {
	for pos, description := range descriptions {
		if serviceApiId == *description.ApiId {
			return pos, description
		}
	}
	return -1, nil
}

func removeServiceDescription(i int, a []*publishserviceapi.ServiceAPIDescription) []*publishserviceapi.ServiceAPIDescription {
	a[i] = a[len(a)-1] // Copy last element to index i.
	a[len(a)-1] = nil  // Erase last element (write zero value).
	a = a[:len(a)-1]   // Truncate slice.
	return a
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
