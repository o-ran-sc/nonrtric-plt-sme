<!--
-
========================LICENSE_START=================================
O-RAN-SC
%%
Copyright (C) 2024 OpenInfra Foundation Europe. All rights reserved.
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

# O-RAN-SC Non-RealTime RIC Service Manager

Service Manager is a Go implementation of a service that calls the CAPIF Core function. When publishing a service we create a Kong route and Kong service, https://konghq.com/. The InterfaceDescription that we return is updated to point to the Kong Data Plane. Therefore, the API interface that we return from Service Discovery has the Kong host and port, and not the original service's host and port. This allows the rApp's API call to be re-directed through Kong.

## O-RAN-SC Non-RealTime RIC CAPIF Core Implementation

Service Manager is a Go implementation of the CAPIF Core function, which is based on the 3GPP "29.222 Common API Framework for 3GPP Northbound APIs (CAPIF)" interfaces, see https://portal.3gpp.org/desktopmodules/Specifications/SpecificationDetails.aspx?specificationId=3450.

See [CAPIF Core](../capifcore/README.md)

## Generation of API Code

The CAPIF APIs are generated from the OpenAPI specifications provided by 3GPP. The `generate.sh` script downloads the
specifications from 3GPP, fixes them and then generates the APIs. While these files are checked into the repo, they can be re-generated using `generate.sh`.

```sh
./generate.sh
```

The specifications are downloaded from the following site; https://www.3gpp.org/ftp/Specs/archive/29_series. To see
the APIs in swagger format, see the following link; https://github.com/jdegre/5GC_APIs/tree/Rel-17#common-api-framework-capif.

To fix the specifications there are three tools.
- `commoncollector`, collects type definitions from peripheral specifications to keep down the number of dependencies to
  other specifications. The types to collect are listed in the `definitions.txt` file. Some fixes are hard-coded.
- `enumfixer`, fixes enumeration definitions so they can be properly generated.
- `specificationfixer`, fixes flaws in the specifications so they can be properly generated. All fixes are hard-coded.

## Set Up

First, we need to run `generate.sh` as described above to generate our API code from the 3GPP spec.

Before we can test or run Service Manager, we need to configure a .env file with the required parameters. Please see the template .env.example in the servicemanager directory.

You can set the environmental variable SERVICE_MANAGER_ENV to specify the .env file. For example, the following command specifies to use the config file
.env.development. If this flag is not set, first we try .env.development and then .env.

```sh
export SERVICE_MANAGER_ENV=development
```

### CAPIFcore and Kong

We also need Kong and CAPIFcore to be running. Please see the examples in the `deploy` folder. You can also use https://gerrit.o-ran-sc.org/r/it/dep for deployment. Please see the notes at https://wiki.o-ran-sc.org/display/RICNR/Release+J%3A+Service+Manager.

## Build

After generating the API code, we can build the application with the following command.

```sh
go build
```

## Unit Tests

To run the unit tests for the application, first ensure that the .env file is configured. In the following example, we specify `.env.test`.

```sh
export SERVICE_MANAGER_ENV=test
go test ./...
```

## Run Locally

To run as a local app, first ensure that the .env file is configured. In the following example, we specify `.env.development`.

```sh
export SERVICE_MANAGER_ENV=development
./servicemanager
```

Service Manager is then available on the port configured in .env.

## Building the Docker Image

The application can also be built as a Docker image, by using the following command. We build the image without a .env file. This is supplied by volume mounting at container run-time. Because we need to include CAPIFcore in the Docker build context, we build from the git repo's root directory, sme.

```sh
docker build -t servicemanager -f servicemanager/Dockerfile .
```

## Kongclearup

Please note that a special executable has been provided for deleting Kong routes and services that have been created by Service Manager in Kong. This executable is called `kongclearup` and is found in the working directory of the Service Manger Docker image, at `/app/servicemanager`. When we create a Kong route or service, we add Kong tags with information as follows.
  * apfId
  * aefId
  * apiId
  * apiVersion
  * resourceName

When we delete Kong routes and services using `kongclearup`, we check for the existance of these tags, specifically, apfId, apiId and aefId. Only if these tags exist and have values do we proceed to delete the Kong service or route. The executable `kongclearup` uses the volume-mounted .env file to load the configuration giving the location of Kong. Please refer to `sme/servicemanager/internal/kongclearup.go`.

## Stand-alone Deployment on Kubernetes

For a stand-alone deployment, please see the `deploy` folder for configurations to deploy to Service Manager to Kubernetes. We need the following steps.
 - Deploy a PV for Kong's Postgres database (depends on your Kubernetes cluster, not needed for Minikube)
 - Deploy a PVC for Kong's Postgres database
 - Deploy Kong with Postgres
 - Deploy CAPIFcore
 - Deploy Service Manager

We consolidate the above steps into the script `deploy-to-k8s.sh`. To delete the full deployment, you can use `delete-from-k8s.sh`. The deploy folder has the following structure.

- sme/
  - servicemanager/
    - deploy/
      - src/
      - manifests/

We store the Kubernetes manifests files in the manifests in the subfolder. We store the shell scripts in the src folder.

In `deploy-to-k8s.sh`, we copy .env.example and use `sed` to replace the template values with values for running the Service Manager container. You will need to update this part of the script with your own values. There is an example sed replacement in function `substitute_manifest()` in `deploy-to-k8s.sh`. Here, you can substitute your own Docker images for CAPIFcore and Service Manager for local development.

In addition there are 2 switches that are added for developer convenience.
 * --repo # allows you to specify your own docker repo, e.g. your Docker Hub id
 * --env  # allows you to specify an additional env file, and sets SERVICE_MANAGER_ENV in the Docker environment to point to this file.

The additional env file needs to exist in the sme/servicemanager folder so that kongclearup can access is. It is specified by its filename. The relative path ../.. is added in the `deploy-to-k8s.sh` script. For example, to use

`./deploy-to-k8s.sh --env .env.development`

 ../../.env.development needs to exist.

## Postman

A Postman collection has been included in this repo at sme/postman/ServiceManager.postman_collection.json.