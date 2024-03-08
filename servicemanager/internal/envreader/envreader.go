// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2024: OpenInfra Foundation Europe
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

package envreader

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type ConfigReader interface {
    ReadDotEnv() (map[string]string, map[string]int, error)
}

// RealConfigReader implements ConfigReader
type RealConfigReader struct {
}

func (r *RealConfigReader) ReadDotEnv() (map[string]string, map[string]int, error) {
	setLogLevel("Info")

	env := os.Getenv("SERVICE_MANAGER_ENV")
	log.Infof("read SERVICE_MANAGER_ENV: %s", env)

	if env == "" {
		env = "development"
	}

	// Root folder of this project
	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Join(filepath.Dir(file), "../..")
	basePath += "/"

	var myEnv map[string]string
	envFile := basePath + ".env." + env
	myEnv, err := godotenv.Read(envFile)

	if err != nil {
		log.Warnf("error reading .env file: %s", err)

		envFile = basePath + ".env"
		myEnv, err = godotenv.Read(envFile)
		if err != nil {
			log.Fatalf("error reading .env file: %s", err)
			return nil, nil, err
		}
	}

	setLogLevel(myEnv["LOG_LEVEL"])
	logConfig(myEnv, envFile)

	myPorts, err := createMapPorts(myEnv)
	return myEnv, myPorts, err
}

func setLogLevel(logLevel string) error {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	loglevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("error loading LOG_LEVEL from .env file: %v", err)
		return err
	}
	log.SetLevel(loglevel)
	return nil
}

func logConfig(myEnv map[string]string, envFile string) {
	log.Infof("imported .env: %s", envFile)

	log.Infof("KONG_DOMAIN %s", myEnv["KONG_DOMAIN"])
	log.Infof("KONG_PROTOCOL %s", myEnv["KONG_PROTOCOL"])
	log.Infof("KONG_IPV4 %s", myEnv["KONG_IPV4"])
	log.Infof("KONG_DATA_PLANE_PORT %s", myEnv["KONG_DATA_PLANE_PORT"])
	log.Infof("KONG_CONTROL_PLANE_PORT %s", myEnv["KONG_CONTROL_PLANE_PORT"])
	log.Infof("CAPIF_PROTOCOL %s", myEnv["CAPIF_PROTOCOL"])
	log.Infof("CAPIF_IPV4 %s", myEnv["CAPIF_IPV4"])
	log.Infof("CAPIF_PORT %s", myEnv["CAPIF_PORT"])
	log.Infof("LOG_LEVEL %s", myEnv["LOG_LEVEL"])
	log.Infof("SERVICE_MANAGER_PORT %s", myEnv["SERVICE_MANAGER_PORT"])
	log.Infof("TEST_SERVICE_IPV4 %s", myEnv["TEST_SERVICE_IPV4"])
	log.Infof("TEST_SERVICE_PORT %s", myEnv["TEST_SERVICE_PORT"])
}

func createMapPorts(myEnv map[string]string) (map[string]int, error) {
    myPorts := make(map[string]int)
	var err error

	myPorts["KONG_DATA_PLANE_PORT"], err = strconv.Atoi(myEnv["KONG_DATA_PLANE_PORT"])
	if err != nil {
		log.Fatalf("error loading KONG_DATA_PLANE_PORT from .env file: %s", err)
		return nil, err
	}

	myPorts["KONG_CONTROL_PLANE_PORT"], err = strconv.Atoi(myEnv["KONG_CONTROL_PLANE_PORT"])
	if err != nil {
		log.Fatalf("error loading KONG_CONTROL_PLANE_PORT from .env file: %s", err)
		return nil, err
	}

	myPorts["CAPIF_PORT"], err = strconv.Atoi(myEnv["CAPIF_PORT"])
	if err != nil {
		log.Fatalf("error loading CAPIF_PORT from .env file: %s", err)
		return nil, err
	}

	myPorts["SERVICE_MANAGER_PORT"], err = strconv.Atoi(myEnv["SERVICE_MANAGER_PORT"])
	if err != nil {
		log.Fatalf("error loading SERVICE_MANAGER_PORT from .env file: %s", err)
		return nil, err
	}

	// TEST_SERVICE_PORT is required for unit testing, but not required for production
	if myEnv["TEST_SERVICE_PORT"] != "" {
		myPorts["TEST_SERVICE_PORT"], err = strconv.Atoi(myEnv["TEST_SERVICE_PORT"])
		if err != nil {
			log.Fatalf("error loading TEST_SERVICE_PORT from .env file: %s", err)
			return nil, err
		}
	}

	return myPorts, err
}

// MockConfigReader is a mock implementation for testing
type MockConfigReader struct {
    MockedConfig map[string]string
}

func (m *MockConfigReader) ReadDotEnv() (map[string]string, map[string]int, error) {
	const envFile = "mock"

	setLogLevel(m.MockedConfig["LOG_LEVEL"])
	logConfig(m.MockedConfig, envFile)

	// Return the mocked configuration for testing
	myPorts, err := createMapPorts(m.MockedConfig)
    return m.MockedConfig, myPorts, err
}