// Package publishserviceapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
package publishserviceapi

import (
	externalRef0 "oransc.org/nonrtric/capifcore/internal/common"
	externalRef1 "oransc.org/nonrtric/capifcore/internal/common29122"
	externalRef2 "oransc.org/nonrtric/capifcore/internal/common29571"
)

// Defines values for CommunicationType.
const (
	CommunicationTypeREQUESTRESPONSE CommunicationType = "REQUEST_RESPONSE"

	CommunicationTypeSUBSCRIBENOTIFY CommunicationType = "SUBSCRIBE_NOTIFY"
)

// Defines values for DataFormat.
const (
	DataFormatJSON DataFormat = "JSON"
)

// Defines values for Operation.
const (
	OperationDELETE Operation = "DELETE"

	OperationGET Operation = "GET"

	OperationPATCH Operation = "PATCH"

	OperationPOST Operation = "POST"

	OperationPUT Operation = "PUT"
)

// Defines values for Protocol.
const (
	ProtocolHTTP11 Protocol = "HTTP_1_1"

	ProtocolHTTP2 Protocol = "HTTP_2"
)

// Defines values for SecurityMethod.
const (
	SecurityMethodOAUTH SecurityMethod = "OAUTH"

	SecurityMethodPKI SecurityMethod = "PKI"

	SecurityMethodPSK SecurityMethod = "PSK"
)

// The location information (e.g. civic address, GPS coordinates, data center ID) where the AEF providing the service API is located.
type AefLocation struct {
	// Indicates a Civic address.
	CivicAddr *externalRef0.CivicAddress `json:"civicAddr,omitempty"`

	// Identifies the data center where the AEF providing the service API is located.
	DcId *string `json:"dcId,omitempty"`

	// Geographic area specified by different shape.
	GeoArea *externalRef0.GeographicArea `json:"geoArea,omitempty"`
}

// Represents the AEF profile data.
type AefProfile struct {
	// Identifier of the API exposing function
	AefId string `json:"aefId"`

	// The location information (e.g. civic address, GPS coordinates, data center ID) where the AEF providing the service API is located.
	AefLocation *AefLocation `json:"aefLocation,omitempty"`

	// Possible values are:
	// - JSON: JavaScript Object Notation
	DataFormat *DataFormat `json:"dataFormat,omitempty"`

	// Domain to which API belongs to
	DomainName *string `json:"domainName,omitempty"`

	// Interface details
	InterfaceDescriptions *[]InterfaceDescription `json:"interfaceDescriptions,omitempty"`

	// Possible values are:
	// - HTTP_1_1: HTTP version 1.1
	// - HTTP_2: HTTP version 2
	Protocol *Protocol `json:"protocol,omitempty"`

	// Security methods supported by the AEF
	SecurityMethods *[]SecurityMethod `json:"securityMethods,omitempty"`

	// API version
	Versions []Version `json:"versions"`
}

// Possible values are:
// - REQUEST_RESPONSE: The communication is of the type request-response
// - SUBSCRIBE_NOTIFY: The communication is of the type subscribe-notify
type CommunicationType string

// Represents the description of a custom operation.
type CustomOperation struct {
	// Possible values are:
	// - REQUEST_RESPONSE: The communication is of the type request-response
	// - SUBSCRIBE_NOTIFY: The communication is of the type subscribe-notify
	CommType CommunicationType `json:"commType"`

	// it is set as {custOpName} part of the URI structure for a custom operation without resource association as defined in clause 5.2.4 of 3GPP TS 29.122.
	CustOpName string `json:"custOpName"`

	// Text description of the custom operation
	Description *string `json:"description,omitempty"`

	// Supported HTTP methods for the API resource. Only applicable when the protocol in AefProfile indicates HTTP.
	Operations *[]Operation `json:"operations,omitempty"`
}

// Possible values are:
// - JSON: JavaScript Object Notation
type DataFormat string

// Represents the description of an API's interface.
type InterfaceDescription struct {
	// string identifying a Ipv4 address formatted in the "dotted decimal" notation as defined in IETF RFC 1166.
	Ipv4Addr *externalRef1.Ipv4Addr `json:"ipv4Addr,omitempty"`

	// string identifying a Ipv6 address formatted according to clause 4 in IETF RFC 5952. The mixed Ipv4 Ipv6 notation according to clause 5 of IETF RFC 5952 shall not be used.
	Ipv6Addr *externalRef1.Ipv6Addr `json:"ipv6Addr,omitempty"`

	// Unsigned integer with valid values between 0 and 65535.
	Port *externalRef1.Port `json:"port,omitempty"`

	// Security methods supported by the interface, it take precedence over the security methods provided in AefProfile, for this specific interface.
	SecurityMethods *[]SecurityMethod `json:"securityMethods,omitempty"`
}

// Possible values are:
// - GET: HTTP GET method
// - POST: HTTP POST method
// - PUT: HTTP PUT method
// - PATCH: HTTP PATCH method
// - DELETE: HTTP DELETE method
type Operation string

// Possible values are:
// - HTTP_1_1: HTTP version 1.1
// - HTTP_2: HTTP version 2
type Protocol string

// Represents the published API path within the same CAPIF provider domain.
type PublishedApiPath struct {
	// A list of CCF identifiers where the service API is already published.
	CcfIds *[]string `json:"ccfIds,omitempty"`
}

// Represents the API resource data.
type Resource struct {
	// Possible values are:
	// - REQUEST_RESPONSE: The communication is of the type request-response
	// - SUBSCRIBE_NOTIFY: The communication is of the type subscribe-notify
	CommType CommunicationType `json:"commType"`

	// it is set as {custOpName} part of the URI structure for a custom operation associated with a resource as defined in clause 5.2.4 of 3GPP TS 29.122.
	CustOpName *string `json:"custOpName,omitempty"`

	// Text description of the API resource
	Description *string `json:"description,omitempty"`

	// Supported HTTP methods for the API resource. Only applicable when the protocol in AefProfile indicates HTTP.
	Operations *[]Operation `json:"operations,omitempty"`

	// Resource name
	ResourceName string `json:"resourceName"`

	// Relative URI of the API resource, it is set as {apiSpecificSuffixes} part of the URI structure as defined in clause 5.2.4 of 3GPP TS 29.122.
	Uri string `json:"uri"`
}

// Possible values are:
// - PSK: Security method 1 (Using TLS-PSK) as described in 3GPP TS 33.122
// - PKI: Security method 2 (Using PKI) as described in 3GPP TS 33.122
// - OAUTH: Security method 3 (TLS with OAuth token) as described in 3GPP TS 33.122
type SecurityMethod string

// Represents the description of a service API as published by the APF.
type ServiceAPIDescription struct {
	// AEF profile information, which includes the exposed API details (e.g. protocol).
	AefProfiles *[]AefProfile `json:"aefProfiles,omitempty"`

	// API identifier assigned by the CAPIF core function to the published service API. Shall not be present in the HTTP POST request from the API publishing function to the CAPIF core function. Shall be present in the HTTP POST response from the CAPIF core function to the API publishing function and in the HTTP GET response from the CAPIF core function to the API invoker (discovery API).
	ApiId *string `json:"apiId,omitempty"`

	// API name, it is set as {apiName} part of the URI structure as defined in clause 5.2.4 of 3GPP TS 29.122.
	ApiName string `json:"apiName"`

	// A string used to indicate the features supported by an API that is used as defined in clause  6.6 in 3GPP TS 29.500. The string shall contain a bitmask indicating supported features in  hexadecimal representation Each character in the string shall take a value of "0" to "9",  "a" to "f" or "A" to "F" and shall represent the support of 4 features as described in  table 5.2.2-3. The most significant character representing the highest-numbered features shall  appear first in the string, and the character representing features 1 to 4 shall appear last  in the string. The list of features and their numbering (starting with 1) are defined  separately for each API. If the string contains a lower number of characters than there are  defined features for an API, all features that would be represented by characters that are not  present in the string are not supported.
	ApiSuppFeats *externalRef2.SupportedFeatures `json:"apiSuppFeats,omitempty"`

	// CAPIF core function identifier.
	CcfId *string `json:"ccfId,omitempty"`

	// Text description of the API
	Description *string `json:"description,omitempty"`

	// Represents the published API path within the same CAPIF provider domain.
	PubApiPath         *PublishedApiPath `json:"pubApiPath,omitempty"`
	ServiceAPICategory *string           `json:"serviceAPICategory,omitempty"`

	// Indicates whether the service API and/or the service API category can be shared to the list of CAPIF provider domains.
	ShareableInfo *ShareableInformation `json:"shareableInfo,omitempty"`

	// A string used to indicate the features supported by an API that is used as defined in clause  6.6 in 3GPP TS 29.500. The string shall contain a bitmask indicating supported features in  hexadecimal representation Each character in the string shall take a value of "0" to "9",  "a" to "f" or "A" to "F" and shall represent the support of 4 features as described in  table 5.2.2-3. The most significant character representing the highest-numbered features shall  appear first in the string, and the character representing features 1 to 4 shall appear last  in the string. The list of features and their numbering (starting with 1) are defined  separately for each API. If the string contains a lower number of characters than there are  defined features for an API, all features that would be represented by characters that are not  present in the string are not supported.
	SupportedFeatures *externalRef2.SupportedFeatures `json:"supportedFeatures,omitempty"`
}

// Represents the parameters to request the modification of an APF published API resource.
type ServiceAPIDescriptionPatch struct {
	AefProfiles *[]AefProfile `json:"aefProfiles,omitempty"`

	// A string used to indicate the features supported by an API that is used as defined in clause  6.6 in 3GPP TS 29.500. The string shall contain a bitmask indicating supported features in  hexadecimal representation Each character in the string shall take a value of "0" to "9",  "a" to "f" or "A" to "F" and shall represent the support of 4 features as described in  table 5.2.2-3. The most significant character representing the highest-numbered features shall  appear first in the string, and the character representing features 1 to 4 shall appear last  in the string. The list of features and their numbering (starting with 1) are defined  separately for each API. If the string contains a lower number of characters than there are  defined features for an API, all features that would be represented by characters that are not  present in the string are not supported.
	ApiSuppFeats *externalRef2.SupportedFeatures `json:"apiSuppFeats,omitempty"`

	// CAPIF core function identifier.
	CcfId *string `json:"ccfId,omitempty"`

	// Text description of the API
	Description *string `json:"description,omitempty"`

	// Represents the published API path within the same CAPIF provider domain.
	PubApiPath         *PublishedApiPath `json:"pubApiPath,omitempty"`
	ServiceAPICategory *string           `json:"serviceAPICategory,omitempty"`

	// Indicates whether the service API and/or the service API category can be shared to the list of CAPIF provider domains.
	ShareableInfo *ShareableInformation `json:"shareableInfo,omitempty"`
}

// Indicates whether the service API and/or the service API category can be shared to the list of CAPIF provider domains.
type ShareableInformation struct {
	// List of CAPIF provider domains to which the service API information to be shared.
	CapifProvDoms *[]string `json:"capifProvDoms,omitempty"`

	// Set to "true" indicates that the service API and/or the service API category can be shared to the list of CAPIF provider domain information. Otherwise set to "false".
	IsShareable bool `json:"isShareable"`
}

// Represents the API version information.
type Version struct {
	// API major version in URI (e.g. v1)
	ApiVersion string `json:"apiVersion"`

	// Custom operations without resource association.
	CustOperations *[]CustomOperation `json:"custOperations,omitempty"`

	// string with format "date-time" as defined in OpenAPI.
	Expiry *externalRef1.DateTime `json:"expiry,omitempty"`

	// Resources supported by the API.
	Resources *[]Resource `json:"resources,omitempty"`
}

// PostApfIdServiceApisJSONBody defines parameters for PostApfIdServiceApis.
type PostApfIdServiceApisJSONBody ServiceAPIDescription

// PutApfIdServiceApisServiceApiIdJSONBody defines parameters for PutApfIdServiceApisServiceApiId.
type PutApfIdServiceApisServiceApiIdJSONBody ServiceAPIDescription

// PostApfIdServiceApisJSONRequestBody defines body for PostApfIdServiceApis for application/json ContentType.
type PostApfIdServiceApisJSONRequestBody PostApfIdServiceApisJSONBody

// PutApfIdServiceApisServiceApiIdJSONRequestBody defines body for PutApfIdServiceApisServiceApiId for application/json ContentType.
type PutApfIdServiceApisServiceApiIdJSONRequestBody PutApfIdServiceApisServiceApiIdJSONBody
