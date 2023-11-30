// Code generated by MockGen. DO NOT EDIT.
// Source: dump.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	models "github.com/freepaddler/yap-metrics/internal/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockDumpDev is a mock of DumpDev interface.
type MockDumpDev struct {
	ctrl     *gomock.Controller
	recorder *MockDumpDevMockRecorder
}

// MockDumpDevMockRecorder is the mock recorder for MockDumpDev.
type MockDumpDevMockRecorder struct {
	mock *MockDumpDev
}

// NewMockDumpDev creates a new mock instance.
func NewMockDumpDev(ctrl *gomock.Controller) *MockDumpDev {
	mock := &MockDumpDev{ctrl: ctrl}
	mock.recorder = &MockDumpDevMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDumpDev) EXPECT() *MockDumpDevMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockDumpDev) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockDumpDevMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockDumpDev)(nil).Close))
}

// Read mocks base method.
func (m *MockDumpDev) Read(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockDumpDevMockRecorder) Read(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockDumpDev)(nil).Read), p)
}

// Truncate mocks base method.
func (m *MockDumpDev) Truncate(size int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Truncate", size)
	ret0, _ := ret[0].(error)
	return ret0
}

// Truncate indicates an expected call of Truncate.
func (mr *MockDumpDevMockRecorder) Truncate(size interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Truncate", reflect.TypeOf((*MockDumpDev)(nil).Truncate), size)
}

// Write mocks base method.
func (m *MockDumpDev) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockDumpDevMockRecorder) Write(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockDumpDev)(nil).Write), p)
}

// MockStoreDumper is a mock of StoreDumper interface.
type MockStoreDumper struct {
	ctrl     *gomock.Controller
	recorder *MockStoreDumperMockRecorder
}

// MockStoreDumperMockRecorder is the mock recorder for MockStoreDumper.
type MockStoreDumperMockRecorder struct {
	mock *MockStoreDumper
}

// NewMockStoreDumper creates a new mock instance.
func NewMockStoreDumper(ctrl *gomock.Controller) *MockStoreDumper {
	mock := &MockStoreDumper{ctrl: ctrl}
	mock.recorder = &MockStoreDumperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStoreDumper) EXPECT() *MockStoreDumperMockRecorder {
	return m.recorder
}

// GetAll mocks base method.
func (m *MockStoreDumper) GetAll() []models.Metrics {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll")
	ret0, _ := ret[0].([]models.Metrics)
	return ret0
}

// GetAll indicates an expected call of GetAll.
func (mr *MockStoreDumperMockRecorder) GetAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockStoreDumper)(nil).GetAll))
}

// Restore mocks base method.
func (m *MockStoreDumper) Restore(metrics []models.Metrics) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Restore", metrics)
}

// Restore indicates an expected call of Restore.
func (mr *MockStoreDumperMockRecorder) Restore(metrics interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Restore", reflect.TypeOf((*MockStoreDumper)(nil).Restore), metrics)
}