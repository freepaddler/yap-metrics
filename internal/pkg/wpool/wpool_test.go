package wpool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	tasks = 6 // number of tasks to run
)

func incPayload(n *int) {
	time.Sleep(time.Second)
	*n++
}

func TestNormalExecution(t *testing.T) {
	pool := New(context.Background(), 2)
	n := 0
	// start 4 tasks
	for i := 0; i < tasks; i++ {
		pool.Task(func() {
			incPayload(&n)
		})
	}
	pool.Stop(context.Background(), true)
	assert.Equal(t, tasks, n, "Expected %d got %d", tasks, n)
}

func TestContextCancel(t *testing.T) {
	pool := New(context.Background(), 2)
	n := 0
	ctx, cancel := context.WithTimeout(context.Background(), tasks/2*time.Second)
	defer cancel()
	go func(ctx context.Context) {
		for i := 0; i < tasks; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				pool.Task(func() {
					incPayload(&n)
				})
			}
		}
	}(ctx)
	<-ctx.Done()
	assert.Greater(t, tasks, n, "Expected %d > %d", tasks, n)
}
