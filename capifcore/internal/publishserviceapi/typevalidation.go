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
package publishserviceapi

import (
	"errors"
	//"fmt"
	"strings"
)

func (sd ServiceAPIDescription) Validate() error {
	if len(strings.TrimSpace(sd.ApiName)) == 0 {
		return errors.New("ServiceAPIDescription missing required apiName")
	}
	return nil
}

func (sd ServiceAPIDescription) ValidateAlreadyPublished(otherService ServiceAPIDescription) error {
	if sd.ApiName == otherService.ApiName {
		return errors.New("service with identical apiName is already published")
	}
	return nil
}
