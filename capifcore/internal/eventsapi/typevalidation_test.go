// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023: Nordix Foundation
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

package eventsapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEventSubscription(t *testing.T) {
	subUnderTest := EventSubscription{}

	err := subUnderTest.Validate()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "required")
	assert.Contains(t, err.Error(), "events")

	var invalidEventType CAPIFEvent = "invalid"
	subUnderTest.Events = []CAPIFEvent{invalidEventType}
	err = subUnderTest.Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid")
	assert.Contains(t, err.Error(), "events")

	subUnderTest.Events = []CAPIFEvent{CAPIFEventAPIINVOKERONBOARDED}
	err = subUnderTest.Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing")
	assert.Contains(t, err.Error(), "notificationDestination")

	subUnderTest.NotificationDestination = "invalid dest"
	err = subUnderTest.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "invalid")
		assert.Contains(t, err.Error(), "notificationDestination")
	}

	subUnderTest.NotificationDestination = "http://golang.cafe/"
	err = subUnderTest.Validate()
	assert.Nil(t, err)
}
