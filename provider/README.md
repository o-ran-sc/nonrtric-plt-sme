<!--
-
========================LICENSE_START=================================
O-RAN-SC
%%
Copyright (C) 2023: Nordix Foundation
%%
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
========================LICENSE_END===================================

-->

# O-RAN-SC Non-RealTime RIC CAPIF Provider Stub

This is a Go implementation of a stub for the CAPIF Provider function, which is based on the 3GPP "29.222 Common API Framework for 3GPP Northbound APIs (CAPIF)" interfaces, see https://portal.3gpp.org/desktopmodules/Specifications/SpecificationDetails.aspx?specificationId=3450.

This stub offers a user interface that helps to test the functionalities implemented in the O-RAN-SC Capif implementation and the supported features are the following:

- Registers a new API Provider domain with API provider domain functions profiles.
- Publish a new API
- Retrieve all published APIs

### Registers a new API Provider domain with API provider domain functions profiles.

This service operation is used by an API management function to register API provider domain functions as a recognized API provider of CAPIF domain.

<img src="docs/Register API Provider Domain.svg">

The request from the provider domain should include API provider Enrolment Details, consisting of details of all API provider domain functions, for example:

```
{
    "apiProvDomInfo": "Provider domain",
    "apiProvFuncs": [
        {
            "apiProvFuncInfo": "rApp as APF",
            "apiProvFuncRole": "APF",
            "regInfo": {
                "apiProvPubKey": "APF-PublicKey"
            }
        },
        {
            "apiProvFuncInfo": "rApp as AEF",
            "apiProvFuncRole": "AEF",
            "regInfo": {
                "apiProvPubKey": "AEF-PublicKey"
            }
        },
        {
            "apiProvFuncInfo": "rApp as AMF",
            "apiProvFuncRole": "AMF",
            "regInfo": {
                "apiProvPubKey": "AMF-PublicKey"
            }
        },
        {
            "apiProvFuncInfo": "Gateway as entrypoint AEF",
            "apiProvFuncRole": "AEF",
            "regInfo": {
                "apiProvPubKey": "AEF-Gateway-PublicKey"
            }
        }
    ],
    "regSec": "PSK"
}
```

The CAPIF core proceeds to register the provider and creates Ids for the API provider domain functions that will be return as part of the response message.

### Publish a new API

This service operation is used by an API publishing function to publish service APIs on the CAPIF core function.

<img src="docs/Publish a new API.svg">

The CAPIF supports publishing service APIs by the API provider. The API publishing function can be within PLMN trust domain or within 3rd party trust domain.

In order to publish a new API, the APF should be registered in the CAPIF core (apfId is required) along with the Service API information.

The Service API information includes:
- Service API name
- API provider name (optional)
- Service API type
- Communication type
- Serving Area Information (optional)
- AEF location (optional)
- Interface details (e.g. IP address, port number, URI)
- Protocols
- Version numbers
- Data format

```
{
    "apiName": "example",
    "description": "Example API of rApp B",
    "aefProfiles": [
        {
            "aefId": "AEF_id_rApp_as_AEF",
            "description": "Example rApp B as AEF",
            "versions": [
                {
                    "apiVersion": "v1",
                    "resources": [
                        {
                            "resourceName": "example",
                            "commType": "REQUEST_RESPONSE",
                            "uri": "/example/subscription/subscription_id_1",
                            "operations": [
                                "GET"
                            ]
                        }
                    ]
                }
            ],
            "protocol": "HTTP_1_1",
			"securityMethods": ["PSK"],
			"interfaceDescriptions": [
				{
				  "ipv4Addr": "string",
				  "port": 65535,
				  "securityMethods": ["PKI"]
				},
				{
				  "ipv4Addr": "string",
				  "port": 65535,
				  "securityMethods": ["PKI"]
				}
			  ]
        }
    ]
}
```


### Retrieve all published APIs

This service operation is used by an API publishing function to retrieve service APIs from the CAPIF core function.

<img src="docs/Retrieve all published APIs.svg">

Respond includes requested API Information.

## Build application

To build the application, run the following command:

    go build

The application can also be built as a Docker image, by using the following command:

    docker build . -t capifprov

## Run

To run the provider from the command line, run the following commands from this folder.

    ./capifprov [-port <port (default 9090)>] [-capifCoreUrl <URL to Capif core (default http://localhost:8090)>] [-loglevel <log level (default Info)>]
