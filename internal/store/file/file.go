package file

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
)

// FileStorage is persistent storage implementation
type FileStorage struct {
	mu   sync.Mutex
	file *os.File
	enc  *json.Encoder
	dec  *json.Decoder
	// temporary solution
	closed bool
}

// NewFileStorage is a constructor
func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.Log.Error().Err(err).Msgf("unable to open file %s", path)
		return nil, err
	}
	return &FileStorage{
		mu:     sync.Mutex{},
		file:   file,
		enc:    json.NewEncoder(file),
		dec:    json.NewDecoder(file),
		closed: false,
	}, nil
}

// SaveMetric id called from storage to indicate metric change
func (f *FileStorage) SaveMetric(m models.Metrics) {
	f.writeMetric(m)
}

// writeMetric internal method to write metric to file
func (f *FileStorage) writeMetric(m models.Metrics) {
	logger.Log.Debug().Msgf("saving metric %s to file", m.Name)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enc.Encode(m)
}

// SaveStorage saves all metrics from storage to file
func (f *FileStorage) SaveStorage(s store.Storage) {
	logger.Log.Debug().Msg("saving store to file...")
	snap := s.Snapshot()
	for _, m := range snap {
		f.writeMetric(m)
	}
}

// RestoreStorage loads metrics from file to storage
func (f *FileStorage) RestoreStorage(s store.Storage) {
	logger.Log.Debug().Msg("starting storage restore")
	var err error
	var m models.Metrics
	f.mu.Lock()
	defer f.mu.Unlock()
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
	if !f.closed {
		f.closed = true
		logger.Log.Debug().Msg("closing file storage")
		if err := f.file.Close(); err != nil {
			logger.Log.Warn().Err(err).Msg("closing file storage error")
			return
		}
	}
}

// SaveLoop regularly saves storage to file
func (f *FileStorage) SaveLoop(s store.Storage, interval int) {
	logger.Log.Debug().Msg("starting file storage loop...")
	t := time.Tick(time.Duration(interval) * time.Second)
	for range t {
		if f.closed {
			logger.Log.Debug().Msg("file storage loop stopped")
			break
		}
		f.SaveStorage(s)
	}
}
