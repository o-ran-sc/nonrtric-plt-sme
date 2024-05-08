#!/bin/bash
# -
#   ========================LICENSE_START=================================
#   O-RAN-SC
#   %%
#   Copyright (C) 2024: OpenInfra Foundation Europe
#   %%
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#   ========================LICENSE_END===================================
#

make_internal_dirs () {
    echo "Make the internal API directory structure"
    mkdir -p internal/discoverserviceapi
    mkdir -p internal/invokermanagementapi
    mkdir -p internal/providermanagementapi
    mkdir -p internal/publishserviceapi
}

set_up_dir_paths () {
    echo "Set up dir paths"
    cwd=$PWD
    sme_dir=$(dirname "$cwd")
    capifcore_dir="$sme_dir/capifcore"
}

curl_api_specs () {
    echo "Curl down the API specs"
    mkdir -p specs
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.222/29222-h60.zip -o specs/apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.122/29122-h70.zip -o specs/common29122apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.508/29508-h80.zip -o specs/common29508apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.510/29510-h70.zip -o specs/common29510apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.512/29512-h80.zip -o specs/common29512apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.514/29514-h60.zip -o specs/common29514apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.517/29517-h70.zip -o specs/common29517apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.518/29518-h70.zip -o specs/common29518apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.522/29522-h70.zip -o specs/common29522apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.523/29523-h80.zip -o specs/common29523apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.554/29554-h40.zip -o specs/common29554apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.571/29571-h70.zip -o specs/common29571apidef.zip
    curl https://www.3gpp.org/ftp/Specs/archive/29_series/29.572/29572-h60.zip -o specs/common29572apidef.zip
}

spec_extraction () {
    cd specs/
    echo "Specifications extraction"
    unzip -o apidef.zip
    unzip -o common29122apidef.zip
    unzip -o common29508apidef.zip
    unzip -o common29510apidef.zip
    unzip -o common29512apidef.zip
    unzip -o common29514apidef.zip
    unzip -o common29517apidef.zip
    unzip -o common29518apidef.zip
    unzip -o common29522apidef.zip
    unzip -o common29523apidef.zip
    unzip -o common29554apidef.zip
    unzip -o common29571apidef.zip
    unzip -o common29572apidef.zip
}

fix_with_sed () {
    echo "Fixing with sed"

    # Remove types that are not used by CAPIF that have dependencies to other specifications.
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\CivicAddress/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\ExternalMbsServiceArea/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\GeographicArea/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\GeoServiceArea/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsMediaComp/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsMediaCompRm/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsMediaInfo/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsServiceInfo/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsSession/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml
    sed -i -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\SpatialValidityCond/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml

    # Remove attributes that cannot be generated easily.
    sed -i '/accessTokenError.*/,+3d' TS29571_CommonData.yaml
    sed -i '/accessTokenRequest.*/,+3d' TS29571_CommonData.yaml
    sed -i '/oneOf.*/,+2d' TS29222_CAPIF_Publish_Service_API.yaml
    sed -i '/oneOf.*/,+2d' TS29222_CAPIF_Security_API.yaml

    # Replace references to external specs that are collected to the common spec by the commoncollector
    # <replacements_start>
    sed -i 's/TS29572_Nlmf_Location/CommonData/g' TS29122_CommonData.yaml
    sed -i 's/TS29554_Npcf_BDTPolicyControl/CommonData/g' TS29122_CommonData.yaml
    sed -i 's/TS29514_Npcf_PolicyAuthorization/CommonData/g' TS29122_CommonData.yaml
    sed -i 's/TS29514_Npcf_PolicyAuthorization/CommonData/g' TS29571_CommonData.yaml
    sed -i 's/TS29572_Nlmf_Location/CommonData/g' TS29571_CommonData.yaml
    sed -i 's/TS29572_Nlmf_Location/CommonData/g' TS29222_CAPIF_Publish_Service_API.yaml
    sed -i 's/TS29520_Nnwdaf_EventsSubscription/CommonData/g' TS29222_CAPIF_Routing_Info_API.yaml
    sed -i 's/TS29510_Nnrf_NFManagement/CommonData/g' TS29222_CAPIF_Routing_Info_API.yaml
    sed -i 's/TS29523_Npcf_EventExposure/CommonData/g' TS29222_CAPIF_Events_API.yaml
    # <new_replacement>

    # This spec has references to itself that need to be removed
    sed -i 's/TS29571_CommonData.yaml//g' TS29571_CommonData.yaml
}

gentools () {
    echo "Fixing enums"

    cd "$capifcore_dir/internal/gentools/enumfixer"
    go build .
    ./enumfixer -apidir="$sme_dir/servicemanager/specs"

    echo "Gathering common references"
    cd "$capifcore_dir/internal/gentools/commoncollector"
    go build .
    ./commoncollector -apidir="$sme_dir/servicemanager/specs"

    echo "Fixing misc in specifications"
    cd "$capifcore_dir/internal/gentools/specificationfixer"
    go build .
    ./specificationfixer -apidir="$sme_dir/servicemanager/specs"
}

generate_apis_from_spec () {
    go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.10.1
    PATH=$PATH:~/go/bin

    cd $cwd

    echo "Generating TS29122_CommonData"
    mkdir -p internal/common29122
    oapi-codegen --config "gogeneratorspecs/common29122/generator_settings.yaml" specs/TS29122_CommonData.yaml

    echo "Generating aggregated CommonData"
    mkdir -p internal/common
    oapi-codegen --config "gogeneratorspecs/common/generator_settings.yaml" specs/CommonData.yaml

    echo "Generating TS29571_CommonData"
    mkdir -p internal/common29571
    oapi-codegen --config "gogeneratorspecs/common29571/generator_settings.yaml" specs/TS29571_CommonData.yaml

    echo "Generating TS29222_CAPIF_Publish_Service_API"
    mkdir -p internal/publishserviceapi
    oapi-codegen --config "gogeneratorspecs/publishserviceapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Publish_Service_API.yaml
    oapi-codegen --config "gogeneratorspecs/publishserviceapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Publish_Service_API.yaml
    oapi-codegen --config "gogeneratorspecs/publishserviceapi/generator_settings_client.yaml" specs/TS29222_CAPIF_Publish_Service_API.yaml

    echo "Generating TS29222_CAPIF_API_Invoker_Management_API"
    mkdir -p internal/invokermanagementapi
    oapi-codegen --config "gogeneratorspecs/invokermanagementapi/generator_settings_types.yaml" specs/TS29222_CAPIF_API_Invoker_Management_API.yaml
    oapi-codegen --config "gogeneratorspecs/invokermanagementapi/generator_settings_server.yaml" specs/TS29222_CAPIF_API_Invoker_Management_API.yaml
    oapi-codegen --config "gogeneratorspecs/invokermanagementapi/generator_settings_client.yaml" specs/TS29222_CAPIF_API_Invoker_Management_API.yaml

    echo "Generating TS29222_CAPIF_API_Provider_Management_API"
    mkdir -p internal/providermanagementapi
    oapi-codegen --config "gogeneratorspecs/providermanagementapi/generator_settings_types.yaml" specs/TS29222_CAPIF_API_Provider_Management_API.yaml
    oapi-codegen --config "gogeneratorspecs/providermanagementapi/generator_settings_server.yaml" specs/TS29222_CAPIF_API_Provider_Management_API.yaml
    oapi-codegen --config "gogeneratorspecs/providermanagementapi/generator_settings_client.yaml" specs/TS29222_CAPIF_API_Provider_Management_API.yaml

    echo "Generating TS29222_CAPIF_Discover_Service_API"
    mkdir -p internal/discoverserviceapi
    oapi-codegen --config "gogeneratorspecs/discoverserviceapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Discover_Service_API.yaml
    oapi-codegen --config "gogeneratorspecs/discoverserviceapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Discover_Service_API.yaml
    oapi-codegen --config "gogeneratorspecs/discoverserviceapi/generator_settings_client.yaml" specs/TS29222_CAPIF_Discover_Service_API.yaml
}

generate_html2_from_spec() {
    cd $cwd

    echo "Generating Provider_Management HTML"
    docker run --rm \
        -u $(id -u ${USER}):$(id -g ${USER}) \
        -v ${PWD}:/local openapitools/openapi-generator-cli generate \
        -i /local/specs/TS29222_CAPIF_API_Provider_Management_API.yaml \
        -g html2 \
        -o /local/openapi/html2/Provider_Management
    mv -v openapi/html2/Provider_Management/index.html openapi/html2/Provider_Management/Provider_Management.html

    echo "Generating Publish_Service HTML"
    docker run --rm \
        -u $(id -u ${USER}):$(id -g ${USER}) \
        -v ${PWD}:/local openapitools/openapi-generator-cli generate \
        -i /local/specs/TS29222_CAPIF_Publish_Service_API.yaml \
        -g html2 \
        -o /local/openapi/html2/Publish_Service
    mv -v openapi/html2/Publish_Service/index.html openapi/html2/Publish_Service/Publish_Service.html

    echo "Generating Invoker_Management HTML"
    docker run --rm \
        -u $(id -u ${USER}):$(id -g ${USER}) \
        -v ${PWD}:/local openapitools/openapi-generator-cli generate \
        -i /local/specs/TS29222_CAPIF_API_Invoker_Management_API.yaml \
        -g html2 \
        -o /local/openapi/html2/Invoker_Management
    mv -v openapi/html2/Invoker_Management/index.html openapi/html2/Invoker_Management/Invoker_Management.html

    echo "Generating Discover_Service HTML"
    docker run --rm \
        -u $(id -u ${USER}):$(id -g ${USER}) \
        -v ${PWD}:/local openapitools/openapi-generator-cli generate \
        -i /local/specs/TS29222_CAPIF_Discover_Service_API.yaml \
        -g html2 \
        -o /local/openapi/html2/Discover_Service
    mv -v openapi/html2/Discover_Service/index.html openapi/html2/Discover_Service/Discover_Service.html

    echo "Generating Events HTML"
    docker run --rm \
        -u $(id -u ${USER}):$(id -g ${USER}) \
        -v ${PWD}:/local openapitools/openapi-generator-cli generate \
        -i /local/specs/TS29222_CAPIF_Events_API.yaml \
        -g html2 \
        -o /local/openapi/html2/Events
    mv -v openapi/html2/Events/index.html openapi/html2/Events/Events.html

    echo "Generating Security HTML"
    docker run --rm \
        -u $(id -u ${USER}):$(id -g ${USER}) \
        -v ${PWD}:/local openapitools/openapi-generator-cli generate \
        -i /local/specs/TS29222_CAPIF_Security_API.yaml \
        -g html2 \
        -o /local/openapi/html2/Security
    mv -v openapi/html2/Security/index.html openapi/html2/Security/Security.html

    mkdir -vp ../docs/openapi
    find openapi/html2 -type f -name "*.html" -exec cp -v {} ../docs/openapi \;
}

run_tests () {
    # Make sure that SERVICE_MANAGER_ENV is configured with the required .env file, e.g.
    # export SERVICE_MANAGER_ENV=development
    cd "$cwd"
    go test ./...
}

# Main code block
echo $(date -u) "generate started"

# Check if the run-tests switch is provided as a command-line argument
RUN_TESTS=false
while [[ "$#" -gt 0 ]]; do
    case "$1" in
        -t|--run-tests)
            RUN_TESTS=true
            shift  # consume the switch
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
    shift
done

make_internal_dirs
set_up_dir_paths
curl_api_specs
spec_extraction
fix_with_sed
gentools
generate_apis_from_spec
generate_html2_from_spec

cd $cwd

# Check if the run-tests switch is enabled
if [ "$RUN_TESTS" = true ]; then
    echo "Running tests!"
    run_tests
else
    echo "Not running tests."
fi

echo $(date -u) "generate completed"