#==================================================================================
#   Copyright (C) 2024: OpenInfra Foundation Europe
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

WORKDIR /app/capifcore

COPY capifcore/go.mod .
COPY capifcore/go.sum .
RUN go mod download

COPY ./capifcore/. .

WORKDIR /app/servicemanager

COPY servicemanager/go.mod .
COPY servicemanager/go.sum .
RUN go mod download

COPY ./servicemanager/. .

RUN go build

WORKDIR /app/servicemanager/internal
RUN go build kongclearup.go

##
## Deploy
##

FROM ubuntu:22.04

# We use the /app/servicemanager directory as this is where our servicemanager build expects to find its .env file.
WORKDIR /app/servicemanager

## Copy from "build" stage
COPY --from=build /app/servicemanager/servicemanager .
COPY --from=build /app/servicemanager/internal/kongclearup .

ENTRYPOINT ["/app/servicemanager/servicemanager"]
