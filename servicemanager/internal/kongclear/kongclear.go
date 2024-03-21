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

package kongclear

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

// Service represents the structure for Kong service creation
type KongService struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Tags []string `json:"tags"`
}

type KongServiceResponse struct {
	ID string `json:"id"`
}

type ServiceResponse struct {
	Data []KongService `json:"data"`
}

type KongRoute struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Paths     []string `json:"paths"`
	Service   Service  `json:"service"`
	StripPath bool     `json:"strip_path"`
	Tags      []string `json:"tags"`
}

type RouteResponse struct {
	Data []KongRoute `json:"data"`
}

type Service struct {
	ID string `json:"id"`
}

func KongClear(myEnv map[string]string, myPorts map[string]int) error {
	log.Info("delete only ServiceManager Kong routes and services")

	kongAdminApiUrl := fmt.Sprintf("%s://%s:%d/", myEnv["KONG_PROTOCOL"], myEnv["KONG_IPV4"], myPorts["KONG_CONTROL_PLANE_PORT"])

	err := deleteRoutes(kongAdminApiUrl)
	if err != nil {
		log.Fatalf("error deleting routes %v", err)
		return err
	}

	err = deleteServices(kongAdminApiUrl)
	if err != nil {
		log.Fatalf("error deleting services %v", err)
		return err
	}

	log.Info("finished deleting only ServiceManger Kong routes and services")
	return err
}

func deleteRoutes(kongAdminApiUrl string) error {
	routes, err := listRoutes(kongAdminApiUrl)
	if err != nil {
		return err
	}

	for _, route := range routes {
		if areServiceManagerTags(route.Tags) {
			if err := deleteRoute(kongAdminApiUrl, route.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteServices(kongAdminApiUrl string) error {
	services, err := listServices(kongAdminApiUrl)
	if err != nil {
		return err
	}

	for _, service := range services {
		if areServiceManagerTags(service.Tags) {
			if err := deleteService(kongAdminApiUrl, service.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func listRoutes(kongAdminApiUrl string) ([]KongRoute, error) {
	client := resty.New()
	resp, err := client.R().
		Get(kongAdminApiUrl + "routes")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		err := fmt.Errorf("failed to list routes, status code %d", resp.StatusCode())
		return nil, err
	}

	var routeResponse RouteResponse
	err = json.Unmarshal(resp.Body(), &routeResponse)
	if err != nil {
		return nil, err
	}

	log.Infof("kong routes %v", routeResponse.Data)
	return routeResponse.Data, nil
}

func listServices(kongAdminApiUrl string) ([]KongService, error) {
	client := resty.New()
	resp, err := client.R().Get(kongAdminApiUrl + "services")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		err := fmt.Errorf("failed to list services, status code %d", resp.StatusCode())
		return nil, err
	}

	var serviceResponse ServiceResponse
	err = json.Unmarshal(resp.Body(), &serviceResponse)
	if err != nil {
		return nil, err
	}

	log.Infof("kong services %v", serviceResponse.Data)
	return serviceResponse.Data, nil
}

func areServiceManagerTags(tags []string) bool {
	tagMap := make(map[string]string)

	for _, tag := range tags {
		log.Debugf("found tag %s", tag)
		tagSlice := strings.Split(tag, ":")
		log.Debugf("tag slice %v", tagSlice)
		if (len(tagSlice) > 0) && (tagSlice[0] != "") {
			if (len(tagSlice) > 1) {
				tagMap[tagSlice[0]] = tagSlice[1]
			} else {
				tagMap[tagSlice[0]] = ""
			}
		}
	}

	if tagMap["apfId"] == "" {
		log.Debug("did NOT find apfId")
		return false
	}
	log.Debugf("found valid apfId %s", tagMap["apfId"])

	if tagMap["aefId"] == "" {
		log.Debug("did NOT find aefId")
		return false
	}
	log.Debugf("found valid aefId %s", tagMap["aefId"])

	if tagMap["apiId"] == "" {
		log.Debug("did NOT find apiId")
		return false
	}
	log.Debugf("found valid apiId %s", tagMap["apiId"])

	return true
}

func deleteRoute(kongAdminApiUrl string, routeID string) error {
	log.Infof("delete kong route id %s", routeID)
	client := resty.New()
	resp, err := client.R().Delete(kongAdminApiUrl + "routes/" + routeID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		err := fmt.Errorf("failed to delete route %s, status code %d", routeID, resp.StatusCode())
		return err
	}

	return nil
}

func deleteService(kongAdminApiUrl string, serviceID string) error {
	log.Infof("delete kong service id %s", serviceID)
	client := resty.New()
	resp, err := client.R().Delete(kongAdminApiUrl + "services/" + serviceID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		err := fmt.Errorf("failed to delete service %s, status code %d", serviceID, resp.StatusCode())
		return err
	}

	return nil
}
