package wpool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

const (
	tasks     = 8 // number of tasks to run
	workers   = 2
	taskSleep = 2 * time.Second
)

type counter struct {
	m sync.Mutex
	n int
}

func (c *counter) incPayload() {
	time.Sleep(taskSleep)
	c.m.Lock()
	defer c.m.Unlock()
	c.n++
}

// all tasks should be executed
func TestGracefulStop(t *testing.T) {
	logger.SetLevel("debug")
	pool := New(context.Background(), workers)
	n := counter{}
	var err error
	for i := 0; i < tasks; i++ {
		err = pool.Task(n.incPayload)
	}
	pool.Stop(true)
	require.NoError(t, err)
	assert.Equal(t, tasks, n.n, "Expected %d got %d", tasks, n.n)
}

// wpool is stopped by context not all tasks should be executed
func TestContextStop(t *testing.T) {
	logger.SetLevel("debug")
	ctx, cancel := context.WithTimeout(context.Background(), tasks/workers/2*taskSleep)
	defer cancel()
	pool := New(ctx, workers)
	n := counter{}
	var err error

	for i := 0; i < tasks; i++ {
		err = pool.Task(n.incPayload)
		if err != nil {
			break
		}
	}

	<-pool.Stopped
	require.ErrorIs(t, err, ErrClosed)
	assert.Greater(t, tasks, n.n, "Expected %d got %d", tasks, n.n)
}

// force stop should prevent all tasks executions
func TestForceStop(t *testing.T) {
	logger.SetLevel("debug")
	ctx, cancel := context.WithTimeout(context.Background(), tasks/workers/2*taskSleep)
	defer cancel()
	pool := New(context.Background(), workers)
	n := counter{}
	go func() {
		for i := 0; i < tasks; i++ {
			if pool.Task(n.incPayload) != nil {
				return
			}
		}
	}()
	<-ctx.Done()
	// stop pool
	pool.Stop(false)
	<-pool.Stopped
	assert.Greater(t, tasks, n.n, "Expected %d got %d", tasks, n.n)
}
