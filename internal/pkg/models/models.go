// Package models defines metrics models and validation functions.
package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

var (
	ErrInvalidMetric = errors.New("invalid metric format")
	ErrInvalidName   = fmt.Errorf("%w: missing metric name", ErrInvalidMetric)
	ErrInvalidType   = fmt.Errorf("%w: invalid metric type", ErrInvalidMetric)
	ErrInvalidValue  = fmt.Errorf("%w: invalid metric value", ErrInvalidMetric)
)

// MetricRequest is a struct for metrics requests
type MetricRequest struct {
	Name string `json:"id"`
	Type string `json:"type"`
}

// NewMetricRequest is used to create MetricRequest struct from string values
func NewMetricRequest(n, t string) (m MetricRequest, err error) {
	if n == "" {
		err = ErrInvalidName
		return
	}
	m.Name = n
	switch t {
	case Counter:
		m.Type = Counter
	case Gauge:
		m.Type = Gauge
	default:
		err = ErrInvalidType
	}
	return
}

// UnmarshalJSON validates MetricRequest data
func (m *MetricRequest) UnmarshalJSON(b []byte) error {
	type _mr MetricRequest
	_m := &struct {
		*_mr
	}{
		_mr: (*_mr)(m),
	}
	if err := json.Unmarshal(b, _m); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMetric, err)
	}
	if _, err := NewMetricRequest(m.Name, m.Type); err != nil {
		return err
	}
	return nil
}

// Metrics is universal struct for all supported metric types.
// Should be used in updates and responses.
type Metrics struct {
	Name   string   `json:"id"`
	Type   string   `json:"type"`
	FValue *float64 `json:"value,omitempty"`
	IValue *int64   `json:"delta,omitempty"`
}

// NewMetric is used to create Metrics struct from string values. Should be used in update requests.
// Correct value is required.
func NewMetric(n, t, v string) (m Metrics, err error) {
	if _, err := NewMetricRequest(n, t); err != nil {
		return m, err
	}
	m.Name = n
	m.Type = t
	switch m.Type {
	case Counter:
		if i, err := strconv.ParseInt(v, 10, 64); err != nil {
			return m, fmt.Errorf("%w: %w", ErrInvalidValue, err)
		} else {
			m.IValue = &i
		}
	case Gauge:
		if f, err := strconv.ParseFloat(v, 64); err != nil {
			return m, fmt.Errorf("%w: %w", ErrInvalidValue, err)
		} else {
			m.FValue = &f
		}
	default:
		err = ErrInvalidType
	}
	return
}

// StringVal returns metric value as a string.
func (m *Metrics) StringVal() string {
	switch m.Type {
	case Counter:
		if m.IValue != nil {
			return strconv.FormatInt(*m.IValue, 10)
		}
	case Gauge:
		if m.FValue != nil {
			return strconv.FormatFloat(*m.FValue, 'f', -1, 64)
		}
	}
	return ""
}

// UnmarshalJSON validates Metrics update request data.
func (m *Metrics) UnmarshalJSON(b []byte) error {
	type _metrics Metrics
	_m := &struct {
		*_metrics
	}{
		_metrics: (*_metrics)(m),
	}
	if err := json.Unmarshal(b, _m); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMetric, err)
	}
	if _, err := NewMetricRequest(m.Name, m.Type); err != nil {
		return err
	}
	switch m.Type {
	case Counter:
		if m.IValue == nil || m.FValue != nil {
			return ErrInvalidMetric
		}
	case Gauge:
		if m.FValue == nil || m.IValue != nil {
			return ErrInvalidMetric
		}
	default:
		return ErrInvalidType
	}
	return nil
}
