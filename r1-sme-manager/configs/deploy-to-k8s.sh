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

echo $(date -u) "deploy-to-k8s started"

kubectl create ns kong

# Set up storage for Postgres, used by Kong
# Minikube uses dynamic provisioning
CURRENT_CONTEXT=$(kubectl config current-context)
if [ "$CURRENT_CONTEXT" != "minikube" ]; then
    kubectl create -f kong-postgres-pv.yaml
fi
kubectl create -f kong-postgres-pvc.yaml

# Deploy Kong
helm repo add kong https://charts.konghq.com
helm repo update
helm install kong kong/kong -n kong -f values.yaml

# Deploy R1-SME-Manager with Capifcore
kubectl create -f capifcore.yaml
kubectl create configmap env-configmap --from-file=../.env -n r1-sme-manager
kubectl create -f r1-sme-manager.yaml

kubectl rollout status deployment capifcore -n r1-sme-manager --timeout=90s
kubectl rollout status deployment r1-sme-manager -n r1-sme-manager --timeout=90s
kubectl rollout status deployment kong-kong -n kong --timeout=90s

echo $(date -u) "deploy-to-k8s completed"