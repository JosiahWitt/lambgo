// Code generated by `ensure mocks generate`. DO NOT EDIT.
// Source: io/fs (interfaces: ReadFileFS)

// Package mock_fs is a generated GoMock package.
package mock_fs

import (
	"github.com/golang/mock/gomock"
	"io/fs"
	"reflect"
)

// MockReadFileFS is a mock of the ReadFileFS interface in io/fs.
type MockReadFileFS struct {
	ctrl     *gomock.Controller
	recorder *MockReadFileFSMockRecorder
}

// MockReadFileFSMockRecorder is the mock recorder for MockReadFileFS.
type MockReadFileFSMockRecorder struct {
	mock *MockReadFileFS
}

// NewMockReadFileFS creates a new mock instance.
func NewMockReadFileFS(ctrl *gomock.Controller) *MockReadFileFS {
	mock := &MockReadFileFS{ctrl: ctrl}
	mock.recorder = &MockReadFileFSMockRecorder{mock}
	return mock
}

// NEW creates a MockReadFileFS. This method is used internally by ensure.
func (*MockReadFileFS) NEW(ctrl *gomock.Controller) *MockReadFileFS {
	return NewMockReadFileFS(ctrl)
}

// EXPECT returns a struct that allows setting up expectations.
func (m *MockReadFileFS) EXPECT() *MockReadFileFSMockRecorder {
	return m.recorder
}

// Open mocks Open on ReadFileFS.
func (m *MockReadFileFS) Open(_name string) (fs.File, error) {
	m.ctrl.T.Helper()
	inputs := []interface{}{_name}
	ret := m.ctrl.Call(m, "Open", inputs...)
	ret0, _ := ret[0].(fs.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Open sets up expectations for calls to Open.
// Calling this method multiple times allows expecting multiple calls to Open with a variety of parameters.
//
// Inputs:
//
//  name string
//
// Outputs:
//
//  fs.File
//  error
func (mr *MockReadFileFSMockRecorder) Open(_name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	inputs := []interface{}{_name}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Open", reflect.TypeOf((*MockReadFileFS)(nil).Open), inputs...)
}

// ReadFile mocks ReadFile on ReadFileFS.
func (m *MockReadFileFS) ReadFile(_name string) ([]byte, error) {
	m.ctrl.T.Helper()
	inputs := []interface{}{_name}
	ret := m.ctrl.Call(m, "ReadFile", inputs...)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadFile sets up expectations for calls to ReadFile.
// Calling this method multiple times allows expecting multiple calls to ReadFile with a variety of parameters.
//
// Inputs:
//
//  name string
//
// Outputs:
//
//  []byte
//  error
func (mr *MockReadFileFSMockRecorder) ReadFile(_name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	inputs := []interface{}{_name}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadFile", reflect.TypeOf((*MockReadFileFS)(nil).ReadFile), inputs...)
}
