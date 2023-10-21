// Package wpool is worker pool implementation with context for correct cancellation
package wpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

// TerminationTimeout is a default value for Pool.termTimeout
const TerminationTimeout = time.Second

// Pool describes worker pool setup
type Pool struct {
	size        int           // number of workers in pool
	chStop      chan struct{} // channel to stop workers
	chTask      chan func()   // channel to add tasks
	wg          sync.WaitGroup
	termTimeout time.Duration // timeout to wait pool tasks to complete on cancel
	closed      bool          // if closed, no new tasks allowed
	Stopped     chan struct{} // when closed, pool is terminated
}

// New is a Pool constructor
func New(ctx context.Context, size int) *Pool {
	pool := &Pool{
		size:        size,
		chStop:      make(chan struct{}),
		chTask:      make(chan func(), size),
		termTimeout: TerminationTimeout,
		Stopped:     make(chan struct{}),
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

// SetStopTimeout changes Pool.termTimeout
func (p *Pool) SetStopTimeout(to time.Duration) {
	p.termTimeout = to
}

// cancelController tries to stop pool gracefully on context event
func (p *Pool) cancelController(ctx context.Context) {
	select {
	// handle context event
	case <-ctx.Done():
		// context with timeout for graceful stop
		ctxStop, cancelStop := context.WithTimeout(context.Background(), p.termTimeout)
		defer cancelStop()
		p.Stop(ctxStop, true)
	// stop cancelController
	case <-p.chStop:
		return
	}
}

// Task adds task to worker
func (p *Pool) Task(f func()) error {
	for {
		select {
		case p.chTask <- f:
			return nil
		default:
			if p.closed {
				return fmt.Errorf("wpool is in stopping or closed state")
			}
		}
	}
}

// Stop stops all workers in pool with context.
// If graceful, then it waits for all tasks in queue to complete
func (p *Pool) Stop(ctx context.Context, graceful bool) {
	p.closed = true
	logger.Log.Debug().Msg("wpoll set closed")
	c := make(chan struct{}, 1)
	go func() {
		defer close(c)
		if graceful {
			close(p.chTask)
			p.wg.Wait()
			close(p.chStop)
		} else {
			close(p.chStop)
			p.wg.Wait()
		}
	}()
	select {
	case <-c: // normal stop
		logger.Log.Debug().Msg("wpool stopped")
		close(p.Stopped)
		return
	case <-ctx.Done(): // context timeout
		logger.Log.Debug().Msg("wpool terminated by context")
		close(p.Stopped)
		return
	}
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
		case f, ok := <-p.chTask:
			if !ok {
				logger.Log.Debug().Msgf("wpool worker %d received graceful stop", n)
				return
			}
			f()
		case <-p.chStop:
			logger.Log.Debug().Msgf("wpool worker %d received force stop", n)
			return

		}
	}
}
