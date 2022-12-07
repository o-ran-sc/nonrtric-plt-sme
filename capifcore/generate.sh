# -
#   ========================LICENSE_START=================================
#   O-RAN-SC
#   %%
#   Copyright (C) 2022: Nordix Foundation
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

cwd=$(pwd)

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

cd specs/

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

# Remove types that are not used by CAPIF that have dependencies to other specifications.
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\CivicAddress/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\ExternalMbsServiceArea/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\GeographicArea/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\GeoServiceArea/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsMediaComp/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsMediaCompRm/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsMediaInfo/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsServiceInfo/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\MbsSession/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed -e 'H;x;/^\(  *\)\n\1/{s/\n.*//;x;d;}' -e 's/.*//;x;/\SpatialValidityCond/{s/^\( *\).*/ \1/;x;d;}' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml

# Remove attributes that can not be generated easily.
sed '/accessTokenError.*/,+3d' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml
sed '/accessTokenRequest.*/,+3d' TS29571_CommonData.yaml >temp.yaml
mv temp.yaml TS29571_CommonData.yaml

sed '/oneOf.*/,+2d' TS29222_CAPIF_Publish_Service_API.yaml >temp.yaml
mv temp.yaml TS29222_CAPIF_Publish_Service_API.yaml

sed '/oneOf.*/,+2d' TS29222_CAPIF_Security_API.yaml >temp.yaml
mv temp.yaml TS29222_CAPIF_Security_API.yaml

# Replace references to external specs that are collected to the common spec by the commoncollector
# <replacements_start>
cat TS29122_CommonData.yaml | sed 's/TS29572_Nlmf_Location/CommonData/g' > temp.yaml
mv temp.yaml TS29122_CommonData.yaml
cat TS29122_CommonData.yaml | sed 's/TS29554_Npcf_BDTPolicyControl/CommonData/g' > temp.yaml
mv temp.yaml TS29122_CommonData.yaml
cat TS29122_CommonData.yaml | sed 's/TS29514_Npcf_PolicyAuthorization/CommonData/g' > temp.yaml
mv temp.yaml TS29122_CommonData.yaml
cat TS29571_CommonData.yaml | sed 's/TS29514_Npcf_PolicyAuthorization/CommonData/g' > temp.yaml
mv temp.yaml TS29571_CommonData.yaml
cat TS29571_CommonData.yaml | sed 's/TS29572_Nlmf_Location/CommonData/g' > temp.yaml
mv temp.yaml TS29571_CommonData.yaml
cat TS29222_CAPIF_Publish_Service_API.yaml | sed 's/TS29572_Nlmf_Location/CommonData/g' > temp.yaml
mv temp.yaml TS29222_CAPIF_Publish_Service_API.yaml
cat TS29222_CAPIF_Routing_Info_API.yaml | sed 's/TS29520_Nnwdaf_EventsSubscription/CommonData/g' > temp.yaml
mv temp.yaml TS29222_CAPIF_Routing_Info_API.yaml
cat TS29222_CAPIF_Routing_Info_API.yaml | sed 's/TS29510_Nnrf_NFManagement/CommonData/g' > temp.yaml
mv temp.yaml TS29222_CAPIF_Routing_Info_API.yaml
cat TS29222_CAPIF_Events_API.yaml | sed 's/TS29523_Npcf_EventExposure/CommonData/g' > temp.yaml
mv temp.yaml TS29222_CAPIF_Events_API.yaml
# <new_replacement>

# This spec has references to itself that need to be removed
cat TS29571_CommonData.yaml | sed 's/TS29571_CommonData.yaml//g' > temp.yaml
mv temp.yaml TS29571_CommonData.yaml

cd $cwd

echo "Fixing enums"
cd internal/gentools/enumfixer
go build .
./enumfixer -apidir=../../../specs

cd $cwd
echo "Gathering common references"
cd internal/gentools/commoncollector
go build .
./commoncollector -apidir=../../../specs

cd $cwd
echo "Fixing misc in specifications"
cd internal/gentools/specificationfixer
go build .
./specificationfixer -apidir=../../../specs

cd $cwd

go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.10.1
PATH=$PATH:~/go/bin

echo "Generating TS29122_CommonData"
mkdir -p internal/common29122
oapi-codegen --config gogeneratorspecs/common29122/generator_settings.yaml specs/TS29122_CommonData.yaml

echo "Generating aggregated CommonData"
mkdir -p internal/common
oapi-codegen --config gogeneratorspecs/common/generator_settings.yaml specs/CommonData.yaml

echo "Generating TS29571_CommonData"
mkdir -p internal/common29571
oapi-codegen --config gogeneratorspecs/common29571/generator_settings.yaml specs/TS29571_CommonData.yaml

echo "Generating TS29222_CAPIF_Publish_Service_API"
mkdir -p internal/publishserviceapi
oapi-codegen --config gogeneratorspecs/publishserviceapi/generator_settings_types.yaml specs/TS29222_CAPIF_Publish_Service_API.yaml
oapi-codegen --config gogeneratorspecs/publishserviceapi/generator_settings_server.yaml specs/TS29222_CAPIF_Publish_Service_API.yaml

echo "Generating TS29222_CAPIF_API_Invoker_Management_API"
mkdir -p internal/invokermanagementapi
oapi-codegen --config gogeneratorspecs/invokermanagementapi/generator_settings_types.yaml specs/TS29222_CAPIF_API_Invoker_Management_API.yaml
oapi-codegen --config gogeneratorspecs/invokermanagementapi/generator_settings_server.yaml specs/TS29222_CAPIF_API_Invoker_Management_API.yaml

echo "Generating TS29222_CAPIF_API_Provider_Management_API"
mkdir -p internal/providermanagementapi
oapi-codegen --config gogeneratorspecs/providermanagementapi/generator_settings_types.yaml specs/TS29222_CAPIF_API_Provider_Management_API.yaml
oapi-codegen --config gogeneratorspecs/providermanagementapi/generator_settings_server.yaml specs/TS29222_CAPIF_API_Provider_Management_API.yaml

echo "Generating TS29222_CAPIF_Discover_Service_API"
mkdir -p internal/discoverserviceapi
oapi-codegen --config gogeneratorspecs/discoverserviceapi/generator_settings_types.yaml specs/TS29222_CAPIF_Discover_Service_API.yaml
oapi-codegen --config gogeneratorspecs/discoverserviceapi/generator_settings_server.yaml specs/TS29222_CAPIF_Discover_Service_API.yaml

echo "Generating TS29222_CAPIF_Security_API"
mkdir -p internal/securityapi
oapi-codegen --config gogeneratorspecs/securityapi/generator_settings_types.yaml specs/TS29222_CAPIF_Security_API.yaml
oapi-codegen --config gogeneratorspecs/securityapi/generator_settings_server.yaml specs/TS29222_CAPIF_Security_API.yaml

echo "Generating TS29222_CAPIF_Logging_API_Invocation_API"
mkdir -p internal/loggingapi
oapi-codegen --config gogeneratorspecs/loggingapi/generator_settings_types.yaml specs/TS29222_CAPIF_Logging_API_Invocation_API.yaml
oapi-codegen --config gogeneratorspecs/loggingapi/generator_settings_server.yaml specs/TS29222_CAPIF_Logging_API_Invocation_API.yaml

echo "Generating TS29222_CAPIF_Routing_Info_API"
mkdir -p internal/routinginfoapi
oapi-codegen --config gogeneratorspecs/routinginfoapi/generator_settings_types.yaml specs/TS29222_CAPIF_Routing_Info_API.yaml
oapi-codegen --config gogeneratorspecs/routinginfoapi/generator_settings_server.yaml specs/TS29222_CAPIF_Routing_Info_API.yaml

echo "Generating TS29222_CAPIF_Access_Control_Policy_API"
mkdir -p internal/accesscontrolpolicyapi
oapi-codegen --config gogeneratorspecs/accesscontrolpolicyapi/generator_settings_types.yaml specs/TS29222_CAPIF_Access_Control_Policy_API.yaml
oapi-codegen --config gogeneratorspecs/accesscontrolpolicyapi/generator_settings_server.yaml specs/TS29222_CAPIF_Access_Control_Policy_API.yaml

echo "Generating TS29222_CAPIF_Events_API"
mkdir -p internal/eventsapi
oapi-codegen --config gogeneratorspecs/eventsapi/generator_settings_types.yaml specs/TS29222_CAPIF_Events_API.yaml
oapi-codegen --config gogeneratorspecs/eventsapi/generator_settings_server.yaml specs/TS29222_CAPIF_Events_API.yaml

echo "Generating TS29222_CAPIF_Auditing_API"
mkdir -p internal/auditingapi
oapi-codegen --config gogeneratorspecs/auditingapi/generator_settings_types.yaml specs/TS29222_CAPIF_Auditing_API.yaml
oapi-codegen --config gogeneratorspecs/auditingapi/generator_settings_server.yaml specs/TS29222_CAPIF_Auditing_API.yaml

echo "Generating TS29222_AEF_Security_API"
mkdir -p internal/aefsecurityapi
oapi-codegen --config gogeneratorspecs/aefsecurityapi/generator_settings_types.yaml specs/TS29222_AEF_Security_API.yaml
oapi-codegen --config gogeneratorspecs/aefsecurityapi/generator_settings_server.yaml specs/TS29222_AEF_Security_API.yaml

echo "Cleanup"
rm -rf specs

echo "Generating mocks."
go generate ./...