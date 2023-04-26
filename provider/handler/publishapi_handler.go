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
	"oransc.org/nonrtric/capifprov/internal/publishserviceapi"
)

func PublishapiHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "publishapi.html", map[string]interface{}{
		"isError":    false,
		"isResponse": false,
	})
}

func PublishApiFormHandler(server string) echo.HandlerFunc {
	return func(c echo.Context) error {

		apfId := c.FormValue("apfId")
		if apfId == "" {
			return c.Render(http.StatusBadRequest, "publishapi.html", map[string]interface{}{
				"isError":    true,
				"isResponse": false,
				"response":   "field apfId is needed",
			})
		}

		//server format: http://localhost:8090
		url := server + "/published-apis/v1/" + apfId + "/service-apis"

		log.Infof("[Publish API] url to capif core %v for aefId: %v", url, apfId)
		var apiDescription publishserviceapi.ServiceAPIDescription

		err := json.Unmarshal([]byte(c.FormValue("apiDescription")), &apiDescription)
		if err != nil {
			log.Error("[Publish API] error unmarshaling parameter ServiceAPIDescription as JSON")
			return c.Render(http.StatusBadRequest, "publishapi.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "error unmarshaling parameter ServiceAPIDescription as JSON",
			})
		}

		headers := map[string]string{
			"Content-Type": "application/json",
		}
		resp, err := makeRequest("POST", url, headers, apiDescription)
		if err != nil {
			log.Errorf("[Publish API] %v", fmt.Sprintf("error: %v", err))
			return c.Render(http.StatusBadRequest, "publishapi.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   fmt.Sprintf("error: %v", err),
			})
		}

		var resAPI publishserviceapi.ServiceAPIDescription
		err = json.Unmarshal(resp, &resAPI)
		if err != nil {
			log.Error("[Publish API] error unmarshaling parameter ServiceAPIDescription as JSON")
			return c.Render(http.StatusBadRequest, "publishapi.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "Error unmarshaling parameter ServiceAPIDescription as JSON",
			})
		}

		bytes, _ := json.Marshal(resAPI)
		log.Infof("[Publish API] API %v with the id: %v has been register", resAPI.ApiName, *resAPI.ApiId)
		return c.Render(http.StatusOK, "publishapi.html", map[string]interface{}{
			"isResponse": true,
			"response":   string(bytes),
		})
	}
}
