#!/bin/bash
##############################################################################
#
#   Copyright (C) 2021: Nordix Foundation
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

echo "--> build-sme-ubuntu.sh"
curdir=`pwd`
# go installs tools like go-acc to $HOME/go/bin
# ubuntu minion path lacks go
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
go version

# install the go coverage tool helper
go get -v github.com/ory/go-acc

export GO111MODULE=on
go get github.com/stretchr/testify/mock@v1.7.1

go-acc ./... --ignore mocks,api,common

sed -i -e 's/oransc\.org\/nonrtric\/sme/sme/' coverage.txt

echo "--> build-sme-ubuntu.sh ends"
