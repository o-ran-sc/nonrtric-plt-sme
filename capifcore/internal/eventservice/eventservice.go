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
	"k8s.io/utils/strings/slices"
	"oransc.org/nonrtric/capifcore/internal/common29122"
	"oransc.org/nonrtric/capifcore/internal/eventsapi"
	"oransc.org/nonrtric/capifcore/internal/publishserviceapi"
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

	log.Debug(es.subscriptions)
	if _, ok := es.subscriptions[subscriptionId]; ok {
		es.deleteSubscription(subscriptionId)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (es *EventService) deleteSubscription(subscriptionId string) {
	log.Debug("Deleting subscription", subscriptionId)
	es.lock.Lock()
	defer es.lock.Unlock()
	delete(es.subscriptions, subscriptionId)
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
	subsIds := es.getMatchingSubs(event)
	for _, subId := range subsIds {
		go es.sendEvent(event, subId)
	}
}

func (es *EventService) sendEvent(event eventsapi.EventNotification, subscriptionId string) {
	event.SubscriptionId = subscriptionId
	e, _ := json.Marshal(event)
	if error := restclient.Put(string(es.subscriptions[subscriptionId].NotificationDestination), []byte(e), es.client); error != nil {
		log.Error("Unable to send event")
	}
}

func (es *EventService) getMatchingSubs(event eventsapi.EventNotification) []string {
	es.lock.Lock()
	defer es.lock.Unlock()
	matchingTypeSubs := es.filterOnEventType(event)
	matchingSubs := []string{}
	for _, subId := range matchingTypeSubs {
		subscription := es.subscriptions[subId]
		if subscription.EventFilters == nil || event.EventDetail == nil {
			matchingSubs = append(matchingSubs, subId)
		} else if matchesFilters(event.EventDetail.ApiIds, *subscription.EventFilters, getApiIdsFromFilter) &&
			matchesFilters(event.EventDetail.ApiInvokerIds, *subscription.EventFilters, getInvokerIdsFromFilter) &&
			matchesFilters(getAefIdsFromEvent(event.EventDetail.ServiceAPIDescriptions), *subscription.EventFilters, getAefIdsFromFilter) {
			matchingSubs = append(matchingSubs, subId)
		}
	}

	return matchingSubs
}

func (es *EventService) filterOnEventType(event eventsapi.EventNotification) []string {
	matchingSubs := []string{}
	for subId, subInfo := range es.subscriptions {
		if slices.Contains(asStrings(subInfo.Events), string(event.Events)) {
			matchingSubs = append(matchingSubs, subId)
		}
	}
	return matchingSubs
}

func matchesFilters(eventIds *[]string, filters []eventsapi.CAPIFEventFilter, getIds func(eventsapi.CAPIFEventFilter) *[]string) bool {
	if len(filters) == 0 || eventIds == nil {
		return true
	}
	for _, id := range *eventIds {
		filter := filters[0]
		filterIds := getIds(filter)
		if filterIds == nil || len(*filterIds) == 0 {
			return matchesFilters(eventIds, filters[1:], getIds)
		}
		return slices.Contains(*getIds(filter), id) && matchesFilters(eventIds, filters[1:], getIds)
	}
	return true
}

func getApiIdsFromFilter(filter eventsapi.CAPIFEventFilter) *[]string {
	return filter.ApiIds
}

func getInvokerIdsFromFilter(filter eventsapi.CAPIFEventFilter) *[]string {
	return filter.ApiInvokerIds
}

func getAefIdsFromEvent(serviceAPIDescriptions *[]publishserviceapi.ServiceAPIDescription) *[]string {
	aefIds := []string{}
	if serviceAPIDescriptions == nil {
		return &aefIds
	}
	for _, serviceDescription := range *serviceAPIDescriptions {
		if serviceDescription.AefProfiles == nil {
			return &aefIds
		}
		for _, profile := range *serviceDescription.AefProfiles {
			aefIds = append(aefIds, profile.AefId)
		}
	}
	return &aefIds
}

func getAefIdsFromFilter(filter eventsapi.CAPIFEventFilter) *[]string {
	return filter.AefIds
}

func asStrings(events []eventsapi.CAPIFEvent) []string {
	asStrings := make([]string, len(events))
	for i, event := range events {
		asStrings[i] = string(event)
	}
	return asStrings
}

func (es *EventService) getSubscriptionId(subscriberId string) string {
	es.idCounter++
	return subscriberId + strconv.FormatUint(uint64(es.idCounter), 10)
}

func (es *EventService) addSubscription(subId string, subscription eventsapi.EventSubscription) {
	es.lock.Lock()
	defer es.lock.Unlock()
	es.subscriptions[subId] = subscription
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
