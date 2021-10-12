// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ViBiOh/ketchup/pkg/model (interfaces: Mailer)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	model "github.com/ViBiOh/mailer/pkg/model"
	gomock "github.com/golang/mock/gomock"
)

// Mailer is a mock of Mailer interface.
type Mailer struct {
	ctrl     *gomock.Controller
	recorder *MailerMockRecorder
}

// MailerMockRecorder is the mock recorder for Mailer.
type MailerMockRecorder struct {
	mock *Mailer
}

// NewMailer creates a new mock instance.
func NewMailer(ctrl *gomock.Controller) *Mailer {
	mock := &Mailer{ctrl: ctrl}
	mock.recorder = &MailerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Mailer) EXPECT() *MailerMockRecorder {
	return m.recorder
}

// Enabled mocks base method.
func (m *Mailer) Enabled() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Enabled indicates an expected call of Enabled.
func (mr *MailerMockRecorder) Enabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enabled", reflect.TypeOf((*Mailer)(nil).Enabled))
}

// Send mocks base method.
func (m *Mailer) Send(arg0 context.Context, arg1 model.MailRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MailerMockRecorder) Send(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*Mailer)(nil).Send), arg0, arg1)
}