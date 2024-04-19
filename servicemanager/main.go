// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023-2024: OpenInfra Foundation Europe
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

package main

import (
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	echo "github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"oransc.org/nonrtric/servicemanager/internal/common29122"
	"oransc.org/nonrtric/servicemanager/internal/envreader"

	"oransc.org/nonrtric/servicemanager/internal/discoverserviceapi"
	"oransc.org/nonrtric/servicemanager/internal/invokermanagementapi"
	"oransc.org/nonrtric/servicemanager/internal/providermanagementapi"
	"oransc.org/nonrtric/servicemanager/internal/publishserviceapi"

	"oransc.org/nonrtric/servicemanager/internal/discoverservice"
	"oransc.org/nonrtric/servicemanager/internal/invokermanagement"
	"oransc.org/nonrtric/servicemanager/internal/providermanagement"
	"oransc.org/nonrtric/servicemanager/internal/publishservice"
)

func main() {
	realConfigReader := &envreader.RealConfigReader{}
	myEnv, myPorts, err := realConfigReader.ReadDotEnv()
	if err != nil {
		log.Fatal("error loading environment file")
		return
	}

	eServiceManager := echo.New()
	err = registerHandlers(eServiceManager, myEnv, myPorts)

	if err != nil {
		log.Fatal("registerHandlers fatal error")
		return
	}

	port := myPorts["SERVICE_MANAGER_PORT"]

	go startWebServer(eServiceManager, port)
	log.Info("server started and listening on port: ", port)
	keepServerAlive()
}

func registerHandlers(e *echo.Echo, myEnv map[string]string, myPorts map[string]int) (err error) {
	// Log all requests
	e.Use(echomiddleware.Logger())

	capifProtocol := myEnv["CAPIF_PROTOCOL"]
	capifIPv4 := common29122.Ipv4Addr(myEnv["CAPIF_IPV4"])
	capifPort := common29122.Port(myPorts["CAPIF_PORT"])
	kongDomain := myEnv["KONG_DOMAIN"]
	kongProtocol := myEnv["KONG_PROTOCOL"]
	kongControlPlaneIPv4 := common29122.Ipv4Addr(myEnv["KONG_CONTROL_PLANE_IPV4"])
	kongControlPlanePort := common29122.Port(myPorts["KONG_CONTROL_PLANE_PORT"])
	kongDataPlaneIPv4 := common29122.Ipv4Addr(myEnv["KONG_DATA_PLANE_IPV4"])
	kongDataPlanePort := common29122.Port(myPorts["KONG_DATA_PLANE_PORT"])

	var group *echo.Group

	// Register ProviderManagement
	providerManagerSwagger, err := providermanagementapi.GetSwagger()
	if err != nil {
		log.Fatalf("error loading ProviderManagement swagger spec\n: %s", err)
		return err
	}
	providerManagerSwagger.Servers = nil
	providerManager := providermanagement.NewProviderManager(capifProtocol, capifIPv4, capifPort)
	group = e.Group("/api-provider-management/v1")
	group.Use(middleware.OapiRequestValidator(providerManagerSwagger))
	providermanagementapi.RegisterHandlersWithBaseURL(e, providerManager, "/api-provider-management/v1")

	// Register PublishService
	publishServiceSwagger, err := publishserviceapi.GetSwagger()
	if err != nil {
		log.Fatalf("error loading PublishService swagger spec\n: %s", err)
		return err
	}
	publishServiceSwagger.Servers = nil
	publishService := publishservice.NewPublishService(
		kongDomain, kongProtocol,
		kongControlPlaneIPv4, kongControlPlanePort,
		kongDataPlaneIPv4, kongDataPlanePort,
		capifProtocol, capifIPv4, capifPort)

	group = e.Group("/published-apis/v1")
	group.Use(middleware.OapiRequestValidator(publishServiceSwagger))
	publishserviceapi.RegisterHandlersWithBaseURL(e, publishService, "/published-apis/v1")

	// Register InvokerManagement
	invokerManagerSwagger, err := invokermanagementapi.GetSwagger()
	if err != nil {
		log.Fatalf("error loading InvokerManagement swagger spec\n: %s", err)
		return err
	}
	invokerManagerSwagger.Servers = nil
	invokerManager := invokermanagement.NewInvokerManager(capifProtocol, capifIPv4, capifPort)
	group = e.Group("/api-invoker-management/v1")
	group.Use(middleware.OapiRequestValidator(invokerManagerSwagger))
	invokermanagementapi.RegisterHandlersWithBaseURL(e, invokerManager, "/api-invoker-management/v1")

	// Register DiscoverService
	discoverServiceSwagger, err := discoverserviceapi.GetSwagger()
	if err != nil {
		log.Fatalf("error loading DiscoverService swagger spec\n: %s", err)
		return err
	}

	discoverServiceSwagger.Servers = nil
	discoverService := discoverservice.NewDiscoverService(capifProtocol, capifIPv4, capifPort)

	group = e.Group("/service-apis/v1")
	group.Use(middleware.OapiRequestValidator(discoverServiceSwagger))
	discoverserviceapi.RegisterHandlersWithBaseURL(e, discoverService, "/service-apis/v1")

	e.GET("/", hello)
	e.GET("/swagger/:apiName", getSwagger)

	return err
}

func startWebServer(e *echo.Echo, port int) {
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", port)))
}

func keepServerAlive() {
	forever := make(chan int)
	<-forever
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
