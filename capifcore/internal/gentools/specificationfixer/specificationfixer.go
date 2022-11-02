// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022: Nordix Foundation
//   %%
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//   ========================LICENSE_END===================================
//

package main

import (
	"flag"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var apiDir *string

func main() {
	apiDir = flag.String("apidir", "", "Directory containing API definitions to fix")
	flag.Parse()

	m := getData("TS29571_CommonData.yaml")
	components := m["components"]
	cMap := components.(map[interface{}]interface{})
	schemas := cMap["schemas"].(map[interface{}]interface{})
	snssaiExtensionData := schemas["SnssaiExtension"].(map[interface{}]interface{})
	props := snssaiExtensionData["properties"].(map[interface{}]interface{})
	wildcardSdData := props["wildcardSd"].(map[interface{}]interface{})
	delete(wildcardSdData, "enum")

	writeFile("TS29571_CommonData.yaml", m)

	m = getData("TS29222_CAPIF_Security_API.yaml")
	components = m["components"]
	cMap = components.(map[interface{}]interface{})
	schemas = cMap["schemas"].(map[interface{}]interface{})
	accessTokenReq := schemas["AccessTokenReq"].(map[interface{}]interface{})
	accessTokenReq["type"] = "object"

	writeFile("TS29222_CAPIF_Security_API.yaml", m)
}

func getData(filename string) map[string]interface{} {
	yamlFile, err := ioutil.ReadFile(*apiDir + "/" + filename)
	if err != nil {
		log.Fatalf("Error reading yamlFile. #%v ", err)
	}
	m := make(map[string]interface{})
	err = yaml.Unmarshal(yamlFile, m)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return m
}

func writeFile(filename string, data map[string]interface{}) {
	modCommon, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalf("Marshal: #%v ", err)
	}
	err = ioutil.WriteFile(*apiDir+"/"+filename, modCommon, 0644)
	if err != nil {
		log.Fatalf("Error writing yamlFile. #%v ", err)
	}
}
