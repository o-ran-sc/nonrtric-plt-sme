// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022-2023: Nordix Foundation
//   Copyright (C) 2024: OpenInfra Foundation Europe
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

	echo "github.com/labstack/echo/v4"
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
	// Checks if the provided API is published.
	// Returns true if the provided API has been published, false otherwise.
	IsAPIPublished(aefId, path string) bool
	// Gets all published APIs.
	// Returns a list of all APIs that has been published.
	GetAllPublishedServices() []publishapi.ServiceAPIDescription
	GetAllowedPublishedServices(invokerApiList []publishapi.ServiceAPIDescription) []publishapi.ServiceAPIDescription
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

func (ps *PublishService) getAllAefIds() []string {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	allIds := []string{}
	for _, descriptions := range ps.publishedServices {
		for _, description := range descriptions {
			allIds = append(allIds, description.GetAefIds()...)
		}
	}
	return allIds
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

func (ps *PublishService) GetAllowedPublishedServices(apiListRequestedServices []publishapi.ServiceAPIDescription) []publishapi.ServiceAPIDescription {
	apiListAllPublished := ps.GetAllPublishedServices()
	if apiListRequestedServices != nil {
		allowedPublishedServices := intersection(apiListAllPublished, apiListRequestedServices)
		return allowedPublishedServices
	}
	return []publishapi.ServiceAPIDescription{}
}

func intersection(a, b []publishapi.ServiceAPIDescription) []publishapi.ServiceAPIDescription {
	var result []publishapi.ServiceAPIDescription

	for _, itemA := range a {
		for _, itemB := range b {
			if *itemA.ApiId == *itemB.ApiId {
				result = append(result, itemA)
				break
			}
		}
	}
	return result
}

// Retrieve all published APIs.
func (ps *PublishService) GetApfIdServiceApis(ctx echo.Context, apfId string) error {
	if !ps.serviceRegister.IsPublishingFunctionRegistered(apfId) {
		errorMsg := fmt.Sprintf("Unable to get the service due to %s api is only available for publishers", apfId)
		return sendCoreError(ctx, http.StatusNotFound, errorMsg)
	}

	serviceDescriptions := ps.publishedServices[apfId]
	err := ctx.JSON(http.StatusOK, serviceDescriptions)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}
	return nil
}

// Publish a new API.
func (ps *PublishService) PostApfIdServiceApis(ctx echo.Context, apfId string) error {
	var newServiceAPIDescription publishapi.ServiceAPIDescription
	errorMsg := "Unable to publish the service due to %s "
	err := ctx.Bind(&newServiceAPIDescription)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errorMsg, "invalid format for service "+apfId))
	}

	if !ps.serviceRegister.IsPublishingFunctionRegistered(apfId) {
		return sendCoreError(ctx, http.StatusForbidden, fmt.Sprintf(errorMsg, "api is only available for publishers "+apfId))
	}

	if err := ps.isServicePublished(newServiceAPIDescription); err != nil {
		return sendCoreError(ctx, http.StatusForbidden, fmt.Sprintf(errorMsg, err))
	}

	if err := newServiceAPIDescription.Validate(); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errorMsg, err))
	}
	ps.lock.Lock()
	defer ps.lock.Unlock()

	registeredFuncs := ps.serviceRegister.GetAefsForPublisher(apfId)
	for _, profile := range *newServiceAPIDescription.AefProfiles {
		if !slices.Contains(registeredFuncs, profile.AefId) {
			return sendCoreError(ctx, http.StatusNotFound, fmt.Sprintf(errorMsg, fmt.Sprintf("function %s not registered", profile.AefId)))
		}
	}

	newServiceAPIDescription.PrepareNewService()

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

func (ps *PublishService) isServicePublished(newService publishapi.ServiceAPIDescription) error {
	for _, services := range ps.publishedServices {
		for _, service := range services {
			if err := service.ValidateAlreadyPublished(newService); err != nil {
				return err
			}
		}
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
		// Check for nil as we had a failure here when running unit tests in parallel against a single Capifcore instance
		if (description.ApiId != nil) && (serviceApiId == *description.ApiId) {
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
	ps.lock.Lock()
	defer ps.lock.Unlock()
	errMsg := "Unable to update service due to %s."

	pos, publishedService, err := ps.checkIfServiceIsPublished(apfId, serviceApiId, ctx)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	updatedServiceDescription, err := getServiceFromRequest(ctx)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	// Additional validation for PUT
	if (updatedServiceDescription.ApiId == nil) || (*updatedServiceDescription.ApiId != serviceApiId) {
		errDetail := "ServiceAPIDescription ApiId doesn't match path parameter"
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, errDetail))
	}

	err = ps.checkProfilesRegistered(apfId, *updatedServiceDescription.AefProfiles)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	ps.updateDescription(pos, apfId, &updatedServiceDescription, &publishedService)

	publishedService.AefProfiles = updatedServiceDescription.AefProfiles
	ps.publishedServices[apfId][pos] = publishedService

	err = ctx.JSON(http.StatusOK, publishedService)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}
	return nil
}

func (ps *PublishService) checkIfServiceIsPublished(apfId string, serviceApiId string, ctx echo.Context) (int, publishapi.ServiceAPIDescription, error) {
	publishedServices, ok := ps.publishedServices[apfId]
	if !ok {
		return 0, publishapi.ServiceAPIDescription{}, fmt.Errorf("service must be published before updating it")
	} else {
		for pos, description := range publishedServices {
			if *description.ApiId == serviceApiId {
				return pos, description, nil
			}
		}
	}
	return 0, publishapi.ServiceAPIDescription{}, fmt.Errorf("service must be published before updating it")
}

func getServiceFromRequest(ctx echo.Context) (publishapi.ServiceAPIDescription, error) {
	var updatedServiceDescription publishapi.ServiceAPIDescription
	err := ctx.Bind(&updatedServiceDescription)
	if err != nil {
		return publishapi.ServiceAPIDescription{}, fmt.Errorf("invalid format for service")
	}
	return updatedServiceDescription, nil
}

func (ps *PublishService) updateDescription(pos int, apfId string, updatedServiceDescription, publishedService *publishapi.ServiceAPIDescription) {
	if updatedServiceDescription.Description != nil {
		publishedService.Description = updatedServiceDescription.Description
		go ps.sendEvent(*publishedService, eventsapi.CAPIFEventSERVICEAPIUPDATE)
	}
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

func (ps *PublishService) checkProfilesRegistered(apfId string, updatedProfiles []publishapi.AefProfile) error {
	registeredFuncs := ps.serviceRegister.GetAefsForPublisher(apfId)
	for _, profile := range updatedProfiles {
		if !slices.Contains(registeredFuncs, profile.AefId) {
			return fmt.Errorf("function %s not registered", profile.AefId)
		}
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
