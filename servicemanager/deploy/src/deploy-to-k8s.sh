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

create_env_from_template(){
    # Set up .env file for Kubernetes Config Map
    echo "create_env_from_template"
    cp -v ../../.env.example ./.env
    sed -i 's/KONG_DOMAIN=<string>/KONG_DOMAIN=kong/' .env
    sed -i 's/KONG_PROTOCOL=<http or https protocol scheme>/KONG_PROTOCOL=http/' .env
    sed -i 's/KONG_IPV4=<host string>/KONG_IPV4=10.101.1.101/' .env
    sed -i 's/KONG_DATA_PLANE_PORT=<port number>/KONG_DATA_PLANE_PORT=32080/' .env
    sed -i 's/KONG_CONTROL_PLANE_PORT=<port number>/KONG_CONTROL_PLANE_PORT=32081/' .env
    sed -i 's/CAPIF_PROTOCOL=<http or https protocol scheme>/CAPIF_PROTOCOL=http/' .env
    sed -i 's/CAPIF_IPV4=<host>/CAPIF_IPV4=10.101.1.101/' .env
    sed -i 's/CAPIF_PORT=<port number>/CAPIF_PORT=31570/' .env
    sed -i 's/LOG_LEVEL=<Trace, Debug, Info, Warning, Error, Fatal or Panic>/LOG_LEVEL=Info/' .env
    sed -i 's/SERVICE_MANAGER_PORT=<port number>/SERVICE_MANAGER_PORT=8095/' .env
    sed -i 's/TEST_SERVICE_IPV4=<host string>/TEST_SERVICE_IPV4=10.101.1.101/' .env
    sed -i 's/TEST_SERVICE_PORT=<port number>/TEST_SERVICE_PORT=30951/' .env
}

substitute_manifest(){
    echo "substitute_manifest"

    # sed -i 's/image: o-ran-sc.org\/nonrtric\/plt\/capifcore/image: mydockeruser\/capifcore:latest/' ../manifests/capifcore.yaml
    # sed -i 's/imagePullPolicy: IfNotPresent/imagePullPolicy: Always/' ../manifests/capifcore.yaml
    # sed -i 's/image: o-ran-sc.org\/nonrtric\/plt\/servicemanager/image: mydockeruser\/servicemanager:latest/' ../manifests/servicemanager.yaml
    # sed -i 's/imagePullPolicy: IfNotPresent/imagePullPolicy: Always/' ../manifests/servicemanager.yaml
}

echo $(date -u) "deploy-to-k8s started"

kubectl create ns kong

# Set up storage for Postgres, used by Kong
# Minikube uses dynamic provisioning
CURRENT_CONTEXT=$(kubectl config current-context)
if [ "$CURRENT_CONTEXT" != "minikube" ]; then
    kubectl create -f ../manifests/kong-postgres-pv.yaml
fi
kubectl create -f ../manifests/kong-postgres-pvc.yaml

# Deploy Kong
helm repo add kong https://charts.konghq.com
helm repo update
helm install kong kong/kong -n kong -f ../manifests/values.yaml

create_env_from_template
substitute_manifest

# Create the Kubernetes resources
kubectl create -f ../manifests/capifcore.yaml
kubectl create configmap env-configmap --from-file=.env -n servicemanager
kubectl create -f ../manifests/servicemanager.yaml

kubectl rollout status deployment capifcore -n servicemanager --timeout=90s
kubectl rollout status deployment servicemanager -n servicemanager --timeout=90s
kubectl rollout status deployment kong-kong -n kong --timeout=90s

echo $(date -u) "deploy-to-k8s completed"
