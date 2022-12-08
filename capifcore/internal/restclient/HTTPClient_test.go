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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"oransc.org/nonrtric/capifcore/internal/restclient/mocks"
)

func TestRequestError_Error(t *testing.T) {
	assertions := require.New(t)
	actualError := RequestError{
		StatusCode: http.StatusBadRequest,
		Body:       []byte("error"),
	}
	assertions.Equal("Request failed due to error response with status: 400 and body: error", actualError.Error())
}

func TestPutOk(t *testing.T) {
	assertions := require.New(t)
	clientMock := mocks.HTTPClient{}

	clientMock.On("Do", mock.Anything).Return(&http.Response{
		StatusCode: http.StatusOK,
	}, nil)

	if err := Put("http://localhost:9990", []byte("body"), &clientMock); err != nil {
		t.Errorf("Put() error = %v, did not want error", err)
	}
	var actualRequest *http.Request
	clientMock.AssertCalled(t, "Do", mock.MatchedBy(func(req *http.Request) bool {
		actualRequest = req
		return true
	}))
	assertions.Equal(http.MethodPut, actualRequest.Method)
	assertions.Equal("http", actualRequest.URL.Scheme)
	assertions.Equal("localhost:9990", actualRequest.URL.Host)
	assertions.Equal("application/json", actualRequest.Header.Get("Content-Type"))
	body, _ := io.ReadAll(actualRequest.Body)
	expectedBody := []byte("body")
	assertions.Equal(expectedBody, body)
	clientMock.AssertNumberOfCalls(t, "Do", 1)
}

func Test_doErrorCases(t *testing.T) {
	assertions := require.New(t)
	type args struct {
		url              string
		mockReturnStatus int
		mockReturnBody   []byte
		mockReturnError  error
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Bad request should get RequestError",
			args: args{
				url:              "badRequest",
				mockReturnStatus: http.StatusBadRequest,
				mockReturnBody:   []byte("bad request"),
				mockReturnError:  nil,
			},
			wantErr: RequestError{
				StatusCode: http.StatusBadRequest,
				Body:       []byte("bad request"),
			},
		},
		{
			name: "Server unavailable should get error",
			args: args{
				url:             "serverUnavailable",
				mockReturnError: fmt.Errorf("Server unavailable"),
			},
			wantErr: fmt.Errorf("Server unavailable"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientMock := mocks.HTTPClient{}
			clientMock.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: tt.args.mockReturnStatus,
				Body:       io.NopCloser(bytes.NewReader(tt.args.mockReturnBody)),
			}, tt.args.mockReturnError)
			err := do("PUT", tt.args.url, nil, "", &clientMock)
			assertions.Equal(tt.wantErr, err, tt.name)
		})
	}
}
