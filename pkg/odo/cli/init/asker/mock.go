// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/odo/cli/init/asker/interface.go

// Package asker is a generated GoMock package.
package asker

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	catalog "github.com/redhat-developer/odo/pkg/catalog"
)

// MockAsker is a mock of Asker interface.
type MockAsker struct {
	ctrl     *gomock.Controller
	recorder *MockAskerMockRecorder
}

// MockAskerMockRecorder is the mock recorder for MockAsker.
type MockAskerMockRecorder struct {
	mock *MockAsker
}

// NewMockAsker creates a new mock instance.
func NewMockAsker(ctrl *gomock.Controller) *MockAsker {
	mock := &MockAsker{ctrl: ctrl}
	mock.recorder = &MockAskerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAsker) EXPECT() *MockAskerMockRecorder {
	return m.recorder
}

// AskLanguage mocks base method.
func (m *MockAsker) AskLanguage(langs []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AskLanguage", langs)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AskLanguage indicates an expected call of AskLanguage.
func (mr *MockAskerMockRecorder) AskLanguage(langs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AskLanguage", reflect.TypeOf((*MockAsker)(nil).AskLanguage), langs)
}

// AskName mocks base method.
func (m *MockAsker) AskName(defaultName string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AskName", defaultName)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AskName indicates an expected call of AskName.
func (mr *MockAskerMockRecorder) AskName(defaultName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AskName", reflect.TypeOf((*MockAsker)(nil).AskName), defaultName)
}

// AskStarterProject mocks base method.
func (m *MockAsker) AskStarterProject(projects []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AskStarterProject", projects)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AskStarterProject indicates an expected call of AskStarterProject.
func (mr *MockAskerMockRecorder) AskStarterProject(projects interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AskStarterProject", reflect.TypeOf((*MockAsker)(nil).AskStarterProject), projects)
}

// AskType mocks base method.
func (m *MockAsker) AskType(types catalog.TypesWithDetails) (catalog.DevfileComponentType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AskType", types)
	ret0, _ := ret[0].(catalog.DevfileComponentType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AskType indicates an expected call of AskType.
func (mr *MockAskerMockRecorder) AskType(types interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AskType", reflect.TypeOf((*MockAsker)(nil).AskType), types)
}
