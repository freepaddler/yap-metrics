// Code generated by MockGen. DO NOT EDIT.
// Source: http.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	models "github.com/freepaddler/yap-metrics/internal/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockHTTPHandlerStorage is a mock of HTTPHandlerStorage interface.
type MockHTTPHandlerStorage struct {
	ctrl     *gomock.Controller
	recorder *MockHTTPHandlerStorageMockRecorder
}

// MockHTTPHandlerStorageMockRecorder is the mock recorder for MockHTTPHandlerStorage.
type MockHTTPHandlerStorageMockRecorder struct {
	mock *MockHTTPHandlerStorage
}

// NewMockHTTPHandlerStorage creates a new mock instance.
func NewMockHTTPHandlerStorage(ctrl *gomock.Controller) *MockHTTPHandlerStorage {
	mock := &MockHTTPHandlerStorage{ctrl: ctrl}
	mock.recorder = &MockHTTPHandlerStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHTTPHandlerStorage) EXPECT() *MockHTTPHandlerStorageMockRecorder {
	return m.recorder
}

// GetAll mocks base method.
func (m *MockHTTPHandlerStorage) GetAll() []models.Metrics {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll")
	ret0, _ := ret[0].([]models.Metrics)
	return ret0
}

// GetAll indicates an expected call of GetAll.
func (mr *MockHTTPHandlerStorageMockRecorder) GetAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockHTTPHandlerStorage)(nil).GetAll))
}

// GetOne mocks base method.
func (m *MockHTTPHandlerStorage) GetOne(request models.MetricRequest) (models.Metrics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOne", request)
	ret0, _ := ret[0].(models.Metrics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOne indicates an expected call of GetOne.
func (mr *MockHTTPHandlerStorageMockRecorder) GetOne(request interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOne", reflect.TypeOf((*MockHTTPHandlerStorage)(nil).GetOne), request)
}

// Ping mocks base method.
func (m *MockHTTPHandlerStorage) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockHTTPHandlerStorageMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockHTTPHandlerStorage)(nil).Ping))
}

// UpdateMany mocks base method.
func (m *MockHTTPHandlerStorage) UpdateMany(metrics []models.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateMany", metrics)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateMany indicates an expected call of UpdateMany.
func (mr *MockHTTPHandlerStorageMockRecorder) UpdateMany(metrics interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMany", reflect.TypeOf((*MockHTTPHandlerStorage)(nil).UpdateMany), metrics)
}

// UpdateOne mocks base method.
func (m *MockHTTPHandlerStorage) UpdateOne(metric *models.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOne", metric)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOne indicates an expected call of UpdateOne.
func (mr *MockHTTPHandlerStorageMockRecorder) UpdateOne(metric interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOne", reflect.TypeOf((*MockHTTPHandlerStorage)(nil).UpdateOne), metric)
}