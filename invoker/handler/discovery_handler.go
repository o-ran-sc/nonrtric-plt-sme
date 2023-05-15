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
	"net/url"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifinvoker/internal/discoverserviceapi"
)

func DiscoverAPIs(server string) echo.HandlerFunc {
	return func(c echo.Context) error {

		invokerId := c.FormValue("api-invoker-id")
		if invokerId == "" {
			return c.Render(http.StatusOK, "discovery.html", map[string]interface{}{
				"isResponse": false,
				"isError":    false,
			})
		}

		//server format: http://localhost:8090
		urlStr, _ := url.Parse(server + "/service-apis/v1/allServiceAPIs")
		log.Infof("[Discovery API] apis to %v for invokerId: %v", urlStr, invokerId)

		//TODO check how to remove empty parameters
		c.Request().ParseForm()
		params := c.Request().Form
		for key, val := range params {
			if len(val) == 0 || val[0] == "" {
				params.Del(key)
			}
		}
		urlStr.RawQuery = params.Encode()

		headers := map[string]string{
			"Content-Type": "text/plain",
		}
		resp, err := makeRequest("GET", urlStr.String(), headers, nil)
		if err != nil {
			log.Errorf("[Discovery API] %v", fmt.Sprintf("error: %v", err))
			return c.Render(http.StatusBadRequest, "discovery.html", map[string]interface{}{
				"response":   fmt.Sprintf("error: %v", err),
				"isError":    true,
				"isResponse": false,
			})
		}
		log.Infof("[Discovery API] Response from service: %+v error: %v\n", string(resp), err)

		var resAPIs discoverserviceapi.DiscoveredAPIs
		err = json.Unmarshal(resp, &resAPIs)
		if err != nil {
			log.Error("[Discovery API] error unmarshaling parameter DiscoveredAPIs as JSON")
			return c.Render(http.StatusBadRequest, "discovery.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "error unmarshaling parameter []DiscoveredAPIs as JSON",
			})
		}

		if len(*resAPIs.ServiceAPIDescriptions) == 0 {
			log.Info("[Discovery API] There are no APIs availables for the specified parameters.")
			return c.Render(http.StatusOK, "discovery.html", map[string]interface{}{
				"isResponse": false,
				"isError":    false,
				"isEmpty":    true,
				"response":   "There are no APIs availables for the specified parameters.",
			})
		}

		bytes, _ := json.Marshal(resAPIs)
		return c.Render(http.StatusOK, "discovery.html", map[string]interface{}{
			"isResponse": true,
			"response":   string(bytes),
		})
	}
}
