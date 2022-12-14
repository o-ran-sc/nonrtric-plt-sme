// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	invokermanagementapi "oransc.org/nonrtric/capifcore/internal/invokermanagementapi"
)

// InvokerRegister is an autogenerated mock type for the InvokerRegister type
type InvokerRegister struct {
	mock.Mock
}

// GetInvokerApiList provides a mock function with given fields: invokerId
func (_m *InvokerRegister) GetInvokerApiList(invokerId string) *invokermanagementapi.APIList {
	ret := _m.Called(invokerId)

	var r0 *invokermanagementapi.APIList
	if rf, ok := ret.Get(0).(func(string) *invokermanagementapi.APIList); ok {
		r0 = rf(invokerId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*invokermanagementapi.APIList)
		}
	}

	return r0
}

// IsInvokerRegistered provides a mock function with given fields: invokerId
func (_m *InvokerRegister) IsInvokerRegistered(invokerId string) bool {
	ret := _m.Called(invokerId)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(invokerId)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// VerifyInvokerSecret provides a mock function with given fields: invokerId, secret
func (_m *InvokerRegister) VerifyInvokerSecret(invokerId string, secret string) bool {
	ret := _m.Called(invokerId, secret)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(invokerId, secret)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type mockConstructorTestingTNewInvokerRegister interface {
	mock.TestingT
	Cleanup(func())
}

// NewInvokerRegister creates a new instance of InvokerRegister. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewInvokerRegister(t mockConstructorTestingTNewInvokerRegister) *InvokerRegister {
	mock := &InvokerRegister{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
