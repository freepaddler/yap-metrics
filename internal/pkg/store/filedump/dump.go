// Package filedump allows to dump storage metrics to file and restore them back to storage.
// Restore overwrites existing in dump metrics in storage. Restore may be run only once.
package filedump

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

const (
	dumpBoundary = "DumpDate="
)

var (
	ErrWrite         = errors.New("dump write failed")
	ErrRead          = errors.New("dump read failed")
	ErrRestoreDenied = errors.New("dump restore denied")
	ErrEmpty         = errors.New("dump is empty")
)

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../../mocks/Dump_mock.go

type FileDump struct {
	file     *os.File
	mu       sync.RWMutex
	once     sync.Once
	encoder  *json.Encoder
	decoder  *json.Decoder
	restored bool
}

func NewFileDump(f *os.File) *FileDump {
	return &FileDump{
		file:    f,
		encoder: json.NewEncoder(f),
		decoder: json.NewDecoder(f),
	}
}

func (fd *FileDump) Dump(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return ErrEmpty
	}
	fd.mu.Lock()
	defer fd.mu.Unlock()
	// no restore after write
	fd.restored = true
	//d.dumpDev.Write(fmt.(dumpDate+time.Now().Format(time.RFC3339Nano)))
	fmt.Fprintf(fd.file, "%s%s\n", dumpBoundary, time.Now().Format(time.RFC3339Nano))

	if err := fd.encoder.Encode(&metrics); err != nil {
		return fmt.Errorf("%w: %w", ErrWrite, err)
	}

	return nil
}

func (fd *FileDump) Restore() (metrics []models.Metrics, lastDump time.Time, err error) {
	if fd.restored {
		err = ErrRestoreDenied
		return
	}
	fd.restored = true
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	// search for the last dump boundary
	scanner := bufio.NewScanner(fd.file)
	var pos, offset int64

	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		pos += int64(advance)
		if timeBytes, ok := bytes.CutPrefix(token, []byte(dumpBoundary)); ok {
			if ts, err := time.Parse(time.RFC3339Nano, string(timeBytes)); err == nil {
				lastDump = ts
				offset = pos
			}
		}
		return
	}
	scanner.Split(scanLines)

	// scan whole file for last dump boundary
	for scanner.Scan() {
	}

	// move start to offset
	fd.file.Seek(offset, io.SeekStart)

	err = fd.decoder.Decode(&metrics)
	if err != nil && err == io.EOF {
		logger.Log().Warn().Err(err).Msg("error parsing file data")
		err = fmt.Errorf("%w: %w", ErrRead, err)
		return
	}
	if len(metrics) == 0 {
		err = ErrEmpty
	}
	return
}
