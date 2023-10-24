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

substitute_repo(){
    echo "substitute_repo"
    docker_repo=$1

    # Use our own Capificore and ServiceManager images
    sed -i "s/image: o-ran-sc.org\/nonrtric\/plt\/capifcore/image: $docker_repo\/capifcore:latest/" ../manifests/capifcore.yaml
    sed -i 's/imagePullPolicy: IfNotPresent/imagePullPolicy: Always/' ../manifests/capifcore.yaml
    sed -i "s/image: o-ran-sc.org\/nonrtric\/plt\/servicemanager/image: $docker_repo\/servicemanager:latest/" ../manifests/servicemanager.yaml
    sed -i 's/imagePullPolicy: IfNotPresent/imagePullPolicy: Always/' ../manifests/servicemanager.yaml
}

add_env(){
    echo "add_env"
    additional_env=$1
    # Add our own .env file
    if [ -f $additional_env ]; then
        echo "found additional env"
        kubectl create configmap env-configmap --from-file=.env --from-file=$additional_env -n servicemanager

        # Add additional env file to volume mounting
        env_filename=$(basename "$additional_env")
        echo "env_filename $env_filename"

        mount_path_wc=$(grep "mountPath: /app/$env_filename" ../manifests/servicemanager.yaml | wc -l)
        env_path_count=$((mount_path_wc))
        if [ $env_path_count -eq 0 ]; then
            echo "Adding mount path"
            sed -i -e '/subPath: .env/a\' -e "        - name: config-volume\n          mountPath: /app/$env_filename\n          subPath: $env_filename" ../manifests/servicemanager.yaml
        fi

        # Update SERVICE_MANAGER_ENV to point to additional env
        env_extension=$(basename "$additional_env" | awk -F. '{print $NF}')
        echo "env_extension $env_extension"

        sed -i "/- name: SERVICE_MANAGER_ENV/{n;s/              value: \"\"/              value: \"$env_extension\"/}" ../manifests/servicemanager.yaml
    fi
}

echo $(date -u) "deploy-to-k8s started"

# Check if the development switch is provided as a command-line argument
USE_OWN_REPO=false
ADD_ENV=false

while [[ "$#" -gt 0 ]]; do
    case "$1" in
        -r|--repo)
            USE_OWN_REPO=true
            shift  # consume the switch
            if [ -n "$1" ]; then
                DOCKER_REPO="$1"
                shift  # consume the value
            else
                echo "Error: Argument for $1 is missing." >&2
                exit 1
            fi
            ;;
        -e|--env)
            ADD_ENV=true
            shift  # consume the switch
            if [ -n "$1" ]; then
                ENV_PATH="$1"
                shift  # consume the value
            else
                echo "Error: Argument for $1 is missing." >&2
                exit 1
            fi
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
done

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

# Check if the development switch is enabled
if [ "$USE_OWN_REPO" = true ]; then
    substitute_repo $DOCKER_REPO
fi

kubectl create ns servicemanager

if [ "$ADD_ENV" = true ]; then
    add_env $ENV_PATH
else
    kubectl create configmap env-configmap --from-file=.env -n servicemanager
fi

# Create the Kubernetes resources
kubectl create -f ../manifests/capifcore.yaml
kubectl create -f ../manifests/servicemanager.yaml

kubectl rollout status deployment capifcore -n servicemanager --timeout=90s
kubectl rollout status deployment servicemanager -n servicemanager --timeout=90s
kubectl rollout status deployment kong-kong -n kong --timeout=90s

echo $(date -u) "deploy-to-k8s completed"
