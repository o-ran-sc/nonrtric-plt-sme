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
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (es EventSubscription) Validate() error {
	if len(es.Events) == 0 {
		return errors.New("required attribute EventSubscription:events must contain at least one element")
	}

	for _, event := range es.Events {
		if err := validateEvent(event); err != nil {
			return errors.New("EventSubscription events contains invalid event")
		}
	}

	if len(strings.TrimSpace(string(es.NotificationDestination))) == 0 {
		return errors.New("EventSubscription missing required notificationDestination")
	}
	if _, err := url.ParseRequestURI(string(es.NotificationDestination)); err != nil {
		return fmt.Errorf("APIInvokerEnrolmentDetails has invalid notificationDestination, err=%s", err)
	}

	return nil
}

func validateEvent(event CAPIFEvent) error {
	switch event {
	case CAPIFEventACCESSCONTROLPOLICYUNAVAILABLE:
	case CAPIFEventACCESSCONTROLPOLICYUPDATE:
	case CAPIFEventAPIINVOKERAUTHORIZATIONREVOKED:
	case CAPIFEventAPIINVOKEROFFBOARDED:
	case CAPIFEventAPIINVOKERONBOARDED:
	case CAPIFEventAPIINVOKERUPDATED:
	case CAPIFEventAPITOPOLOGYHIDINGCREATED:
	case CAPIFEventAPITOPOLOGYHIDINGREVOKED:
	case CAPIFEventSERVICEAPIAVAILABLE:
	case CAPIFEventSERVICEAPIINVOCATIONFAILURE:
	case CAPIFEventSERVICEAPIINVOCATIONSUCCESS:
	case CAPIFEventSERVICEAPIUNAVAILABLE:
	case CAPIFEventSERVICEAPIUPDATE:
	default:
		return errors.New("wrong event type")
	}
	return nil
}
