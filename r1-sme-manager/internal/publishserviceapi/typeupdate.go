// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023-2024: OpenInfra Foundation Europe
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

package publishserviceapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	common29122 "oransc.org/nonrtric/r1-sme-manager/internal/common29122"
)

func (sd *ServiceAPIDescription) PrepareNewService() {
	apiName := "api_id_" + strings.ReplaceAll(sd.ApiName, " ", "_")
	sd.ApiId = &apiName
}

func (sd *ServiceAPIDescription) RegisterKong(kongDomain string, kongProtocol string, kongIPv4 common29122.Ipv4Addr, kongDataPlanePort common29122.Port, kongControlPlanePort common29122.Port) (int, error) {
	log.Info("Entering RegisterKong")
	var (
		statusCode int
		err        error
	)
	kongControlPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongIPv4, kongControlPlanePort)

	statusCode, err = sd.createKongRoutes(kongControlPlaneURL)
	if (err != nil) || (statusCode != http.StatusCreated) {
		return statusCode, err
	}

	sd.updateInterfaceDescription(kongIPv4, kongDataPlanePort, kongDomain)

	log.Info("Exiting from RegisterKong")
	return statusCode, nil
}

func (sd *ServiceAPIDescription) createKongRoutes(kongControlPlaneURL string) (int, error) {
	log.Info("Entering createKongRoutes")
	var (
		statusCode int
		err        error
	)

	client := resty.New()

	profiles := *sd.AefProfiles
	for _, profile := range profiles {
		log.Infof("createKongRoutes, AefId %s", profile.AefId)
		for _, version := range profile.Versions {
			log.Infof("createKongRoutes, apiVersion \"%s\"", version.ApiVersion)
			for _, resource := range *version.Resources {
				statusCode, err = sd.createKongRoute(kongControlPlaneURL, client, resource, profile.AefId, version.ApiVersion)
				if (err != nil) || (statusCode != http.StatusCreated) {
					return statusCode, err
				}
			}
		}
	}
	return statusCode, nil
}

func (sd *ServiceAPIDescription) createKongRoute(kongControlPlaneURL string, client *resty.Client, resource Resource, aefId string, apiVersion string) (int, error) {
	log.Info("Entering createKongRoute")
	uri := resource.Uri

	if apiVersion != "" {
		if apiVersion[0] != '/' {
			apiVersion = "/" + apiVersion
		}
		if apiVersion[len(apiVersion)-1] != '/' && resource.Uri[0] != '/' {
			apiVersion = apiVersion + "/"
		}
		uri = apiVersion + resource.Uri
	}



	log.Infof("createKongRoute, uri %s", uri)

	serviceName := *sd.ApiId + "_" + resource.ResourceName
	log.Infof("createKongRoute, serviceName %s", serviceName)
	log.Infof("createKongRoute, aefId %s", aefId)

	statusCode, err := sd.createKongService(kongControlPlaneURL, serviceName, uri, aefId)
	if (err != nil) || (statusCode != http.StatusCreated) {
		return statusCode, err
	}

	routeName := serviceName
	kongRoutesURL := kongControlPlaneURL + "/services/" + serviceName + "/routes"

	// Define the route information for Kong
	kongRouteInfo := map[string]interface{}{
		"name":       routeName,
		"paths":      []string{uri},
		"methods":    resource.Operations,
		"tags":       []string{aefId},
		"strip_path": true,
	}

	// Make the POST request to create the Kong service
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(kongRouteInfo).
		Post(kongRoutesURL)

	// Check for errors in the request
	if err != nil {
		log.Infof("Error: %v", err)
		return resp.StatusCode(), err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusCreated {
		log.Infof("Kong route %s created successfully!", routeName)
	} else {
		err = fmt.Errorf("the Kong service already exists. Status code: %d", resp.StatusCode()) 
		log.Error(err.Error())
		log.Infof("Response body: %s", resp.Body())
		return resp.StatusCode(), err
	}

	return resp.StatusCode(), nil
}

func (sd *ServiceAPIDescription) createKongService(kongControlPlaneURL string, kongServiceName string, kongServiceUri string, aefId string) (int, error) {
	log.Info("Entering createKongService")
	log.Infof("createKongService, kongServiceName %s", kongServiceName)

	// Define the service information for Kong
	firstAEFProfileIpv4Addr, firstAEFProfilePort, err := sd.findFirstAEFProfile()
	if err != nil {
		return http.StatusBadRequest, err
	}

	const kongProtocol = "http"

	kongServiceInfo := map[string]interface{}{
		"host":     firstAEFProfileIpv4Addr,
		"name":     kongServiceName,
		"port":     firstAEFProfilePort,
		"protocol": kongProtocol,
		"path":     kongServiceUri,
		"tags":     []string{aefId},
	}

	// Kong admin API endpoint for creating a service
	kongServicesURL := kongControlPlaneURL + "/services"

	// Create a new Resty client
	client := resty.New()

	// Make the POST request to create the Kong service
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(kongServiceInfo).
		Post(kongServicesURL)

	// Check for errors in the request
	if err != nil {
		log.Infof("Create Kong Service Request Error: %v", err)
		return http.StatusInternalServerError, err
	}

	// Check the response status code
	statusCode := resp.StatusCode()
	if statusCode == http.StatusCreated {
		log.Infof("Kong service %s created successfully!", kongServiceName)
	} else if resp.StatusCode() == http.StatusConflict {
		log.Errorf("the Kong service already exists. Status code: %d", resp.StatusCode())
		err = fmt.Errorf("service with identical apiName is already published") // for compatibilty with Capif error message on a duplicate service
		statusCode = http.StatusForbidden // for compatibilty with the spec, TS29222_CAPIF_Publish_Service_API
	} else {
		err = fmt.Errorf("error creating Kong service. Status code: %d", resp.StatusCode())
	}
	if err != nil {
		log.Errorf(err.Error())
		log.Errorf("response body: %s", resp.Body())
	}

	return statusCode, err
}

func (sd *ServiceAPIDescription) findFirstAEFProfile() (common29122.Ipv4Addr, common29122.Port, error) {
	log.Infof("Entering findFirstAEFProfile")
	aefProfiles := *sd.AefProfiles
	aefProfile := aefProfiles[0]

	var err error

	if aefProfile.InterfaceDescriptions == nil {
		log.Infof("Cannot read interfaceDescription")
		err = errors.New("cannot read interfaceDescription")
		return "", common29122.Port(0), err
	}

	interfaceDescription := (*aefProfile.InterfaceDescriptions)[0]
	firstIpv4Addr := *interfaceDescription.Ipv4Addr
	firstPort := *interfaceDescription.Port

	log.Infof("findFirstAEFProfile firstIpv4Addr %s firstPort %d", firstIpv4Addr, firstPort)

	return firstIpv4Addr, firstPort, err
}

// Update our exposures to point to Kong by replacing in incoming interface description with Kong interface descriptions.
func (sd *ServiceAPIDescription) updateInterfaceDescription(kongIPv4 common29122.Ipv4Addr, kongDataPlanePort common29122.Port, kongDomain string) {
	log.Info("Updating InterfaceDescriptions!")

	interfaceDesc := InterfaceDescription{
		Ipv4Addr: &kongIPv4,
		Port:     &kongDataPlanePort,
	}
	interfaceDescs := []InterfaceDescription{interfaceDesc}

	profiles := *sd.AefProfiles
	for i, profile := range profiles {
		profile.DomainName = &kongDomain
		profile.InterfaceDescriptions = &interfaceDescs
		profiles[i] = profile
	}
}

func (sd *ServiceAPIDescription) UnregisterKong(kongDomain string, kongProtocol string, kongIPv4 common29122.Ipv4Addr, kongDataPlanePort common29122.Port, kongControlPlanePort common29122.Port) (int, error) {
	log.Info("Entering UnregisterKong")

	var (
		statusCode int
		err        error
	)
	kongControlPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongIPv4, kongControlPlanePort)

	statusCode, err = sd.deleteKongRoutes(kongControlPlaneURL)
	if (err != nil) || (statusCode != http.StatusNoContent) {
		return statusCode, err
	}

	log.Info("Exiting from UnregisterKong")
	return statusCode, nil
}

func (sd *ServiceAPIDescription) deleteKongRoutes(kongControlPlaneURL string) (int, error) {
	log.Info("Entering deleteKongRoutes")

	var (
		statusCode int
		err        error
	)

	client := resty.New()

	profiles := *sd.AefProfiles
	for _, profile := range profiles {
		log.Infof("deleteKongRoutes, AefId %s", profile.AefId)
		for _, version := range profile.Versions {
			log.Infof("deleteKongRoutes, apiVersion \"%s\"", version.ApiVersion)
			for _, resource := range *version.Resources {
				statusCode, err = sd.deleteKongRoute(kongControlPlaneURL, client, resource, profile.AefId, version.ApiVersion)
				if (err != nil) || (statusCode != http.StatusNoContent) {
					return statusCode, err
				}
			}
		}
	}
	return statusCode, nil
}

func (sd *ServiceAPIDescription) deleteKongRoute(kongControlPlaneURL string, client *resty.Client, resource Resource, aefId string, apiVersion string) (int, error) {
	log.Info("Entering deleteKongRoute")
	routeName := *sd.ApiId + "_" + resource.ResourceName
	kongRoutesURL := kongControlPlaneURL + "/routes/" + routeName + "?tags=" + aefId
	log.Infof("deleteKongRoute, routeName %s, tag %s", routeName, aefId)

	// Make the DELETE request to delete the Kong route
	resp, err := client.R().Delete(kongRoutesURL)

	// Check for errors in the request
	if err != nil {
		log.Infof("Error on Kong route delete: %v", err)
		return resp.StatusCode(), err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusNoContent {
		log.Infof("Kong route %s deleted successfully!", routeName)
	} else {
		log.Infof("Error deleting Kong route. Status code: %d", resp.StatusCode())
		log.Infof("Response body: %s", resp.Body())
		return resp.StatusCode(), err
	}

	statusCode, err := sd.deleteKongService(kongControlPlaneURL, routeName, aefId)
	if (err != nil) || (statusCode != http.StatusNoContent) {
		return statusCode, err
	}
	return statusCode, err
}

func (sd *ServiceAPIDescription) deleteKongService(kongControlPlaneURL string, serviceName string, aefId string) (int, error) {
	log.Info("Entering deleteKongService")
	// Define the service information for Kong
	// Kong admin API endpoint for deleting a service
	kongServicesURL := kongControlPlaneURL + "/services/" + serviceName + "?tags=" + aefId

	// Create a new Resty client
	client := resty.New()

	// Make the DELETE request to delete the Kong service
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Delete(kongServicesURL)

	// Check for errors in the request
	if err != nil {
		log.Infof("Delete Kong Service Request Error: %v", err)
		return http.StatusInternalServerError, err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusNoContent {
		log.Infof("Kong service %s deleted successfully!", serviceName)
	} else {
		log.Infof("Error deleting Kong service. Status code: %d", resp.StatusCode())
		log.Infof("Response body: %s", resp.Body())
	}
	return resp.StatusCode(), nil
}
