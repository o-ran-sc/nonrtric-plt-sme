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
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var apiDir *string
var common = map[interface{}]interface{}{
	"openapi": "3.0.0",
	"info": map[interface{}]interface{}{
		"title":   "Common",
		"version": "1.0.0",
	},
	"components": map[interface{}]interface{}{
		"schemas": map[interface{}]interface{}{},
	},
}

func main() {
	apiDir = flag.String("apidir", "", "Directory containing API definitions to fix")
	flag.Parse()

	file, err := os.Open("definitions.txt")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	components := common["components"]
	cMap := components.(map[interface{}]interface{})
	schemas := cMap["schemas"].(map[interface{}]interface{})
	for scanner.Scan() {
		name, data := getDependency(scanner.Text())
		if name == "EthFlowDescription" {
			changeToLocalReference("fDir", "FlowDirection", data)
		}
		if name == "ReportingInformation" {
			changeToLocalReference("notifMethod", "NotificationMethod", data)
		}
		if name == "RelativeCartesianLocation" {
			properties := data["properties"].(map[interface{}]interface{})
			delete(properties, true)
			data["required"] = remove(data["required"].([]interface{}), 1)
		}
		schemas[name] = data
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	modCommon, err := yaml.Marshal(common)
	if err != nil {
		log.Fatalf("Marshal: #%v ", err)
	}
	err = ioutil.WriteFile(*apiDir+"/"+"CommonData.yaml", modCommon, 0644)
	if err != nil {
		log.Fatalf("Error writing yamlFile. #%v ", err)
	}
}

func changeToLocalReference(attrname, refName string, data map[interface{}]interface{}) {
	properties := data["properties"].(map[interface{}]interface{})
	ref := properties[attrname].(map[interface{}]interface{})
	ref["$ref"] = "#/components/schemas/" + refName
}

func getDependency(s string) (string, map[interface{}]interface{}) {
	info := strings.Split(s, "#")
	yamlFile, err := ioutil.ReadFile(*apiDir + "/" + info[0])
	if err != nil {
		log.Fatalf("Error reading yamlFile. #%v ", err)
	}
	m := make(map[string]interface{})
	err = yaml.Unmarshal(yamlFile, m)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	components := m["components"]
	cMap := components.(map[interface{}]interface{})
	schemas := cMap["schemas"].(map[interface{}]interface{})
	component := strings.Split(info[1], "/")
	dep := schemas[component[3]].(map[interface{}]interface{})
	return component[3], dep
}

func remove(slice []interface{}, s int) []interface{} {
	return append(slice[:s], slice[s+1:]...)
}
