// Package models defines metrics models and validation functions
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
	ErrBadMetric = errors.New("invalid metric format")
)

// Metrics model
type Metrics struct {
	Name   string   `json:"id"`
	Type   string   `json:"type"`
	FValue *float64 `json:"value,omitempty"`
	IValue *int64   `json:"delta,omitempty"`
}

// New is a helper to create metric from string values.
// May be used in handlers to validate unstructured requests.
// Allows empty (zero length string) values (v).
func New(n, t, v string) (m Metrics, err error) {
	if n == "" {
		err = fmt.Errorf("noname: %w", ErrBadMetric)
		return
	}
	m.Name = n
	switch t {
	case Counter:
		m.Type = Counter
		if v != "" {
			if i, err := strconv.ParseInt(v, 10, 64); err != nil {
				return m, fmt.Errorf("%w (%w)", ErrBadMetric, err)
			} else {
				m.IValue = &i
			}
		}
	case Gauge:
		m.Type = Gauge
		if v != "" {
			if f, err := strconv.ParseFloat(v, 64); err != nil {
				return m, fmt.Errorf("%w (%w)", ErrBadMetric, err)
			} else {
				m.FValue = &f
			}
		}
	default:
		err = fmt.Errorf("%w (invalid metric type)", ErrBadMetric)
	}
	return
}

// StringVal returns metric value as a string
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

// UnmarshalJSON also validates metric type and value
func (m *Metrics) UnmarshalJSON(b []byte) error {
	type _metrics Metrics
	_m := &struct {
		*_metrics
	}{
		_metrics: (*_metrics)(m),
	}
	if err := json.Unmarshal(b, &_m); err != nil {
		return err
	}
	if m.Name == "" {
		return ErrBadMetric
	}
	switch m.Type {
	case Counter:
		if m.FValue != nil {
			return ErrBadMetric
		}
	case Gauge:
		if m.IValue != nil {
			return ErrBadMetric
		}
	default:
		return ErrBadMetric
	}
	return nil
}
