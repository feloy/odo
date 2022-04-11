// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/binding/backend/interface.go

// Package backend is a generated GoMock package.
package backend

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockCreateBindingBackend is a mock of CreateBindingBackend interface
type MockCreateBindingBackend struct {
	ctrl     *gomock.Controller
	recorder *MockCreateBindingBackendMockRecorder
}

// MockCreateBindingBackendMockRecorder is the mock recorder for MockCreateBindingBackend
type MockCreateBindingBackendMockRecorder struct {
	mock *MockCreateBindingBackend
}

// NewMockCreateBindingBackend creates a new mock instance
func NewMockCreateBindingBackend(ctrl *gomock.Controller) *MockCreateBindingBackend {
	mock := &MockCreateBindingBackend{ctrl: ctrl}
	mock.recorder = &MockCreateBindingBackendMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCreateBindingBackend) EXPECT() *MockCreateBindingBackendMockRecorder {
	return m.recorder
}

// Validate mocks base method
func (m *MockCreateBindingBackend) Validate(flags map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", flags)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate
func (mr *MockCreateBindingBackendMockRecorder) Validate(flags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockCreateBindingBackend)(nil).Validate), flags)
}

// SelectServiceInstance mocks base method
func (m *MockCreateBindingBackend) SelectServiceInstance(flags map[string]string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectServiceInstance", flags)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectServiceInstance indicates an expected call of SelectServiceInstance
func (mr *MockCreateBindingBackendMockRecorder) SelectServiceInstance(flags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectServiceInstance", reflect.TypeOf((*MockCreateBindingBackend)(nil).SelectServiceInstance), flags)
}

// AskBindingName mocks base method
func (m *MockCreateBindingBackend) AskBindingName(componentName string, flags map[string]string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AskBindingName", componentName, flags)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AskBindingName indicates an expected call of AskBindingName
func (mr *MockCreateBindingBackendMockRecorder) AskBindingName(componentName, flags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AskBindingName", reflect.TypeOf((*MockCreateBindingBackend)(nil).AskBindingName), componentName, flags)
}

// AskBindAsFiles mocks base method
func (m *MockCreateBindingBackend) AskBindAsFiles(flags map[string]string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AskBindAsFiles", flags)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AskBindAsFiles indicates an expected call of AskBindAsFiles
func (mr *MockCreateBindingBackendMockRecorder) AskBindAsFiles(flags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AskBindAsFiles", reflect.TypeOf((*MockCreateBindingBackend)(nil).AskBindAsFiles), flags)
}
