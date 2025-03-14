// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/jackc/pgx/v5 (interfaces: Row,Rows)
//
// Generated by this command:
//
//	mockgen -destination pkg/mocks/pgx.go -package mocks -mock_names Row=Row,Rows=Rows github.com/jackc/pgx/v5 Row,Rows
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	pgx "github.com/jackc/pgx/v5"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	gomock "go.uber.org/mock/gomock"
)

// Row is a mock of Row interface.
type Row struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *RowMockRecorder
}

// RowMockRecorder is the mock recorder for Row.
type RowMockRecorder struct {
	mock *Row
}

// NewRow creates a new mock instance.
func NewRow(ctrl *gomock.Controller) *Row {
	mock := &Row{ctrl: ctrl}
	mock.recorder = &RowMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Row) EXPECT() *RowMockRecorder {
	return m.recorder
}

// Scan mocks base method.
func (m *Row) Scan(dest ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range dest {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *RowMockRecorder) Scan(dest ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*Row)(nil).Scan), dest...)
}

// Rows is a mock of Rows interface.
type Rows struct {
	isgomock struct{}
	ctrl     *gomock.Controller
	recorder *RowsMockRecorder
}

// RowsMockRecorder is the mock recorder for Rows.
type RowsMockRecorder struct {
	mock *Rows
}

// NewRows creates a new mock instance.
func NewRows(ctrl *gomock.Controller) *Rows {
	mock := &Rows{ctrl: ctrl}
	mock.recorder = &RowsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Rows) EXPECT() *RowsMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *Rows) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *RowsMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*Rows)(nil).Close))
}

// CommandTag mocks base method.
func (m *Rows) CommandTag() pgconn.CommandTag {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommandTag")
	ret0, _ := ret[0].(pgconn.CommandTag)
	return ret0
}

// CommandTag indicates an expected call of CommandTag.
func (mr *RowsMockRecorder) CommandTag() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommandTag", reflect.TypeOf((*Rows)(nil).CommandTag))
}

// Conn mocks base method.
func (m *Rows) Conn() *pgx.Conn {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Conn")
	ret0, _ := ret[0].(*pgx.Conn)
	return ret0
}

// Conn indicates an expected call of Conn.
func (mr *RowsMockRecorder) Conn() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Conn", reflect.TypeOf((*Rows)(nil).Conn))
}

// Err mocks base method.
func (m *Rows) Err() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

// Err indicates an expected call of Err.
func (mr *RowsMockRecorder) Err() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Err", reflect.TypeOf((*Rows)(nil).Err))
}

// FieldDescriptions mocks base method.
func (m *Rows) FieldDescriptions() []pgconn.FieldDescription {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FieldDescriptions")
	ret0, _ := ret[0].([]pgconn.FieldDescription)
	return ret0
}

// FieldDescriptions indicates an expected call of FieldDescriptions.
func (mr *RowsMockRecorder) FieldDescriptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FieldDescriptions", reflect.TypeOf((*Rows)(nil).FieldDescriptions))
}

// Next mocks base method.
func (m *Rows) Next() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Next indicates an expected call of Next.
func (mr *RowsMockRecorder) Next() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*Rows)(nil).Next))
}

// RawValues mocks base method.
func (m *Rows) RawValues() [][]byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RawValues")
	ret0, _ := ret[0].([][]byte)
	return ret0
}

// RawValues indicates an expected call of RawValues.
func (mr *RowsMockRecorder) RawValues() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RawValues", reflect.TypeOf((*Rows)(nil).RawValues))
}

// Scan mocks base method.
func (m *Rows) Scan(dest ...any) error {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range dest {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Scan", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Scan indicates an expected call of Scan.
func (mr *RowsMockRecorder) Scan(dest ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scan", reflect.TypeOf((*Rows)(nil).Scan), dest...)
}

// Values mocks base method.
func (m *Rows) Values() ([]any, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Values")
	ret0, _ := ret[0].([]any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Values indicates an expected call of Values.
func (mr *RowsMockRecorder) Values() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Values", reflect.TypeOf((*Rows)(nil).Values))
}
