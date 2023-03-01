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

package securityapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"oransc.org/nonrtric/capifcore/internal/common29122"
	publishapi "oransc.org/nonrtric/capifcore/internal/publishserviceapi"
)

func TestPrepareNewSecurityContext(t *testing.T) {
	apiId := "app-management"
	aefId := "aefId"
	description := "Description"
	services := []publishapi.ServiceAPIDescription{
		{
			AefProfiles: &[]publishapi.AefProfile{
				{
					AefId: aefId,
					Versions: []publishapi.Version{
						{
							Resources: &[]publishapi.Resource{
								{
									CommType: "REQUEST_RESPONSE",
								},
							},
						},
					},
					SecurityMethods: &[]publishapi.SecurityMethod{
						publishapi.SecurityMethodPKI,
					},
				},
			},
			ApiId:       &apiId,
			Description: &description,
		},
	}

	servSecurityUnderTest := ServiceSecurity{
		NotificationDestination: common29122.Uri("http://golang.cafe/"),
		SecurityInfo: []SecurityInformation{
			{
				PrefSecurityMethods: []publishapi.SecurityMethod{
					publishapi.SecurityMethodOAUTH,
				},
			},
		},
	}

	err := servSecurityUnderTest.PrepareNewSecurityContext(services)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not found ")
	assert.Contains(t, err.Error(), "security method")

	servSecurityUnderTest.SecurityInfo = []SecurityInformation{
		{
			ApiId: &apiId,
			AefId: &aefId,
			PrefSecurityMethods: []publishapi.SecurityMethod{
				publishapi.SecurityMethodOAUTH,
			},
		},
	}

	servSecurityUnderTest.PrepareNewSecurityContext(services)
	assert.Equal(t, publishapi.SecurityMethodPKI, *servSecurityUnderTest.SecurityInfo[0].SelSecurityMethod)

	servSecurityUnderTest.SecurityInfo = []SecurityInformation{
		{
			ApiId: &apiId,
			PrefSecurityMethods: []publishapi.SecurityMethod{
				publishapi.SecurityMethodOAUTH,
			},
			InterfaceDetails: &publishapi.InterfaceDescription{
				SecurityMethods: &[]publishapi.SecurityMethod{
					publishapi.SecurityMethodPSK,
				},
			},
		},
	}

	servSecurityUnderTest.PrepareNewSecurityContext(services)
	assert.Equal(t, publishapi.SecurityMethodPSK, *servSecurityUnderTest.SecurityInfo[0].SelSecurityMethod)

}
