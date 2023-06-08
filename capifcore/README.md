<!--
 -
   ========================LICENSE_START=================================
   O-RAN-SC
   %%
   Copyright (C) 2022: Nordix Foundation
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

# O-RAN-SC Non-RealTime RIC CAPIF Core implementation

This product is a Go implementation of the CAPIF Core function, based on the 3GPP "29.222 Common API Framework for 3GPP Northbound APIs (CAPIF)" interfaces, see https://portal.3gpp.org/desktopmodules/Specifications/SpecificationDetails.aspx?specificationId=3450.

The, almost, complete data model for CAPIF is shown in the diagram below.

<img src="docs/diagrams/Information model for CAPIF.svg">

The data used within CAPIF Core for registering rApps that both provides and consumes services is shown in the diagram below.

<img src="docs/diagrams/Information in rApp registration.svg">

Some examples of interactions between components using the CAPIF interface are shown in the sequence diagram below.

***NOTE!*** It has not been decided that CAPIF Core will actually do the Helm chart installation. This is just provided in the prototype as an example of what CAPIF Core could do.

<img src="docs/diagrams/Register Provider.svg">

If Helm is used, before publishing a service, the chart that belongs to the service must be registered in ChartMuseum. When publishing the service the following information should be provided in the `ServiceAPIDescription::description` attribute; "namespace", "repoName", "chartName", "releaseName". An example of the information: "Description of rApp helloWorld,namespace,repoName,chartName,releaseName".

## Generation of API code

The CAPIF APIs are generated from the OpenAPI specifications provided by 3GPP. The `generate.sh` script downloads the
specifications from 3GPP, fixes them and then generates the APIs. It also generates the mocks needed for unit testing.
The specifications are downloaded from the following site; https://www.3gpp.org/ftp/Specs/archive/29_series. To see
the APIs in swagger format, see the following link; https://github.com/jdegre/5GC_APIs/tree/Rel-16#common-api-framework-capif.
**NOTE!** The documentation in this link is for release 16 of CAPIF, the downloaded specifications are for release 17.

To fix the specifications there are three tools:
- `commoncollector`, collects type definitions from peripheral specifications to keep down the number of dependencies to
  other specifications. The types to collect are listed in the `definitions.txt`file. Some fixes are hard coded.
- `enumfixer`, fixes enumeration definitions so they can be properly generated.
- `specificationfixer`, fixes flaws in the specifications so they can be properly generated. All fixes are hard coded.

### Steps to add a new dependency to the commoncollector

When a dependency to a new specification is introduced in any of the CAPIF specifications, see example below, the following steps should be performed:

For the CAPIF specification "TS29222_CAPIF_Discover_Service_API" a new dependency like the following has been introduced.

    websockNotifConfig:
        $ref: ✅TS29122_CommonData.yaml#/components/schemas/WebsockNotifConfig✅'

1. Copy the part between the checkboxes of the reference and add it to the `definitions.txt` file. This step is not needed if the type is already defined in the file.
2. Look in the `generate.sh` script, between the "<replacements_start>" and "<new_replacement>" tags, to see if "TS29122_CommonData"
   has already been replaced in "TS29222_CAPIF_Discover_Service_API".
3. If it has not been replaced, add a replacement above the "<new_replacement>" tag by copying and adapting the two rows above the tag.

### Security in CAPIF

Security requirements that are applicable to all CAPIF entities includes provide authorization mechanism for service APIs from the 3rd party API providers and support a common security mechanism for all API implementations to provide confidentiality and integrity protection.

In the current implementation Keycloak is being used as identity and access management (IAM) solution that provides authentication, authorization, and user management for applications and services. Keycloak provides robust authentication mechanisms, including username/password, two-factor authentication, and client certificate authentication that complies with CAPIF security requirements.

A docker-compose file is included to start up keycloak.

## Build and test

To generate mocks manually, run the following command:

    go generate ./...

**NOTE!** The `helmmanagement` package contains two mocks from the `helm.sh/helm/v3` product. If they need to be
regenerated, their interfaces must be copied into the `helm.go` file and a generation annotation added before running
the generation script.

To build the application, run the following command:

    go build

To run the unit tests for the application, run the following command:

    go test ./...

The application can also be built as a Docker image, by using the following command:

    docker build . -t capifcore

## Run

To run the Core Function from the command line, run the following commands from this folder. For the parameter `chartMuseumUrl`, if it is not provided CAPIF Core will not do any Helm integration, i.e. try to start any Halm chart when publishing a service.

    ./capifcore [-port <port (default 8090)>] [-secPort <Secure port (default 4433)>] [-chartMuseumUrl <URL to ChartMuseum>] [-repoName <Helm repo name (default capifcore)>] [-loglevel <log level (default Info)>] [-certPath <Path to certificate>] [-keyPath <Path to private key>]

Use docker compose file to start CAPIF core together with Keycloak:

    docker-compose up

**NOTE!** There is a configuration file in configs/keycloak.yaml with information related to keycloak host, when running locally the host value must be set to localhost (Eg. host: "localhost") and when using docker-compose set value of host to keycloak (Eg. host:"keycloak")

Before using CAPIF API invoker management, an invoker realm must be created in keycloak. Make sure it is created before running CAPIF core. After creating the realm in keycloak, set the name in the keycloak.yaml configuration file.

To run CAPIF Core as a K8s pod together with ChartMuseum, start and stop scripts are provided. The pod configurations are provided in the `configs` folder. CAPIF Core is then available on port `31570`.
