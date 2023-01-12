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
	"strings"
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
}

type ProviderManager struct {
	onboardedProviders map[string]provapi.APIProviderEnrolmentDetails
	lock               sync.Mutex
}

func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		onboardedProviders: make(map[string]provapi.APIProviderEnrolmentDetails),
	}
}

func (pm *ProviderManager) IsFunctionRegistered(functionId string) bool {
	registered := false
out:
	for _, provider := range pm.onboardedProviders {
		for _, registeredFunc := range *provider.ApiProvFuncs {
			if *registeredFunc.ApiProvFuncId == functionId {
				registered = true
				break out
			}
		}
	}

	return registered
}

func (pm *ProviderManager) GetAefsForPublisher(apfId string) []string {
	for _, provider := range pm.onboardedProviders {
		for _, registeredFunc := range *provider.ApiProvFuncs {
			if *registeredFunc.ApiProvFuncId == apfId && registeredFunc.ApiProvFuncRole == provapi.ApiProviderFuncRoleAPF {
				return getExposedFuncs(provider.ApiProvFuncs)
			}
		}
	}
	return nil
}

func getExposedFuncs(providerFuncs *[]provapi.APIProviderFunctionDetails) []string {
	exposedFuncs := []string{}
	for _, registeredFunc := range *providerFuncs {
		if registeredFunc.ApiProvFuncRole == provapi.ApiProviderFuncRoleAEF {
			exposedFuncs = append(exposedFuncs, *registeredFunc.ApiProvFuncId)
		}
	}
	return exposedFuncs
}

func (pm *ProviderManager) PostRegistrations(ctx echo.Context) error {
	var newProvider provapi.APIProviderEnrolmentDetails
	err := ctx.Bind(&newProvider)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, "Invalid format for provider")
	}

	if newProvider.ApiProvDomInfo == nil || *newProvider.ApiProvDomInfo == "" {
		return sendCoreError(ctx, http.StatusBadRequest, "Provider missing required ApiProvDomInfo")
	}

	pm.prepareNewProvider(&newProvider)

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, *newProvider.ApiProvDomId))
	err = ctx.JSON(http.StatusCreated, newProvider)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (pm *ProviderManager) prepareNewProvider(newProvider *provapi.APIProviderEnrolmentDetails) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	newProvider.ApiProvDomId = pm.getDomainId(newProvider.ApiProvDomInfo)

	pm.registerFunctions(newProvider.ApiProvFuncs)
	pm.onboardedProviders[*newProvider.ApiProvDomId] = *newProvider
}

func (pm *ProviderManager) DeleteRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {

	log.Debug(pm.onboardedProviders)
	if _, ok := pm.onboardedProviders[registrationId]; ok {
		pm.deleteProvider(registrationId)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (pm *ProviderManager) deleteProvider(registrationId string) {
	log.Debug("Deleting provider", registrationId)
	pm.lock.Lock()
	defer pm.lock.Unlock()
	delete(pm.onboardedProviders, registrationId)
}

func (pm *ProviderManager) PutRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	errMsg := "Unable to update provider due to %s."
	registeredProvider, err := pm.checkIfProviderIsRegistered(registrationId, ctx)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	updatedProvider, err := getProviderFromRequest(ctx)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	updateDomainInfo(&updatedProvider, registeredProvider)

	registeredProvider.ApiProvFuncs, err = updateFuncs(updatedProvider.ApiProvFuncs, registeredProvider.ApiProvFuncs)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}

	pm.onboardedProviders[*registeredProvider.ApiProvDomId] = *registeredProvider
	err = ctx.JSON(http.StatusOK, *registeredProvider)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (pm *ProviderManager) checkIfProviderIsRegistered(registrationId string, ctx echo.Context) (*provapi.APIProviderEnrolmentDetails, error) {
	registeredProvider, ok := pm.onboardedProviders[registrationId]
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

func updateDomainInfo(updatedProvider, registeredProvider *provapi.APIProviderEnrolmentDetails) {
	if updatedProvider.ApiProvDomInfo != nil {
		registeredProvider.ApiProvDomInfo = updatedProvider.ApiProvDomInfo
	}
}

func updateFuncs(updatedFuncs, registeredFuncs *[]provapi.APIProviderFunctionDetails) (*[]provapi.APIProviderFunctionDetails, error) {
	addedFuncs := []provapi.APIProviderFunctionDetails{}
	changedFuncs := []provapi.APIProviderFunctionDetails{}
	for _, function := range *updatedFuncs {
		if function.ApiProvFuncId == nil {
			function.ApiProvFuncId = getFuncId(function.ApiProvFuncRole, function.ApiProvFuncInfo)
			addedFuncs = append(addedFuncs, function)
		} else {
			registeredFunction, ok := getApiFunc(*function.ApiProvFuncId, registeredFuncs)
			if !ok {
				return nil, fmt.Errorf("function with ID %s is not registered for the provider", *function.ApiProvFuncId)
			}
			if function.ApiProvFuncInfo != nil {
				registeredFunction.ApiProvFuncInfo = function.ApiProvFuncInfo
			}
			changedFuncs = append(changedFuncs, function)
		}
	}
	modifiedFuncs := append(changedFuncs, addedFuncs...)
	return &modifiedFuncs, nil
}

func getApiFunc(funcId string, apiFunctions *[]provapi.APIProviderFunctionDetails) (provapi.APIProviderFunctionDetails, bool) {
	for _, function := range *apiFunctions {
		if *function.ApiProvFuncId == funcId {
			return function, true
		}
	}
	return provapi.APIProviderFunctionDetails{}, false
}

func (pm *ProviderManager) ModifyIndApiProviderEnrolment(ctx echo.Context, registrationId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (pm *ProviderManager) registerFunctions(provFuncs *[]provapi.APIProviderFunctionDetails) {
	if provFuncs == nil {
		return
	}
	for i, provFunc := range *provFuncs {
		(*provFuncs)[i].ApiProvFuncId = getFuncId(provFunc.ApiProvFuncRole, provFunc.ApiProvFuncInfo)
	}
}

func (pm *ProviderManager) getDomainId(domainInfo *string) *string {
	idAsString := "domain_id_" + strings.ReplaceAll(*domainInfo, " ", "_")
	return &idAsString
}

func getFuncId(role provapi.ApiProviderFuncRole, funcInfo *string) *string {
	var idPrefix string
	switch role {
	case provapi.ApiProviderFuncRoleAPF:
		idPrefix = "APF_id_"
	case provapi.ApiProviderFuncRoleAMF:
		idPrefix = "AMF_id_"
	case provapi.ApiProviderFuncRoleAEF:
		idPrefix = "AEF_id_"
	default:
		idPrefix = "function_id_"
	}
	idAsString := idPrefix + strings.ReplaceAll(*funcInfo, " ", "_")
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
