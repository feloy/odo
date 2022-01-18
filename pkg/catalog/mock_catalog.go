// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/catalog/catalog.go

// Package catalog is a generated GoMock package.
package catalog

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	kclient "github.com/redhat-developer/odo/pkg/kclient"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// GetDevfileRegistries mocks base method.
func (m *MockClient) GetDevfileRegistries(registryName string) ([]Registry, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDevfileRegistries", registryName)
	ret0, _ := ret[0].([]Registry)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDevfileRegistries indicates an expected call of GetDevfileRegistries.
func (mr *MockClientMockRecorder) GetDevfileRegistries(registryName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDevfileRegistries", reflect.TypeOf((*MockClient)(nil).GetDevfileRegistries), registryName)
}

// GetStarterProjectsNames mocks base method.
func (m *MockClient) GetStarterProjectsNames(details DevfileComponentType) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStarterProjectsNames", details)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStarterProjectsNames indicates an expected call of GetStarterProjectsNames.
func (mr *MockClientMockRecorder) GetStarterProjectsNames(details interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStarterProjectsNames", reflect.TypeOf((*MockClient)(nil).GetStarterProjectsNames), details)
}

// ListDevfileComponents mocks base method.
func (m *MockClient) ListDevfileComponents(registryName string) (DevfileComponentTypeList, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListDevfileComponents", registryName)
	ret0, _ := ret[0].(DevfileComponentTypeList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDevfileComponents indicates an expected call of ListDevfileComponents.
func (mr *MockClientMockRecorder) ListDevfileComponents(registryName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDevfileComponents", reflect.TypeOf((*MockClient)(nil).ListDevfileComponents), registryName)
}

// SearchComponent mocks base method.
func (m *MockClient) SearchComponent(client kclient.ClientInterface, name string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchComponent", client, name)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchComponent indicates an expected call of SearchComponent.
func (mr *MockClientMockRecorder) SearchComponent(client, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchComponent", reflect.TypeOf((*MockClient)(nil).SearchComponent), client, name)
}
