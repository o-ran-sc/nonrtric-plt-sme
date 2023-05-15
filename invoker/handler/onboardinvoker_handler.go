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
	invokerapi "oransc.org/nonrtric/capifinvoker/internal/invokermanagementapi"
)

func OnboardingInvokerHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "onboardinvoker.html", map[string]interface{}{
		"isError":    false,
		"isResponse": false,
	})
}

func OnboardInvoker(server string) echo.HandlerFunc {
	return func(c echo.Context) error {

		url := server + "/api-invoker-management/v1/onboardedInvokers"
		log.Infof("[Register invoker] url to capif core %v\n", url)

		var newInvoker invokerapi.APIInvokerEnrolmentDetails
		fmt.Printf("Getting enrolment details from UI:: %+v", c.FormValue("enrolmentDetails"))
		err := json.Unmarshal([]byte(c.FormValue("enrolmentDetails")), &newInvoker)
		if err != nil {
			fmt.Printf("error: %+v", err)
			log.Error("[Register invoker] error unmarshaling parameter enrolmentDetails as JSON")
			return c.Render(http.StatusBadRequest, "onboardinvoker.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "Error unmarshaling parameter enrolmentDetails as JSON",
			})
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}
		jsonBytes, err := json.Marshal(newInvoker)
		if err != nil {
			return c.Render(http.StatusBadRequest, "onboardinvoker.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "Error marshaling parameter enrolmentDetails before doing request",
			})
		}

		resp, err := makeRequest("POST", url, headers, bytes.NewReader(jsonBytes))
		if err != nil {
			log.Errorf("[Register invoker] %v", fmt.Sprintf("error: %v", err))
			return c.Render(http.StatusBadRequest, "onboardinvoker.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   fmt.Sprintf("error: %v", err),
			})
		}

		var resInvoker invokerapi.APIInvokerEnrolmentDetails
		err = json.Unmarshal(resp, &resInvoker)
		if err != nil {
			log.Error("[Register invoker] error unmarshaling response enrolmentDetails as JSON")
			return c.Render(http.StatusBadRequest, "onboardinvoker.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "error unmarshaling parameter enrolmentDetails as JSON",
			})
		}

		bytes, _ := json.Marshal(resInvoker)
		log.Infof("[Register invoker] new invoker register with id %v\n", *resInvoker.ApiInvokerId)
		return c.Render(http.StatusOK, "onboardinvoker.html", map[string]interface{}{
			"isResponse": true,
			"isError":    false,
			"response":   string(bytes),
		})
	}
}
