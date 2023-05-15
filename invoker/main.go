// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023: Nordix Foundation
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
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
	"oransc.org/nonrtric/capifinvoker/handler"
)

type TemplateRegistry struct {
	templates map[string]*template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "base", data)
}

func main() {

	// Echo instance
	e := echo.New()
	e.Static("/", "view")
	var capifCoreUrl string
	flag.StringVar(&capifCoreUrl, "capifCoreUrl", "http://localhost:8090", "Url for CAPIF core")
	var logLevelStr = flag.String("loglevel", "Info", "Log level")
	var port = flag.Int("port", 9090, "Port for CAPIF Provider")

	flag.Parse()

	if loglevel, err := log.ParseLevel(*logLevelStr); err == nil {
		log.SetLevel(loglevel)
	}

	templates := make(map[string]*template.Template)
	templates["home.html"] = template.Must(template.ParseFiles("view/home.html", "view/base.html"))
	templates["discovery.html"] = template.Must(template.ParseFiles("view/discovery.html", "view/base.html"))
	templates["onboardinvoker.html"] = template.Must(template.ParseFiles("view/onboardinvoker.html", "view/base.html"))
	templates["securitymethod.html"] = template.Must(template.ParseFiles("view/securitymethod.html", "view/base.html"))
	templates["gettoken.html"] = template.Must(template.ParseFiles("view/gettoken.html", "view/base.html"))

	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	// Route => handler
	e.GET("/", handler.HomeHandler)
	e.POST("/", handler.HomeHandler)

	e.GET("/discovery", handler.DiscoverAPIs(capifCoreUrl))

	e.GET("/onboardinvoker", handler.OnboardingInvokerHandler)
	e.POST("/onboardinvoker", handler.OnboardInvoker(capifCoreUrl))

	e.GET("/securitymethod", handler.SecurityMethodHandler)
	e.POST("/securitymethod", handler.ObtainSecurityMethod(capifCoreUrl))

	e.GET("/gettoken", handler.GetTokenHandler)
	e.POST("/gettoken", handler.ObtainToken(capifCoreUrl))

	// Start the web server
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
}
