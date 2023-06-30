package file

import (
	"encoding/json"
	"io"
	"os"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
)

// FileStorage is persistent storage implementation
type FileStorage struct {
	file *os.File
	enc  *json.Encoder
	dec  *json.Decoder
}

// NewFileStorage is a constructor
func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Log.Error().Err(err).Msgf("unable to open file %s", path)
		return nil, err
	}
	return &FileStorage{
		file: file,
		enc:  json.NewEncoder(file),
		dec:  json.NewDecoder(file),
	}, nil
}

// Updated id called from storage to indicate metric change
func (f *FileStorage) Updated(s store.Storage, m models.Metrics) {
	switch m.Type {
	// get metric values
	case models.Gauge:
		m.FValue, _ = s.GetGauge(m.Name)
	case models.Counter:
		m.IValue, _ = s.GetCounter(m.Name)
	}
	f.writeMetric(m)
}

// writeMetric internal method to write metric to file
func (f *FileStorage) writeMetric(m models.Metrics) {
	if f.file != nil {
		logger.Log.Debug().Msgf("saving metric %s to file", m.Name)
		f.enc.Encode(m)
	}
}

// SaveStorage saves all metrics from storage to file
func (f *FileStorage) SaveStorage(s store.Storage) {
	logger.Log.Debug().Msg("Saving store to file...")
	for _, m := range s.GetAllMetrics() {
		f.writeMetric(m)
	}
}

// RestoreStorage loads metrics from file to storage
func (f *FileStorage) RestoreStorage(s store.Storage) {
	logger.Log.Debug().Msg("starting storage restore")
	var err error
	var m models.Metrics
	for {
		err = f.dec.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Log.Warn().Err(err).Msg("error parsing file data")
		}
		switch m.Type {
		case models.Gauge:
			s.SetGauge(m.Name, *m.FValue)
		case models.Counter:
			s.DelCounter(m.Name)
			s.IncCounter(m.Name, *m.IValue)
		}
	}
	logger.Log.Debug().Msg("done storage restore")
}

// Close closes file
func (f *FileStorage) Close() {
	f.file.Close()
}
