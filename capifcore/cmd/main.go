// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022: Nordix Foundation. All rights reserved.
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

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"helm.sh/helm/v3/pkg/cli"
	log "github.com/sirupsen/logrus"

	"oransc.org/nonrtric/capifcore/internal/helmmanagement"
	config "oransc.org/nonrtric/capifcore/internal/config"
	"oransc.org/nonrtric/capifcore/internal/keycloak"

	"oransc.org/nonrtric/capifcore"
)

func main() {
	var url string
	var helmManager helmmanagement.HelmManager
	var repoName string

	var port = flag.Int("port", 8090, "Port for CAPIF Core Function HTTP server")
	var secPort = flag.Int("secPort", 4433, "Port for CAPIF Core Function HTTPS server")
	flag.StringVar(&url, "chartMuseumUrl", "", "ChartMuseum URL")
	flag.StringVar(&repoName, "repoName", "capifcore", "Repository name")
	var logLevelStr = flag.String("loglevel", "Info", "Log level")
	var certPath = flag.String("certPath", "certs/cert.pem", "Path for server certificate")
	var keyPath = flag.String("keyPath", "certs/key.pem", "Path for server private key")

	flag.Parse()

	if loglevel, err := log.ParseLevel(*logLevelStr); err == nil {
		log.SetLevel(loglevel)
	}

	// Add Helm repo
	helmManager = helmmanagement.NewHelmManager(cli.New())
	err := helmManager.SetUpRepo(repoName, url)
	if err != nil {
		log.Warnf("No Helm repo added due to: %s", err.Error())
	}

	// Read configuration file
	cfg, err := config.ReadKeycloakConfigFile("configs")
	if err != nil {
		log.Fatalf("Error loading configuration file\n: %s", err)
	}
	km := keycloak.NewKeycloakManager(cfg, &http.Client{})

	eWeb := echo.New()
	capifcore.RegisterHandlers(eWeb, helmManager, km)
	go startWebServer(eWeb, *port)

	eHttpsWeb := echo.New()
	capifcore.RegisterHandlers(eHttpsWeb, helmManager, km)
	go startHttpsWebServer(eHttpsWeb, *secPort, *certPath, *keyPath)

	log.Info("Server started and listening on port: ", *port)
	keepServerAlive()
}

func startWebServer(e *echo.Echo, port int) {
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", port)))
}

func startHttpsWebServer(e *echo.Echo, port int, certPath string, keyPath string) {
	e.Logger.Fatal(e.StartTLS(fmt.Sprintf("0.0.0.0:%d", port), certPath, keyPath))
}

func keepServerAlive() {
	forever := make(chan int)
	<-forever
}
