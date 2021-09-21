// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ViBiOh/ketchup/pkg/model (interfaces: RepositoryService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/ViBiOh/ketchup/pkg/model"
	semver "github.com/ViBiOh/ketchup/pkg/semver"
	gomock "github.com/golang/mock/gomock"
)

// RepositoryService is a mock of RepositoryService interface.
type RepositoryService struct {
	ctrl     *gomock.Controller
	recorder *RepositoryServiceMockRecorder
}

// RepositoryServiceMockRecorder is the mock recorder for RepositoryService.
type RepositoryServiceMockRecorder struct {
	mock *RepositoryService
}

// NewRepositoryService creates a new mock instance.
func NewRepositoryService(ctrl *gomock.Controller) *RepositoryService {
	mock := &RepositoryService{ctrl: ctrl}
	mock.recorder = &RepositoryServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *RepositoryService) EXPECT() *RepositoryServiceMockRecorder {
	return m.recorder
}

// Clean mocks base method.
func (m *RepositoryService) Clean(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Clean", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Clean indicates an expected call of Clean.
func (mr *RepositoryServiceMockRecorder) Clean(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Clean", reflect.TypeOf((*RepositoryService)(nil).Clean), arg0)
}

// GetOrCreate mocks base method.
func (m *RepositoryService) GetOrCreate(arg0 context.Context, arg1 model.RepositoryKind, arg2, arg3, arg4 string) (model.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrCreate", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(model.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrCreate indicates an expected call of GetOrCreate.
func (mr *RepositoryServiceMockRecorder) GetOrCreate(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrCreate", reflect.TypeOf((*RepositoryService)(nil).GetOrCreate), arg0, arg1, arg2, arg3, arg4)
}

// LatestVersions mocks base method.
func (m *RepositoryService) LatestVersions(arg0 model.Repository) (map[string]semver.Version, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LatestVersions", arg0)
	ret0, _ := ret[0].(map[string]semver.Version)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LatestVersions indicates an expected call of LatestVersions.
func (mr *RepositoryServiceMockRecorder) LatestVersions(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LatestVersions", reflect.TypeOf((*RepositoryService)(nil).LatestVersions), arg0)
}

// List mocks base method.
func (m *RepositoryService) List(arg0 context.Context, arg1 uint, arg2 string) ([]model.Repository, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model.Repository)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// List indicates an expected call of List.
func (mr *RepositoryServiceMockRecorder) List(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*RepositoryService)(nil).List), arg0, arg1, arg2)
}

// ListByKinds mocks base method.
func (m *RepositoryService) ListByKinds(arg0 context.Context, arg1 uint, arg2 string, arg3 ...model.RepositoryKind) ([]model.Repository, uint64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListByKinds", varargs...)
	ret0, _ := ret[0].([]model.Repository)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ListByKinds indicates an expected call of ListByKinds.
func (mr *RepositoryServiceMockRecorder) ListByKinds(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByKinds", reflect.TypeOf((*RepositoryService)(nil).ListByKinds), varargs...)
}

// Suggest mocks base method.
func (m *RepositoryService) Suggest(arg0 context.Context, arg1 []uint64, arg2 uint64) ([]model.Repository, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Suggest", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model.Repository)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Suggest indicates an expected call of Suggest.
func (mr *RepositoryServiceMockRecorder) Suggest(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Suggest", reflect.TypeOf((*RepositoryService)(nil).Suggest), arg0, arg1, arg2)
}

// Update mocks base method.
func (m *RepositoryService) Update(arg0 context.Context, arg1 model.Repository) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *RepositoryServiceMockRecorder) Update(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*RepositoryService)(nil).Update), arg0, arg1)
}
