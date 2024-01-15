// Code generated by mockery v2.35.4. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	keycloak "oransc.org/nonrtric/capifcore/internal/keycloak"
)

// AccessManagement is an autogenerated mock type for the AccessManagement type
type AccessManagement struct {
	mock.Mock
}

// AddClient provides a mock function with given fields: clientId, realm
func (_m *AccessManagement) AddClient(clientId string, realm string) error {
	ret := _m.Called(clientId, realm)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(clientId, realm)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetClientRepresentation provides a mock function with given fields: clientId, realm
func (_m *AccessManagement) GetClientRepresentation(clientId string, realm string) (*keycloak.Client, error) {
	ret := _m.Called(clientId, realm)

	var r0 *keycloak.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (*keycloak.Client, error)); ok {
		return rf(clientId, realm)
	}
	if rf, ok := ret.Get(0).(func(string, string) *keycloak.Client); ok {
		r0 = rf(clientId, realm)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*keycloak.Client)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(clientId, realm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetToken provides a mock function with given fields: realm, data
func (_m *AccessManagement) GetToken(realm string, data map[string][]string) (keycloak.Jwttoken, error) {
	ret := _m.Called(realm, data)

	var r0 keycloak.Jwttoken
	var r1 error
	if rf, ok := ret.Get(0).(func(string, map[string][]string) (keycloak.Jwttoken, error)); ok {
		return rf(realm, data)
	}
	if rf, ok := ret.Get(0).(func(string, map[string][]string) keycloak.Jwttoken); ok {
		r0 = rf(realm, data)
	} else {
		r0 = ret.Get(0).(keycloak.Jwttoken)
	}

	if rf, ok := ret.Get(1).(func(string, map[string][]string) error); ok {
		r1 = rf(realm, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAccessManagement creates a new instance of AccessManagement. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAccessManagement(t interface {
	mock.TestingT
	Cleanup(func())
}) *AccessManagement {
	mock := &AccessManagement{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
