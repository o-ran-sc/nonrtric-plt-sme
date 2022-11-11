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

	pm.lock.Lock()
	defer pm.lock.Unlock()

	newProvider.ApiProvDomId = pm.getDomainId(newProvider.ApiProvDomInfo)

	pm.registerFunctions(newProvider.ApiProvFuncs)
	pm.onboardedProviders[*newProvider.ApiProvDomId] = newProvider

	uri := ctx.Request().Host + ctx.Request().URL.String()
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, *newProvider.ApiProvDomId))
	err = ctx.JSON(http.StatusCreated, newProvider)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (pm *ProviderManager) DeleteRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	log.Debug(pm.onboardedProviders)
	if _, ok := pm.onboardedProviders[registrationId]; ok {
		log.Debug("Deleting provider", registrationId)
		delete(pm.onboardedProviders, registrationId)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (pm *ProviderManager) PutRegistrationsRegistrationId(ctx echo.Context, registrationId string) error {
	registeredProvider, ok := pm.onboardedProviders[registrationId]
	if !ok {
		return sendCoreError(ctx, http.StatusBadRequest, "Provider must be onboarded before updating it")

	}

	var updatedProvider provapi.APIProviderEnrolmentDetails
	err := ctx.Bind(&updatedProvider)
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, "Invalid format for provider")
	}

	pm.lock.Lock()
	defer pm.lock.Unlock()

	for _, function := range *updatedProvider.ApiProvFuncs {
		if function.ApiProvFuncId == nil {
			function.ApiProvFuncId = pm.getFuncId(function.ApiProvFuncRole, function.ApiProvFuncInfo)
			registeredFuncs := *registeredProvider.ApiProvFuncs
			newFuncs := append(registeredFuncs, function)
			registeredProvider.ApiProvFuncs = &newFuncs
			pm.onboardedProviders[*registeredProvider.ApiProvDomId] = registeredProvider
		}
	}

	err = ctx.JSON(http.StatusOK, registeredProvider)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (pm *ProviderManager) ModifyIndApiProviderEnrolment(ctx echo.Context, registrationId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func (pm *ProviderManager) registerFunctions(provFuncs *[]provapi.APIProviderFunctionDetails) {
	if provFuncs == nil {
		return
	}
	for i, provFunc := range *provFuncs {
		(*provFuncs)[i].ApiProvFuncId = pm.getFuncId(provFunc.ApiProvFuncRole, provFunc.ApiProvFuncInfo)
	}
}

func (pm *ProviderManager) getDomainId(domainInfo *string) *string {
	idAsString := "domain_id_" + strings.ReplaceAll(*domainInfo, " ", "_")
	return &idAsString
}

func (pm *ProviderManager) getFuncId(role provapi.ApiProviderFuncRole, funcInfo *string) *string {
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
