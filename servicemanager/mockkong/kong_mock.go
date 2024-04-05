// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2024: OpenInfra Foundation Europe
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

package mockKong

import (
	"io"
	"net/http"

	echo "github.com/labstack/echo/v4"
)

func RegisterHandlers(e *echo.Echo) {
	// Handle Kong service and route endpoint mock responses here
	e.POST("/services", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading request body")
		}
		return c.String(http.StatusCreated, string(body))
	})

	e.POST("/services/api_id_apiName_helloworld/routes", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading request body")
		}
		return c.String(http.StatusCreated, string(body))
	})

	e.POST("/services/api_id_apiName1_helloworld/routes", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading request body")
		}
		return c.String(http.StatusCreated, string(body))
	})

	e.POST("/services/api_id_apiName2_helloworld/routes", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading request body")
		}
		return c.String(http.StatusCreated, string(body))
	})

	e.POST("/services/api_id_apiName1_app/routes", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading request body")
		}
		return c.String(http.StatusCreated, string(body))
	})

	e.POST("/services/api_id_apiName2_app/routes", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading request body")
		}
		return c.String(http.StatusCreated, string(body))
	})

	e.POST("/routes", func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error reading request body")
		}
		return c.String(http.StatusCreated, string(body))
	})

	e.GET("/services", func(c echo.Context) error {
		return c.String(http.StatusOK, "{}")
	})

	e.GET("/routes", func(c echo.Context) error {
		return c.String(http.StatusOK, "{}")
	})

	e.DELETE("/routes/api_id_apiName_helloworld", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/services/api_id_apiName_helloworld", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/routes/api_id_apiName1_helloworld", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/routes/api_id_apiName2_helloworld", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/routes/api_id_apiName1_app", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/routes/api_id_apiName2_app", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/services/api_id_apiName1_helloworld", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/services/api_id_apiName2_helloworld", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/services/api_id_apiName1_app", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	e.DELETE("/services/api_id_apiName2_app", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
}