package controller

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../../mocks/agentController_mock.go
type AgentController interface {
	CollectCounter(name string, val int64)
	CollectGauge(name string, val float64)
}
