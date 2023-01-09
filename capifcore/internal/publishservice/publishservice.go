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
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"k8s.io/utils/strings/slices"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/eventsapi"
	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/helmmanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"

	log "github.com/sirupsen/logrus"
)

//go:generate mockery --name PublishRegister
type PublishRegister interface {
	// Checks if the provided APIs are published.
	// Returns true if all provided APIs have been published, false otherwise.
	AreAPIsPublished(serviceDescriptions *[]publishapi.ServiceAPIDescription) bool
	// Checks if the provided API is published.
	// Returns true if the provided API has been published, false otherwise.
	IsAPIPublished(aefId, path string) bool
	// Gets all published APIs.
	// Returns a list of all APIs that has been published.
	GetAllPublishedServices() []publishapi.ServiceAPIDescription
}

type PublishService struct {
	publishedServices map[string][]publishapi.ServiceAPIDescription
	serviceRegister   providermanagement.ServiceRegister
	helmManager       helmmanagement.HelmManager
	eventChannel      chan<- eventsapi.EventNotification
	lock              sync.Mutex
}

// Creates a service that implements both the PublishRegister and the publishserviceapi.ServerInterface interfaces.
func NewPublishService(serviceRegister providermanagement.ServiceRegister, hm helmmanagement.HelmManager, eventChannel chan<- eventsapi.EventNotification) *PublishService {
	return &PublishService{
		helmManager:       hm,
		publishedServices: make(map[string][]publishapi.ServiceAPIDescription),
		serviceRegister:   serviceRegister,
		eventChannel:      eventChannel,
	}
}

func (ps *PublishService) AreAPIsPublished(serviceDescriptions *[]publishapi.ServiceAPIDescription) bool {

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
			allIds = append(allIds, getIdsFromDescription(description)...)
		}
	}
	return allIds
}

func getIdsFromDescription(description publishapi.ServiceAPIDescription) []string {
	allIds := []string{}
	if description.AefProfiles != nil {
		for _, aefProfile := range *description.AefProfiles {
			allIds = append(allIds, aefProfile.AefId)
		}
	}
	return allIds
}

func checkNewDescriptions(newDescriptions []publishapi.ServiceAPIDescription, registeredAefIds []string) bool {
	registered := true
	for _, newApi := range newDescriptions {
		if !checkProfiles(newApi.AefProfiles, registeredAefIds) {
			registered = false
			break
		}
	}
	return registered
}

func checkProfiles(newProfiles *[]publishapi.AefProfile, registeredAefIds []string) bool {
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

func (ps *PublishService) IsAPIPublished(aefId, path string) bool {
	return slices.Contains(ps.getAllAefIds(), aefId)
}

func (ps *PublishService) GetAllPublishedServices() []publishapi.ServiceAPIDescription {
	publishedDescriptions := []publishapi.ServiceAPIDescription{}
	for _, descriptions := range ps.publishedServices {
		publishedDescriptions = append(publishedDescriptions, descriptions...)
	}
	return publishedDescriptions
}

// Retrieve all published APIs.
func (ps *PublishService) GetApfIdServiceApis(ctx echo.Context, apfId string) error {
	serviceDescriptions, ok := ps.publishedServices[apfId]
	if ok {
		err := ctx.JSON(http.StatusOK, serviceDescriptions)
		if err != nil {
			// Something really bad happened, tell Echo that our handler failed
			return err
		}
	} else {
		return sendCoreError(ctx, http.StatusNotFound, fmt.Sprintf("Provider %s not registered", apfId))
	}

	return nil
}

// Publish a new API.
func (ps *PublishService) PostApfIdServiceApis(ctx echo.Context, apfId string) error {
	var newServiceAPIDescription publishapi.ServiceAPIDescription
	err := ctx.Bind(&newServiceAPIDescription)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, "Invalid format for service "+apfId)
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	registeredFuncs := ps.serviceRegister.GetAefsForPublisher(apfId)
	for _, profile := range *newServiceAPIDescription.AefProfiles {
		if !slices.Contains(registeredFuncs, profile.AefId) {
			return sendCoreError(ctx, http.StatusNotFound, fmt.Sprintf("Function %s not registered", profile.AefId))
		}
	}

	newId := "api_id_" + newServiceAPIDescription.ApiName
	newServiceAPIDescription.ApiId = &newId

	shouldReturn, returnValue := ps.installHelmChart(newServiceAPIDescription, ctx)
	if shouldReturn {
		return returnValue
	}
	go ps.sendEvent(newServiceAPIDescription, eventsapi.CAPIFEventSERVICEAPIAVAILABLE)

	_, ok := ps.publishedServices[apfId]
	if ok {
		ps.publishedServices[apfId] = append(ps.publishedServices[apfId], newServiceAPIDescription)
	} else {
		ps.publishedServices[apfId] = append([]publishapi.ServiceAPIDescription{}, newServiceAPIDescription)
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

func (ps *PublishService) installHelmChart(newServiceAPIDescription publishapi.ServiceAPIDescription, ctx echo.Context) (bool, error) {
	info := strings.Split(*newServiceAPIDescription.Description, ",")
	if len(info) == 5 {
		err := ps.helmManager.InstallHelmChart(info[1], info[2], info[3], info[4])
		if err != nil {
			return true, sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf("Unable to install Helm chart %s due to: %s", info[3], err.Error()))
		}
		log.Debug("Installed service: ", newServiceAPIDescription.ApiId)
	}
	return false, nil
}

// Unpublish a published service API.
func (ps *PublishService) DeleteApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	serviceDescriptions, ok := ps.publishedServices[string(apfId)]
	if ok {
		pos, description := getServiceDescription(serviceApiId, serviceDescriptions)
		if description != nil {
			info := strings.Split(*description.Description, ",")
			if len(info) == 5 {
				ps.helmManager.UninstallHelmChart(info[1], info[3])
				log.Debug("Deleted service: ", serviceApiId)
			}
			ps.lock.Lock()
			ps.publishedServices[string(apfId)] = removeServiceDescription(pos, serviceDescriptions)
			ps.lock.Unlock()
			go ps.sendEvent(*description, eventsapi.CAPIFEventSERVICEAPIUNAVAILABLE)
		}
	}
	return ctx.NoContent(http.StatusNoContent)
}

// Retrieve a published service API.
func (ps *PublishService) GetApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	ps.lock.Lock()
	serviceDescriptions, ok := ps.publishedServices[apfId]
	ps.lock.Unlock()

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

func getServiceDescription(serviceApiId string, descriptions []publishapi.ServiceAPIDescription) (int, *publishapi.ServiceAPIDescription) {
	for pos, description := range descriptions {
		if serviceApiId == *description.ApiId {
			return pos, &description
		}
	}
	return -1, nil
}

func removeServiceDescription(i int, a []publishapi.ServiceAPIDescription) []publishapi.ServiceAPIDescription {
	a[i] = a[len(a)-1]                               // Copy last element to index i.
	a[len(a)-1] = publishapi.ServiceAPIDescription{} // Erase last element (write zero value).
	a = a[:len(a)-1]                                 // Truncate slice.
	return a
}

// Modify an existing published service API.
func (ps *PublishService) ModifyIndAPFPubAPI(ctx echo.Context, apfId string, serviceApiId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

// Update a published service API.
func (ps *PublishService) PutApfIdServiceApisServiceApiId(ctx echo.Context, apfId string, serviceApiId string) error {
	pos, publishedService, shouldReturn, returnValue := ps.checkIfServiceIsPublished(apfId, serviceApiId, ctx)
	if shouldReturn {
		return returnValue
	}

	updatedServiceDescription, shouldReturn, returnValue := getServiceFromRequest(ctx)
	if shouldReturn {
		return returnValue
	}

	if updatedServiceDescription.Description != nil {
		ps.lock.Lock()
		defer ps.lock.Unlock()

		publishedService.Description = updatedServiceDescription.Description
		ps.publishedServices[apfId][pos] = publishedService
		go ps.sendEvent(publishedService, eventsapi.CAPIFEventSERVICEAPIUPDATE)
	}

	pos, shouldReturn, returnValue = ps.updateProfiles(pos, apfId, updatedServiceDescription, publishedService, ctx)
	if shouldReturn {
		return returnValue
	}

	err := ctx.JSON(http.StatusOK, ps.publishedServices[apfId][pos])
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}
	return nil
}

func (ps *PublishService) checkIfServiceIsPublished(apfId string, serviceApiId string, ctx echo.Context) (int, publishapi.ServiceAPIDescription, bool, error) {
	publishedServices, ok := ps.publishedServices[apfId]
	if !ok {
		return 0, publishapi.ServiceAPIDescription{}, true, sendCoreError(ctx, http.StatusBadRequest, "Service must be published before updating it")
	} else {
		for pos, description := range publishedServices {
			if *description.ApiId == serviceApiId {
				return pos, description, false, nil

			}

		}

	}
	return 0, publishapi.ServiceAPIDescription{}, true, sendCoreError(ctx, http.StatusBadRequest, "Service must be published before updating it")
}

func getServiceFromRequest(ctx echo.Context) (publishapi.ServiceAPIDescription, bool, error) {
	var updatedServiceDescription publishapi.ServiceAPIDescription
	err := ctx.Bind(&updatedServiceDescription)
	if err != nil {
		return publishapi.ServiceAPIDescription{}, true, sendCoreError(ctx, http.StatusBadRequest, "Invalid format for service")
	}
	return updatedServiceDescription, false, nil
}

func (ps *PublishService) updateProfiles(pos int, apfId string, updatedServiceDescription publishapi.ServiceAPIDescription, publishedService publishapi.ServiceAPIDescription, ctx echo.Context) (int, bool, error) {
	registeredFuncs := ps.serviceRegister.GetAefsForPublisher(apfId)
	for _, profile := range *updatedServiceDescription.AefProfiles {
		if !slices.Contains(registeredFuncs, profile.AefId) {
			return 0, false, sendCoreError(ctx, http.StatusNotFound, fmt.Sprintf("Function %s not registered", profile.AefId))
		}
		if ps.checkIfProfileIsNew(profile.AefId, *publishedService.AefProfiles) {

			publishedService.AefProfiles = ps.addProfile(profile, publishedService)
			ps.publishedServices[apfId][pos] = publishedService

		} else {
			pos, shouldReturn, returnValue := ps.updateProfile(profile, publishedService, ctx)
			if shouldReturn {
				return pos, true, returnValue
			}
		}

	}
	return 0, false, nil
}

func (ps *PublishService) sendEvent(service publishapi.ServiceAPIDescription, eventType eventsapi.CAPIFEvent) {
	apiIds := []string{*service.ApiId}
	apis := []publishapi.ServiceAPIDescription{service}
	event := eventsapi.EventNotification{
		EventDetail: &eventsapi.CAPIFEventDetail{
			ApiIds:                 &apiIds,
			ServiceAPIDescriptions: &apis,
		},
		Events: eventType,
	}
	ps.eventChannel <- event
}

func (ps *PublishService) checkIfProfileIsNew(aefId string, publishedPofiles []publishapi.AefProfile) bool {
	for _, profile := range publishedPofiles {
		if profile.AefId == aefId {
			return false
		}
	}
	return true
}
func (ps *PublishService) addProfile(profile publishapi.AefProfile, publishedService publishapi.ServiceAPIDescription) *[]publishapi.AefProfile {
	registeredProfiles := *publishedService.AefProfiles
	newProfiles := append(registeredProfiles, profile)
	publishedService.AefProfiles = &newProfiles
	return &newProfiles

}

func (*PublishService) updateProfile(profile publishapi.AefProfile, publishedService publishapi.ServiceAPIDescription, ctx echo.Context) (int, bool, error) {
	pos, registeredProfile, err := getProfile(profile.AefId, publishedService.AefProfiles)
	if err != nil {
		return pos, true, sendCoreError(ctx, http.StatusBadRequest, "Unable to update service due to: "+err.Error())
	}
	if profile.DomainName != nil {
		registeredProfile.DomainName = profile.DomainName
		(*publishedService.AefProfiles)[pos] = registeredProfile
	}
	return -1, false, nil
}

func getProfile(profileId string, apiProfiles *[]publishapi.AefProfile) (int, publishapi.AefProfile, error) {
	for pos, profile := range *apiProfiles {
		if profile.AefId == profileId {
			return pos, profile, nil
		}
	}
	return 0, publishapi.AefProfile{}, fmt.Errorf("profile with ID %s is not registered for the service", profileId)
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
