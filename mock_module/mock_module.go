// Code generated by MockGen. DO NOT EDIT.
// Source: ./module (interfaces: Module)
//
// Generated by this command:
//
//	mockgen ./module Module
//

// Package mock_module is a generated GoMock package.
package mock_module

import (
	reflect "reflect"

	module "github.com/sneat-co/sneat-go-core/module"
	gomock "go.uber.org/mock/gomock"
)

// MockModule is a mock of Module interface.
type MockModule struct {
	ctrl     *gomock.Controller
	recorder *MockModuleMockRecorder
	isgomock struct{}
}

// MockModuleMockRecorder is the mock recorder for MockModule.
type MockModuleMockRecorder struct {
	mock *MockModule
}

// NewMockModule creates a new mock instance.
func NewMockModule(ctrl *gomock.Controller) *MockModule {
	mock := &MockModule{ctrl: ctrl}
	mock.recorder = &MockModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModule) EXPECT() *MockModuleMockRecorder {
	return m.recorder
}

// ID mocks base method.
func (m *MockModule) ID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "userID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ID indicates an expected call of ID.
func (mr *MockModuleMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "userID", reflect.TypeOf((*MockModule)(nil).ID))
}

// Register mocks base method.
func (m *MockModule) Register(args module.RegistrationArgs) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Register", args)
}

// Register indicates an expected call of Register.
func (mr *MockModuleMockRecorder) Register(args any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockModule)(nil).Register), args)
}
