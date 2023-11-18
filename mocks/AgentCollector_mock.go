// Code generated by MockGen. DO NOT EDIT.
// Source: setup.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	models "github.com/freepaddler/yap-metrics/internal/pkg/models"
	gomock "github.com/golang/mock/gomock"
)

// MockStoreCollector is a mock of StoreCollector interface.
type MockStoreCollector struct {
	ctrl     *gomock.Controller
	recorder *MockStoreCollectorMockRecorder
}

// MockStoreCollectorMockRecorder is the mock recorder for MockStoreCollector.
type MockStoreCollectorMockRecorder struct {
	mock *MockStoreCollector
}

// NewMockStoreCollector creates a new mock instance.
func NewMockStoreCollector(ctrl *gomock.Controller) *MockStoreCollector {
	mock := &MockStoreCollector{ctrl: ctrl}
	mock.recorder = &MockStoreCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStoreCollector) EXPECT() *MockStoreCollectorMockRecorder {
	return m.recorder
}

// CollectCounter mocks base method.
func (m *MockStoreCollector) CollectCounter(name string, val int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CollectCounter", name, val)
}

// CollectCounter indicates an expected call of CollectCounter.
func (mr *MockStoreCollectorMockRecorder) CollectCounter(name, val interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectCounter", reflect.TypeOf((*MockStoreCollector)(nil).CollectCounter), name, val)
}

// CollectGauge mocks base method.
func (m *MockStoreCollector) CollectGauge(name string, val float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CollectGauge", name, val)
}

// CollectGauge indicates an expected call of CollectGauge.
func (mr *MockStoreCollectorMockRecorder) CollectGauge(name, val interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectGauge", reflect.TypeOf((*MockStoreCollector)(nil).CollectGauge), name, val)
}

// MockStoreReporter is a mock of StoreReporter interface.
type MockStoreReporter struct {
	ctrl     *gomock.Controller
	recorder *MockStoreReporterMockRecorder
}

// MockStoreReporterMockRecorder is the mock recorder for MockStoreReporter.
type MockStoreReporterMockRecorder struct {
	mock *MockStoreReporter
}

// NewMockStoreReporter creates a new mock instance.
func NewMockStoreReporter(ctrl *gomock.Controller) *MockStoreReporter {
	mock := &MockStoreReporter{ctrl: ctrl}
	mock.recorder = &MockStoreReporterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStoreReporter) EXPECT() *MockStoreReporterMockRecorder {
	return m.recorder
}

// ReportAll mocks base method.
func (m *MockStoreReporter) ReportAll() ([]models.Metrics, time.Time) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReportAll")
	ret0, _ := ret[0].([]models.Metrics)
	ret1, _ := ret[1].(time.Time)
	return ret0, ret1
}

// ReportAll indicates an expected call of ReportAll.
func (mr *MockStoreReporterMockRecorder) ReportAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReportAll", reflect.TypeOf((*MockStoreReporter)(nil).ReportAll))
}

// RestoreReport mocks base method.
func (m *MockStoreReporter) RestoreReport(metrics []models.Metrics, ts time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RestoreReport", metrics, ts)
}

// RestoreReport indicates an expected call of RestoreReport.
func (mr *MockStoreReporterMockRecorder) RestoreReport(metrics, ts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreReport", reflect.TypeOf((*MockStoreReporter)(nil).RestoreReport), metrics, ts)
}

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

// Send mocks base method.
func (m *MockReporter) Send(arg0 []models.Metrics) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockReporterMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockReporter)(nil).Send), arg0)
}

// MockStore is a mock of Store interface.
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore.
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance.
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// CollectCounter mocks base method.
func (m *MockStore) CollectCounter(name string, val int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CollectCounter", name, val)
}

// CollectCounter indicates an expected call of CollectCounter.
func (mr *MockStoreMockRecorder) CollectCounter(name, val interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectCounter", reflect.TypeOf((*MockStore)(nil).CollectCounter), name, val)
}

// CollectGauge mocks base method.
func (m *MockStore) CollectGauge(name string, val float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CollectGauge", name, val)
}

// CollectGauge indicates an expected call of CollectGauge.
func (mr *MockStoreMockRecorder) CollectGauge(name, val interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectGauge", reflect.TypeOf((*MockStore)(nil).CollectGauge), name, val)
}

// ReportAll mocks base method.
func (m *MockStore) ReportAll() ([]models.Metrics, time.Time) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReportAll")
	ret0, _ := ret[0].([]models.Metrics)
	ret1, _ := ret[1].(time.Time)
	return ret0, ret1
}

// ReportAll indicates an expected call of ReportAll.
func (mr *MockStoreMockRecorder) ReportAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReportAll", reflect.TypeOf((*MockStore)(nil).ReportAll))
}

// RestoreReport mocks base method.
func (m *MockStore) RestoreReport(metrics []models.Metrics, ts time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RestoreReport", metrics, ts)
}

// RestoreReport indicates an expected call of RestoreReport.
func (mr *MockStoreMockRecorder) RestoreReport(metrics, ts interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreReport", reflect.TypeOf((*MockStore)(nil).RestoreReport), metrics, ts)
}
