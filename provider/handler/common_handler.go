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
	"io"
	"net/http"
)

func makeRequest(method, url string, headers map[string]string, data interface{}) ([]byte, error) {
	client := &http.Client{}

	// Create a new HTTP request with the specified method and URL
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Set any headers specified in the headers map
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// If there is data to send, marshal it to JSON and set it as the request body
	if data != nil {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(jsonBytes))
	}

	// Send the request and get the response
	if resp, err := client.Do(req); err == nil {
		if isResponseSuccess(resp.StatusCode) {
			defer resp.Body.Close()

			// Read the response body
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return respBody, nil
		} else {
			return nil, getRequestError(resp)
		}
	} else {
		return nil, err
	}
}

func isResponseSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode <= 299
}

func getRequestError(response *http.Response) error {
	defer response.Body.Close()
	responseData, _ := io.ReadAll(response.Body)

	return fmt.Errorf("message:  %v code: %v", string(responseData), response.StatusCode)
}
