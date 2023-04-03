// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2021: Nordix Foundation
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

package restclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

const ContentTypeJSON = "application/json"
const ContentTypePlain = "text/plain"

//go:generate mockery --name HTTPClient
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type RequestError struct {
	StatusCode int
	Body       []byte
}

func (pe RequestError) Error() string {
	return fmt.Sprintf("Request failed due to error response with status: %v and body: %v", pe.StatusCode, string(pe.Body))
}

func Put(url string, body []byte, client HTTPClient) error {
	var header = map[string]string{"Content-Type": ContentTypeJSON}
	return do(http.MethodPut, url, body, header, client)
}

func Post(url string, body []byte, header map[string]string, client HTTPClient) error {
	return do(http.MethodPost, url, body, header, client)
}

func do(method string, url string, body []byte, header map[string]string, client HTTPClient) error {
	if req, reqErr := http.NewRequest(method, url, bytes.NewBuffer(body)); reqErr == nil {
		if len(header) > 0 {
			setHeader(req, header)
		}
		if response, respErr := client.Do(req); respErr == nil {
			if isResponseSuccess(response.StatusCode) {
				return nil
			} else {
				return getRequestError(response)
			}
		} else {
			return respErr
		}
	} else {
		return reqErr
	}
}

func setHeader(req *http.Request, header map[string]string) {
	for key, element := range header {
		req.Header.Set(key, element)
	}
}

func isResponseSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode <= 299
}

func getRequestError(response *http.Response) RequestError {
	defer response.Body.Close()
	responseData, _ := io.ReadAll(response.Body)
	putError := RequestError{
		StatusCode: response.StatusCode,
		Body:       responseData,
	}
	return putError
}
