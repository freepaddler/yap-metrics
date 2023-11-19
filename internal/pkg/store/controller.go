// Package store defines metrics realtime storage. It describes required basic store methods.
//
// Also in provides Storage controller with high level methods to interact with Store, hiding basics implementation.

package store

import (
	"errors"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

var (
	ErrMetricNotFound = errors.New("metric not found in store")
)

// Controller implements high level functions over basic store implementation
type Controller struct {
	store    Store
	mu       sync.RWMutex
	gaugesTS map[string]time.Time // timestamps of gauges updates
}

// NewStorageController is a Controller constructor
func NewStorageController(store Store) *Controller {
	return &Controller{
		store:    store,
		gaugesTS: make(map[string]time.Time),
	}
}

// Collector methods

// CollectCounter creates or updates counter value in store
func (c *Controller) CollectCounter(name string, val int64) {
	c.store.IncCounter(name, val)
}

// CollectGauge creates or updates gauge value in store
func (c *Controller) CollectGauge(name string, val float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store.SetGauge(name, val)
	c.gaugesTS[name] = time.Now()
}

// Reporter methods

// ReportAll returns all metrics from store and flushes them.
// Returns time when report was created to be able to restore metrics in case of unsuccessful send.
func (c *Controller) ReportAll() ([]models.Metrics, time.Time) {
	return c.store.Snapshot(true), time.Now()
}

// RestoreLatest restores metrics in case of reporting failed.
// Counter values are always incremented, not to lose data.
// Gauges values are restored if there was no update. Gauge should always have the latest value.
func (c *Controller) RestoreLatest(metrics []models.Metrics, ts time.Time) {
	logger.Log().Debug().Msg("restoring metrics to store")
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range metrics {
		switch v.Type {
		case models.Gauge:
			// gauge value should always be latest
			if ts.After(c.gaugesTS[v.Name]) || ts.Equal(c.gaugesTS[v.Name]) {
				c.store.SetGauge(v.Name, *v.FValue)
				c.gaugesTS[v.Name] = ts
			} else {
				logger.Log().Debug().Msgf("skip gauge '%s' restore, have newer value", v.Name)
			}
		case models.Counter:
			// counter always increments
			c.store.IncCounter(v.Name, *v.IValue)
		}
	}
}

// Handlers methods

// GetAll returns all metrics from store
func (c *Controller) GetAll() []models.Metrics {
	return c.store.Snapshot(false)
}

// GetOne gets metric value from store and updates requested metric with it
func (c *Controller) GetOne(request models.MetricRequest) (m models.Metrics, err error) {
	switch request.Type {
	case models.Counter:
		v, ok := c.store.GetCounter(request.Name)
		if !ok {
			err = ErrMetricNotFound
			return
		}
		m.IValue = &v
	case models.Gauge:
		v, ok := c.store.GetGauge(request.Name)
		if !ok {
			err = ErrMetricNotFound
			return
		}
		m.FValue = &v
	default:
		err = models.ErrInvalidMetric
	}
	if err == nil {
		m.Name = request.Name
		m.Type = request.Type
	}
	return
}

// updateMetric creates or updates metric in store, set new value to requested metric.
// It is a helper function to be called from public methods. It is not write safe.
func (c *Controller) updateMetric(metric *models.Metrics) error {
	switch metric.Type {
	case models.Counter:
		if metric.IValue == nil {
			return models.ErrInvalidMetric
		}
		v := c.store.IncCounter(metric.Name, *metric.IValue)
		metric.IValue = &v
	case models.Gauge:
		if metric.FValue == nil {
			return models.ErrInvalidMetric
		}
		v := c.store.SetGauge(metric.Name, *metric.FValue)
		metric.FValue = &v
	default:
		return models.ErrInvalidMetric
	}
	return nil
}

// UpdateOne updates one metric in store, set new value to requested metric.
func (c *Controller) UpdateOne(metric *models.Metrics) error {
	return c.updateMetric(metric)
}

// UpdateMany updates batch of metric in store, set new value to requested metrics.
func (c *Controller) UpdateMany(metrics []models.Metrics) error {
	for _, v := range metrics {
		if err := c.updateMetric(&v); err != nil {
			return err
		}
	}
	return nil
}

// Ping is used to check store accessibility
func (c *Controller) Ping() error {
	return c.store.Ping()
}
