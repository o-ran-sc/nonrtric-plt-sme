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
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	invokerapi "oransc.org/nonrtric/capifcore/internal/invokermanagementapi"

	"oransc.org/nonrtric/capifcore/internal/publishservice"

	"github.com/labstack/echo/v4"
)

//go:generate mockery --name InvokerRegister
type InvokerRegister interface {
	IsInvokerRegistered(invokerId string) bool
	VerifyInvokerSecret(invokerId, secret string) bool
	GetInvokerApiList(invokerId string) *invokerapi.APIList
}

type InvokerManager struct {
	onboardedInvokers map[string]invokerapi.APIInvokerEnrolmentDetails
	apiRegister       publishservice.APIRegister
	nextId            int64
	lock              sync.Mutex
}

func NewInvokerManager(apiRegister publishservice.APIRegister) *InvokerManager {
	return &InvokerManager{
		onboardedInvokers: make(map[string]invokerapi.APIInvokerEnrolmentDetails),
		apiRegister:       apiRegister,
		nextId:            1000,
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
	invoker, ok := im.onboardedInvokers[invokerId]
	if ok {
		return invoker.ApiList
	}
	return nil
}

func (im *InvokerManager) PostOnboardedInvokers(ctx echo.Context) error {
	var newInvoker invokerapi.APIInvokerEnrolmentDetails
	err := ctx.Bind(&newInvoker)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, "Invalid format for invoker")
	}

	shouldReturn, coreError := im.validateInvoker(newInvoker, ctx)
	if shouldReturn {
		return coreError
	}

	im.lock.Lock()
	defer im.lock.Unlock()

	newInvoker.ApiInvokerId = im.getId(newInvoker.ApiInvokerInformation)
	onboardingSecret := "onboarding_secret_"
	if newInvoker.ApiInvokerInformation != nil {
		onboardingSecret = onboardingSecret + strings.ReplaceAll(*newInvoker.ApiInvokerInformation, " ", "_")
	} else {
		onboardingSecret = onboardingSecret + *newInvoker.ApiInvokerId
	}
	newInvoker.OnboardingInformation.OnboardingSecret = &onboardingSecret

	im.onboardedInvokers[*newInvoker.ApiInvokerId] = newInvoker

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, *newInvoker.ApiInvokerId))
	err = ctx.JSON(http.StatusCreated, newInvoker)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (im *InvokerManager) DeleteOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error {
	im.lock.Lock()
	defer im.lock.Unlock()

	delete(im.onboardedInvokers, onboardingId)

	return ctx.NoContent(http.StatusNoContent)
}

func (im *InvokerManager) PutOnboardedInvokersOnboardingId(ctx echo.Context, onboardingId string) error {
	var invoker invokerapi.APIInvokerEnrolmentDetails
	err := ctx.Bind(&invoker)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, "Invalid format for invoker")
	}

	if onboardingId != *invoker.ApiInvokerId {
		return sendCoreError(ctx, http.StatusBadRequest, "Invoker ApiInvokerId not matching")
	}

	shouldReturn, coreError := im.validateInvoker(invoker, ctx)
	if shouldReturn {
		return coreError
	}

	im.lock.Lock()
	defer im.lock.Unlock()

	if _, ok := im.onboardedInvokers[onboardingId]; ok {
		im.onboardedInvokers[*invoker.ApiInvokerId] = invoker
	} else {
		return sendCoreError(ctx, http.StatusNotFound, "The invoker to update has not been onboarded")
	}

	err = ctx.JSON(http.StatusOK, invoker)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (im *InvokerManager) ModifyIndApiInvokeEnrolment(ctx echo.Context, onboardingId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (im *InvokerManager) validateInvoker(invoker invokerapi.APIInvokerEnrolmentDetails, ctx echo.Context) (bool, error) {
	if invoker.NotificationDestination == "" {
		return true, sendCoreError(ctx, http.StatusBadRequest, "Invoker missing required NotificationDestination")
	}

	if invoker.OnboardingInformation.ApiInvokerPublicKey == "" {
		return true, sendCoreError(ctx, http.StatusBadRequest, "Invoker missing required OnboardingInformation.ApiInvokerPublicKey")
	}

	if !im.areAPIsRegistered(invoker.ApiList) {
		return true, sendCoreError(ctx, http.StatusBadRequest, "Some APIs needed by invoker are not registered")
	}

	return false, nil
}

func (im *InvokerManager) areAPIsRegistered(apis *invokerapi.APIList) bool {
	if apis == nil {
		return true
	}
	return im.apiRegister.AreAPIsRegistered((*[]publishapi.ServiceAPIDescription)(apis))
}

func (im *InvokerManager) getId(invokerInfo *string) *string {
	idAsString := "api_invoker_id_"
	if invokerInfo != nil {
		idAsString = idAsString + strings.ReplaceAll(*invokerInfo, " ", "_")
	} else {
		idAsString = idAsString + strconv.FormatInt(im.nextId, 10)
		im.nextId = im.nextId + 1
	}
	return &idAsString
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
