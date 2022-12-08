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
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/eventsapi"
	"oransc.org/nonrtric/capifcore/internal/restclient"
)

type EventService struct {
	notificationChannel chan eventsapi.EventNotification
	client              restclient.HTTPClient
	subscriptions       map[string]eventsapi.EventSubscription
	idCounter           uint
	lock                sync.Mutex
}

func NewEventService(c restclient.HTTPClient) *EventService {
	es := EventService{
		notificationChannel: make(chan eventsapi.EventNotification),
		client:              c,
		subscriptions:       make(map[string]eventsapi.EventSubscription),
	}
	es.start()
	return &es
}

func (es *EventService) start() {
	go es.handleIncomingEvents()
}

func (es *EventService) handleIncomingEvents() {
	for event := range es.notificationChannel {
		es.handleEvent(event)
	}
}
func (es *EventService) GetNotificationChannel() chan<- eventsapi.EventNotification {
	return es.notificationChannel
}

func (es *EventService) PostSubscriberIdSubscriptions(ctx echo.Context, subscriberId string) error {
	newSubscription, err := getEventSubscriptionFromRequest(ctx)
	errMsg := "Unable to register subscription due to %s."
	if err != nil {
		return sendCoreError(ctx, http.StatusBadRequest, fmt.Sprintf(errMsg, err))
	}
	uri := ctx.Request().Host + ctx.Request().URL.String()
	subId := es.getSubscriptionId(subscriberId)
	es.addSubscription(subId, newSubscription)
	ctx.Response().Header().Set(echo.HeaderLocation, ctx.Scheme()+`://`+path.Join(uri, subId))
	err = ctx.JSON(http.StatusCreated, newSubscription)
	if err != nil {
		// Something really bad happened, tell Echo that our handler failed
		return err
	}

	return nil
}

func (es *EventService) DeleteSubscriberIdSubscriptionsSubscriptionId(ctx echo.Context, subscriberId string, subscriptionId string) error {
	return ctx.NoContent(http.StatusNotImplemented)
}

func getEventSubscriptionFromRequest(ctx echo.Context) (eventsapi.EventSubscription, error) {
	var subscription eventsapi.EventSubscription
	err := ctx.Bind(&subscription)
	if err != nil {
		return eventsapi.EventSubscription{}, fmt.Errorf("invalid format for subscription")
	}
	return subscription, nil
}

func (es *EventService) handleEvent(event eventsapi.EventNotification) {
	subscription := es.getSubscription(event.SubscriptionId)
	if subscription != nil {
		e, _ := json.Marshal(event)
		if error := restclient.Put(string(subscription.NotificationDestination), []byte(e), es.client); error != nil {
			log.Error("Unable to send event")
		}
	}
}

func (es *EventService) getSubscriptionId(subscriberId string) string {
	es.idCounter++
	return subscriberId + strconv.FormatUint(uint64(es.idCounter), 10)
}

func (es *EventService) addSubscription(subId string, subscription eventsapi.EventSubscription) {
	es.lock.Lock()
	es.subscriptions[subId] = subscription
	es.lock.Unlock()
}

func (es *EventService) getSubscription(subId string) *eventsapi.EventSubscription {
	es.lock.Lock()
	defer es.lock.Unlock()
	if sub, ok := es.subscriptions[subId]; ok {
		return &sub
	} else {
		return nil
	}
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendCoreError(ctx echo.Context, code int, message string) error {
	pd := common29122.ProblemDetails{
		Cause:  &message,
		Status: &code,
	}
	err := ctx.JSON(code, pd)
	return err
}
