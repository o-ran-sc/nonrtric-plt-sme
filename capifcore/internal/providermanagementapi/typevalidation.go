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

package providermanagementapi

import (
	"errors"
	"fmt"
	"strings"
)

func (ri RegistrationInformation) Validate() error {
	if len(strings.TrimSpace(ri.ApiProvPubKey)) == 0 {
		return errors.New("RegistrationInformation missing required apiProvPubKey")
	}
	return nil
}

func (fd APIProviderFunctionDetails) Validate() error {
	if len(strings.TrimSpace(string(fd.ApiProvFuncRole))) == 0 {
		return errors.New("APIProviderFunctionDetails missing required apiProvFuncRole")
	}
	switch role := fd.ApiProvFuncRole; role {
	case ApiProviderFuncRoleAEF:
	case ApiProviderFuncRoleAPF:
	case ApiProviderFuncRoleAMF:
	default:
		return errors.New("APIProviderFunctionDetails has invalid apiProvFuncRole")
	}

	return fd.RegInfo.Validate()
}

func (pd APIProviderEnrolmentDetails) Validate() error {
	if len(strings.TrimSpace(pd.RegSec)) == 0 {
		return errors.New("APIProviderEnrolmentDetails missing required regSec")
	}
	if pd.ApiProvFuncs != nil {
		return pd.validateFunctions()
	}
	return nil
}

func (pd APIProviderEnrolmentDetails) validateFunctions() error {
	for _, function := range *pd.ApiProvFuncs {
		err := function.Validate()
		if err != nil {
			return fmt.Errorf("apiProvFuncs contains invalid function: %s", err)
		}
	}
	return nil
}

func (pd APIProviderEnrolmentDetails) ValidateAlreadyRegistered(otherProvider APIProviderEnrolmentDetails) error {
	if pd.RegSec == otherProvider.RegSec {
		return errors.New("provider with identical regSec already registered")
	}
	return nil
}
