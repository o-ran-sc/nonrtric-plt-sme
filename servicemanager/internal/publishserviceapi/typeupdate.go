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
	"regexp"
	"strconv"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	common29122 "oransc.org/nonrtric/servicemanager/internal/common29122"
)

func (sd *ServiceAPIDescription) PrepareNewService() {
	apiName := "api_id_" + strings.ReplaceAll(sd.ApiName, " ", "_")
	sd.ApiId = &apiName
}

func (sd *ServiceAPIDescription) RegisterKong(
	kongDomain string,
	kongProtocol string,
	kongControlPlaneIPv4 common29122.Ipv4Addr,
	kongControlPlanePort common29122.Port,
	kongDataPlaneIPv4 common29122.Ipv4Addr,
	kongDataPlanePort common29122.Port,
	apfId string) (int, error) {

	log.Trace("entering RegisterKong")
	log.Debugf("RegisterKong kongDataPlaneIPv4 %s", kongDataPlaneIPv4)

	var (
		statusCode int
		err        error
	)
	kongControlPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongControlPlaneIPv4, kongControlPlanePort)

	statusCode, err = sd.createKongInterfaceDescriptions(kongControlPlaneURL, apfId)
	if (err != nil) || (statusCode != http.StatusCreated) {
		return statusCode, err
	}

	sd.updateInterfaceDescription(kongDataPlaneIPv4, kongDataPlanePort, kongDomain)

	log.Trace("exiting from RegisterKong")
	return statusCode, nil
}

func (sd *ServiceAPIDescription) createKongInterfaceDescriptions(kongControlPlaneURL string, apfId string) (int, error) {
	log.Trace("entering createKongInterfaceDescriptions")

	var (
		statusCode int
		err        error
	)
	client := resty.New()
	outputUris := []string{}

	if sd == nil {
		err = errors.New("cannot read ServiceAPIDescription")
		log.Errorf(err.Error())
		return http.StatusBadRequest, err
	}

	if (sd.AefProfiles == nil) || (len(*sd.AefProfiles) < 1) {
		err = errors.New("cannot read AefProfiles")
		log.Errorf(err.Error())
		return http.StatusBadRequest, err
	}

	profiles := *sd.AefProfiles
	for _, profile := range profiles {
		log.Debugf("createKongInterfaceDescriptions, AefId %s", profile.AefId)

		if (profile.Versions == nil) || (len(profile.Versions) < 1) {
			err := errors.New("cannot read Versions")
			log.Errorf(err.Error())
			return http.StatusBadRequest, err
		}

		for _, version := range profile.Versions {
			log.Debugf("createKongInterfaceDescriptions, apiVersion \"%s\"", version.ApiVersion)

			if (profile.InterfaceDescriptions == nil) || (len(*profile.InterfaceDescriptions) < 1) {
				err := errors.New("cannot read InterfaceDescriptions")
				log.Errorf(err.Error())
				return http.StatusBadRequest, err
			}

			for _, interfaceDescription := range *profile.InterfaceDescriptions {
				log.Debugf("createKongInterfaceDescriptions, Ipv4Addr %s", *interfaceDescription.Ipv4Addr)
				log.Debugf("createKongInterfaceDescriptions, Port %d", *interfaceDescription.Port)
				if uint(*interfaceDescription.Port) > 65535 {
					err := errors.New("invalid Port")
					log.Errorf(err.Error())
					return http.StatusBadRequest, err
				}

				if interfaceDescription.SecurityMethods == nil {
					log.Debugf("createKongInterfaceDescriptions, SecurityMethods: null")
				} else if len(*interfaceDescription.SecurityMethods) < 1 {
					err := errors.New("cannot read any SecurityMethod")
					log.Errorf(err.Error())
					return http.StatusBadRequest, err
				} else {
					for _, securityMethod := range *interfaceDescription.SecurityMethods {
						log.Debugf("createKongInterfaceDescriptions, SecurityMethod %s", securityMethod)

						if (securityMethod != SecurityMethodOAUTH) && (securityMethod != SecurityMethodPKI) && (securityMethod != SecurityMethodPSK) {
							msg := fmt.Sprintf("invalid SecurityMethod %s", securityMethod)
							err := errors.New(msg)
							log.Errorf(err.Error())
							return http.StatusBadRequest, err
						}
					}
				}

				if (version.Resources == nil) || (len(*version.Resources) < 1) {
					err := errors.New("cannot read Resources")
					log.Errorf(err.Error())
					return http.StatusBadRequest, err
				}

				for _, resource := range *version.Resources {
					var specUri string
					specUri, statusCode, err = sd.createKongServiceRoutePrecheck(kongControlPlaneURL, client, interfaceDescription, resource, apfId, profile.AefId, version.ApiVersion)
					if (err != nil) || (statusCode != http.StatusCreated) {
						return statusCode, err
					}
					log.Debugf("createKongInterfaceDescriptions, specUri %s", specUri)
					outputUris = append(outputUris, specUri)
					log.Tracef("createKongInterfaceDescriptions, len(outputUris) %d", len(outputUris))
					log.Tracef("createKongInterfaceDescriptions, outputUris %v", outputUris)
				}
			}
		}
	}

	// Our list of returned resources has the new resource with the hash code and version number
	m := 0
	for i, profile := range profiles {
		for j, version := range profile.Versions {
			var newResources []Resource
			for range *profile.InterfaceDescriptions {
				log.Tracef("createKongInterfaceDescriptions, range over *profile.InterfaceDescriptions")
				for _, resource := range *version.Resources {
					log.Tracef("createKongInterfaceDescriptions, m %d outputUris[m] %s", m, outputUris[m])
					resource.Uri = outputUris[m]
					m = m + 1
					// Build a new list of resources with updated uris
					newResources = append(newResources, resource)
					log.Tracef("createKongInterfaceDescriptions, newResources %v", newResources)
				}
			}
			// Swap over to the new list of uris
			*profiles[i].Versions[j].Resources = newResources
			log.Tracef("createKongInterfaceDescriptions, assigned *profiles[i].Versions[j].Resources %v", *profiles[i].Versions[j].Resources)
		}
	}
	log.Tracef("exiting createKongInterfaceDescriptions statusCode %d", statusCode)

	return statusCode, nil
}

func (sd *ServiceAPIDescription) createKongServiceRoutePrecheck(
	kongControlPlaneURL string,
	client *resty.Client,
	interfaceDescription InterfaceDescription,
	resource Resource,
	apfId string,
	aefId string,
	apiVersion string) (string, int, error) {
	log.Trace("entering createKongServiceRoutePrecheck")
	log.Debugf("createKongServiceRoutePrecheck, aefId %s", aefId)

	if (resource.Operations == nil) || (len(*resource.Operations) < 1) {
		err := errors.New("cannot read Resource.Operations")
		log.Errorf(err.Error())
		return "", http.StatusBadRequest, err
	}

	log.Debugf("createKongServiceRoutePrecheck, resource.Uri %s", resource.Uri)
	if resource.Uri == "" {
		err := errors.New("cannot read Resource.Uri")
		log.Errorf(err.Error())
		return "", http.StatusBadRequest, err
	}

	log.Debugf("createKongServiceRoutePrecheck, ResourceName %v", resource.ResourceName)

	if resource.ResourceName == "" {
		err := errors.New("cannot read Resource.ResourceName")
		log.Errorf(err.Error())
		return "", http.StatusBadRequest, err
	}

	if (resource.CommType != CommunicationTypeREQUESTRESPONSE) && (resource.CommType != CommunicationTypeSUBSCRIBENOTIFY) {
		err := errors.New("invalid Resource.CommType")
		log.Errorf(err.Error())
		return "", http.StatusBadRequest, err
	}

	specUri := resource.Uri
	kongRegexUri, _ := deriveKongPattern(resource.Uri)

	specUri, statusCode, err := sd.createKongServiceRoute(kongControlPlaneURL, client, interfaceDescription, kongRegexUri, specUri, apfId, aefId, apiVersion, resource)
	if (err != nil) || (statusCode != http.StatusCreated) {
		// We carry on if we tried to create a duplicate service. We depend on Kong route matching.
		return specUri, statusCode, err
	}

	return specUri, statusCode, err
}

func insertVersion(version string, route string) string {
	versionedRoute := route

	if version != "" {
		sep := "/"
		n := 3

		foundRegEx := false
		if strings.HasPrefix(route, "~") {
			log.Debug("insertVersion, found regex prefix")
			foundRegEx = true
			route = strings.TrimPrefix(route, "~")
		}

		log.Debugf("insertVersion route %s", route)
		split := strings.SplitAfterN(route, sep, n)
		log.Debugf("insertVersion split %q", split)

		versionedRoute = split[0]
		if len(split) == 2 {
			versionedRoute = split[0] + split[1]
		} else if len(split) > 2 {
			versionedRoute = split[0] + split[1] + version + sep + split[2]
		}

		if foundRegEx {
			versionedRoute = "~" + versionedRoute
		}
	}
	log.Debugf("insertVersion versionedRoute %s", versionedRoute)

	return versionedRoute
}

func (sd *ServiceAPIDescription) createKongServiceRoute(
	kongControlPlaneURL string,
	client *resty.Client,
	interfaceDescription InterfaceDescription,
	kongRegexUri string,
	specUri string,
	apfId string,
	aefId string,
	apiVersion string,
	resource Resource) (string, int, error) {
	log.Tracef("entering createKongServiceRoute")

	var (
		statusCode int
		err        error
	)

	kongControlPlaneURLParsed, err := url.Parse(kongControlPlaneURL)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	log.Debugf("createKongServiceRoute, kongControlPlaneURL %s", kongControlPlaneURL)
	log.Debugf("createKongServiceRoute, kongControlPlaneURLParsed.Scheme %s", kongControlPlaneURLParsed.Scheme)

	log.Debugf("createKongServiceRoute, kongRegexUri %s", kongRegexUri)
	log.Debugf("createKongServiceRoute, specUri %s", specUri)

	kongRegexUri = insertVersion(apiVersion, kongRegexUri)
	kongServiceUri := kongRegexUri
	log.Debugf("createKongServiceRoute, kongServiceUri after insertVersion, %s", kongServiceUri)

	specUri = insertVersion(apiVersion, specUri)
	log.Debugf("createKongServiceRoute, specUri after insertVersion, %s", specUri)

	if strings.HasPrefix(kongServiceUri, "~") {
		log.Debug("createKongServiceRoute, found regex prefix")

		// For our Kong Service path, we omit the leading ~ and take the path up to the regex, not including the '('
		kongServiceUri = kongServiceUri[1:]
		index := strings.Index(kongServiceUri, "(?")
		if index != -1 {
			kongServiceUri = kongServiceUri[:index]
		} else {
			log.Errorf("createKongServiceRoute, regex characters '(?' not found in the regex %s", kongServiceUri)
			return "", http.StatusBadRequest, err
		}
	} else {
		log.Debug("createKongServiceRoute, no regex prefix found")
	}
	log.Debugf("createKongServiceRoute, kongServiceUri, path up to regex %s", kongServiceUri)

	ipv4Addr := *interfaceDescription.Ipv4Addr
	port := *interfaceDescription.Port

	portAsInt := int(port)
	interfaceDescriptionSeed := string(ipv4Addr) + strconv.Itoa(portAsInt)
	interfaceDescUuid := uuid.NewSHA1(uuid.NameSpaceURL, []byte(interfaceDescriptionSeed))
	uriPrefix := "port-" + strconv.Itoa(portAsInt) + "-hash-" + interfaceDescUuid.String()

	resourceName := resource.ResourceName

	apiId := *sd.ApiId
	kongServiceName := apiId + "-" + resourceName
	kongServiceNamePrefix := kongServiceName + "-" + uriPrefix

	log.Debugf("createKongServiceRoute, kongServiceName %s", kongServiceName)
	log.Debugf("createKongServiceRoute, kongServiceNamePrefix %s", kongServiceNamePrefix)

	tags := buildTags(apfId, aefId, apiId, apiVersion, resourceName)
	log.Debugf("createKongServiceRoute, tags %s", tags)

	kongServiceInfo := map[string]interface{}{
		"host":     ipv4Addr,
		"name":     kongServiceNamePrefix,
		"port":     port,
		"protocol": kongControlPlaneURLParsed.Scheme,
		"path":     kongServiceUri,
		"tags":     tags,
	}

	// Kong admin API endpoint for creating a service
	kongServicesURL := kongControlPlaneURL + "/services"

	// Make the POST request to create the Kong service
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(kongServiceInfo).
		Post(kongServicesURL)

	// Check for errors in the request
	if err != nil {
		log.Errorf("createKongServiceRoute, Request Error: %v", err)
		return "", http.StatusInternalServerError, err
	}

	// Check the response status code
	statusCode = resp.StatusCode()
	if statusCode == http.StatusCreated {
		log.Infof("kong service %s created successfully", kongServiceNamePrefix)
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
		return "", statusCode, err
	}

	// Create matching route
	routeName := kongServiceNamePrefix

	kongRouteUri := prependUri(uriPrefix, kongRegexUri)
	log.Debugf("createKongServiceRoute, kongRouteUri with uriPrefix %s", kongRouteUri)

	kongRouteUri = prependUri(sd.ApiName, kongRouteUri)
	log.Debugf("createKongServiceRoute, kongRouteUri with apiName %s", kongRouteUri)

	specUri = prependUri(uriPrefix, specUri)
	log.Debugf("createKongServiceRoute, specUri with uriPrefix %s", specUri)

	specUri = prependUri(sd.ApiName, specUri)
	log.Debugf("createKongServiceRoute, specUri with apiName %s", specUri)

	statusCode, err = sd.createRouteForService(kongControlPlaneURL, client, resource, routeName, kongRouteUri, kongRegexUri, tags)
	if err != nil {
		log.Errorf(err.Error())
		return kongRouteUri, statusCode, err
	}

	return specUri, statusCode, err
}

func buildTags(apfId string, aefId string, apiId string, apiVersion string, resourceName string) []string {
	tagsMap := map[string]string{
		"apfId":        apfId,
		"aefId":        aefId,
		"apiId":        apiId,
		"apiVersion":   apiVersion,
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

func prependUri(prependUri string, uri string) string {
	if prependUri != "" {
		trimmedUri := uri
		foundRegEx := false
		if strings.HasPrefix(uri, "~") {
			log.Debug("prependUri, found regex prefix")
			foundRegEx = true
			trimmedUri = strings.TrimPrefix(uri, "~")
			log.Debugf("prependUri, TrimPrefix trimmedUri %s", trimmedUri)
		}

		if prependUri[0] != '/' {
			prependUri = "/" + prependUri
		}
		if prependUri[len(prependUri)-1] != '/' && trimmedUri[0] != '/' {
			prependUri = prependUri + "/"
		}
		uri = prependUri + trimmedUri
		if foundRegEx {
			uri = "~" + uri
		}
	}
	return uri
}

func (sd *ServiceAPIDescription) createRouteForService(
	kongControlPlaneURL string,
	client *resty.Client,
	resource Resource,
	routeName string,
	kongRouteUri string,
	kongRegexUri string,
	tags []string) (int, error) {

	log.Debugf("createRouteForService, kongRouteUri %s", kongRouteUri)

	// Create a url.Values map to hold the form data
	data := url.Values{}
	data.Set("strip_path", "true")
	log.Debugf("createRouteForService, strip_path %s", data.Get("strip_path"))
	data.Set("name", routeName)

	routeUriPaths := []string{kongRouteUri}
	for _, path := range routeUriPaths {
		log.Debugf("createRouteForService, path %s", path)
		data.Add("paths", path)
	}

	for _, tag := range tags {
		log.Debugf("createRouteForService, tag %s", tag)
		data.Add("tags", tag)
	}

	for _, op := range *resource.Operations {
		log.Debugf("createRouteForService, op %s", string(op))
		data.Add("methods", string(op))
	}

	// Encode the data to application/x-www-form-urlencoded format
	encodedData := data.Encode()

	// Make the POST request to create the Kong service
	serviceName := routeName
	kongRoutesURL := kongControlPlaneURL + "/services/" + serviceName + "/routes"
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(strings.NewReader(encodedData)).
		Post(kongRoutesURL)

	// Check for errors in the request
	if err != nil {
		log.Debugf("createRouteForService POST Error: %v", err)
		return resp.StatusCode(), err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusCreated {
		log.Infof("kong route %s created successfully", routeName)

		index := strings.Index(kongRegexUri, "(?")
		if index != -1 {
			log.Debugf("createRouteForService, found regex in %s", kongRegexUri)
			requestTransformerUri := strings.TrimPrefix(kongRegexUri, "~")
			log.Debugf("createRouteForService, requestTransformerUri %s", requestTransformerUri)

			statusCode, err := sd.createRequestTransformer(kongControlPlaneURL, client, routeName, requestTransformerUri)
			if (err != nil) || ((statusCode != http.StatusCreated) && (statusCode != http.StatusForbidden)) {
				return statusCode, err
			}
		} else {
			log.Debug("createRouteForService, no variable name found")
		}
	} else {
		log.Debugf("kongRoutesURL %s", kongRoutesURL)
		err = fmt.Errorf("error creating Kong route. Status code: %d", resp.StatusCode())
		log.Error(err.Error())
		log.Errorf("response body: %s", resp.Body())
		return resp.StatusCode(), err
	}

	return resp.StatusCode(), nil
}

func (sd *ServiceAPIDescription) createRequestTransformer(
	kongControlPlaneURL string,
	client *resty.Client,
	routeName string,
	routePattern string) (int, error) {

	log.Trace("entering createRequestTransformer")

	// Make the POST request to create the Kong Request Transformer
	kongRequestTransformerURL := kongControlPlaneURL + "/routes/" + routeName + "/plugins"

	transformPattern, _ := deriveTransformPattern(routePattern)

	// Create the form data
	formData := url.Values{
		"name":               {"request-transformer"},
		"config.replace.uri": {transformPattern},
	}
	encodedData := formData.Encode()

	// Create a new HTTP POST request
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetBody(strings.NewReader(encodedData)).
		Post(kongRequestTransformerURL)

	// Check for errors in the request
	if err != nil {
		log.Debugf("createRequestTransformer POST Error: %v", err)
		return resp.StatusCode(), err
	}

	// Check the response status code
	if resp.StatusCode() == http.StatusCreated {
		log.Infof("kong request transformer for route %s created successfully", routeName)
	} else {
		log.Debugf("kongRequestTransformerURL %s", kongRequestTransformerURL)
		err = fmt.Errorf("error creating Kong request transformer. Status code: %d", resp.StatusCode())
		log.Error(err.Error())
		log.Errorf("response body: %s", resp.Body())
		return resp.StatusCode(), err
	}

	return resp.StatusCode(), nil
}

// Function to derive the transform pattern from the route pattern
func deriveKongPattern(routePattern string) (string, error) {
	log.Trace("entering deriveKongPattern")
	log.Debugf("deriveKongPattern routePattern %s", routePattern)

	// Regular expression to match variable names
	re := regexp.MustCompile(`\{([a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*)\}`)
	log.Debugf("deriveKongPattern MustCompile %v", re)

	// Find all matches in the route pattern
	matches := re.FindAllStringSubmatch(routePattern, -1)
	log.Debugf("deriveKongPattern FindAllStringSubmatch %v", re)

	transformPattern := routePattern
	for _, match := range matches {
		// match[0] is the full match with braces
		// match[1] is the uri variable name
		log.Debugf("deriveKongPattern match %v", match)
		log.Debugf("deriveKongPattern match[0] %v", match[0])
		log.Debugf("deriveKongPattern match[1] %v", match[1])
		placeholder := fmt.Sprintf("(?<%s>[a-zA-Z0-9]+([-_][a-zA-Z0-9]+)*)", match[1])
		// Replace the variable with the Kong regex placeholder
		transformPattern = strings.Replace(transformPattern, match[0], placeholder, 1)
	}
	log.Debugf("deriveKongPattern transformPattern %s", transformPattern)

	if len(matches) != 0 {
		transformPattern = "~" + transformPattern
		log.Debugf("deriveKongPattern transformPattern with prefix %s", transformPattern)
	}

	return transformPattern, nil
}

// Function to derive the transform pattern from the route pattern
func deriveTransformPattern(routePattern string) (string, error) {
	log.Trace("entering deriveTransformPattern")
	log.Debugf("deriveTransformPattern routePattern %s", routePattern)

	// Append a slash to handle an edge case for matching a trailing capture group.
	appendedSlash := false
	if routePattern[len(routePattern)-1] != '/' {
		routePattern = routePattern + "/"
		appendedSlash = true
		log.Debugf("deriveTransformPattern, append / routePattern %s", routePattern)
	}

	// Regular expression to match named capture groups
	re := regexp.MustCompile(`\(\?<([^>]+)>([^\/]+)`)
	// Find all matches in the route pattern
	matches := re.FindAllStringSubmatch(routePattern, -1)

	transformPattern := routePattern
	for _, match := range matches {
		// match[0] is the full match, match[1] is the capture group name, match[2] is the pattern
		placeholder := fmt.Sprintf("$(uri_captures[\"%s\"])", match[1])
		// Replace the capture group with the corresponding placeholder
		transformPattern = strings.Replace(transformPattern, match[0], placeholder, 1)
	}
	log.Debugf("deriveTransformPattern transformPattern %s", transformPattern)

	if appendedSlash {
		transformPattern = strings.TrimSuffix(transformPattern, "/")
		log.Debugf("deriveTransformPattern, remove / transformPattern %s", transformPattern)
	}

	return transformPattern, nil
}

// Update our exposures to point to Kong by replacing in incoming interface description with Kong interface descriptions.
func (sd *ServiceAPIDescription) updateInterfaceDescription(kongDataPlaneIPv4 common29122.Ipv4Addr, kongDataPlanePort common29122.Port, kongDomain string) {
	log.Trace("updating InterfaceDescriptions")
	log.Debugf("InterfaceDescriptions kongDataPlaneIPv4 %s", kongDataPlaneIPv4)

	interfaceDesc := InterfaceDescription{
		Ipv4Addr: &kongDataPlaneIPv4,
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

func (sd *ServiceAPIDescription) UnregisterKong(kongDomain string, kongProtocol string, kongControlPlaneIPv4 common29122.Ipv4Addr, kongControlPlanePort common29122.Port) (int, error) {
	log.Trace("entering UnregisterKong")

	var (
		statusCode int
		err        error
	)
	kongControlPlaneURL := fmt.Sprintf("%s://%s:%d", kongProtocol, kongControlPlaneIPv4, kongControlPlanePort)

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
