// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// HelmManager is an autogenerated mock type for the HelmManager type
type HelmManager struct {
	mock.Mock
}

// InstallHelmChart provides a mock function with given fields: namespace, repoName, chartName, releaseName
func (_m *HelmManager) InstallHelmChart(namespace string, repoName string, chartName string, releaseName string) error {
	ret := _m.Called(namespace, repoName, chartName, releaseName)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(namespace, repoName, chartName, releaseName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetUpRepo provides a mock function with given fields: repoName, url
func (_m *HelmManager) SetUpRepo(repoName string, url string) error {
	ret := _m.Called(repoName, url)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(repoName, url)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UninstallHelmChart provides a mock function with given fields: namespace, chartName
func (_m *HelmManager) UninstallHelmChart(namespace string, chartName string) {
	_m.Called(namespace, chartName)
}

type mockConstructorTestingTNewHelmManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewHelmManager creates a new instance of HelmManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewHelmManager(t mockConstructorTestingTNewHelmManager) *HelmManager {
	mock := &HelmManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}