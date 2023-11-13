package controller

import "github.com/freepaddler/yap-metrics/internal/pkg/models"

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../../mocks/serverController_mock.go

// Handler implements metrics methods for handling server requests
type Handler interface {
	GetAll() []models.Metrics
	GetOne(metric *models.Metrics) error
	UpdateOne(metric *models.Metrics) error
	UpdateMany(metrics []models.Metrics) error
}
