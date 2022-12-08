// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022: Nordix Foundation
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

package eventservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/eventsapi"
	"oransc.org/nonrtric/capifcore/internal/restclient"
)

func TestRegisterSubscriptions(t *testing.T) {
	subscription1 := eventsapi.EventSubscription{
		Events: []eventsapi.CAPIFEvent{
			eventsapi.CAPIFEventSERVICEAPIAVAILABLE,
		},
		NotificationDestination: common29122.Uri("notificationUrl"),
	}
	serviceUnderTest, requestHandler := getEcho(nil)
	subscriberId := "subscriberId"

	result := testutil.NewRequest().Post("/"+subscriberId+"/subscriptions").WithJsonBody(subscription1).Go(t, requestHandler)
	assert.Equal(t, http.StatusCreated, result.Code())
	var resultEvent eventsapi.EventSubscription
	err := result.UnmarshalBodyToObject(&resultEvent)
	assert.NoError(t, err, "error unmarshaling response")
	assert.Equal(t, resultEvent, subscription1)
	assert.Regexp(t, "http://example.com/"+subscriberId+"/subscriptions/"+subscriberId+"[0-9]+", result.Recorder.Header().Get(echo.HeaderLocation))
	subscriptionId1 := path.Base(result.Recorder.Header().Get(echo.HeaderLocation))

	subscription2 := subscription1
	subscription2.Events = []eventsapi.CAPIFEvent{
		eventsapi.CAPIFEventAPIINVOKERUPDATED,
	}
	result = testutil.NewRequest().Post("/"+subscriberId+"/subscriptions").WithJsonBody(subscription2).Go(t, requestHandler)
	assert.Regexp(t, "http://example.com/"+subscriberId+"/subscriptions/"+subscriberId+"[0-9]+", result.Recorder.Header().Get(echo.HeaderLocation))
	subscriptionId2 := path.Base(result.Recorder.Header().Get(echo.HeaderLocation))

	assert.NotEqual(t, subscriptionId1, subscriptionId2)
	registeredSub1 := serviceUnderTest.getSubscription(subscriptionId1)
	assert.Equal(t, subscription1, *registeredSub1)
	registeredSub2 := serviceUnderTest.getSubscription(subscriptionId2)
	assert.Equal(t, subscription2, *registeredSub2)
}

func TestSendEvent(t *testing.T) {
	notificationUrl := "url"
	apiIds := []string{"apiId"}
	subId := "sub1"
	newEvent := eventsapi.EventNotification{
		SubscriptionId: subId,
		EventDetail: &eventsapi.CAPIFEventDetail{
			ApiIds: &apiIds,
		},
		Events: eventsapi.CAPIFEventSERVICEAPIAVAILABLE,
	}
	wg := sync.WaitGroup{}
	clientMock := NewTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() == notificationUrl {
			assert.Equal(t, req.Method, "PUT")
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, newEvent, getBodyAsEvent(req, t))
			wg.Done()
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
				Header:     make(http.Header), // Must be set to non-nil value or it panics
			}
		}
		t.Error("Wrong call to client: ", req)
		t.Fail()
		return nil
	})
	serviceUnderTest, _ := getEcho(clientMock)

	subscription := eventsapi.EventSubscription{
		Events: []eventsapi.CAPIFEvent{
			eventsapi.CAPIFEventSERVICEAPIAVAILABLE,
		},
		NotificationDestination: common29122.Uri(notificationUrl),
	}
	serviceUnderTest.addSubscription(subId, subscription)

	wg.Add(1)
	go func() {
		serviceUnderTest.GetNotificationChannel() <- newEvent
	}()

	if waitTimeout(&wg, 1*time.Second) {
		t.Error("Not all calls to server were made")
		t.Fail()
	}

}

func getEcho(client restclient.HTTPClient) (*EventService, *echo.Echo) {
	swagger, err := eventsapi.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	es := NewEventService(client)

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(middleware.OapiRequestValidator(swagger))

	eventsapi.RegisterHandlers(e, es)
	return es, e
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func getBodyAsEvent(req *http.Request, t *testing.T) eventsapi.EventNotification {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(req.Body); err != nil {
		t.Fail()
	}
	var event eventsapi.EventNotification
	err := json.Unmarshal(buf.Bytes(), &event)
	if err != nil {
		t.Fail()
	}
	return event
}
