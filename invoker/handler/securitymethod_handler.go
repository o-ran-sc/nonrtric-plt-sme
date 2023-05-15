// -
//
//	========================LICENSE_START=================================
//	O-RAN-SC
//	%%
//	Copyright (C) 2023: Nordix Foundation
//	%%
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	     http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.
//	========================LICENSE_END===================================
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifinvoker/internal/securityapi"
)

func SecurityMethodHandler(c echo.Context) error {
	log.Info("[Security API] in security method handler")
	return c.Render(http.StatusOK, "securitymethod.html", map[string]interface{}{
		"isError":    false,
		"isResponse": false,
	})
}

func ObtainSecurityMethod(server string) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info("[Security API] in ObtainSecurityMethod")
		invokerId := c.FormValue("invokerId")
		if invokerId == "" {
			log.Error("[Security API] field invokerId is needed")
			return c.Render(http.StatusBadRequest, "securitymethod.html", map[string]interface{}{
				"isError":    true,
				"isResponse": false,
				"response":   "field invokerId is needed",
			})
		}

		//server format: http://localhost:8090
		url := server + "/capif-security/v1/trustedInvokers/" + invokerId

		log.Infof("[Security API] url to capif core %v for invokerId: %v", url, invokerId)
		var servSecurity securityapi.ServiceSecurity

		err := json.Unmarshal([]byte(c.FormValue("servSecurity")), &servSecurity)
		if err != nil {
			log.Error("[Security API] error unmarshaling parameter ServiceSecurity as JSON")
			return c.Render(http.StatusBadRequest, "securitymethod.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "error unmarshaling parameter ServiceSecurity as JSON",
			})
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}
		jsonBytes, err := json.Marshal(servSecurity)
		if err != nil {
			return c.Render(http.StatusBadRequest, "securitymethod.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "Error marshaling parameter ServiceSecurity before doing request",
			})
		}
		resp, err := makeRequest("PUT", url, headers, bytes.NewReader(jsonBytes))
		if err != nil {
			log.Errorf("[Security API] %v", fmt.Sprintf("error: %v", err))
			return c.Render(http.StatusBadRequest, "securitymethod.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   fmt.Sprintf("error: %v", err),
			})
		}

		var resAPI securityapi.ServiceSecurity
		err = json.Unmarshal(resp, &resAPI)
		if err != nil {
			log.Error("[Security API] error unmarshaling parameter ServiceSecurity as JSON")
			return c.Render(http.StatusBadRequest, "securitymethod.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "Error unmarshaling parameter ServiceSecurity as JSON",
			})
		}

		// Return the rendered response HTML
		bytes, _ := json.Marshal(resAPI)
		return c.Render(http.StatusOK, "securitymethod.html", map[string]interface{}{
			"isResponse": true,
			"response":   string(bytes),
		})
	}
}
