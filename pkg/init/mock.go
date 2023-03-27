// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/init/interface.go

// Package init is a generated GoMock package.
package init

import (
	context "context"
	reflect "reflect"

	v1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	parser "github.com/devfile/library/v2/pkg/devfile/parser"
	gomock "github.com/golang/mock/gomock"
	api "github.com/redhat-developer/odo/pkg/api"
	filesystem "github.com/redhat-developer/odo/pkg/testingutil/filesystem"
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

// DownloadDevfile mocks base method.
func (m *MockClient) DownloadDevfile(ctx context.Context, devfileLocation *api.DetectionResult, destDir string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadDevfile", ctx, devfileLocation, destDir)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadDevfile indicates an expected call of DownloadDevfile.
func (mr *MockClientMockRecorder) DownloadDevfile(ctx, devfileLocation, destDir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadDevfile", reflect.TypeOf((*MockClient)(nil).DownloadDevfile), ctx, devfileLocation, destDir)
}

// DownloadStarterProject mocks base method.
func (m *MockClient) DownloadStarterProject(project *v1alpha2.StarterProject, dest string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadStarterProject", project, dest)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadStarterProject indicates an expected call of DownloadStarterProject.
func (mr *MockClientMockRecorder) DownloadStarterProject(project, dest interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadStarterProject", reflect.TypeOf((*MockClient)(nil).DownloadStarterProject), project, dest)
}

// GetFlags mocks base method.
func (m *MockClient) GetFlags(flags map[string]string) map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFlags", flags)
	ret0, _ := ret[0].(map[string]string)
	return ret0
}

// GetFlags indicates an expected call of GetFlags.
func (mr *MockClientMockRecorder) GetFlags(flags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFlags", reflect.TypeOf((*MockClient)(nil).GetFlags), flags)
}

// HandleApplicationPorts mocks base method.
func (m *MockClient) HandleApplicationPorts(devfileobj parser.DevfileObj, ports []int, flags map[string]string, fs filesystem.Filesystem, dir string) (parser.DevfileObj, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleApplicationPorts", devfileobj, ports, flags, fs, dir)
	ret0, _ := ret[0].(parser.DevfileObj)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HandleApplicationPorts indicates an expected call of HandleApplicationPorts.
func (mr *MockClientMockRecorder) HandleApplicationPorts(devfileobj, ports, flags, fs, dir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleApplicationPorts", reflect.TypeOf((*MockClient)(nil).HandleApplicationPorts), devfileobj, ports, flags, fs, dir)
}

// InitDevfile mocks base method.
func (m *MockClient) InitDevfile(ctx context.Context, flags map[string]string, contextDir string, preInitHandlerFunc func(bool), newDevfileHandlerFunc func(parser.DevfileObj) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitDevfile", ctx, flags, contextDir, preInitHandlerFunc, newDevfileHandlerFunc)
	ret0, _ := ret[0].(error)
	return ret0
}

// InitDevfile indicates an expected call of InitDevfile.
func (mr *MockClientMockRecorder) InitDevfile(ctx, flags, contextDir, preInitHandlerFunc, newDevfileHandlerFunc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitDevfile", reflect.TypeOf((*MockClient)(nil).InitDevfile), ctx, flags, contextDir, preInitHandlerFunc, newDevfileHandlerFunc)
}

// PersonalizeDevfileConfig mocks base method.
func (m *MockClient) PersonalizeDevfileConfig(devfileobj parser.DevfileObj, flags map[string]string, fs filesystem.Filesystem, dir string) (parser.DevfileObj, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PersonalizeDevfileConfig", devfileobj, flags, fs, dir)
	ret0, _ := ret[0].(parser.DevfileObj)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PersonalizeDevfileConfig indicates an expected call of PersonalizeDevfileConfig.
func (mr *MockClientMockRecorder) PersonalizeDevfileConfig(devfileobj, flags, fs, dir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PersonalizeDevfileConfig", reflect.TypeOf((*MockClient)(nil).PersonalizeDevfileConfig), devfileobj, flags, fs, dir)
}

// PersonalizeName mocks base method.
func (m *MockClient) PersonalizeName(devfile parser.DevfileObj, flags map[string]string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PersonalizeName", devfile, flags)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PersonalizeName indicates an expected call of PersonalizeName.
func (mr *MockClientMockRecorder) PersonalizeName(devfile, flags interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PersonalizeName", reflect.TypeOf((*MockClient)(nil).PersonalizeName), devfile, flags)
}

// SelectAndPersonalizeDevfile mocks base method.
func (m *MockClient) SelectAndPersonalizeDevfile(ctx context.Context, flags map[string]string, contextDir string) (parser.DevfileObj, string, *api.DetectionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectAndPersonalizeDevfile", ctx, flags, contextDir)
	ret0, _ := ret[0].(parser.DevfileObj)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(*api.DetectionResult)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// SelectAndPersonalizeDevfile indicates an expected call of SelectAndPersonalizeDevfile.
func (mr *MockClientMockRecorder) SelectAndPersonalizeDevfile(ctx, flags, contextDir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectAndPersonalizeDevfile", reflect.TypeOf((*MockClient)(nil).SelectAndPersonalizeDevfile), ctx, flags, contextDir)
}

// SelectDevfile mocks base method.
func (m *MockClient) SelectDevfile(ctx context.Context, flags map[string]string, fs filesystem.Filesystem, dir string) (*api.DetectionResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectDevfile", ctx, flags, fs, dir)
	ret0, _ := ret[0].(*api.DetectionResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectDevfile indicates an expected call of SelectDevfile.
func (mr *MockClientMockRecorder) SelectDevfile(ctx, flags, fs, dir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectDevfile", reflect.TypeOf((*MockClient)(nil).SelectDevfile), ctx, flags, fs, dir)
}

// SelectStarterProject mocks base method.
func (m *MockClient) SelectStarterProject(devfile parser.DevfileObj, flags map[string]string, isEmptyDir bool) (*v1alpha2.StarterProject, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SelectStarterProject", devfile, flags, isEmptyDir)
	ret0, _ := ret[0].(*v1alpha2.StarterProject)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SelectStarterProject indicates an expected call of SelectStarterProject.
func (mr *MockClientMockRecorder) SelectStarterProject(devfile, flags, isEmptyDir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SelectStarterProject", reflect.TypeOf((*MockClient)(nil).SelectStarterProject), devfile, flags, isEmptyDir)
}

// Validate mocks base method.
func (m *MockClient) Validate(flags map[string]string, fs filesystem.Filesystem, dir string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", flags, fs, dir)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate.
func (mr *MockClientMockRecorder) Validate(flags, fs, dir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockClient)(nil).Validate), flags, fs, dir)
}
