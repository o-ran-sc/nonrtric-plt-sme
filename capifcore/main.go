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

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"helm.sh/helm/v3/pkg/cli"
	"oransc.org/nonrtric/capifcore/internal/discoverserviceapi"
	"oransc.org/nonrtric/capifcore/internal/invokermanagementapi"
	"oransc.org/nonrtric/capifcore/internal/providermanagementapi"
	"oransc.org/nonrtric/capifcore/internal/securityapi"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifcore/internal/discoverservice"
	"oransc.org/nonrtric/capifcore/internal/helmmanagement"
	"oransc.org/nonrtric/capifcore/internal/invokermanagement"
	"oransc.org/nonrtric/capifcore/internal/providermanagement"
	"oransc.org/nonrtric/capifcore/internal/publishservice"
	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"
	security "oransc.org/nonrtric/capifcore/internal/securityservice"
)

var url string
var helmManager helmmanagement.HelmManager
var repoName string

func main() {
	var port = flag.Int("port", 8090, "Port for CAPIF Core Function HTTP server")
	flag.StringVar(&url, "url", "http://chartmuseum:8080", "ChartMuseum url")
	flag.StringVar(&repoName, "repoName", "local-dev", "Repository name")
	var logLevelStr = flag.String("loglevel", "Info", "Log level")
	flag.Parse()

	if loglevel, err := log.ParseLevel(*logLevelStr); err == nil {
		log.SetLevel(loglevel)
	}

	// Add repo
	fmt.Printf("Adding %s to Helm Repo\n", url)
	helmManager = helmmanagement.NewHelmManager(cli.New())
	err := helmManager.AddToRepo(repoName, url)
	if err != nil {
		log.Fatal(err.Error())
	}

	go startWebServer(getEcho(), *port)

	log.Info("Server started and listening on port: ", *port)

	keepServerAlive()
}

func getEcho() *echo.Echo {
	// This is how you set up a basic Echo router
	e := echo.New()
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

	// Register PublishService
	publishServiceSwagger, err := publishserviceapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading PublishService swagger spec\n: %s", err)
	}
	publishServiceSwagger.Servers = nil
	publishService := publishservice.NewPublishService(providerManager, helmManager)
	group = e.Group("/published-apis/v1")
	group.Use(middleware.OapiRequestValidator(publishServiceSwagger))
	publishserviceapi.RegisterHandlersWithBaseURL(e, publishService, "/published-apis/v1")

	// Register DiscoverService
	discoverServiceSwagger, err := discoverserviceapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading DiscoverService swagger spec\n: %s", err)
	}
	discoverServiceSwagger.Servers = nil
	discoverService := discoverservice.NewDiscoverService(publishService)
	group = e.Group("/service-apis/v1")
	group.Use(middleware.OapiRequestValidator(discoverServiceSwagger))
	discoverserviceapi.RegisterHandlersWithBaseURL(e, discoverService, "/service-apis/v1")

	// Register InvokerManagement
	invokerManagerSwagger, err := invokermanagementapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading InvokerManagement swagger spec\n: %s", err)
	}
	invokerManagerSwagger.Servers = nil
	invokerManager := invokermanagement.NewInvokerManager(publishService)
	group = e.Group("/api-invoker-management/v1")
	group.Use(middleware.OapiRequestValidator(invokerManagerSwagger))
	invokermanagementapi.RegisterHandlersWithBaseURL(e, invokerManager, "/api-invoker-management/v1")

	// Register Security
	securitySwagger, err := publishserviceapi.GetSwagger()
	if err != nil {
		log.Fatalf("Error loading Security swagger spec\n: %s", err)
	}
	securitySwagger.Servers = nil
	securityService := security.NewSecurity(providerManager, publishService, invokerManager)
	group = e.Group("/capif-security/v1")
	group.Use(middleware.OapiRequestValidator(securitySwagger))
	securityapi.RegisterHandlersWithBaseURL(e, securityService, "/capif-security/v1")

	e.GET("/", hello)

	return e
}

func startWebServer(e *echo.Echo, port int) {
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", port)))
}

func keepServerAlive() {
	forever := make(chan int)
	<-forever
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}
