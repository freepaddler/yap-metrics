package wpool

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tasks     = 8 // number of tasks to run
	workers   = 2
	taskSleep = time.Second
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
func TestStop(t *testing.T) {
	//logger.SetLevel("debug")
	pool := New(context.Background(), workers)
	n := new(counter)
	var err error
	for i := 0; i < tasks; i++ {
		err = pool.Task(n.incPayload)
		if err != nil {
			t.Log(err)
		}
	}
	<-pool.Stop()
	require.NoError(t, err)
	// multiple runs should not panic
	<-pool.Stop()
	assert.Equal(t, tasks, n.n, "Expected %d got %d", tasks, n.n)
}

func TestStoppedError(t *testing.T) {
	pool := New(context.Background(), workers)
	pool.Stop()
	// require error trying to add to stopped wpool
	err := pool.Task(func() {
	})
	require.ErrorIs(t, err, ErrClosed, "Expected error '%v', got '%v'", ErrClosed, err)
}

// wpool is stopped by context not all tasks should be executed
func TestContextStop(t *testing.T) {
	//logger.SetLevel("debug")
	ctx, cancel := context.WithTimeout(context.Background(), tasks/workers/2*taskSleep)
	defer cancel()
	pool := New(ctx, workers)
	n := new(counter)

	for i := 0; i < tasks; i++ {
		pool.Task(n.incPayload)
	}

	<-pool.Stop()
	assert.Greater(t, tasks, n.n, "Expected %d got %d", tasks, n.n)
}
