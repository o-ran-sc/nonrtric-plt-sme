#!/bin/bash
##############################################################################
#
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
##############################################################################
set -eux

echo "--> build-servicemanager-ubuntu.sh"

# go installs tools like go-acc to $HOME/go/bin
# ubuntu minion path lacks go
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
export GO111MODULE=on
go version
cd servicemanager

# Get the go coverage tool helper
go install github.com/ory/go-acc@v0.2.8
go get github.com/stretchr/testify/mock@v1.8.1
go-acc ./... --ignore mockkong,common,discoverserviceapi,invokermanagementapi,publishserviceapi,providermanagementapi

sed -i 's/oransc\.org\/nonrtric\/servicemanager/servicemanager/' coverage.txt

echo "--> build-servicemanager-ubuntu.sh ends"
