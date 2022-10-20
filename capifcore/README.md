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

# O-RAN-SC Non-RealTime CAPIF implementation

This product is a Go implementation of the CAPIF Core function, based on the 3GPP CAPIF interfaces.

## Generation of API code

The CAPIF APIs are generated from the OpenAPI specification provided by 3GPP. The `generate.sh` script downloads the
specifications from 3GPP, fixes them and then generates the APIs. It also generates the mocks needed for unit testing.

To fix the specifications there are three tools:
- `commoncollector`, collects type definitions from peripheral specifications to keep down the number of dependencies to
  other specifications. The types to collect are listed in the `definitions.txt`file.
- `enumfixer`, fixes enumeration definitions so they can be properly generated.
- `specificationfixer`, fixes flaws in the specifications so they can be properly generated. All fixes are hard coded.

## Build and test

To build the application, run the following command:

    go build

To generate mocks, run the following command:

    go generate ./...

To run the unit tests for the application, run the following command:

    go test ./...

## Run

To run the Core Function run the following commands from this folder.

    ./capifcore [-port <port (default 8080)>]