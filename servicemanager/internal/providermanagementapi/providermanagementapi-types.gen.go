// Package providermanagementapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.10.1 DO NOT EDIT.
package providermanagementapi

import (
	externalRef1 "oransc.org/nonrtric/servicemanager/internal/common29571"
)

// Defines values for ApiProviderFuncRole.
const (
	ApiProviderFuncRoleAEF ApiProviderFuncRole = "AEF"

	ApiProviderFuncRoleAMF ApiProviderFuncRole = "AMF"

	ApiProviderFuncRoleAPF ApiProviderFuncRole = "APF"
)

// Represents an API provider domain's enrolment details.
type APIProviderEnrolmentDetails struct {
	// API provider domain ID assigned by the CAPIF core function to the API management function while registering the API provider domain. Shall not be present in the HTTP POST request from the API Management function to the CAPIF core function, to on-board itself. Shall be present in all other HTTP requests and responses.
	ApiProvDomId *string `json:"apiProvDomId,omitempty"`

	// Generic information related to the API provider domain such as details of the API provider applications.
	ApiProvDomInfo *string `json:"apiProvDomInfo,omitempty"`

	// A list of individual API provider domain functions details. When included by the API management function in the HTTP request message, it lists the API provider domain functions that the API management function intends to register/update in registration or update registration procedure. When included by the CAPIF core function in the HTTP response message, it lists the API domain functions details that are registered or updated successfully.
	ApiProvFuncs *[]APIProviderFunctionDetails `json:"apiProvFuncs,omitempty"`

	// Registration or update specific failure information of failed API provider domain function registrations.Shall be present in the HTTP response body if atleast one of the API provider domain function registration or update registration fails.
	FailReason *string `json:"failReason,omitempty"`

	// Security information necessary for the CAPIF core function to validate the registration of the API provider domain. Shall be present in HTTP POST request from API management function to CAPIF core function for API provider domain registration.
	RegSec string `json:"regSec"`

	// A string used to indicate the features supported by an API that is used as defined in clause  6.6 in 3GPP TS 29.500. The string shall contain a bitmask indicating supported features in  hexadecimal representation Each character in the string shall take a value of "0" to "9",  "a" to "f" or "A" to "F" and shall represent the support of 4 features as described in  table 5.2.2-3. The most significant character representing the highest-numbered features shall  appear first in the string, and the character representing features 1 to 4 shall appear last  in the string. The list of features and their numbering (starting with 1) are defined  separately for each API. If the string contains a lower number of characters than there are  defined features for an API, all features that would be represented by characters that are not  present in the string are not supported.
	SuppFeat *externalRef1.SupportedFeatures `json:"suppFeat,omitempty"`
}

// Represents an API provider domain's enrolment details.
type APIProviderEnrolmentDetailsPatch struct {
	// Generic information related to the API provider domain such as details of the API provider applications.
	ApiProvDomInfo *string `json:"apiProvDomInfo,omitempty"`

	// A list of individual API provider domain functions details. When included by the API management function in the HTTP request message, it lists the API provider domain functions that the API management function intends to register/update in registration or update registration procedure.
	ApiProvFuncs *[]APIProviderFunctionDetails `json:"apiProvFuncs,omitempty"`
}

// Represents API provider domain function's details.
type APIProviderFunctionDetails struct {
	// API provider domain functionID assigned by the CAPIF core function to the API provider domain function while registering/updating the API provider domain. Shall not be present in the HTTP POST request from the API management function to the CAPIF core function, to register itself. Shall be present in all other HTTP requests and responses.
	ApiProvFuncId *string `json:"apiProvFuncId,omitempty"`

	// Generic information related to the API provider domain function such as details of the API provider applications.
	ApiProvFuncInfo *string `json:"apiProvFuncInfo,omitempty"`

	// Possible values are:
	// - AEF: API provider function is API Exposing Function.
	// - APF: API provider function is API Publishing Function.
	// - AMF: API Provider function is API Management Function.
	ApiProvFuncRole ApiProviderFuncRole `json:"apiProvFuncRole"`

	// Represents registration information of an individual API provider domain function.
	RegInfo RegistrationInformation `json:"regInfo"`
}

// Possible values are:
// - AEF: API provider function is API Exposing Function.
// - APF: API provider function is API Publishing Function.
// - AMF: API Provider function is API Management Function.
type ApiProviderFuncRole string

// Represents registration information of an individual API provider domain function.
type RegistrationInformation struct {
	// API provider domain function's client certificate
	ApiProvCert *string `json:"apiProvCert,omitempty"`

	// Public Key of API Provider domain function.
	ApiProvPubKey string `json:"apiProvPubKey"`
}

// PostRegistrationsJSONBody defines parameters for PostRegistrations.
type PostRegistrationsJSONBody APIProviderEnrolmentDetails

// PutRegistrationsRegistrationIdJSONBody defines parameters for PutRegistrationsRegistrationId.
type PutRegistrationsRegistrationIdJSONBody APIProviderEnrolmentDetails

// PostRegistrationsJSONRequestBody defines body for PostRegistrations for application/json ContentType.
type PostRegistrationsJSONRequestBody PostRegistrationsJSONBody

// PutRegistrationsRegistrationIdJSONRequestBody defines body for PutRegistrationsRegistrationId for application/json ContentType.
type PutRegistrationsRegistrationIdJSONRequestBody PutRegistrationsRegistrationIdJSONBody
