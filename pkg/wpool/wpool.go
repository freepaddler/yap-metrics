// Package wpool is worker pool implementation with context for correct cancellation
package wpool

import (
	"context"
	"errors"
	"sync"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

var (
	ErrClosed = errors.New("wpool is stopped")
)

// Pool describes a worker pool setup
type Pool struct {
	size           int           // number of workers in pool
	chStop         chan struct{} // channel to stop all pool goroutines
	chTask         chan func()   // channel to add tasks
	chStopGraceful chan struct{} // channel to stop all pool goroutines gracefully
	wg             sync.WaitGroup
	stop           sync.Once     // only one attempt to stop pool
	Stopped        chan struct{} // exported channel to check if pool is stopped (in case of context cancel)
}

// New is a Pool constructor
func New(ctx context.Context, size int) *Pool {
	pool := &Pool{
		size:           size,
		chStop:         make(chan struct{}),
		chTask:         make(chan func(), size),
		chStopGraceful: make(chan struct{}),
		Stopped:        make(chan struct{}),
	}
	// start cancelController
	go func() {
		pool.cancelController(ctx)
	}()
	// start workers
	for i := 0; i < size; i++ {
		go pool.worker(i)
	}
	return pool
}

// cancelController stops pool gracefully on context event
func (p *Pool) cancelController(ctx context.Context) {
	select {
	// handle context event
	case <-ctx.Done():
		p.Stop(true)
	// stop cancelController
	case <-p.chStop:
	}
	return
}

// Task adds a task for execution by pool workers
func (p *Pool) Task(f func()) error {
	for {
		select {
		case <-p.chStopGraceful:
		case <-p.chStop:
			return ErrClosed
		default:
			select {
			case p.chTask <- f:
				logger.Log.Debug().Msg("wpool task queued")
				return nil
			default:
			}
		}
	}

}

// Stop terminates all workers in the pool
// If graceful=true, then waits until all tasks in queue are done
func (p *Pool) Stop(graceful bool) {
	p.stop.Do(func() {
		logger.Log.Debug().Msgf("wpool request stop, graceful: %t", graceful)
		close(p.chStopGraceful)
		close(p.chTask)
		if graceful {
			// wait workers to complete all queue tasks
			p.wg.Wait()
			// stop worker pool
			close(p.chStop)
		} else {
			// stop workers
			close(p.chStop)
			// wait workers to complete current tasks
			p.wg.Wait()
		}
		close(p.Stopped)
		logger.Log.Debug().Msg("wpool stopped")
	})
}

// worker is a unique pool worker process
func (p *Pool) worker(n int) {
	p.wg.Add(1)
	defer func() {
		p.wg.Done()
		logger.Log.Debug().Msgf("wpool worker %d stopped", n)
	}()
	logger.Log.Debug().Msgf("wpool worker %d started", n)
	for {
		select {
		case <-p.chStop:
			logger.Log.Debug().Msgf("wpool worker %d received stop request", n)
			return
		case f, ok := <-p.chTask:
			if !ok {
				logger.Log.Debug().Msgf("wpool worker %d: tasks chan closed", n)
				return
			}
			f()
		}
	}
}
