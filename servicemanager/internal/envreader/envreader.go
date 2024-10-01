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
	"fmt"
	"net/url"
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
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
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
	if err == nil {
		err = validateEnv(myEnv)
	}

	if err == nil {
		err = validateUrls(myEnv, myPorts)
	}

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
	log.Infof("KONG_CONTROL_PLANE_IPV4 %s", myEnv["KONG_CONTROL_PLANE_IPV4"])
	log.Infof("KONG_CONTROL_PLANE_PORT %s", myEnv["KONG_CONTROL_PLANE_PORT"])
	log.Infof("KONG_DATA_PLANE_IPV4 %s", myEnv["KONG_DATA_PLANE_IPV4"])
	log.Infof("KONG_DATA_PLANE_PORT %s", myEnv["KONG_DATA_PLANE_PORT"])
	log.Infof("CAPIF_PROTOCOL %s", myEnv["CAPIF_PROTOCOL"])
	log.Infof("CAPIF_IPV4 %s", myEnv["CAPIF_IPV4"])
	log.Infof("CAPIF_PORT %s", myEnv["CAPIF_PORT"])
	log.Infof("LOG_LEVEL %s", myEnv["LOG_LEVEL"])
	log.Infof("SERVICE_MANAGER_PORT %s", myEnv["SERVICE_MANAGER_PORT"])
	log.Infof("TEST_SERVICE_IPV4 %s", myEnv["TEST_SERVICE_IPV4"])
	log.Infof("TEST_SERVICE_PORT %s", myEnv["TEST_SERVICE_PORT"])
}

func validateUrls(myEnv map[string]string, myPorts map[string]int) error {
	capifProtocol := myEnv["CAPIF_PROTOCOL"]
	capifIPv4 := myEnv["CAPIF_IPV4"]
	capifPort := myPorts["CAPIF_PORT"]
	capifcoreUrl := fmt.Sprintf("%s://%s:%d", capifProtocol, capifIPv4, capifPort)

	kongProtocol := myEnv["KONG_PROTOCOL"]
	kongControlPlaneIPv4 := myEnv["KONG_CONTROL_PLANE_IPV4"]
	kongControlPlanePort := myPorts["KONG_CONTROL_PLANE_PORT"]
	kongControlPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongControlPlaneIPv4, kongControlPlanePort)

	kongDataPlaneIPv4 := myEnv["KONG_DATA_PLANE_IPV4"]
	kongDataPlanePort := myPorts["KONG_DATA_PLANE_PORT"]
	kongDataPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongDataPlaneIPv4, kongDataPlanePort)

	log.Infof("Capifcore URL %s", capifcoreUrl)
	log.Infof("Kong Control Plane URL %s", kongControlPlaneURL)
	log.Infof("Kong Data Plane URL %s", kongDataPlaneURL)

	// Very basic checks
	_, err := url.ParseRequestURI(capifcoreUrl)
	if err != nil {
		err = fmt.Errorf("error parsing Capifcore URL: %s", err)
		return err
	}
	_, err = url.ParseRequestURI(kongControlPlaneURL)
	if err != nil {
		err = fmt.Errorf("error parsing Kong Control Plane URL: %s", err)
		return err
	}
	_, err = url.ParseRequestURI(kongDataPlaneURL)
	if err != nil {
		err = fmt.Errorf("error parsing Kong Data Plane URL: %s", err)
		return err
	}

	return nil
}

func validateEnv(myEnv map[string]string) error {
	var err error = nil

	kongDomain := myEnv["KONG_DOMAIN"]
	kongProtocol := myEnv["KONG_PROTOCOL"]
	kongControlPlaneIPv4 := myEnv["KONG_CONTROL_PLANE_IPV4"]
	kongDataPlaneIPv4 := myEnv["KONG_DATA_PLANE_IPV4"]
	capifProtocol := myEnv["CAPIF_PROTOCOL"]
	capifIPv4 := myEnv["CAPIF_IPV4"]

	if kongDomain == "" || kongDomain == "<string>" {
		err = fmt.Errorf("error loading KONG_DOMAIN from .env file: %s", kongDomain)
	} else if kongProtocol == "" || kongProtocol == "<http or https protocol scheme>" {
		err = fmt.Errorf("error loading KONG_PROTOCOL from .env file: %s", kongProtocol)
	} else if kongControlPlaneIPv4 == "" || kongControlPlaneIPv4 == "<host string>" {
		err = fmt.Errorf("error loading KONG_CONTROL_PLANE_IPV4 from .env file: %s", kongControlPlaneIPv4)
	} else if kongDataPlaneIPv4 == "" || kongDataPlaneIPv4 == "<host string>" {
		err = fmt.Errorf("error loading KONG_DATA_PLANE_IPV4 from .env file: %s", kongDataPlaneIPv4)
	} else if capifProtocol == "" || capifProtocol == "<http or https protocol scheme>" {
		err = fmt.Errorf("error loading CAPIF_PROTOCOL from .env file: %s", capifProtocol)
	} else if capifIPv4 == "" || capifIPv4 == "<host string>" || capifIPv4 == "<host>" {
		err = fmt.Errorf("error loading CAPIF_IPV4 from .env file: %s", capifIPv4)
	}
	// TEST_SERVICE_IPV4 is used only by the unit tests and are validated in the unit tests.

	return err
}

func createMapPorts(myEnv map[string]string) (map[string]int, error) {
    myPorts := make(map[string]int)
	var err error

	myPorts["KONG_DATA_PLANE_PORT"], err = strconv.Atoi(myEnv["KONG_DATA_PLANE_PORT"])
	if err != nil {
		err = fmt.Errorf("error loading KONG_DATA_PLANE_PORT from .env file: %s", err)
		return nil, err
	}

	myPorts["KONG_CONTROL_PLANE_PORT"], err = strconv.Atoi(myEnv["KONG_CONTROL_PLANE_PORT"])
	if err != nil {
		err = fmt.Errorf("error loading KONG_CONTROL_PLANE_PORT from .env file: %s", err)
		return nil, err
	}

	myPorts["CAPIF_PORT"], err = strconv.Atoi(myEnv["CAPIF_PORT"])
	if err != nil {
		err = fmt.Errorf("error loading CAPIF_PORT from .env file: %s", err)
		return nil, err
	}

	myPorts["SERVICE_MANAGER_PORT"], err = strconv.Atoi(myEnv["SERVICE_MANAGER_PORT"])
	if err != nil {
		err = fmt.Errorf("error loading SERVICE_MANAGER_PORT from .env file: %s", err)
		return nil, err
	}

	// TEST_SERVICE_PORT is required for unit testing, but not required for production
	if myEnv["TEST_SERVICE_PORT"] != "" {
		myPorts["TEST_SERVICE_PORT"], err = strconv.Atoi(myEnv["TEST_SERVICE_PORT"])
		if err != nil {
			err = fmt.Errorf("error loading TEST_SERVICE_PORT from .env file: %s", err)
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
