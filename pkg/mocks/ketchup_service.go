// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ViBiOh/ketchup/pkg/model (interfaces: KetchupService)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/ViBiOh/ketchup/pkg/model"
	gomock "github.com/golang/mock/gomock"
)

// KetchupService is a mock of KetchupService interface.
type KetchupService struct {
	ctrl     *gomock.Controller
	recorder *KetchupServiceMockRecorder
}

// KetchupServiceMockRecorder is the mock recorder for KetchupService.
type KetchupServiceMockRecorder struct {
	mock *KetchupService
}

// NewKetchupService creates a new mock instance.
func NewKetchupService(ctrl *gomock.Controller) *KetchupService {
	mock := &KetchupService{ctrl: ctrl}
	mock.recorder = &KetchupServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *KetchupService) EXPECT() *KetchupServiceMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *KetchupService) Create(arg0 context.Context, arg1 model.Ketchup) (model.Ketchup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(model.Ketchup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *KetchupServiceMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*KetchupService)(nil).Create), arg0, arg1)
}

// Delete mocks base method.
func (m *KetchupService) Delete(arg0 context.Context, arg1 model.Ketchup) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *KetchupServiceMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*KetchupService)(nil).Delete), arg0, arg1)
}

// List mocks base method.
func (m *KetchupService) List(arg0 context.Context, arg1 uint, arg2 string) ([]model.Ketchup, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model.Ketchup)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// List indicates an expected call of List.
func (mr *KetchupServiceMockRecorder) List(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*KetchupService)(nil).List), arg0, arg1, arg2)
}

// ListForRepositories mocks base method.
func (m *KetchupService) ListForRepositories(arg0 context.Context, arg1 []model.Repository, arg2 model.KetchupFrequency) ([]model.Ketchup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListForRepositories", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model.Ketchup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListForRepositories indicates an expected call of ListForRepositories.
func (mr *KetchupServiceMockRecorder) ListForRepositories(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListForRepositories", reflect.TypeOf((*KetchupService)(nil).ListForRepositories), arg0, arg1, arg2)
}

// ListOutdatedByFrequency mocks base method.
func (m *KetchupService) ListOutdatedByFrequency(arg0 context.Context, arg1 model.KetchupFrequency, arg2 ...model.User) ([]model.Ketchup, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListOutdatedByFrequency", varargs...)
	ret0, _ := ret[0].([]model.Ketchup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOutdatedByFrequency indicates an expected call of ListOutdatedByFrequency.
func (mr *KetchupServiceMockRecorder) ListOutdatedByFrequency(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOutdatedByFrequency", reflect.TypeOf((*KetchupService)(nil).ListOutdatedByFrequency), varargs...)
}

// Update mocks base method.
func (m *KetchupService) Update(arg0 context.Context, arg1 string, arg2 model.Ketchup) (model.Ketchup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2)
	ret0, _ := ret[0].(model.Ketchup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *KetchupServiceMockRecorder) Update(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*KetchupService)(nil).Update), arg0, arg1, arg2)
}

// UpdateAll mocks base method.
func (m *KetchupService) UpdateAll(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAll", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAll indicates an expected call of UpdateAll.
func (mr *KetchupServiceMockRecorder) UpdateAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAll", reflect.TypeOf((*KetchupService)(nil).UpdateAll), arg0)
}

// UpdateVersion mocks base method.
func (m *KetchupService) UpdateVersion(arg0 context.Context, arg1, arg2 uint64, arg3, arg4 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateVersion", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateVersion indicates an expected call of UpdateVersion.
func (mr *KetchupServiceMockRecorder) UpdateVersion(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateVersion", reflect.TypeOf((*KetchupService)(nil).UpdateVersion), arg0, arg1, arg2, arg3, arg4)
}
