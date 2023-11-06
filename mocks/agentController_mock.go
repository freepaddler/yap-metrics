// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	models "github.com/freepaddler/yap-metrics/internal/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockReporter is a mock of Reporter interface.
type MockReporter struct {
	ctrl     *gomock.Controller
	recorder *MockReporterMockRecorder
}

// MockReporterMockRecorder is the mock recorder for MockReporter.
type MockReporterMockRecorder struct {
	mock *MockReporter
}

// NewMockReporter creates a new mock instance.
func NewMockReporter(ctrl *gomock.Controller) *MockReporter {
	mock := &MockReporter{ctrl: ctrl}
	mock.recorder = &MockReporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockReporter) EXPECT() *MockReporterMockRecorder {
	return m.recorder
}

// ReportAll mocks base method.
func (m *MockReporter) ReportAll() ([]models.Metrics, time.Time) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReportAll")
	ret0, _ := ret[0].([]models.Metrics)
	ret1, _ := ret[1].(time.Time)
	return ret0, ret1
}

// ReportAll indicates an expected call of ReportAll.
func (mr *MockReporterMockRecorder) ReportAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReportAll", reflect.TypeOf((*MockReporter)(nil).ReportAll))
}

// RestoreReport mocks base method.
func (m *MockReporter) RestoreReport(metrics []models.Metrics, ts time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RestoreReport", metrics, ts)
}

// RestoreReport indicates an expected call of RestoreReport.
func (mr *MockReporterMockRecorder) RestoreReport(metrics, ts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreReport", reflect.TypeOf((*MockReporter)(nil).RestoreReport), metrics, ts)
}

// MockCollector is a mock of Collector interface.
type MockCollector struct {
	ctrl     *gomock.Controller
	recorder *MockCollectorMockRecorder
}

// MockCollectorMockRecorder is the mock recorder for MockCollector.
type MockCollectorMockRecorder struct {
	mock *MockCollector
}

// NewMockCollector creates a new mock instance.
func NewMockCollector(ctrl *gomock.Controller) *MockCollector {
	mock := &MockCollector{ctrl: ctrl}
	mock.recorder = &MockCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCollector) EXPECT() *MockCollectorMockRecorder {
	return m.recorder
}

// CollectCounter mocks base method.
func (m *MockCollector) CollectCounter(name string, val int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CollectCounter", name, val)
}

// CollectCounter indicates an expected call of CollectCounter.
func (mr *MockCollectorMockRecorder) CollectCounter(name, val interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectCounter", reflect.TypeOf((*MockCollector)(nil).CollectCounter), name, val)
}

// CollectGauge mocks base method.
func (m *MockCollector) CollectGauge(name string, val float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CollectGauge", name, val)
}

// CollectGauge indicates an expected call of CollectGauge.
func (mr *MockCollectorMockRecorder) CollectGauge(name, val interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectGauge", reflect.TypeOf((*MockCollector)(nil).CollectGauge), name, val)
}