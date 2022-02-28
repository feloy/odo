// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/component/delete/interface.go

// Package delete is a generated GoMock package.
package delete

import (
	parser "github.com/devfile/library/pkg/devfile/parser"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// UnDeploy mocks base method
func (m *MockClient) UnDeploy(devfileObj parser.DevfileObj, path string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnDeploy", devfileObj, path)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnDeploy indicates an expected call of UnDeploy
func (mr *MockClientMockRecorder) UnDeploy(devfileObj, path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnDeploy", reflect.TypeOf((*MockClient)(nil).UnDeploy), devfileObj, path)
}

// DeleteComponent mocks base method
func (m *MockClient) DeleteComponent(devfileObj parser.DevfileObj, componentName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteComponent", devfileObj, componentName)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteComponent indicates an expected call of DeleteComponent
func (mr *MockClientMockRecorder) DeleteComponent(devfileObj, componentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteComponent", reflect.TypeOf((*MockClient)(nil).DeleteComponent), devfileObj, componentName)
}
