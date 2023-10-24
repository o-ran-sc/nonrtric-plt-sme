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

# O-RAN-SC Non-RealTime RIC Service Management and Exposure

This product is a Go implementation of a service that calls the CAPIF Core function. When publishing a service we create a Kong route and Kong service, https://konghq.com/. The InterfaceDescription that we return is updated to point to the Kong Data Plane. Therefore, the API interface that we return from Service Discovery has the Kong host and port, and not the original service's host and port. This allows the rApp's API call to be re-directed through Kong.

## O-RAN-SC Non-RealTime RIC CAPIF Core Implementation

This product is a Go implementation of the CAPIF Core function, which is based on the 3GPP "29.222 Common API Framework for 3GPP Northbound APIs (CAPIF)" interfaces, see https://portal.3gpp.org/desktopmodules/Specifications/SpecificationDetails.aspx?specificationId=3450.

See [CAPIF Core](../capifcore/README.md)

## Generation of API Code

The CAPIF APIs are generated from the OpenAPI specifications provided by 3GPP. The `generate.sh` script downloads the
specifications from 3GPP, fixes them and then generates the APIs.

```sh
./generate.sh
```

The specifications are downloaded from the following site; https://www.3gpp.org/ftp/Specs/archive/29_series. To see
the APIs in swagger format, see the following link; https://github.com/jdegre/5GC_APIs/tree/Rel-16#common-api-framework-capif.
**NOTE!** The documentation in this link is for release 16 of CAPIF, the downloaded specifications are for release 17.

To fix the specifications there are three tools.
- `commoncollector`, collects type definitions from peripheral specifications to keep down the number of dependencies to
  other specifications. The types to collect are listed in the `definitions.txt`file. Some fixes are hard coded.
- `enumfixer`, fixes enumeration definitions so they can be properly generated.
- `specificationfixer`, fixes flaws in the specifications so they can be properly generated. All fixes are hard coded.

## Set Up

First, we need to run `generate.sh` as described above to generate our API code from the 3GPP spec.

Before we can test or run R1-SME-Manager, we need to configure a .env file with the required parameters. Please see the template .env.example in the r1-sme-manager directory.

You can set the environmental variable R1SME_ENV to specify the .env file. For example, the following command specifies to use the config file
.env.development. If this flag is not set, first we try .env.development and then .env.

```sh
export R1SME_ENV=development
```

### Capifcore and Kong

We also need Kong and Capifcore to be running. Please see the examples in the `configs` folder.

## Build

After generating the API code, we can build the application with the following command.

```sh
go build
```

## Unit Tests

To run the unit tests for the application, first ensure that the .env file is configured. In the following example, we specify `.env.test`. For now, we need to disable parallelism in the unit tests with -p=1.

```sh
export R1SME_ENV=test
go test -p=1 -count=1 ./...
```

## Run Locally

To run as a local app, first ensure that the .env file is configured. In the following example, we specify `.env.development`.

```sh
export R1SME_ENV=development
./r1-sme-manager
```

R1-SME-Manager is then available on the port configured in .env.

## Building the Docker Image

The application can also be built as a Docker image, by using the following command. We build the image without a .env file. This is supplied by volume mounting at container run-time.

```sh
docker image build . -t r1-sme-manager
```

## Stand-alone Deployment on Kubernetes

For a stand-alone deployment, please see the `deploy` folder for configurations to deploy to R1-SME-Manager to Kubernetes. We need the following steps.
 - Deploy a PV for Kong's Postgres database (depends on your Kubernetes cluster)
 - Deploy a PVC for Kong's Postgres database
 - Deploy Kong with Postgres
 - Deploy Capifcore
 - Deploy R1-SME-Manager

We consolidate the above steps into the script `deploy-to-k8s.sh`. To delete the full deployment, you can use `delete-from-k8s.sh`. The deploy folder has the following structure.

- sme/
  - r1-sme-manager/
    - deploy/
      - src/
      - manifests/

We store the Kubernetes manifests files in the manifests in the subfolder. We store the sh scripts in the src folder.

In `deploy-to-k8s.sh`, we copy .env.example and use sed to replace the template values with values for testing/production. You will need to update this part of the script with your own values.

There is a placeholder in r1-sme-manager.yaml called REPLACE_ME. You will need to replace this value with the path to your r1-sme-manager image from your Docker repo. There is an example sed replacement in function `update_r1_sme_manager_manifest()` in `deploy-to-k8s.sh`.