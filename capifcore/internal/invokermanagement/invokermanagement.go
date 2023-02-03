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

package invokermanagement

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"sync"

	"oransc.org/nonrtric/capifcore/internal/eventsapi"
	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	invokerapi "oransc.org/nonrtric/capifcore/internal/invokermanagementapi"

	"oransc.org/nonrtric/capifcore/internal/publishservice"

	"github.com/labstack/echo/v4"
)

//go:generate mockery --name InvokerRegister
type InvokerRegister interface {
	// Checks if the invoker is registered.
	// Returns true of the provided invoker is registered, false otherwise.
	IsInvokerRegistered(invokerId string) bool
	// Verifies that the provided secret is the invoker's registered secret.
	// Returns true if the provided secret is the registered invoker's secret, false otherwise.
	VerifyInvokerSecret(invokerId, secret string) bool
	// Gets the provided invoker's registered APIs.
	// Returns a list of all the invoker's registered APIs.
	GetInvokerApiList(invokerId string) *invokerapi.APIList
}

type InvokerManager struct {
	onboardedInvokers map[string]invokerapi.APIInvokerEnrolmentDetails
	publishRegister   publishservice.PublishRegister
	nextId            int64
	eventChannel      chan<- eventsapi.EventNotification
	lock              sync.Mutex
}

// Creates a manager that implements both the InvokerRegister and the invokermanagementapi.ServerInterface interfaces.
func NewInvokerManager(publishRegister publishservice.PublishRegister, eventChannel chan<- eventsapi.EventNotification) *InvokerManager {
	return &InvokerManager{
		onboardedInvokers: make(map[string]invokerapi.APIInvokerEnrolmentDetails),
		publishRegister:   publishRegister,
		nextId:            1000,
		eventChannel:      eventChannel,
	}
}

func (im *InvokerManager) IsInvokerRegistered(invokerId string) bool {
	im.lock.Lock()
	defer im.lock.Unlock()

	_, registered := im.onboardedInvokers[invokerId]
	return registered
}

func (im *InvokerManager) VerifyInvokerSecret(invokerId, secret string) bool {
	im.lock.Lock()
	defer im.lock.Unlock()

	verified := false
	if invoker, registered := im.onboardedInvokers[invokerId]; registered {
		verified = *invoker.OnboardingInformation.OnboardingSecret == secret
	}
	return verified
}

func (im *InvokerManager) GetInvokerApiList(invokerId string) *invokerapi.APIList {
	var apiList invokerapi.APIList = im.publishRegister.GetAllPublishedServices()
	im.lock.Lock()
	defer im.lock.Unlock()
	invoker, ok := im.onboardedInvokers[invokerId]
	if ok {
		invoker.ApiList = &apiList
		return &apiList
	}
	return nil
}

// Creates a new individual API Invoker profile.
func (im *InvokerManager) PostOnboardedInvokers(ctx echo.Context) error {
	var newInvoker invokerapi.APIInvokerEnrolmentDetails
	errMsg := "Unable to onboard invoker due to %s"
	if err := ctx.Bind(&newInvoker); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, "invalid format for invoker"))
	}

	if im.isInvokerOnboarded(newInvoker) {
		return sendCoreError(ctx, http.StatusForbidden, fmt.Sprintf(errMsg, "invoker already onboarded"))
	}

	if err := im.validateInvoker(newInvoker, ctx); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	im.prepareNewInvoker(&newInvoker)

	go im.sendEvent(*newInvoker.ApiInvokerId, eventsapi.CAPIFEventAPIINVOKERONBOARDED)

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, *newInvoker.ApiInvokerId))
	err := ctx.JSON(http.StatusCreated, newInvoker)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (im *InvokerManager) isInvokerOnboarded(newInvoker invokerapi.APIInvokerEnrolmentDetails) bool {
	for _, invoker := range im.onboardedInvokers {
		if newInvoker.OnboardingInformation.ApiInvokerPublicKey == invoker.OnboardingInformation.ApiInvokerPublicKey {
			return true
		}
	}
	return false
}

func (im *InvokerManager) prepareNewInvoker(newInvoker *invokerapi.APIInvokerEnrolmentDetails) {
	var apiList invokerapi.APIList = im.publishRegister.GetAllPublishedServices()
	newInvoker.ApiList = &apiList

	im.lock.Lock()
	defer im.lock.Unlock()

	newInvoker.PrepareNewInvoker()

	im.onboardedInvokers[*newInvoker.ApiInvokerId] = *newInvoker
}

// Deletes an individual API Invoker.
func (im *InvokerManager) DeleteOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error {
	if _, ok := im.onboardedInvokers[onboardingId]; ok {
		im.deleteInvoker(onboardingId)
	}

	go im.sendEvent(onboardingId, eventsapi.CAPIFEventAPIINVOKEROFFBOARDED)

	return ctx.NoContent(http.StatusNoContent)
}

func (im *InvokerManager) deleteInvoker(onboardingId string) {
	im.lock.Lock()
	defer im.lock.Unlock()
	delete(im.onboardedInvokers, onboardingId)
}

// Updates an individual API invoker details.
func (im *InvokerManager) PutOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error {
	var invoker invokerapi.APIInvokerEnrolmentDetails
	errMsg := "Unable to update invoker due to %s"
	if err := ctx.Bind(&invoker); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, "invalid format for invoker"))
	}

	if onboardingId != *invoker.ApiInvokerId {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, "ApiInvokerId not matching"))
	}

	if err := im.validateInvoker(invoker, ctx); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	if _, ok := im.onboardedInvokers[onboardingId]; ok {
		im.updateInvoker(invoker)
	} else {
		return sendCoreError(ctx, http.StatusNotFound, "The invoker to update has not been onboarded")
	}

	err := ctx.JSON(http.StatusOK, invoker)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (im *InvokerManager) updateInvoker(invoker invokerapi.APIInvokerEnrolmentDetails) {
	im.lock.Lock()
	defer im.lock.Unlock()
	im.onboardedInvokers[*invoker.ApiInvokerId] = invoker
}

func (im *InvokerManager) ModifyIndApiInvokeEnrolment(ctx echo.Context, onboardingId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (im *InvokerManager) validateInvoker(invoker invokerapi.APIInvokerEnrolmentDetails, ctx echo.Context) error {
	if err := invoker.Validate(); err != nil {
		return err
	}
	if !im.areAPIsPublished(invoker.ApiList) {
		return errors.New("some APIs needed by invoker are not registered")
	}

	return nil
}

func (im *InvokerManager) areAPIsPublished(apis *invokerapi.APIList) bool {
	if apis == nil {
		return true
	}
	return im.publishRegister.AreAPIsPublished((*[]publishapi.ServiceAPIDescription)(apis))
}

func (im *InvokerManager) sendEvent(invokerId string, eventType eventsapi.CAPIFEvent) {
	invokerIds := []string{invokerId}
	event := eventsapi.EventNotification{
		EventDetail: &eventsapi.CAPIFEventDetail{
			ApiInvokerIds: &invokerIds,
		},
		Events: eventType,
	}
	im.eventChannel <- event
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
