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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifprov/internal/providermanagementapi"
)

func RegistrationHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "registration.html", map[string]interface{}{
		"isError":    false,
		"isResponse": false,
	})
}

func RegistrationFormHandler(server string) echo.HandlerFunc {
	return func(c echo.Context) error {

		url := server + "/api-provider-management/v1/registrations"
		log.Infof("[Register provider] url to capif core %v\n", url)

		var newProvider providermanagementapi.APIProviderEnrolmentDetails
		err := json.Unmarshal([]byte(c.FormValue("enrolmentDetails")), &newProvider)
		if err != nil {
			log.Error("[Register provider] error unmarshaling parameter enrolmentDetails as JSON")
			return c.Render(http.StatusBadRequest, "registration.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "Error unmarshaling parameter enrolmentDetails as JSON",
			})
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}
		resp, err := makeRequest("POST", url, headers, newProvider)
		if err != nil {
			log.Errorf("[Register provider] %v", fmt.Sprintf("error: %v", err))
			return c.Render(http.StatusBadRequest, "registration.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   fmt.Sprintf("error: %v", err),
			})
		}

		var resProvider providermanagementapi.APIProviderEnrolmentDetails
		err = json.Unmarshal(resp, &resProvider)
		if err != nil {
			log.Error("[Register provider] error unmarshaling parameter enrolmentDetails as JSON")
			return c.Render(http.StatusBadRequest, "registration.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "error unmarshaling parameter enrolmentDetails as JSON",
			})
		}

		bytes, _ := json.Marshal(resProvider)
		log.Infof("[Register provider] Api Provider domain %v has been register\n", resProvider.ApiProvDomId)
		return c.Render(http.StatusOK, "registration.html", map[string]interface{}{
			"isResponse": true,
			"isError":    false,
			"response":   string(bytes),
		})
	}
}
