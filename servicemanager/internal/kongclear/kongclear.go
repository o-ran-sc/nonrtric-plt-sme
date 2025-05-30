// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2024-2025: OpenInfra Foundation Europe
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
	"net/url"
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
	Offset string        `json:"offset"`
	Data   []KongService `json:"data"`
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
	Offset string      `json:"offset"`
	Data   []KongRoute `json:"data"`
}

type Service struct {
	ID string `json:"id"`
}

func KongClear(myEnv map[string]string, myPorts map[string]int) error {
	log.Info("delete only ServiceManager Kong routes and services")

	kongAdminApiUrl := fmt.Sprintf("%s://%s:%d/", myEnv["KONG_PROTOCOL"], myEnv["KONG_CONTROL_PLANE_IPV4"], myPorts["KONG_CONTROL_PLANE_PORT"])

	err := DeleteRoutes(kongAdminApiUrl, "", "")
	if err != nil {
		log.Fatalf("error deleting routes %v", err)
		return err
	}

	err = DeleteServices(kongAdminApiUrl, "", "")
	if err != nil {
		log.Fatalf("error deleting services %v", err)
		return err
	}

	log.Info("finished deleting only ServiceManger Kong routes and services")
	return err
}

func DeleteRoutes(kongAdminApiUrl string, offset string, tags string) error {
	kongRoutesApiUrl := kongAdminApiUrl + "routes"

	params := url.Values{}
	if offset != "" {
		log.Tracef("Using offset %s for kong routes", offset)
		params.Add("offset", offset)
	}
	if tags != "" {
		log.Tracef("Using tags %s for kong routes", tags)
		params.Add("tags", tags)
	}

	if len(params) > 0 {
		log.Debugf("Using params %s for kong routes", params.Encode())
		kongRoutesApiUrl += "?" + params.Encode()
	}

	routes, nextOffset, err := listRoutes(kongRoutesApiUrl)
	if err != nil {
		return err
	}
	log.Infof("Fetched kong routes size is %d", len(routes))

	for _, route := range routes {
		if areServiceManagerTags(route.Tags) {
			if err := deleteRoute(kongAdminApiUrl, route.Name); err != nil {
				return err
			}
		}
	}

	// If the offset is not empty, it means there are more routes to process
	if nextOffset != "" {
		log.Tracef("More routes to process, offset is %s", nextOffset)
		if err := DeleteRoutes(kongAdminApiUrl, nextOffset, tags); err != nil {
			return err
		}
	}

	return nil
}

func DeleteServices(kongAdminApiUrl string, offset string, tags string) error {
	kongServiceApiUrl := kongAdminApiUrl + "services"

	params := url.Values{}
	if offset != "" {
		log.Tracef("Using offset %s for kong services", offset)
		params.Add("offset", offset)
	}
	if tags != "" {
		log.Tracef("Using tags %s for kong services", tags)
		params.Add("tags", tags)
	}

	if len(params) > 0 {
		log.Debugf("Using params %s for kong services", params.Encode())
		kongServiceApiUrl += "?" + params.Encode()
	}

	services, nextOffset, err := listServices(kongServiceApiUrl)
	if err != nil {
		return err
	}

	log.Infof("Fetched Kong services size is %d", len(services))

	for _, service := range services {
		if areServiceManagerTags(service.Tags) {
			if err := deleteService(kongAdminApiUrl, service.Name); err != nil {
				return err
			}
		}
	}

	// If the offset is not empty, it means there are more services to process
	if nextOffset != "" {
		log.Tracef("More services to process, offset is %s", nextOffset)
		if err := DeleteServices(kongAdminApiUrl, nextOffset, tags); err != nil {
			return err
		}
	}

	return nil
}

func listRoutes(kongRoutesApiUrl string) ([]KongRoute, string, error) {
	log.Debugf("List kong routes from %s", kongRoutesApiUrl)
	client := resty.New()
	resp, err := client.R().
		Get(kongRoutesApiUrl)

	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode() != http.StatusOK {
		err := fmt.Errorf("failed to list routes, status code %d", resp.StatusCode())
		return nil, "", err
	}

	var routeResponse RouteResponse
	err = json.Unmarshal(resp.Body(), &routeResponse)
	if err != nil {
		return nil, "", err
	}

	log.Debugf("Kong routes %v", routeResponse.Data)
	return routeResponse.Data, routeResponse.Offset, nil
}

func listServices(kongServicesApiUrl string) ([]KongService, string, error) {
	log.Debugf("List kong services from %s", kongServicesApiUrl)
	client := resty.New()
	resp, err := client.R().Get(kongServicesApiUrl)

	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode() != http.StatusOK {
		err := fmt.Errorf("failed to list services, status code %d", resp.StatusCode())
		return nil, "", err
	}

	var serviceResponse ServiceResponse
	err = json.Unmarshal(resp.Body(), &serviceResponse)
	if err != nil {
		return nil, "", err
	}

	log.Debugf("Kong services %v", serviceResponse.Data)
	return serviceResponse.Data, serviceResponse.Offset, nil
}

func areServiceManagerTags(tags []string) bool {
	tagMap := make(map[string]string)

	for _, tag := range tags {
		log.Debugf("found tag %s", tag)
		tagSlice := strings.Split(tag, ":")
		log.Debugf("tag slice %v", tagSlice)
		if (len(tagSlice) > 0) && (tagSlice[0] != "") {
			if len(tagSlice) > 1 {
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
	log.Debugf("delete kong route %s", routeID)
	client := resty.New()
	resp, err := client.R().Delete(kongAdminApiUrl + "routes/" + routeID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		err := fmt.Errorf("failed to delete route %s, status code %d", routeID, resp.StatusCode())
		return err
	}

	log.Infof("kong route %s deleted successfully", routeID)
	return nil
}

func deleteService(kongAdminApiUrl string, serviceID string) error {
	log.Debugf("delete kong service %s", serviceID)
	client := resty.New()
	resp, err := client.R().Delete(kongAdminApiUrl + "services/" + serviceID)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusNoContent {
		err := fmt.Errorf("failed to delete service %s, status code %d", serviceID, resp.StatusCode())
		return err
	}

	log.Infof("kong service %s deleted successfully", serviceID)
	return nil
}
