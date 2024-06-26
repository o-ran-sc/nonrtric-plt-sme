#==================================================================================
#   Copyright (C) 2022-2023: Nordix Foundation. All rights reserved.
#   Copyright (C) 2024: OpenInfra Foundation Europe. All rights reserved.
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#
#==================================================================================

##
## Build
##
FROM nexus3.o-ran-sc.org:10001/golang:1.19.2-bullseye AS build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /capifcore cmd/main.go
##
## Deploy
##
FROM ubuntu
WORKDIR /
## Copy from "build" stage
COPY --from=build /capifcore .

RUN mkdir /configs
RUN mkdir /certs

COPY configs/keycloak.yaml /configs/keycloak.yaml
COPY certs/cert.pem /certs/cert.pem
COPY certs/key.pem /certs/key.pem

ENTRYPOINT ["/capifcore"]
