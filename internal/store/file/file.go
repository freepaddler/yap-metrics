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
	sw   bool
}

// NewFileStorage is a constructor
func NewFileStorage(path string, storeint int) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Log.Error().Err(err).Msgf("unable to open file %s", path)
		return nil, err
	}
	return &FileStorage{
		file: file,
		enc:  json.NewEncoder(file),
		dec:  json.NewDecoder(file),
		sw:   storeint == 0,
	}, nil
}

// SaveMetric called from storage to indicate data change
func (f *FileStorage) SaveMetric(m models.Metrics) {
	if f.sw {
		f.writeMetric(m)
	}
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
	for err != io.EOF {
		err = f.dec.Decode(&m)
		if err != nil {
			logger.Log.Warn().Err(err).Msg("error parsing file data")
		}
		s.SetMetric(&m)
	}
	logger.Log.Debug().Msg("done storage restore")
}

// Close closes file
func (f *FileStorage) Close() {
	f.file.Close()
}
