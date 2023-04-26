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

	log "github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
	"oransc.org/nonrtric/capifprov/internal/publishserviceapi"
)

func GetApiRequest(server string) echo.HandlerFunc {
	return func(c echo.Context) error {

		aefId := c.FormValue("apfId")
		if aefId == "" {
			return c.Render(http.StatusOK, "getapi.html", map[string]interface{}{
				"isResponse": false,
				"isError":    false,
			})
		}

		//server format: http://localhost:8090
		url := server + "/published-apis/v1/" + aefId + "/service-apis"
		log.Infof("[Get API] to %v for aefId: %v", url, aefId)

		headers := map[string]string{
			"Content-Type": "text/plain",
		}
		resp, err := makeRequest("GET", url, headers, nil)
		if err != nil {
			log.Errorf("[Get API] %v", fmt.Sprintf("error: %v", err))
			return c.Render(http.StatusBadRequest, "getapi.html", map[string]interface{}{
				"response":   fmt.Sprintf("error: %v", err),
				"isError":    true,
				"isResponse": false,
			})
		}
		log.Infof("[Get API] Response from service: %+v error: %v\n", string(resp), err)

		var resAPIs []publishserviceapi.ServiceAPIDescription
		err = json.Unmarshal(resp, &resAPIs)
		if err != nil {
			log.Error("[Get API] error unmarshaling parameter ServiceAPIDescription as JSON")
			return c.Render(http.StatusBadRequest, "getapi.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "error unmarshaling parameter ServiceAPIDescription as JSON",
			})
		}

		bytes, _ := json.Marshal(resAPIs)
		log.Infof("[Get API] There are %v ServiceAPIDescription objects available", len(resAPIs))
		return c.Render(http.StatusOK, "getapi.html", map[string]interface{}{
			"isResponse": true,
			"response":   string(bytes),
		})
	}
}
