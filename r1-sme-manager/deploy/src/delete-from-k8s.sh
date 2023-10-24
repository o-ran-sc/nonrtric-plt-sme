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

echo $(date -u) "delete-from-k8s started"

# Delete R1-SME-Manager with Capifcore
kubectl delete -f ../manifests/r1-sme-manager.yaml
kubectl delete configmap env-configmap -n r1-sme-manager
kubectl delete -f ../manifests/capifcore.yaml

# Delete Kong
go run ../../internal/kongclearup.go
helm uninstall kong -n kong
helm repo remove kong
kubectl wait deploy/kong-kong --for=delete --timeout=-300s -n kong 

# Delete storage for the Postgres used by Kong
kubectl delete -f ../manifests/kong-postgres-pvc.yaml
kubectl delete ns kong

echo $(date -u) "delete-from-k8s completed"