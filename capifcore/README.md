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

## Run

To run the Core Function run the following commands from the top of the repo:

    go build
    ./capifcore [-port <port (default 8080)>]