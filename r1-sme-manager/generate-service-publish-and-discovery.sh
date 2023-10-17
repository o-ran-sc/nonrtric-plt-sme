# -
#   ========================LICENSE_START=================================
#   O-RAN-SC
#   %%
#   Copyright (C) 2023: Nordix Foundation
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

#!/bin/bash

make_internal_dirs () {
    echo "Make the internal directory structure"
    mkdir -p internal/config
    mkdir -p internal/discoverservice
    mkdir -p internal/discoverserviceapi
    mkdir -p internal/eventsapi
    mkdir -p internal/eventservice
    mkdir -p internal/helmmanagement
    mkdir -p internal/invokermanagement
    mkdir -p internal/invokermanagementapi
    mkdir -p internal/keycloak
    mkdir -p internal/loggingapi
    mkdir -p internal/providermanagement
    mkdir -p internal/providermanagementapi
    mkdir -p internal/publishservice
    mkdir -p internal/publishserviceapi
    mkdir -p internal/restclient
}

tear_down () {
    echo "Tear down the internal directory"
    rm -rf internal/
}

set_up_dir_paths () {
    echo "Set up dir paths"
    cwd=$PWD
    sme_dir=$(dirname "$cwd")
    capifcore_dir="$sme_dir/capifcore"
}

copy_test_wrappers () {
    echo "Copy the test wrappers"
    cp -v $capifcore_dir/internal/config/*.go internal/config
    cp -v $capifcore_dir/internal/discoverservice/*.go internal/discoverservice
    cp -v $capifcore_dir/internal/eventsapi/type*.go internal/eventsapi
    cp -v $capifcore_dir/internal/eventservice/*.go internal/eventservice
    cp -v $capifcore_dir/internal/helmmanagement/*.go internal/helmmanagement
    cp -v $capifcore_dir/internal/invokermanagement/*.go internal/invokermanagement
    cp -v $capifcore_dir/internal/invokermanagementapi/type*.go internal/invokermanagementapi
    cp -v $capifcore_dir/internal/keycloak/*.go internal/keycloak
    cp -v $capifcore_dir/internal/providermanagement/*.go internal/providermanagement
    cp -v $capifcore_dir/internal/providermanagementapi/*.go internal/providermanagementapi
    cp -v $capifcore_dir/internal/publishservice/*.go internal/publishservice
    cp -v $capifcore_dir/internal/publishserviceapi/type*.go internal/publishserviceapi
    cp -v $capifcore_dir/internal/restclient/*.go internal/restclient
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

jar_extraction () {
    cd specs/
    echo "Jar extraction"
    jar xvf apidef.zip
    jar xvf common29122apidef.zip
    jar xvf common29508apidef.zip
    jar xvf common29510apidef.zip
    jar xvf common29512apidef.zip
    jar xvf common29514apidef.zip
    jar xvf common29517apidef.zip
    jar xvf common29518apidef.zip
    jar xvf common29522apidef.zip
    jar xvf common29523apidef.zip
    jar xvf common29554apidef.zip
    jar xvf common29571apidef.zip
    jar xvf common29572apidef.zip
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

    # Remove attributes that can not be generated easily.
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
    ./enumfixer -apidir="$sme_dir/r1-sme-manager/specs"

    echo "Gathering common references"
    cd "$capifcore_dir/internal/gentools/commoncollector"
    go build .
    ./commoncollector -apidir="$sme_dir/r1-sme-manager/specs"

    echo "Fixing misc in specifications"
    cd "$capifcore_dir/internal/gentools/specificationfixer"
    go build .
    ./specificationfixer -apidir="$sme_dir/r1-sme-manager/specs"
}

code_generation () {
    go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.10.1
    PATH=$PATH:~/go/bin

    cd $cwd

    echo "Generating TS29122_CommonData"
    mkdir -p internal/common29122
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/common29122/generator_settings.yaml" specs/TS29122_CommonData.yaml

    echo "Generating aggregated CommonData"
    mkdir -p internal/common
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/common/generator_settings.yaml" specs/CommonData.yaml

    echo "Generating TS29571_CommonData"
    mkdir -p internal/common29571
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/common29571/generator_settings.yaml" specs/TS29571_CommonData.yaml

    echo "Generating TS29222_CAPIF_Publish_Service_API"
    mkdir -p internal/publishserviceapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/publishserviceapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Publish_Service_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/publishserviceapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Publish_Service_API.yaml

    echo "Generating TS29222_CAPIF_API_Invoker_Management_API"
    mkdir -p internal/invokermanagementapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/invokermanagementapi/generator_settings_types.yaml" specs/TS29222_CAPIF_API_Invoker_Management_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/invokermanagementapi/generator_settings_server.yaml" specs/TS29222_CAPIF_API_Invoker_Management_API.yaml

    echo "Generating TS29222_CAPIF_API_Provider_Management_API"
    mkdir -p internal/providermanagementapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/providermanagementapi/generator_settings_types.yaml" specs/TS29222_CAPIF_API_Provider_Management_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/providermanagementapi/generator_settings_server.yaml" specs/TS29222_CAPIF_API_Provider_Management_API.yaml

    echo "Generating TS29222_CAPIF_Discover_Service_API"
    mkdir -p internal/discoverserviceapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/discoverserviceapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Discover_Service_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/discoverserviceapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Discover_Service_API.yaml

    echo "Generating TS29222_CAPIF_Logging_API_Invocation_API"
    mkdir -p internal/loggingapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/loggingapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Logging_API_Invocation_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/loggingapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Logging_API_Invocation_API.yaml

    echo "Generating TS29222_CAPIF_Routing_Info_API"
    mkdir -p internal/routinginfoapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/routinginfoapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Routing_Info_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/routinginfoapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Routing_Info_API.yaml

    echo "Generating TS29222_CAPIF_Access_Control_Policy_API"
    mkdir -p internal/accesscontrolpolicyapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/accesscontrolpolicyapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Access_Control_Policy_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/accesscontrolpolicyapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Access_Control_Policy_API.yaml

    echo "Generating TS29222_CAPIF_Events_API"
    mkdir -p internal/eventsapi
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/eventsapi/generator_settings_types.yaml" specs/TS29222_CAPIF_Events_API.yaml
    oapi-codegen --config "$capifcore_dir/gogeneratorspecs/eventsapi/generator_settings_server.yaml" specs/TS29222_CAPIF_Events_API.yaml

    echo "Clean up"
    rm -rf specs
}


fix_package_imports () {
    echo "Fix package imports"
    if find "$cwd/internal" -type f -exec sed -i 's/oransc.org\/nonrtric\/capifcore/oransc.org\/nonrtric\/r1-sme-manager/g' {} +; then
        echo "Package import replacement successful"
    else
        echo "Package import replacement failed"
    fi
}

generate_mocks () {
    echo "Generating mocks"
    go generate ./...
}

running_tests () {
    echo "Running tests"
    cd internal
    go clean -testcache
    go test ./publishservice ./discoverservice
}

# Main code block

tear_down
make_internal_dirs
set_up_dir_paths
copy_test_wrappers
curl_api_specs
jar_extraction
fix_with_sed
gentools
code_generation
fix_package_imports
generate_mocks
running_tests
