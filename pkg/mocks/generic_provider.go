// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ViBiOh/ketchup/pkg/model (interfaces: GenericProvider)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	semver "github.com/ViBiOh/ketchup/pkg/semver"
	gomock "github.com/golang/mock/gomock"
)

// GenericProvider is a mock of GenericProvider interface.
type GenericProvider struct {
	ctrl     *gomock.Controller
	recorder *GenericProviderMockRecorder
}

// GenericProviderMockRecorder is the mock recorder for GenericProvider.
type GenericProviderMockRecorder struct {
	mock *GenericProvider
}

// NewGenericProvider creates a new mock instance.
func NewGenericProvider(ctrl *gomock.Controller) *GenericProvider {
	mock := &GenericProvider{ctrl: ctrl}
	mock.recorder = &GenericProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *GenericProvider) EXPECT() *GenericProviderMockRecorder {
	return m.recorder
}

// LatestVersions mocks base method.
func (m *GenericProvider) LatestVersions(arg0 string, arg1 []string) (map[string]semver.Version, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LatestVersions", arg0, arg1)
	ret0, _ := ret[0].(map[string]semver.Version)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LatestVersions indicates an expected call of LatestVersions.
func (mr *GenericProviderMockRecorder) LatestVersions(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LatestVersions", reflect.TypeOf((*GenericProvider)(nil).LatestVersions), arg0, arg1)
}
