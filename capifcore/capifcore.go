// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022-2023: Nordix Foundation. All rights reserved.
//   Copyright (C) 2023-2024 OpenInfra Foundation Europe. All rights reserved.
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

package capifcore

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/discoverserviceapi"
	"oransc.org/nonrtric/capifcore/internal/eventsapi"
	"oransc.org/nonrtric/capifcore/internal/invokermanagementapi"
	"oransc.org/nonrtric/capifcore/internal/providermanagementapi"
	"oransc.org/nonrtric/capifcore/internal/securityapi"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifcore/internal/helmmanagement"

	"oransc.org/nonrtric/capifcore/internal/discoverservice"
	"oransc.org/nonrtric/capifcore/internal/eventservice"
	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"
	"oransc.org/nonrtric/capifcore/internal/publishservice"
	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"
	security "oransc.org/nonrtric/capifcore/internal/securityservice"
	"oransc.org/nonrtric/capifcore/internal/keycloak"
)

func RegisterHandlers(e *echo.Echo, helmManager helmmanagement.HelmManager, km *keycloak.KeycloakManager) {
	// Log all requests
	e.Use(echomiddleware.Logger())

	var group *echo.Group
	// Register ProviderManagement
	providerManagerSwagger, err := providermanagementapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading ProviderManagement swagger spec\n: %s", err)
	}
	providerManagerSwagger.Servers = nil
	providerManager := providermanagement.NewProviderManager()
	group = e.Group("/api-provider-management/v1")
	group.Use(middleware.OapiRequestValidator(providerManagerSwagger))
	providermanagementapi.RegisterHandlersWithBaseURL(e, providerManager, "/api-provider-management/v1")

	// Register EventService
	eventServiceSwagger, err := eventsapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading EventService swagger spec\n: %s", err)
	}
	eventServiceSwagger.Servers = nil
	eventService := eventservice.NewEventService(&http.Client{})
	group = e.Group("/capif-events/v1")
	group.Use(middleware.OapiRequestValidator(eventServiceSwagger))
	eventsapi.RegisterHandlersWithBaseURL(e, eventService, "/capif-events/v1")
	eventChannel := eventService.GetNotificationChannel()

	// Register PublishService
	publishServiceSwagger, err := publishserviceapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading PublishService swagger spec\n: %s", err)
	}
	publishServiceSwagger.Servers = nil
	publishService := publishservice.NewPublishService(providerManager, helmManager, eventChannel)
	group = e.Group("/published-apis/v1")
	group.Use(middleware.OapiRequestValidator(publishServiceSwagger))
	publishserviceapi.RegisterHandlersWithBaseURL(e, publishService, "/published-apis/v1")

	// Register InvokerManagement
	invokerManagerSwagger, err := invokermanagementapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading InvokerManagement swagger spec\n: %s", err)
	}
	invokerManagerSwagger.Servers = nil
	invokerManager := invokermanagement.NewInvokerManager(publishService, km, eventChannel)
	group = e.Group("/api-invoker-management/v1")
	group.Use(middleware.OapiRequestValidator(invokerManagerSwagger))
	invokermanagementapi.RegisterHandlersWithBaseURL(e, invokerManager, "/api-invoker-management/v1")

	// Register DiscoverService
	discoverServiceSwagger, err := discoverserviceapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading DiscoverService swagger spec\n: %s", err)
	}
	discoverServiceSwagger.Servers = nil
	discoverService := discoverservice.NewDiscoverService(invokerManager)
	group = e.Group("/service-apis/v1")
	group.Use(middleware.OapiRequestValidator(discoverServiceSwagger))
	discoverserviceapi.RegisterHandlersWithBaseURL(e, discoverService, "/service-apis/v1")

	// Register Security
	securitySwagger, err := securityapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading Security swagger spec\n: %s", err)
	}
	securitySwagger.Servers = nil
	securityService := security.NewSecurity(providerManager, publishService, invokerManager, km)
	group = e.Group("/capif-security/v1")
	group.Use(middleware.OapiRequestValidator(securitySwagger))
	securityapi.RegisterHandlersWithBaseURL(e, securityService, "/capif-security/v1")

	e.GET("/", hello)

	e.GET("/swagger/:apiName", getSwagger)
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func getSwagger(c echo.Context) error {
	var swagger *openapi3.T
	var err error
	switch api := c.Param("apiName"); api {
	case "provider":
		swagger, err = providermanagementapi.GetSwagger()
	case "publish":
		swagger, err = publishserviceapi.GetSwagger()
	case "invoker":
		swagger, err = invokermanagementapi.GetSwagger()
	case "discover":
		swagger, err = discoverserviceapi.GetSwagger()
	case "events":
		swagger, err = eventsapi.GetSwagger()
	case "security":
		swagger, err = securityapi.GetSwagger()
	default:
		return c.JSON(http.StatusBadRequest, getProblemDetails("Invalid API name "+api, http.StatusBadRequest))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, getProblemDetails("Unable to get swagger for API", http.StatusInternalServerError))
	}
	return c.JSON(http.StatusOK, swagger)
}

func getProblemDetails(cause string, status int) common29122.ProblemDetails {
	return common29122.ProblemDetails{
		Cause:  &cause,
		Status: &status,
	}
}
