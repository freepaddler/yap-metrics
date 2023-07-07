package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

// FileStorage is persistent storage implementation
type FileStorage struct {
	mu   sync.Mutex
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
		mu:   sync.Mutex{},
		file: file,
		enc:  json.NewEncoder(file),
		dec:  json.NewDecoder(file),
	}, nil
}

// SaveMetric id called from storage to indicate metric change
func (f *FileStorage) SaveMetric(m models.Metrics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writeMetric(m)
}

// writeMetric internal method to write metric to file
func (f *FileStorage) writeMetric(m models.Metrics) {
	logger.Log.Debug().Msgf("saving metric %s to file", m.Name)
	f.enc.Encode(m)
}

// SaveStorage saves all metrics from storage to file
func (f *FileStorage) SaveStorage(s store.Storage) {
	logger.Log.Debug().Msg("saving store to file")
	f.mu.Lock()
	defer f.mu.Unlock()
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
	logger.Log.Debug().Msg("closing file storage")
	if err := f.file.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
		logger.Log.Warn().Err(err).Msg("closing file storage error")
		return
	}
}

// SaveLoop regularly saves storage to file
func (f *FileStorage) SaveLoop(ctx context.Context, s store.Storage, interval int) {
	logger.Log.Debug().Msg("starting file storage loop")
	t := time.NewTicker(time.Duration(interval) * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug().Msg("file storage loop stopped")
			return
		case <-t.C:
			f.SaveStorage(s)
		}
	}
}
