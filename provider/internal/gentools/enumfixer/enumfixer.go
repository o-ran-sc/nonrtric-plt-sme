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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Enum struct {
	Enum        []string `yaml:"enum"`
	Type        string   `yaml:"type"`
	Description string   `yaml:"description"`
}

func main() {
	var apiDir = flag.String("apidir", "", "Directory containing API definitions to fix")
	flag.Parse()
	err := filepath.Walk(*apiDir, fixEnums)
	if err != nil {
		fmt.Println(err)
	}
}

func fixEnums(path string, info os.FileInfo, _ error) error {
	if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("yamlFile. Get err   #%v ", err)
		}
		m := make(map[string]interface{})
		err = yaml.Unmarshal(yamlFile, m)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
		components := m["components"]
		if components != nil {
			cMap := components.(map[interface{}]interface{})
			if _, ok := cMap["schemas"].(map[interface{}]interface{}); ok {
				schemas := cMap["schemas"].(map[interface{}]interface{})
				for typeName, typeDef := range schemas {
					tDMap := typeDef.(map[interface{}]interface{})
					anyOf, ok := tDMap["anyOf"]
					if ok {
						aOSlice := anyOf.([]interface{})
						correctEnum := Enum{}
						mapInterface := aOSlice[0].(map[interface{}]interface{})
						enumInterface := mapInterface["enum"]
						if enumInterface != nil {
							is := enumInterface.([]interface{})
							var enumVals []string
							for i := 0; i < len(is); i++ {
								if reflect.TypeOf(is[i]).Kind() == reflect.String {
									enumVals = append(enumVals, is[i].(string))

								} else if reflect.TypeOf(is[1]).Kind() == reflect.Int {
									enumVals = append(enumVals, strconv.Itoa(is[i].(int)))
								}
							}
							correctEnum.Enum = enumVals
							correctEnum.Type = "string"
							description := tDMap["description"]
							if description != nil {
								correctEnum.Description = description.(string)
							} else {
								if aOSlice[1] != nil {
									mapInterface = aOSlice[1].(map[interface{}]interface{})
									description := mapInterface["description"]
									if description != nil {
										correctEnum.Description = description.(string)
									}
								}
							}
							schemas[typeName] = correctEnum
						}
					}
				}
				modM, err := yaml.Marshal(m)
				if err != nil {
					log.Printf("yamlFile. Get err   #%v ", err)
				}
				err = ioutil.WriteFile(path, modM, 0644)
				if err != nil {
					log.Printf("yamlFile. Get err   #%v ", err)
				}
			}
		}
	}
	return nil
}
