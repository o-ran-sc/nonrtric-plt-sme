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

package providermanagement

import (
	"fmt"
	"net/http"
	"path"
	"sync"

	"github.com/labstack/echo/v4"

	"oransc.org/nonrtric/capifcore/internal/common29122"
	provapi "oransc.org/nonrtric/capifcore/internal/providermanagementapi"

	log "github.com/sirupsen/logrus"
)

//go:generate mockery --name ServiceRegister
type ServiceRegister interface {
	IsFunctionRegistered(functionId string) bool
	GetAefsForPublisher(apfId string) []string
	IsPublishingFunctionRegistered(apiProvFuncId string) bool
}

type ProviderManager struct {
	registeredProviders map[string]provapi.APIProviderEnrolmentDetails
	lock                sync.Mutex
}

func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		registeredProviders: make(map[string]provapi.APIProviderEnrolmentDetails),
	}
}

func (pm *ProviderManager) IsFunctionRegistered(functionId string) bool {
	for _, provider := range pm.registeredProviders {
		if provider.IsFunctionRegistered(functionId) {
			return true
		}
	}
	return false
}

func (pm *ProviderManager) GetAefsForPublisher(apfId string) []string {
	for _, provider := range pm.registeredProviders {
		if aefs := provider.GetExposingFunctionIdsForPublisher(apfId); aefs != nil {
			return aefs
		}
	}
	return nil
}

func (pm *ProviderManager) IsPublishingFunctionRegistered(apiProvFuncId string) bool {
	for _, provider := range pm.registeredProviders {
		if provider.IsPublishingFunctionRegistered(apiProvFuncId) {
			return true
		}
	}
	return false
}

func (pm *ProviderManager) PostRegistrations(ctx echo.Context) error {
	var newProvider provapi.APIProviderEnrolmentDetails
	errMsg := "Unable to register provider due to %s"
	if err := ctx.Bind(&newProvider); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, "invalid format for provider"))
	}

	if err := pm.isProviderRegistered(newProvider); err != nil {
		return sendCoreError(ctx, http.StatusForbidden, fmt.Sprintf(errMsg, err))
	}

	if err := newProvider.Validate(); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	pm.prepareNewProvider(&newProvider)

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, *newProvider.ApiProvDomId))
	if err := ctx.JSON(http.StatusCreated, newProvider); err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}
	return nil
}

func (pm *ProviderManager) isProviderRegistered(newProvider provapi.APIProviderEnrolmentDetails) error {
	for _, prov := range pm.registeredProviders {
		if err := prov.ValidateAlreadyRegistered(newProvider); err != nil {
			return err
		}
	}
	return nil
}

func (pm *ProviderManager) prepareNewProvider(newProvider *provapi.APIProviderEnrolmentDetails) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	newProvider.PrepareNewProvider()
	pm.registeredProviders[*newProvider.ApiProvDomId] = *newProvider
}

func (pm *ProviderManager) DeleteRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {
	log.Debug(pm.registeredProviders)
	if _, ok := pm.registeredProviders[registrationId]; ok {
		pm.deleteProvider(registrationId)
	}
	return ctx.NoContent(http.StatusNoContent)
}

func (pm *ProviderManager) deleteProvider(registrationId string) {
	log.Debug("Deleting provider", registrationId)
	pm.lock.Lock()
	defer pm.lock.Unlock()
	delete(pm.registeredProviders, registrationId)
}

func (pm *ProviderManager) PutRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {
	errMsg := "Unable to update provider due to %s."
	registeredProvider, err := pm.checkIfProviderIsRegistered(registrationId, ctx)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	updatedProvider, err := getProviderFromRequest(ctx)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	if updatedProvider.Validate() != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	// Additional validation for PUT
	if updatedProvider.ApiProvDomId == nil {
		errDetail := "APIProviderEnrolmentDetails missing required ApiProvDomId"
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, errDetail))
	}

	if err = pm.updateProvider(updatedProvider, registeredProvider); err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	if err = ctx.JSON(http.StatusOK, updatedProvider); err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}
	return nil
}

func (pm *ProviderManager) ModifyIndApiProviderEnrolment(ctx echo.Context, registrationId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (pm *ProviderManager) checkIfProviderIsRegistered(registrationId string, ctx echo.Context) (*provapi.APIProviderEnrolmentDetails, error) {
	registeredProvider, ok := pm.registeredProviders[registrationId]
	if !ok {
		return nil, fmt.Errorf("provider not onboarded")
	}
	return &registeredProvider, nil
}

func getProviderFromRequest(ctx echo.Context) (provapi.APIProviderEnrolmentDetails, error) {
	var updatedProvider provapi.APIProviderEnrolmentDetails
	err := ctx.Bind(&updatedProvider)
	if err != nil {
		return provapi.APIProviderEnrolmentDetails{}, fmt.Errorf("invalid format for provider")
	}
	return updatedProvider, nil
}

func (pm *ProviderManager) updateProvider(updatedProvider provapi.APIProviderEnrolmentDetails, registeredProvider *provapi.APIProviderEnrolmentDetails) error {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	if err := updatedProvider.UpdateFuncs(*registeredProvider); err == nil {
		pm.registeredProviders[*updatedProvider.ApiProvDomId] = updatedProvider
		return nil
	} else {
		return err
	}
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
