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

# O-RAN-SC Non-RealTime RIC CAPIF Invoker Stub

This is a Go implementation of a stub for the CAPIF Invoker function, which is based on the 3GPP "29.222 Common API Framework for 3GPP Northbound APIs (CAPIF)" interfaces, see https://portal.3gpp.org/desktopmodules/Specifications/SpecificationDetails.aspx?specificationId=3450.

This stub offers a user interface that helps to test the functionalities implemented in the O-RAN-SC CAPIF implementation and the supported features are as follows:

- Onboard API Invoker
- Discover published service APIs and retrieve a collection of APIs according to certain filter criteria.
- Obtain Security method
- Obtain Authorization

### Onboard API Invoker

This service operation is used by an API invoker to on-board itself as a recognized user of CAPIF

<img src="docs/Onboarding new invoker.svg">

To onboard, the Invoker should send a request to the CAPIF core, including an API invoker Enrolment Details, API List and a Notification Destination URI for on-boarding notification.

```
{
    "apiInvokerInformation": "rApp as API invoker",
	 "apiList": [
		{}
	],
    "NotificationDestination": "http://invoker-app:8086/callback",
    "onboardingInformation": {
		"apiInvokerPublicKey": "{PUBLIC_KEY_INVOKER}",
		"apiInvokerCertificate": "apiInvokerCertificate"
  },
  "requestTestNotification": true
}
```

After receiving the request, the CAPIF core should check if the invoker can be onboarded. If the invoker is eligible for onboarding, the CAPIF core will create the API invoker Profile, which includes an API invoker Identifier, Authentication Information, Authorization Information, and CAPIF Identity Information. Keycloak is utilized in this implementation to manage identity information.

### Discover published service APIs and retrieve a collection of APIs according to certain filter criteria.

This service operation is used by an API invoker to discover service API available at the CAPIF core function.

<img src="docs/Discover Service API.svg">

If the invoker is authorized to discover the service APIs, the CAPIF core function search the API registry for APIs matching the query criteria and return the filtered search results in the response message.


### Obtain Security method

This service operation is used by an API invoker to negotiate and obtain information about service API security method for itself with CAPIF core function.

<img src="docs/Obtain Security Method.svg">

The invoker sends a request to the CAPIF core including Security Method Request and a Notification Destination URI for security related notifications. The Security Method Request contains the unique interface details of the service APIs and may contain a preferred security method for each unique service API interface.

Example of SecurityService:

```
{
  "notificationDestination": "http://invoker-app:8086/callback",
  "supportedFeatures": "fffffff",
  "securityInfo": [
    {
      "aefId": "AEF_id_rApp_as_AEF",
      "apiId": "api_id_example",
      "prefSecurityMethods": [
        "PSK"
      ],
      "selSecurityMethod": "PSK"
    }
  ],
  "requestTestNotification": true
}
```


### Obtain Authorization

This service operation is used by an API invoker to obtain authorization to access service APIs.

<img src="docs/Obtain Access Token.svg">

On success, "200 OK" will be returned. The payload body of the response contains the requested access token, the token type and the expiration time for the token. The access token is a JSON Web Token (JWT).

## Build application

To build the application, run the following command:

    go build

The application can also be built as a Docker image, by using the following command:

    docker build . -t capifprov

## Run

To run the provider from the command line, run the following commands from this folder.

    ./capifprov [-port <port (default 9090)>] [-capifCoreUrl <URL to Capif core (default http://localhost:8090)>] [-loglevel <log level (default Info)>]
