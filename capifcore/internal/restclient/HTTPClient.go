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
	Get(url string) (*http.Response, error)
}

type RequestError struct {
	StatusCode int
	Body       []byte
}

func (pe RequestError) Error() string {
	return fmt.Sprintf("Request failed due to error response with status: %v and body: %v", pe.StatusCode, string(pe.Body))
}

func Get(url string, header map[string]string, client HTTPClient) ([]byte, error) {
	return do(http.MethodGet, url, nil, header, client)
}

func Put(url string, body []byte, client HTTPClient) error {
	var header = map[string]string{"Content-Type": ContentTypeJSON}
	_, err := do(http.MethodPut, url, body, header, client)
	return err
}

func Post(url string, body []byte, header map[string]string, client HTTPClient) error {
	_, err := do(http.MethodPost, url, body, header, client)
	return err
}

func do(method string, url string, body []byte, header map[string]string, client HTTPClient) ([]byte, error) {
	if req, reqErr := http.NewRequest(method, url, nil); reqErr == nil {
		if len(header) > 0 {
			setHeader(req, header)
		}
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		if response, respErr := client.Do(req); respErr == nil {
			if isResponseSuccess(response.StatusCode) {
				fmt.Printf("HTTP client:: response statuscode:: %v body:: %v\n", response.StatusCode, response.Body)
				defer response.Body.Close()

				// Read the response body
				respBody, err := io.ReadAll(response.Body)
				if err != nil {
					return nil, err
				}
				return respBody, nil
			} else {
				return nil, getRequestError(response)
			}
		} else {
			return nil, respErr
		}
	} else {
		return nil, reqErr
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
