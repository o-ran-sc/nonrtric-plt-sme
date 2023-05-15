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
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifinvoker/internal/securityapi"
)

func GetTokenHandler(c echo.Context) error {
	log.Info("[Security API] in get token handler")
	return c.Render(http.StatusOK, "gettoken.html", map[string]interface{}{
		"isError":    false,
		"isResponse": false,
	})
}

func ObtainToken(server string) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Info("[Security API] in ObtainToken")
		securityId := c.FormValue("securityId")
		if securityId == "" {
			log.Error("[Security API] field securityId is needed")
			return c.Render(http.StatusBadRequest, "gettoken.html", map[string]interface{}{
				"isError":    true,
				"isResponse": false,
				"response":   "field securityId is needed",
			})
		}

		//server format: http://localhost:8090
		urlStr := server + "/capif-security/v1/securities/" + securityId + "/token"

		log.Infof("[Security API] url to capif core %v for securityId: %v", urlStr, securityId)

		data := url.Values{}
		data.Set("client_id", c.FormValue("clientId"))
		data.Set("client_secret", c.FormValue("clientSecret"))
		data.Set("grant_type", "client_credentials")
		data.Set("scope", c.FormValue("scope"))

		headers := map[string]string{
			"Content-Type":   "application/x-www-form-urlencoded",
			"Content-Length": strconv.Itoa(len(data.Encode())),
		}
		resp, err := makeRequest("POST", urlStr, headers, strings.NewReader(data.Encode()))
		if err != nil {
			log.Errorf("[Security API] %v", fmt.Sprintf("error: %v", err))
			return c.Render(http.StatusBadRequest, "gettoken.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   fmt.Sprintf("error: %v", err),
			})
		}

		var resToken securityapi.AccessTokenRsp
		if err = json.Unmarshal(resp, &resToken); err != nil {
			log.Error("[Security API] error unmarshaling parameter AccessTokenRsp as JSON")
			return c.Render(http.StatusBadRequest, "gettoken.html", map[string]interface{}{
				"isResponse": false,
				"isError":    true,
				"response":   "Error unmarshaling parameter AccessTokenRsp as JSON",
			})
		}

		// Return the rendered response HTML
		bytes, _ := json.Marshal(resToken)
		log.Infof("[Security API] jwt token fetch AccessTokenRsp is %v\n", resToken)
		return c.Render(http.StatusOK, "gettoken.html", map[string]interface{}{
			"isResponse": true,
			"isError":    false,
			"response":   string(bytes),
		})
	}
}
