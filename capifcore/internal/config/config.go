// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023: Nordix Foundation
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

package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type AdminUser struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type AuthorizationServer struct {
	Port      string            `yaml:"port"`
	Host      string            `yaml:"host"`
	AdminUser AdminUser         `yaml:"admin"`
	Realms    map[string]string `yaml:"realms"`
}

type Config struct {
	AuthorizationServer AuthorizationServer `yaml:"authorizationServer"`
}

func ReadKeycloakConfigFile(configFolder string) (*Config, error) {

	f, err := os.Open(filepath.Join(configFolder, "keycloak.yaml"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg *Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
