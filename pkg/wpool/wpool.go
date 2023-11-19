// Package wpool is worker pool implementation with context for correct cancellation
package wpool

import (
	"context"
	"errors"
	"sync"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

var (
	ErrClosed = errors.New("wpool is stopped") // error, indicating stopped wpool
)

// Pool describes a worker pool setup
type Pool struct {
	size     int            // number of workers in pool
	chStop   chan struct{}  // channel to stop all pool goroutines
	chTask   chan func()    // channel to add tasks
	wg       sync.WaitGroup // waitGroup for workers
	stopOnce sync.Once      // only one attempt to stop pool
	stopped  chan struct{}  // external info channel that pool is actually stopped
}

// New is a Pool constructor
func New(ctx context.Context, size int) *Pool {
	pool := &Pool{
		size:    size,
		chStop:  make(chan struct{}),
		chTask:  make(chan func()),
		stopped: make(chan struct{}),
	}
	// start workers
	for i := 0; i < size; i++ {
		go pool.worker(ctx, i)
	}
	return pool
}

// Task adds a task for execution by pool workers.
func (p *Pool) Task(f func()) error {
	select {
	case <-p.chStop:
	case p.chTask <- f:
		logger.Log().Debug().Msg("wpool task accepted")
		return nil
	}
	return ErrClosed

}

// Stop terminates all workers in the pool. Runs only once.
// Returns channel: if closed, then pool is stopped
func (p *Pool) Stop() <-chan struct{} {
	p.stopOnce.Do(func() {
		logger.Log().Debug().Msg("wpool request stop")
		close(p.chStop)
		p.wg.Wait()
		logger.Log().Debug().Msg("wpool stopped")
		close(p.stopped)
	})
	return p.stopped
}

// worker is a unique pool worker process
func (p *Pool) worker(ctx context.Context, n int) {
	p.wg.Add(1)
	defer func() {
		p.wg.Done()
		logger.Log().Debug().Msgf("wpool worker %d stopped", n)
	}()
	logger.Log().Debug().Msgf("wpool worker %d started", n)
	for {
		select {
		// cancel pool by context
		case <-ctx.Done():
			logger.Log().Debug().Msgf("wpool worker %d received context cancel", n)
			go func() { p.Stop() }()
			return
		case <-p.chStop:
			logger.Log().Debug().Msgf("wpool worker %d received stop request", n)
			return
		case f, ok := <-p.chTask:
			if !ok {
				logger.Log().Debug().Msgf("wpool worker %d: tasks chan closed", n)
				return
			}
			logger.Log().Debug().Msgf("wpool worker %d doing job", n)
			f()
		}
	}
}
