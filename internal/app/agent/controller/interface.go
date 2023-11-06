package controller

import (
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../../mocks/agentController_mock.go
type AgentController interface {
	CollectCounter(name string, val int64)
	CollectGauge(name string, val float64)
	ReportAll() ([]models.Metrics, time.Time)
	RestoreReport(metrics []models.Metrics, ts time.Time)
}
