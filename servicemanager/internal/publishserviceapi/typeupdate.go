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
	"net/url"
	"strings"

	resty "github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	common29122 "oransc.org/nonrtric/servicemanager/internal/common29122"
)

func (sd *ServiceAPIDescription) PrepareNewService() {
	apiName := "api_id_" + strings.ReplaceAll(sd.ApiName, " ", "_")
	sd.ApiId = &apiName
}

func (sd *ServiceAPIDescription) RegisterKong(kongDomain string,
		kongProtocol string,
		kongIPv4 common29122.Ipv4Addr,
		kongDataPlanePort common29122.Port,
		kongControlPlanePort common29122.Port,
		apfId string) (int, error) {

	log.Trace("entering RegisterKong")
	var (
		statusCode int
		err        error
	)
	kongControlPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongIPv4, kongControlPlanePort)

	statusCode, err = sd.createKongRoutes(kongControlPlaneURL, apfId)
	if (err != nil) || (statusCode != http.StatusCreated) {
		return statusCode, err
	}

	sd.updateInterfaceDescription(kongIPv4, kongDataPlanePort, kongDomain)

	log.Trace("exiting from RegisterKong")
	return statusCode, nil
}

func (sd *ServiceAPIDescription) createKongRoutes(kongControlPlaneURL string, apfId string) (int, error) {
	log.Trace("entering createKongRoutes")
	var (
		statusCode int
		err        error
	)

	client := resty.New()

	profiles := *sd.AefProfiles
	for _, profile := range profiles {
		log.Debugf("createKongRoutes, AefId %s", profile.AefId)
		for _, version := range profile.Versions {
			log.Debugf("createKongRoutes, apiVersion \"%s\"", version.ApiVersion)
			for _, resource := range *version.Resources {
				statusCode, err = sd.createKongRoute(kongControlPlaneURL, client, resource, apfId, profile.AefId, version.ApiVersion)
				if (err != nil) || (statusCode != http.StatusCreated) {
					return statusCode, err
				}
			}
		}
	}
	return statusCode, nil
}

func (sd *ServiceAPIDescription) createKongRoute(
		kongControlPlaneURL string,
		client *resty.Client,
		resource Resource,
		apfId string,
		aefId string,
		apiVersion string ) (int, error) {
	log.Trace("entering createKongRoute")

	resourceName := resource.ResourceName
	apiId := *sd.ApiId

	tags := buildTags(apfId, aefId, apiId, apiVersion, resourceName)
	log.Debugf("createKongRoute, tags %s", tags)

	serviceName := apiId + "_" + resourceName
	routeName := serviceName

	log.Debugf("createKongRoute, serviceName %s", serviceName)
	log.Debugf("createKongRoute, routeName %s", routeName)
	log.Debugf("createKongRoute, aefId %s", aefId)

	uri := buildUriWithVersion(apiVersion, resource.Uri)
	log.Debugf("createKongRoute, uri %s", uri)

	statusCode, err := sd.createKongService(kongControlPlaneURL, serviceName, uri, tags)
	if (err != nil) || (statusCode != http.StatusCreated) {
		return statusCode, err
	}

	kongRoutesURL := kongControlPlaneURL + "/services/" + serviceName + "/routes"

	// Define the route information for Kong
	kongRouteInfo := map[string]interface{}{
		"name":       routeName,
		"paths":      []string{uri},
		"methods":    resource.Operations,
		"tags":       tags,
		"strip_path": true,
	}

	// Make the POST request to create the Kong service
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(kongRouteInfo).
		Post(kongRoutesURL)

	// Check for errors in the request
	if err != nil {
		log.Debugf("createKongRoute POST Error: %v", err)
		return resp.StatusCode(), err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusCreated {
		log.Infof("kong route %s created successfully", routeName)
	} else {
		log.Debugf("kongRoutesURL %s", kongRoutesURL)
		err = fmt.Errorf("error creating Kong route. Status code: %d", resp.StatusCode())
		log.Error(err.Error())
		log.Errorf("response body: %s", resp.Body())
		return resp.StatusCode(), err
	}

	return resp.StatusCode(), nil
}

func buildUriWithVersion(apiVersion string, uri string) string {
	if apiVersion != "" {
		if apiVersion[0] != '/' {
			apiVersion = "/" + apiVersion
		}
		if apiVersion[len(apiVersion)-1] != '/' && uri[0] != '/' {
			apiVersion = apiVersion + "/"
		}
		uri = apiVersion + uri
	}
	return uri
}

func buildTags(apfId string, aefId string, apiId string, apiVersion string, resourceName string) []string  {
	tagsMap := map[string]string{
		"apfId": apfId,
		"aefId": aefId,
		"apiId": apiId,
		"apiVersion": apiVersion,
		"resourceName": resourceName,
	}

	// Convert the map to a slice of strings
	var tagsSlice []string
	for key, value := range tagsMap {
		str := fmt.Sprintf("%s: %s", key, value)
		tagsSlice = append(tagsSlice, str)
	}

	return tagsSlice
}

func (sd *ServiceAPIDescription) createKongService(kongControlPlaneURL string, kongServiceName string, kongServiceUri string, tags []string) (int, error) {
	log.Tracef("entering createKongService")
	log.Tracef("createKongService, kongServiceName %s", kongServiceName)

	// Define the service information for Kong
	firstAEFProfileIpv4Addr, firstAEFProfilePort, err := sd.findFirstAEFProfile()
	if err != nil {
		return http.StatusBadRequest, err
	}

	kongControlPlaneURLParsed, err := url.Parse(kongControlPlaneURL)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	log.Debugf("kongControlPlaneURL %s", kongControlPlaneURL)
	log.Debugf("kongControlPlaneURLParsed.Scheme %s", kongControlPlaneURLParsed.Scheme)

	kongServiceInfo := map[string]interface{}{
		"host":     firstAEFProfileIpv4Addr,
		"name":     kongServiceName,
		"port":     firstAEFProfilePort,
		"protocol": kongControlPlaneURLParsed.Scheme,
		"path":     kongServiceUri,
		"tags":     tags,
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
		log.Errorf("create Kong Service Request Error: %v", err)
		return http.StatusInternalServerError, err
	}

	// Check the response status code
	statusCode := resp.StatusCode()
	if statusCode == http.StatusCreated {
		log.Infof("kong service %s created successfully", kongServiceName)
	} else if resp.StatusCode() == http.StatusConflict {
		log.Errorf("kong service already exists. Status code: %d", resp.StatusCode())
		err = fmt.Errorf("service with identical apiName is already published") // for compatibilty with Capif error message on a duplicate service
		statusCode = http.StatusForbidden                                       // for compatibilty with the spec, TS29222_CAPIF_Publish_Service_API
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
	log.Tracef("entering findFirstAEFProfile")
	var aefProfile AefProfile
	if *sd.AefProfiles != nil {
		aefProfile = (*sd.AefProfiles)[0]
	}
	if (*sd.AefProfiles == nil) || (aefProfile.InterfaceDescriptions == nil) {
		err := errors.New("cannot read interfaceDescription")
		log.Errorf(err.Error())
		return "", common29122.Port(0), err
	}

	interfaceDescription := (*aefProfile.InterfaceDescriptions)[0]
	firstIpv4Addr := *interfaceDescription.Ipv4Addr
	firstPort := *interfaceDescription.Port

	log.Debugf("findFirstAEFProfile firstIpv4Addr %s firstPort %d", firstIpv4Addr, firstPort)

	return firstIpv4Addr, firstPort, nil
}

// Update our exposures to point to Kong by replacing in incoming interface description with Kong interface descriptions.
func (sd *ServiceAPIDescription) updateInterfaceDescription(kongIPv4 common29122.Ipv4Addr, kongDataPlanePort common29122.Port, kongDomain string) {
	log.Trace("updating InterfaceDescriptions")
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
	log.Trace("entering UnregisterKong")

	var (
		statusCode int
		err        error
	)
	kongControlPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongIPv4, kongControlPlanePort)

	statusCode, err = sd.deleteKongRoutes(kongControlPlaneURL)
	if (err != nil) || (statusCode != http.StatusNoContent) {
		return statusCode, err
	}

	log.Trace("exiting from UnregisterKong")
	return statusCode, nil
}

func (sd *ServiceAPIDescription) deleteKongRoutes(kongControlPlaneURL string) (int, error) {
	log.Trace("entering deleteKongRoutes")

	var (
		statusCode int
		err        error
	)

	client := resty.New()

	profiles := *sd.AefProfiles
	for _, profile := range profiles {
		log.Debugf("deleteKongRoutes, AefId %s", profile.AefId)
		for _, version := range profile.Versions {
			log.Debugf("deleteKongRoutes, apiVersion \"%s\"", version.ApiVersion)
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
	log.Trace("entering deleteKongRoute")
	routeName := *sd.ApiId + "_" + resource.ResourceName
	kongRoutesURL := kongControlPlaneURL + "/routes/" + routeName + "?tags=" + aefId
	log.Debugf("deleteKongRoute, routeName %s, tag %s", routeName, aefId)

	// Make the DELETE request to delete the Kong route
	resp, err := client.R().Delete(kongRoutesURL)

	// Check for errors in the request
	if err != nil {
		log.Errorf("error on Kong route delete: %v", err)
		return resp.StatusCode(), err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusNoContent {
		log.Infof("kong route %s deleted successfully", routeName)
	} else {
		log.Debugf("kongRoutesURL: %s", kongRoutesURL)
		log.Errorf("error deleting Kong route. Status code: %d", resp.StatusCode())
		log.Errorf("response body: %s", resp.Body())
		return resp.StatusCode(), err
	}

	statusCode, err := sd.deleteKongService(kongControlPlaneURL, routeName, aefId)
	if (err != nil) || (statusCode != http.StatusNoContent) {
		return statusCode, err
	}
	return statusCode, err
}

func (sd *ServiceAPIDescription) deleteKongService(kongControlPlaneURL string, serviceName string, aefId string) (int, error) {
	log.Trace("entering deleteKongService")
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
		log.Errorf("delete kong service request: %v", err)
		return http.StatusInternalServerError, err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusNoContent {
		log.Infof("kong service %s deleted successfully", serviceName)
	} else {
		log.Debugf("kongServicesURL: %s", kongServicesURL)
		log.Errorf("deleting Kong service, status code: %d", resp.StatusCode())
		log.Errorf("response body: %s", resp.Body())
	}
	return resp.StatusCode(), nil
}
