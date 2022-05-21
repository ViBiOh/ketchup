// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ViBiOh/ketchup/pkg/model (interfaces: KetchupStore)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/ViBiOh/ketchup/pkg/model"
	gomock "github.com/golang/mock/gomock"
)

// KetchupStore is a mock of KetchupStore interface.
type KetchupStore struct {
	ctrl     *gomock.Controller
	recorder *KetchupStoreMockRecorder
}

// KetchupStoreMockRecorder is the mock recorder for KetchupStore.
type KetchupStoreMockRecorder struct {
	mock *KetchupStore
}

// NewKetchupStore creates a new mock instance.
func NewKetchupStore(ctrl *gomock.Controller) *KetchupStore {
	mock := &KetchupStore{ctrl: ctrl}
	mock.recorder = &KetchupStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *KetchupStore) EXPECT() *KetchupStoreMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *KetchupStore) Create(arg0 context.Context, arg1 model.Ketchup) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *KetchupStoreMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*KetchupStore)(nil).Create), arg0, arg1)
}

// Delete mocks base method.
func (m *KetchupStore) Delete(arg0 context.Context, arg1 model.Ketchup) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *KetchupStoreMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*KetchupStore)(nil).Delete), arg0, arg1)
}

// DoAtomic mocks base method.
func (m *KetchupStore) DoAtomic(arg0 context.Context, arg1 func(context.Context) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DoAtomic", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DoAtomic indicates an expected call of DoAtomic.
func (mr *KetchupStoreMockRecorder) DoAtomic(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DoAtomic", reflect.TypeOf((*KetchupStore)(nil).DoAtomic), arg0, arg1)
}

// GetByRepository mocks base method.
func (m *KetchupStore) GetByRepository(arg0 context.Context, arg1 uint64, arg2 string, arg3 bool) (model.Ketchup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByRepository", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(model.Ketchup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByRepository indicates an expected call of GetByRepository.
func (mr *KetchupStoreMockRecorder) GetByRepository(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByRepository", reflect.TypeOf((*KetchupStore)(nil).GetByRepository), arg0, arg1, arg2, arg3)
}

// List mocks base method.
func (m *KetchupStore) List(arg0 context.Context, arg1 uint, arg2 string) ([]model.Ketchup, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1, arg2)
	ret0, _ := ret[0].([]model.Ketchup)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// List indicates an expected call of List.
func (mr *KetchupStoreMockRecorder) List(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*KetchupStore)(nil).List), arg0, arg1, arg2)
}

// ListByRepositoriesIDAndFrequencies mocks base method.
func (m *KetchupStore) ListByRepositoriesIDAndFrequencies(arg0 context.Context, arg1 []uint64, arg2 ...model.KetchupFrequency) ([]model.Ketchup, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListByRepositoriesIDAndFrequencies", varargs...)
	ret0, _ := ret[0].([]model.Ketchup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByRepositoriesIDAndFrequencies indicates an expected call of ListByRepositoriesIDAndFrequencies.
func (mr *KetchupStoreMockRecorder) ListByRepositoriesIDAndFrequencies(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByRepositoriesIDAndFrequencies", reflect.TypeOf((*KetchupStore)(nil).ListByRepositoriesIDAndFrequencies), varargs...)
}

// ListOutdated mocks base method.
func (m *KetchupStore) ListOutdated(arg0 context.Context, arg1 ...uint64) ([]model.Ketchup, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListOutdated", varargs...)
	ret0, _ := ret[0].([]model.Ketchup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListOutdated indicates an expected call of ListOutdated.
func (mr *KetchupStoreMockRecorder) ListOutdated(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListOutdated", reflect.TypeOf((*KetchupStore)(nil).ListOutdated), varargs...)
}

// Update mocks base method.
func (m *KetchupStore) Update(arg0 context.Context, arg1 model.Ketchup, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *KetchupStoreMockRecorder) Update(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*KetchupStore)(nil).Update), arg0, arg1, arg2)
}

// UpdateAll mocks base method.
func (m *KetchupStore) UpdateAll(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAll", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAll indicates an expected call of UpdateAll.
func (mr *KetchupStoreMockRecorder) UpdateAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAll", reflect.TypeOf((*KetchupStore)(nil).UpdateAll), arg0)
}

// UpdateVersion mocks base method.
func (m *KetchupStore) UpdateVersion(arg0 context.Context, arg1, arg2 uint64, arg3, arg4 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateVersion", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateVersion indicates an expected call of UpdateVersion.
func (mr *KetchupStoreMockRecorder) UpdateVersion(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateVersion", reflect.TypeOf((*KetchupStore)(nil).UpdateVersion), arg0, arg1, arg2, arg3, arg4)
}
