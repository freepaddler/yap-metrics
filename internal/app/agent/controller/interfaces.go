package controller

import (
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../../mocks/agentController_mock.go

// Reporter implements methods to send metrics to server
type Reporter interface {
	ReportAll() ([]models.Metrics, time.Time)
	RestoreReport(metrics []models.Metrics, ts time.Time)
}

// Collector implements methods for metrics collection
type Collector interface {
	CollectCounter(name string, val int64)
	CollectGauge(name string, val float64)
}

// AgentController implements all agent app methods
type AgentController interface {
	Reporter
	Collector
}
